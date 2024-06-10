package collectors

import (
	"fmt"
	"os"
	"regexp"

	"github.com/SUSE/connect-ng/internal/util"
)

type SAP struct{}

// sap follows directory structure of /usr/sap
// systemId and workloadId will be under the main directory
// eg /usr/sap/AB3/ERS1
var sapSystemId = regexp.MustCompile("([A-Z][A-Z0-9]{2})")
var workloadsRegex = regexp.MustCompile("([A-Z]+)([0-9]{2})/")

func (sap SAP) run(arch string) (Result, error) {
	if !util.FileExists("/usr/sap") {
		return NoResult, nil
	}

	sapDirs, err := os.ReadDir("/usr/sap")
	if err != nil {
		return NoResult, err
	}
	var systemId, instanceType, instanceId string

	// We assume that sapDirs contains only one entry,
	// that entry is the unique systemId for the machine
	for _, sapDir := range sapDirs {
		systemId = sapSystemId.FindStringSubmatch(sapDir.Name())[0]
	}
	workloadsPath := fmt.Sprintf("/usr/sap/%s/", systemId)

	workloadsDir, err := os.ReadDir(workloadsPath)

	if err != nil {
		return NoResult, err
	}

	// TODO: Handle multiple workloads
	for _, workloadDir := range workloadsDir {
		matches := workloadsRegex.FindStringSubmatch(workloadDir.Name())
		instanceType = matches[0]
		instanceId = matches[1]
	}
	sapDetected := map[string]string{
		"systemId":     systemId,
		"instanceType": instanceType,
		"instanceId":   instanceId,
	}
	return Result{"sap": sapDetected}, nil
}
