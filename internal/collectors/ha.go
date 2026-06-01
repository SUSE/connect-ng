package collectors

import (
	"bufio"
	"bytes"
	"errors"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

const (
	// Pacemaker constants
	pacemakerCmdPath = "/usr/sbin/pacemakerd"
)

type HA struct{}

// Get Pacemaker version
func getPacemakerVersion() (string, error) {
	pacemakerdOut, err := util.Execute([]string{pacemakerCmdPath, "--version"}, []int{0})
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(bytes.NewReader(pacemakerdOut))
	for scanner.Scan() {
		testLine := scanner.Text()
		fields := strings.Fields(testLine)
		if len(fields) >= 2 {
			if fields[0] == "Pacemaker" {
				return fields[1], nil
			}
		}
	}

	return "", scanner.Err()
}

func isPacemakerRunning(systemdClient util.SystemdClient) (bool, error) {
	// first try to use ListUnitsByPatterns, failing back to ListUnits if not available
	units, err := systemdClient.ListUnitsByPatterns("pacemaker.service")
	if err != nil {
		if !errors.Is(err, util.SystemdMethodNotAvailable) {
			util.Debug.Printf("systemdClient.ListUnitsByPatterns() failed: %s\n", err)
			return false, err
		}
		units, err = systemdClient.ListUnits()
	}

	util.Debug.Printf("pacemakerActive: units: %+v err: %+v\n", units, err)
	if err != nil {
		util.Debug.Printf("systemdClient.ListUnits() failed: %s\n", err)
		return false, err
	}

	for _, unit := range units {
		util.Debug.Printf("pacemakerActive: unit: %+v \n", unit)
		// is pacemaker running
		if unit.Name.Match("pacemaker.service") && unit.SubState == "running" {
			return true, nil
		}
	}
	return false, nil

}

func isSystemHA(systemdClient util.SystemdClient) (Result, error) {
	haActive, err := isPacemakerRunning(systemdClient)
	util.Debug.Print("ha:haActive ", haActive)
	if !haActive {
		return Result{}, err
	}
	pacemakerVersion, err := getPacemakerVersion()
	if pacemakerVersion != "" {
		result := Result{"ha_active": pacemakerVersion}
		util.Debug.Printf("ha:pacemakerVersion:Result: %+v \n", result)
		return result, err
	}
	return Result{}, err
}

func (HA) run(arch string) (Result, error) {
	systemdClient, err := util.NewDbusSystemdClient()
	if err != nil {
		return nil, err
	}
	defer systemdClient.Close()
	result, err := isSystemHA(systemdClient)
	return result, err
}
