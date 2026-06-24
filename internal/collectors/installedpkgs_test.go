package collectors

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/SUSE/connect-ng/pkg/profiles"
	"github.com/stretchr/testify/assert"
)

type pkgsEnvelope struct {
	InstalledPkgs struct {
		Id   string `json:"identifier"`
		Data any    `json:"data"`
	} `json:"installed_pkgs"`
}

var pkgsTestData string

func setupTestData() {
	testProfilePath, _ := os.MkdirTemp("/tmp/", "__suseconnect")
	profiles.SetProfileFilePath(testProfilePath + "/")

	pkgsTestData = "" +
		"SUSE LLC\tglibc\t2.31\t150300.63.1\tx86_64\n" +
		"SUSE\tzypper\t1.14.70\t150400.3.15.1\tx86_64\n" +
		"Example Inc.\tmy-app\t1.0\t1\tx86_64\n" +
		"SUSE LLC\tglibc\t2.31\t150300.63.1\tx86_64\n" +
		"Another Corp\tmysuse-app\t1.0\t1\tx86_64\n"
}

func TestRunSuccessNoUpdate(t *testing.T) {
	assert := assert.New(t)
	setupTestData()

	mockUtilExecute(pkgsTestData, nil)

	collector := InstalledPackages{UpdateDataIDs: false}
	result, err := collector.run(ARCHITECTURE_X86_64)
	assert.NoError(err)

	raw, err := json.Marshal(result)
	assert.NoError(err)

	var pkgs pkgsEnvelope
	assert.NoError(json.Unmarshal(raw, &pkgs))

	assert.NotEmpty(pkgs.InstalledPkgs.Id)
	assert.NotNil(pkgs.InstalledPkgs.Data)

	expectedData := []any{
		[]any{"glibc", "2.31", "150300.63.1", "x86_64"},
		[]any{"zypper", "1.14.70", "150400.3.15.1", "x86_64"},
	}
	assert.Equal(expectedData, pkgs.InstalledPkgs.Data)
}

func TestRunSuccessUpdate_SendsOnlyIdOnSecondRun(t *testing.T) {
	assert := assert.New(t)
	setupTestData()

	mockUtilExecute(pkgsTestData, nil)

	collector := InstalledPackages{UpdateDataIDs: true}

	first, err := collector.run(ARCHITECTURE_X86_64)
	assert.NoError(err)
	raw1, _ := json.Marshal(first)
	var pkgs1 pkgsEnvelope
	assert.NoError(json.Unmarshal(raw1, &pkgs1))
	assert.NotEmpty(pkgs1.InstalledPkgs.Id)
	assert.NotNil(pkgs1.InstalledPkgs.Data)

	second, err := collector.run(ARCHITECTURE_X86_64)
	assert.NoError(err)
	raw2, _ := json.Marshal(second)
	var pkgs2 pkgsEnvelope
	assert.NoError(json.Unmarshal(raw2, &pkgs2))
	assert.NotEmpty(pkgs2.InstalledPkgs.Id)
	assert.Nil(pkgs2.InstalledPkgs.Data)
}

func TestRunFail(t *testing.T) {
	assert := assert.New(t)
	setupTestData()

	mockUtilExecute("", fmt.Errorf("forced rpm error"))

	collector := InstalledPackages{}
	result, err := collector.run(ARCHITECTURE_X86_64)
	profiles.DeleteProfileCache("*")

	assert.Equal(Result{}, result)
	assert.ErrorContains(err, "forced rpm error")
}

func TestFilterPackages(t *testing.T) {
	assert := assert.New(t)

	rawOutput := []byte(
		"SUSE LLC\tglibc\t2.31\t150300.63.1\tx86_64\n" +
			"SUSE\tzypper\t1.14.70\t150400.3.15.1\tx86_64\n" +
			"Example Inc.\tmy-app\t1.0\t1\tx86_64\n" +
			"SUSE LLC\tglibc\t2.31\t150300.63.1\tx86_64\n",
	)

	expected := [][]string{
		{"glibc", "2.31", "150300.63.1", "x86_64"},
		{"zypper", "1.14.70", "150400.3.15.1", "x86_64"},
	}

	result, err := filterPackages(rawOutput)
	assert.NoError(err)
	assert.Equal(expected, result)
}
