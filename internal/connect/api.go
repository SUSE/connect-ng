package connect

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"

	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/internal/zypper"
	"github.com/SUSE/connect-ng/pkg/registration"
)

const (
	UptimeLogFilePath = "/etc/zypp/suse-uptime.log"
)

func upgradeProduct(product registration.Product) (Service, error) {
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

// TODO
func downgradeProduct(product registration.Product) (Service, error) {
	return upgradeProduct(product)
}

func deactivateProduct(product registration.Product) (Service, error) {
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

func syncProducts(products []registration.Product) ([]registration.Product, error) {
	remoteProducts := make([]registration.Product, 0)
	var payload struct {
		Products []registration.Product `json:"products"`
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

func productMigrations(installed []registration.Product) ([]MigrationPath, error) {
	migrations := make([]MigrationPath, 0)
	var payload struct {
		InstalledProducts []registration.Product `json:"installed_products"`
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

func offlineProductMigrations(installed []registration.Product, target registration.Product) ([]MigrationPath, error) {
	migrations := make([]MigrationPath, 0)
	var payload struct {
		InstalledProducts []registration.Product `json:"installed_products"`
		TargetBaseProduct registration.Product   `json:"target_base_product"`
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
