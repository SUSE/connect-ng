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
// eg /usr/sap/AB3/ERS1
var sapSystemId = regexp.MustCompile("([A-Z][A-Z0-9]{2})")
var workloadsRegex = regexp.MustCompile("([A-Z]+)[0-9]{2}")
var localOsReaddir = os.ReadDir
var sapInstallationDir = "/usr/sap"

func getMatchedSubdirectories(absolutePath string, matcher *regexp.Regexp) ([]string, error) {
	subDirectories, err := localOsReaddir(absolutePath)
	//go:nocover
	if err != nil || len(subDirectories) == 0 {
		return []string{}, err
	}
	var match []string
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
	if !util.FileExists(sapInstallationDir) {
		return NoResult, nil
	}
	systemIds, err := getMatchedSubdirectories(sapInstallationDir, sapSystemId)

	if err != nil {
		return NoResult, err
	}

	var detector []map[string]interface{}
	for _, systemId := range systemIds {
		systemPath := path.Join(sapInstallationDir, systemId)
		workloads, _ := getMatchedSubdirectories(systemPath, workloadsRegex)
		detector = append(detector, map[string]interface{}{
			"systemId":      systemId,
			"instanceTypes": workloads,
		})
	}
	return Result{"sap": detector}, nil
}
