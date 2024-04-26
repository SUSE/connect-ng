package collectors

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockDirEntry struct {
	name  string
	isDir bool
	info  os.FileInfo
}

func (m MockDirEntry) Name() string {
	return m.name
}

func (m MockDirEntry) IsDir() bool {
	return m.isDir
}
func (m MockDirEntry) Info() (os.FileInfo, error) {
	return m.info, nil
}
func (m MockDirEntry) Type() os.FileMode {
	if m.isDir {
		return os.ModeDir
	}
	return 0
}

func mockLocalOsReaddir(mockedPath map[string][]string) {
	var mockExpected []os.DirEntry
	localOsReaddir = func(path string) ([]os.DirEntry, error) {
		expectedPaths := mockedPath[path]
		for _, p := range expectedPaths {
			mockExpected = append(mockExpected,
				MockDirEntry{
					name:  p,
					isDir: true,
				})
		}
		return mockExpected, nil
	}
}

func TestGetMatchedSubdirectoriesWithEmptySubdirs(t *testing.T) {
	assert := assert.New(t)
	absolutePath := "/path/abc"
	matcher := regexp.MustCompile("abc")
	expected := []string{}

	mockLocalOsReaddir(map[string][]string{
		"/tmp": []string{},
	})

	result, err := getMatchedSubdirectories(absolutePath, matcher)
	assert.NoError(err)
	assert.Equal(expected, result)
}

func TestGetMatchedSubdirectoriesWithSubDirs(t *testing.T) {
	assert := assert.New(t)
	absolutePath := "/tmp"
	expected := []string{"hi5", "mx0"}
	matcher := regexp.MustCompile("([a-z]{2}[0-9]{1})")

	mockLocalOsReaddir(map[string][]string{
		"/tmp": []string{"hi5", "mx0"},
	})

	result, err := getMatchedSubdirectories(absolutePath, matcher)
	assert.NoError(err)
	assert.Equal(expected, result)
}

func TestSapDetectorCollectorRunWithNoSAP(t *testing.T) {
	assert := assert.New(t)
	sap := SAP{}

	mockUtilFileExists(false)

	res, err := sap.run(ARCHITECTURE_X86_64)
	assert.NoError(err)
	assert.Equal(res, Result{}, "Result mismatch")
}

func TestSAPDetectSingleWorkload(t *testing.T) {
	assert := assert.New(t)
	sap := SAP{}
	expected := Result{"sap": []map[string]interface{}{
		{
			"system_id":      "DEV",
			"instance_types": []string{"ASCS"},
		},
	}}

	mockUtilFileExists(true)
	mockLocalOsReaddir(map[string][]string{
		"/usr/sap":     []string{"DEV"},
		"/usr/sap/DEV": []string{"ASCS01"},
	})

	res, err := sap.run(ARCHITECTURE_X86_64)
	assert.NoError(err)
	assert.Equal(expected, res, "Result mismatch when there are workloads")

}

func TestSAPRunWithMultipleWorkloads(t *testing.T) {
	assert := assert.New(t)
	sap := SAP{}
	expected := Result{"sap": []map[string]interface{}{
		{
			"system_id":      "DEV",
			"instance_types": []string{"ASCS", "J"},
		},
	}}

	mockUtilFileExists(true)
	mockLocalOsReaddir(map[string][]string{
		"/usr/sap":     []string{"DEV", ".config"},
		"/usr/sap/DEV": []string{"ASCS01", "J01"},
	})

	res, err := sap.run(ARCHITECTURE_X86_64)
	assert.NoError(err)
	assert.Equal(expected, res, "Result mismatch when there are workloads")
}
