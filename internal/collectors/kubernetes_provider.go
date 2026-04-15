package collectors

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/SUSE/connect-ng/internal/util"
)

const (
	// recognised kubernetes providers
	RKE2_PROVIDER = "rke2"
	K3S_PROVIDER  = "k3s"

	// recognised kubernetes provider node roles
	KUBERNETES_SERVER = "server"
	KUBERNETES_AGENT  = "agent"
)

// Custom Errors
var (
	KubernetesProviderNotRecognised    = errors.New("kubernetes provider not recognised")
	KubernetesRoleNotRecognised        = errors.New("kubernetes role not recognised")
	KubernetesMultipleProvidersEnabled = errors.New("multiple kubernetes providers enabled")
)

var (
	//
	// regexp matchers
	//

	// match lines with the format "<app> version <version> (<hash>)" and extract
	// <version> and <hash> fields
	versionMatcher = regexp.MustCompile(`^\S+\s+version\s+(\S+)\s+\(([[:xdigit:]]+)\)$`)

	//
	// kubernetes providers and roles that this collector recognises and reports
	//
	recognisedKubernetesProviders = []string{
		RKE2_PROVIDER,
		K3S_PROVIDER,
	}
	recognisedKubernetesRoles = []string{
		KUBERNETES_SERVER,
		KUBERNETES_AGENT,
	}
)

type K8S struct{}

func (K8S) run(arch string) (Result, error) {
	systemdClient, err := util.NewDbusSystemdClient()
	if err != nil {
		return nil, err
	}
	defer systemdClient.Close()

	return generateKubernetesProvider(systemdClient)
}

func generateKubernetesProvider(systemdClient util.SystemdClient) (Result, error) {
	provider, err := getKubernetesProviderData(systemdClient)
	if err != nil {
		util.Debug.Printf("Failed to get kubernetes provider: %s", err)
		return Result{}, err
	}

	// nothing to report
	if provider == nil {
		return Result{}, nil
	}

	res := Result{"kubernetes_provider": provider}
	return res, nil
}

func getKubernetesProviderData(systemdClient util.SystemdClient) (map[string]any, error) {
	kSvcs, err := getKubernetesServices(systemdClient)
	if err != nil {
		return nil, err
	}

	// return nil without error if no services found
	if len(kSvcs) == 0 {
		util.Debug.Println("No kubernetes providers found")
		return nil, nil
	}

	// only one kubernetes provider should be enabled on a system
	if len(kSvcs) > 1 {
		err := fmt.Errorf(
			"too many kubernetes providers enabled (%d): %w",
			len(kSvcs), KubernetesMultipleProvidersEnabled,
		)
		return nil, err
	}

	provider := map[string]any{
		"type":    kSvcs[0].Type,
		"role":    kSvcs[0].Role,
		"version": kSvcs[0].Version,
	}

	return provider, nil
}

func getKubernetesServices(systemdClient util.SystemdClient) ([]*KubernetesService, error) {
	kSvcs := []*KubernetesService{}

	// first try to use ListUnitsByPatterns, failing back to ListUnits if not available
	svcPatterns := []string{"rke2-*.service", "k3s.service", "k3s-*.service*"}
	filterNames := false // expecting ListUnitsByPatterns to handle name filtering for us
	units, err := systemdClient.ListUnitsByPatterns(svcPatterns...)
	if err != nil {
		if !errors.Is(err, util.SystemdMethodNotAvailable) {
			util.Debug.Printf("systemdClient.ListUnitsByPatterns() failed: %s\n", err)
			return nil, err
		}

		// fail back to ListUnits, and ensure name filtering is completed below
		filterNames = true
		units, err = systemdClient.ListUnits()
		if err != nil {
			util.Debug.Printf("systemdClient.ListUnits() failed: %s\n", err)
			return nil, err
		}
	}

	for _, unit := range units {
		// skip if names don't match, if not already filtered by ListUnitsByPatterns
		if filterNames && !unit.Name.Match(svcPatterns...) {
			util.Debug.Printf("Skipping non-matching unit %q\n", unit.Name)
			continue
		}

		state, err := systemdClient.GetUnitFileState(unit.Name)
		if err != nil {
			util.Debug.Printf("systemdClient.GetUnitFileState(%q) failed: %s\n", unit.Name, err)
			// skip on failure
			continue
		}

		// skip if not enabled
		if state != "enabled" {
			util.Debug.Printf("Skipping disabled unit %q\n", unit.Name)
			continue
		}

		ks, err := NewKubernetesService(systemdClient, unit)
		if err != nil {
			util.Debug.Printf("NewKubernetesService(%q) failed: %s\n", unit.Name, err)
			// skip on failure
			continue
		}
		kSvcs = append(kSvcs, ks)
	}

	return kSvcs, nil
}

type KubernetesService struct {
	Name        string
	Type        string
	Binary      string
	Role        string
	Version     string
	VersionHash string
}

func NewKubernetesService(systemdClient util.SystemdClient, unit *util.SystemdUnit) (*KubernetesService, error) {
	ks := new(KubernetesService)

	// retrieve the ExecStart property value for the specified unit file
	execStarts, err := systemdClient.GetExecStart(unit.ObjectPath)
	if err != nil {
		return nil, err
	}

	// proceed only if a single exec start was retrieved
	if len(execStarts) != 1 {
		err = fmt.Errorf("Multiple (%d) ExecStarts found for unit %q", len(execStarts), unit.Name)
		return nil, err
	}
	execStart := execStarts[0]

	// ensure that there are at least 2 arguments, binary path and role, in the args list
	if len(execStart.Args) < 2 {
		err = fmt.Errorf("Missing second argument in ExecStart for unit %q", unit.Name)
		return nil, err
	}

	// store retrieved values
	ks.Name = string(unit.Name)
	ks.Binary = execStart.Path
	ks.Role = execStart.Args[1]
	ks.Type = filepath.Base(ks.Binary)

	// validate type is recognised
	if !slices.Contains(recognisedKubernetesProviders, ks.Type) {
		err = fmt.Errorf(
			"type %q not in recognised providers list %v: %w",
			ks.Type, recognisedKubernetesProviders,
			KubernetesProviderNotRecognised,
		)
		return nil, err
	}

	// validate role is recognised
	if !slices.Contains(recognisedKubernetesRoles, ks.Role) {
		err = fmt.Errorf(
			"role %q not in recognised roles list %v: %w",
			ks.Role, recognisedKubernetesRoles,
			KubernetesRoleNotRecognised,
		)
		return nil, err
	}

	// retrieve the version using the binary
	if err = ks.setVersion(); err != nil {
		return nil, err
	}

	return ks, nil
}

func (ks *KubernetesService) setVersion() error {
	output, err := util.Execute(
		[]string{ks.Binary, "--version"},
		[]int{0},
	)
	if err != nil {
		// report failure if unable to retrieve version
		err := fmt.Errorf("failed to run %s --version: %w", ks.Binary, err)
		return err
	}

	// split into lines and extract version info from first line
	outputLines := bytes.Split(output, []byte("\n"))
	matches := versionMatcher.FindSubmatch(outputLines[0])
	if len(matches) != 3 {
		// report an error if version output doesn't match expected format
		err = fmt.Errorf("failed to parse %s --version output", ks.Binary)
		return err
	}

	ks.Version = string(matches[1])
	ks.VersionHash = string(matches[2])

	return nil
}
