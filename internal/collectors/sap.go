package collectors

import (
	"os"
	"path"
	"regexp"

	"github.com/SUSE/connect-ng/internal/util"
)

type SAP struct{}

// sap follows directory structure of /usr/sap
// systemId and workloadId will be under the main directory
// eg /usr/sap/AB3/ERS12
var sapSystemId = regexp.MustCompile("^([A-Z][A-Z0-9]{2})$")
var workloadsRegex = regexp.MustCompile("^([a-zA-Z]+)[0-9]{2}$")
var localOsReaddir = os.ReadDir
var sapInstallationDir = "/usr/sap"

func getMatchedSubdirectories(absolutePath string, matcher *regexp.Regexp) ([]string, error) {
	subDirectories, err := localOsReaddir(absolutePath)
	//go:nocover
	if err != nil || len(subDirectories) == 0 {
		return []string{}, err
	}
	match := []string{}
	for _, subDirectory := range subDirectories {
		// filter for nil values from FindStringSubmatch
		matches := matcher.FindStringSubmatch(subDirectory.Name())
		if len(matches) >= 2 {
			match = append(match, matches[1])
		}
	}
	return match, nil
}

func (sap SAP) run(arch string) (Result, error) {
	detected := []map[string]interface{}{}

	if !util.FileExists(sapInstallationDir) {
		return NoResult, nil
	}
	systemIds, err := getMatchedSubdirectories(sapInstallationDir, sapSystemId)

	if err != nil {
		return NoResult, err
	}

	for _, systemId := range systemIds {
		systemPath := path.Join(sapInstallationDir, systemId)
		workloads, _ := getMatchedSubdirectories(systemPath, workloadsRegex)

		if len(workloads) > 0 {
			detected = append(detected, map[string]interface{}{
				"system_id":      systemId,
				"instance_types": workloads,
			})
		}
	}

	if len(detected) > 0 {
		return Result{"sap": detected}, nil
	}

	return NoResult, nil
}
