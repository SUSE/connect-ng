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
	// supported kubernetes providers
	RKE2_PROVIDER = "rke2"
	K3S_PROVIDER  = "k3s"

	// supported kubernetes provider node roles
	KUBERNETES_SERVER = "server"
	KUBERNETES_AGENT  = "agent"
)

// Custom Errors
var (
	KubernetesProviderNotSupported = errors.New("kubernetes provider not supported")
	KubernetesRoleNotSupported     = errors.New("kubernetes role not supported")
)

var (
	//
	// regexp matchers
	//

	// match lines with the format: "<unit_file> <state> <preset>"
	listUnitFilesOutPutMatcher = regexp.MustCompile(`^(\S+)\s+(\S+)\s+\S+$`)

	// rke2 service name matcher
	rke2ServiceNameMatcher = regexp.MustCompile(`^rke2-(?:server|agent)\.service$`)

	// rke2 service name matcher
	k3sServiceNameMatcher = regexp.MustCompile(`^k3s(?:|-[^.]+)\.service$`)

	// match lines with the format: "ExecStart={ <binary> ; argv[]=<binary> <role> ... ; ... }"
	//execStartPropertyMatcher = regexp.MustCompile(`^ExecStart=[{]\s+path=(\S+)\s+;\s+argv\[\]=\S+\s+(\S+)(?:\s+\S+)*\s+;\s+.*\s+[}]$`)
	execStartPropertyMatcher = regexp.MustCompile(
		`^ExecStart=\{\s+path=(\S+)\s+;\s+argv\[\]=\S+\s+([^ ;}]+)(?:\s+[^ ;}]+)*\s+` +
			`(?:;\s+\S+=(?:[^ ;}]+\s+)*)*\}$`,
	)

	// match lines with the format: "<app> version <version> (<hash>)"
	versionMatcher = regexp.MustCompile(`^\S+\s+version\s+(\S+)\s+\(([[:xdigit:]]+)\)$`)

	//
	// systemctl settings
	//

	// path to the systemctl command
	systemctlBin = "/usr/bin/systemctl"

	// extra options to pass to the
	systemctlBaseCmd = []string{
		systemctlBin,
		"--no-pager",
		"--legend=false",
	}

	//
	// kubernetes providers and roles
	//
	supportedKubernetesProviders = []string{
		RKE2_PROVIDER,
		K3S_PROVIDER,
	}
	supportedKubernetesRoles = []string{
		KUBERNETES_SERVER,
		KUBERNETES_AGENT,
	}
)

func systemctlCmd(cmd string, args ...string) ([]byte, error) {
	// construct systemctl command line
	cmdLine := []string{}
	cmdLine = append(cmdLine, systemctlBaseCmd...)
	cmdLine = append(cmdLine, cmd)
	cmdLine = append(cmdLine, args...)

	exitCodes := []int{0}

	// execute systemctl command
	output, err := util.Execute(cmdLine, exitCodes)
	if err != nil {
		return nil, err
	}

	return output, nil
}

type K8S struct{}

func (K8S) run(arch string) (Result, error) {

	provider, err := getKubernetesProvider()
	if err != nil {
		util.Info.Printf("Failed to get kubernetes provider: %s", err)
		return Result{}, err
	}

	// nothing to report
	if provider == nil {
		return Result{}, nil
	}

	res := Result{"kubernetes_provider": provider}
	return res, nil
}

func getKubernetesProvider() (map[string]any, error) {
	kSvcs, err := getKubernetesServices()
	if err != nil {
		return nil, err
	}

	// only one kubernetes provider should be enabled on a system
	if len(kSvcs) > 1 {
		err := fmt.Errorf("multiple kubernetes providers enabled [%v]", kSvcs)
		return nil, err
	}

	// return nil without error if no services found
	if len(kSvcs) == 0 {
		return nil, nil
	}

	provider := map[string]any{
		"type":    kSvcs[0].Type,
		"role":    kSvcs[0].Role,
		"version": kSvcs[0].Version,
	}

	return provider, nil
}

func getKubernetesServices() ([]*KubernetesService, error) {
	// search for RKE2 and k3s provider services
	svcFilters := []struct {
		pattern     string
		nameMatcher *regexp.Regexp
	}{
		{"rke2-*", rke2ServiceNameMatcher},
		{"k3s*", k3sServiceNameMatcher},
	}
	kSvcs := []*KubernetesService{}

	for _, filters := range svcFilters {
		unitFiles, err := listMatchingUnitFilesOfTypeAndState(filters.pattern, "service", "enabled", filters.nameMatcher)
		if err != nil {
			util.Debug.Printf(
				"listMatchingUnitFilesOfTypeAndState() for pattern %q nameMatcher %s failed to match services: %s\n",
				filters.pattern, filters.nameMatcher, err,
			)
			return nil, err
		}

		for _, unitFile := range unitFiles {
			ks, err := NewKubernetesService(unitFile)
			if errors.Is(err, KubernetesProviderNotSupported) {
				util.Debug.Printf(
					"NewKubernetesService() failed for unitFile %q: %s\n",
					unitFile, err,
				)
				continue
			} else if errors.Is(err, KubernetesRoleNotSupported) {
				util.Debug.Printf(
					"NewKubernetesService() failed for unitFile %q: %s\n",
					unitFile, err,
				)
				continue
			} else if err != nil {
				util.Debug.Printf(
					"NewKubernetesService() failed for unitFile %q: %s\n",
					unitFile, err,
				)
				return nil, err
			}

			kSvcs = append(kSvcs, ks)
		}
	}

	return kSvcs, nil
}

func listMatchingUnitFilesOfTypeAndState(pattern, unitType, state string, nameFilter *regexp.Regexp) ([]*UnitFile, error) {
	output, err := systemctlCmd(
		"list-unit-files", "--type", unitType, "--state", state, pattern,
	)
	if err != nil {
		return nil, err
	}

	unitFiles := []*UnitFile{}
	for _, line := range bytes.Split(output, []byte("\n")) {
		// filter on lines that match target pattern
		enabledMatches := listUnitFilesOutPutMatcher.FindSubmatch(line)

		// skip line if line doesn't match expected pattern
		if len(enabledMatches) != 3 {
			continue
		}

		// skip line if state field is not desired state
		if string(enabledMatches[2]) != state {
			continue
		}

		// skip unit file name if it doesn't match the expected pattern
		if !nameFilter.Match(enabledMatches[1]) {
			continue
		}

		unitFiles = append(unitFiles, NewUnitFile(string(enabledMatches[1])))
	}

	return unitFiles, nil
}

type UnitFile struct {
	Name string
}

func NewUnitFile(name string) *UnitFile {
	return &UnitFile{Name: name}
}

func (uf *UnitFile) String() string {
	return uf.Name
}

func (uf *UnitFile) Property(property string) ([]byte, error) {
	output, err := systemctlCmd(
		"show", "--property", property, uf.Name,
	)
	if err != nil {
		return nil, err
	}

	return output, nil
}

type KubernetesService struct {
	// public attributes
	Type        string
	Binary      string
	Role        string
	Version     string
	VersionHash string

	// private attributes
	unitFile *UnitFile
}

func NewKubernetesService(unitFile *UnitFile) (*KubernetesService, error) {
	ks := new(KubernetesService)

	ks.unitFile = unitFile

	// retrieve the ExecStart property value for the specified unit file
	execStart, err := ks.unitFile.Property("ExecStart")
	if err != nil {
		return nil, err
	}

	// extract the binary and role from the ExecStart property value
	matches := execStartPropertyMatcher.FindSubmatch(bytes.TrimSpace(execStart))
	if len(matches) != 3 {
		err = fmt.Errorf("failed to parse ExecStart property value")
		return nil, err
	}

	// store extracted values
	ks.Binary = string(matches[1])
	ks.Role = string(matches[2])
	ks.Type = filepath.Base(ks.Binary)

	// validate type is supported
	if !slices.Contains(supportedKubernetesProviders, ks.Type) {
		err = fmt.Errorf(
			"type %q not in supported providers list %v: %w",
			ks.Type, supportedKubernetesProviders,
			KubernetesProviderNotSupported,
		)
		return nil, err
	}

	// validate role is supported
	if !slices.Contains(supportedKubernetesRoles, ks.Role) {
		err = fmt.Errorf(
			"role %q not in supported providers list %v: %w",
			ks.Role, supportedKubernetesRoles,
			KubernetesRoleNotSupported,
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
		return err
	}

	// split into lines and extract version info from first line
	outputLines := bytes.Split(output, []byte("\n"))
	matches := versionMatcher.FindSubmatch(outputLines[0])
	if len(matches) != 3 {
		// report an error if
		err = fmt.Errorf("failed to parse %s --version output", ks.Binary)
		return err
	}

	ks.Version = string(matches[1])
	ks.VersionHash = string(matches[2])

	return nil
}
