package connect

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"

	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/internal/zypper"
)

const (
	UptimeLogFilePath = "/etc/zypp/suse-uptime.log"
)

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
