package connect

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/internal/zypper"
)

const (
	UptimeLogFilePath = "/etc/zypp/suse-uptime.log"
)

func upToDate() bool {
	// REVIST 404 case - see original
	// Should fail in any case. 422 error means that the endpoint is there and working right
	_, err := callHTTP("GET", "/connect/repositories/installer", nil, nil, authNone)
	if err == nil {
		return false
	}
	if ae, ok := err.(APIError); ok {
		if ae.Code == http.StatusUnprocessableEntity {
			return true
		}
	}
	return false
}

// systemActivations returns a map keyed by "Identifier/Version/Arch"
func systemActivations() (map[string]Activation, error) {
	activeMap := make(map[string]Activation)
	resp, err := callHTTP("GET", "/connect/systems/activations", nil, nil, authSystem)
	if err != nil {
		return activeMap, err
	}
	var activations []Activation
	if err = json.Unmarshal(resp, &activations); err != nil {
		return activeMap, JSONError{err}
	}
	for _, activation := range activations {
		activeMap[activation.toTriplet()] = activation
	}
	return activeMap, nil
}

func showProduct(productQuery Product) (Product, error) {
	resp, err := callHTTP("GET", "/connect/systems/products", nil, productQuery.toQuery(), authSystem)
	remoteProduct := Product{}
	if err != nil {
		return remoteProduct, err
	}
	if err = json.Unmarshal(resp, &remoteProduct); err != nil {
		return remoteProduct, JSONError{err}
	}
	return remoteProduct, nil
}

func upgradeProduct(product Product) (Service, error) {
	// NOTE: this can add some extra attributes to json payload which
	//       seem to be safely ignored by the API.
	payload, err := json.Marshal(product)
	remoteService := Service{}
	if err != nil {
		return remoteService, err
	}
	resp, err := callHTTP("PUT", "/connect/systems/products", payload, nil, authSystem)
	if err != nil {
		return remoteService, err
	}
	if err = json.Unmarshal(resp, &remoteService); err != nil {
		return remoteService, JSONError{err}
	}
	return remoteService, nil
}

func downgradeProduct(product Product) (Service, error) {
	return upgradeProduct(product)
}

func activateProduct(product Product, email string) (Service, error) {
	var payload = struct {
		Indentifier string `json:"identifier"`
		Version     string `json:"version"`
		Arch        string `json:"arch"`
		ReleaseType string `json:"release_type"`
		Token       string `json:"token"`
		Email       string `json:"email"`
	}{
		product.Name,
		product.Version,
		product.Arch,
		product.ReleaseType,
		CFG.Token,
		email,
	}

	service := Service{}
	body, err := json.Marshal(payload)
	if err != nil {
		return service, err
	}
	resp, err := callHTTP("POST", "/connect/systems/products", body, nil, authSystem)
	if err != nil {
		return service, err
	}
	err = json.Unmarshal(resp, &service)
	if err != nil {
		return service, JSONError{err}
	}
	return service, nil
}

func deactivateProduct(product Product) (Service, error) {
	// NOTE: this can add some extra attributes to json payload which
	//       seem to be safely ignored by the API.
	payload, err := json.Marshal(product)
	remoteService := Service{}
	if err != nil {
		return remoteService, err
	}
	resp, err := callHTTP("DELETE", "/connect/systems/products", payload, nil, authSystem)
	if err != nil {
		return remoteService, err
	}
	if err = json.Unmarshal(resp, &remoteService); err != nil {
		return remoteService, JSONError{err}
	}
	return remoteService, nil
}

func syncProducts(products []Product) ([]Product, error) {
	remoteProducts := make([]Product, 0)
	var payload struct {
		Products []Product `json:"products"`
	}
	payload.Products = products
	body, err := json.Marshal(payload)
	if err != nil {
		return remoteProducts, err
	}
	resp, err := callHTTP("POST", "/connect/systems/products/synchronize", body, nil, authSystem)
	if err != nil {
		return remoteProducts, err
	}
	err = json.Unmarshal(resp, &remoteProducts)
	if err != nil {
		return remoteProducts, JSONError{err}
	}
	return remoteProducts, nil
}

// updateSystem updates the system's hardware information on SCC
// https://scc.suse.com/connect/v4/documentation#/systems/put_systems
// The body parameter is produced by makeSysInfoBody()
func updateSystem(body []byte) error {
	_, err := callHTTP("PUT", "/connect/systems", body, nil, authSystem)
	return err
}

// readUptimeLogFile reads the system uptime log from a given file and
// returns them as a string array. If the given file does not exist,
// it will be interpreted as if the system uptime log feature is not
// enabled. Hence an empty array will be returned.
func readUptimeLogFile(uptimeLogFilePath string) ([]string, error) {
	// NOTE: the uptime log file is produced by the suse-uptime-tracker
	// (https://github.com/SUSE/uptime-tracker) service. If the service
	// is installed and enabled, barring any unforeseen errors, the
	// uptime log file is expected to there and updated on the regular
	// basis. If the service is not installed or otherwise disabled, the
	// uptime log file may not exist. In that case we assume the uptime
	// tracking feature is disabled.
	_, err := os.Stat(uptimeLogFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	uptimeLogFile, err := os.Open(uptimeLogFilePath)
	if err != nil {
		return nil, err
	}
	defer uptimeLogFile.Close()
	fileScanner := bufio.NewScanner(uptimeLogFile)
	var logEntries []string

	for fileScanner.Scan() {
		logEntries = append(logEntries, fileScanner.Text())
	}
	if err = fileScanner.Err(); err != nil {
		return nil, err
	}
	return logEntries, nil
}

// makeSysInfoBody returns the JSON payload needed for the announce/update system calls
func makeSysInfoBody(distroTarget, namespace string, instanceData []byte, includeUptimeLog bool) ([]byte, error) {
	var payload struct {
		Hostname     string            `json:"hostname"`
		DistroTarget string            `json:"distro_target"`
		InstanceData string            `json:"instance_data,omitempty"`
		Namespace    string            `json:"namespace,omitempty"`
		Hwinfo       collectors.Result `json:"hwinfo"`
		OnlineAt     []string          `json:"online_at,omitempty"`
	}
	if distroTarget != "" {
		payload.DistroTarget = distroTarget
	} else {
		var err error
		payload.DistroTarget, err = zypper.DistroTarget()
		if err != nil {
			return nil, err
		}
	}
	payload.InstanceData = string(instanceData)
	payload.Namespace = namespace

	if includeUptimeLog {
		uptimeLog, err := readUptimeLogFile(UptimeLogFilePath)
		if err != nil {
			util.Debug.Printf("Unable to read uptime log: %v", err)
			util.Info.Print("Unable to read system uptime log")
		} else {
			payload.OnlineAt = uptimeLog
		}
	}

	sysinfo, err := FetchSystemInformation()
	if err != nil {
		return nil, err
	}

	payload.Hwinfo = sysinfo
	payload.Hostname = collectors.FromResult(sysinfo, "hostname", "")

	return json.Marshal(payload)
}

var mandatoryCollectors = []collectors.Collector{
	collectors.CPU{},
	collectors.Hostname{},
	collectors.Memory{},
	collectors.UUID{},
	collectors.Virtualization{},
	collectors.CloudProvider{},
	collectors.Architecture{},
	collectors.ContainerRuntime{},

	// Optional collectors
	collectors.Uname{},
	collectors.SAP{},
}

func FetchSystemInformation() (collectors.Result, error) {
	arch, err := collectors.DetectArchitecture()

	if err != nil {
		return collectors.NoResult, err
	}
	return collectors.CollectInformation(arch, mandatoryCollectors)
}

func productMigrations(installed []Product) ([]MigrationPath, error) {
	migrations := make([]MigrationPath, 0)
	var payload struct {
		InstalledProducts []Product `json:"installed_products"`
	}
	payload.InstalledProducts = installed
	body, err := json.Marshal(payload)
	if err != nil {
		return migrations, err
	}
	resp, err := callHTTP("POST", "/connect/systems/products/migrations", body, nil, authSystem)
	if err != nil {
		return migrations, err
	}
	if err = json.Unmarshal(resp, &migrations); err != nil {
		return migrations, JSONError{err}
	}
	return migrations, nil
}

func offlineProductMigrations(installed []Product, target Product) ([]MigrationPath, error) {
	migrations := make([]MigrationPath, 0)
	var payload struct {
		InstalledProducts []Product `json:"installed_products"`
		TargetBaseProduct Product   `json:"target_base_product"`
	}
	payload.InstalledProducts = installed
	payload.TargetBaseProduct = target
	body, err := json.Marshal(payload)
	if err != nil {
		return migrations, err
	}
	resp, err := callHTTP("POST", "/connect/systems/products/offline_migrations", body, nil, authSystem)
	if err != nil {
		return migrations, err
	}
	if err = json.Unmarshal(resp, &migrations); err != nil {
		return migrations, JSONError{err}
	}
	return migrations, nil
}

func installerUpdates(product Product) ([]zypper.Repository, error) {
	repos := make([]zypper.Repository, 0)
	resp, err := callHTTP("GET", "/connect/repositories/installer", nil, product.toQuery(), authNone)
	if err != nil {
		return repos, err
	}
	if err = json.Unmarshal(resp, &repos); err != nil {
		return repos, JSONError{err}
	}
	return repos, nil
}

func setLabels(labels []Label) error {
	var payload struct {
		Labels []Label `json:"labels"`
	}
	payload.Labels = labels
	body, err := json.Marshal(payload)

	if err != nil {
		return err
	}
	_, err = callHTTP("POST", "/connect/systems/labels", body, nil, authSystem)
	return err
}
