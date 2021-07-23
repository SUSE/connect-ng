package connect

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
)

func TestPrintExtensions(t *testing.T) {
	extensions := make([]Product, 0)
	extensionsData := readTestFile("extensions.json", t)
	json.Unmarshal(extensionsData, &extensions)
	activations := map[string]Activation{}
	activations["SUSE-Manager-Server/3.2/x86_64"] = Activation{}
	// writable root FS test
	output, _ := printExtensions(extensions, activations, true)

	patterns := []string{
		// header
		"\x1b\\[1mAVAILABLE EXTENSIONS AND MODULES\x1b\\[0m",
		// top level product name
		"    \x1b\\[1mSUSE Manager Proxy 3.2 x86_64\x1b\\[0m",
		// top level product hint
		"    Activate with: SUSEConnect -p SUSE-Manager-Proxy/3.2/x86_64",
		// second level product name (not available)
		"        \x1b\\[1mSUSE Manager Retail Branch Server 3.2 x86_64\x1b\\[0m \x1b\\[31m\\(Not available\\)\x1b\\[0m",
		// second level product hint
		"        Activate with: SUSEConnect -p SUSE-Manager-Retail-Branch-Server/3.2/x86_64",
		// activated product name
		"    \x1b\\[1mSUSE Manager Server 3.2 x86_64\x1b\\[0m \x1b\\[33m\\(Activated\\)\x1b\\[0m",
		// activated product hint
		"    Deactivate with: SUSEConnect \x1b\\[31m-d\x1b\\[0m -p SUSE-Manager-Server/3.2/x86_64",
		// non-free product hint
		"    Activate with: SUSEConnect -p suse-openstack-cloud/8/x86_64 -r \x1b\\[32m\x1b\\[1mADDITIONAL REGCODE\x1b\\[0m",
	}
	for _, pattern := range patterns {
		if found, _ := regexp.MatchString(fmt.Sprintf("(?m)^%s$", pattern), output); !found {
			t.Errorf("Pattern: '%s' not found in output", pattern)
		}
	}
	// count empty lines
	if match := regexp.MustCompile("(?m)^$").FindAllStringIndex(output, -1); len(match) != 10 {
		t.Errorf("Expected 10 empty lines, found: %v", len(match))
	}
	// read-only root FS test
	output2, _ := printExtensions(extensions, activations, false)
	cmd := "transactional-update register"
	if found, _ := regexp.MatchString(cmd, output2); !found {
		t.Errorf("'%s' not found in output", cmd)
	}
	cmd = "SUSEConnect"
	if found, _ := regexp.MatchString(cmd, output2); found {
		t.Errorf("Unexpected '%s' found in output", cmd)
	}

	// test "(Not available)" is not printed when using SCC
	CFG.BaseURL = defaultBaseURL
	output3, _ := printExtensions(extensions, activations, true)
	pattern := `(?m)^.*SUSE Manager Retail Branch Server 3.2 x86_64.*(Not available).*$`
	if found, _ := regexp.MatchString(pattern, output3); found {
		t.Errorf("Pattern: '%s' should not be found in output", pattern)
	}
}
