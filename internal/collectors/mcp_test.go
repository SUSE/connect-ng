package collectors

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/SUSE/connect-ng/pkg/profiles"
	"github.com/stretchr/testify/assert"
)

var mcpDataBlob profiles.Profile
var mcpTestData string

func setupMcpTestData() {
	testProfilePath, _ := os.MkdirTemp("", "suseconnect-*")
	profiles.SetProfileFilePath(testProfilePath + "/")

	mcpTestData = `[{"name":"RegistrationStatus","count":1}]`
	var actions []Action
	_ = json.Unmarshal([]byte(mcpTestData), &actions)
	mcpDataBlob.Data = actions

	jsonBytes, _ := json.Marshal(mcpDataBlob.Data)
	hash := sha256.Sum256(jsonBytes)
	mcpDataBlob.Id = hex.EncodeToString(hash[:])
}

func mockMcpOsReadfile(t *testing.T, expectedPath string, content string, err error) {
	mcpOsReadfile = func(path string) ([]byte, error) {
		assert.Equal(t, expectedPath, path)
		return []byte(content), err
	}
	t.Cleanup(func() { mcpOsReadfile = os.ReadFile })
}

func TestMCPRunSuccessNoUpdate(t *testing.T) {
	assert := assert.New(t)
	setupMcpTestData()
	mockMcpOsReadfile(t, mcpPath, mcpTestData, nil)
	expected := Result{mcpTag: mcpDataBlob}

	collector := MCP{UpdateDataIDs: false}
	result, err := collector.run(ARCHITECTURE_X86_64)
	assert.Equal(expected, result)
	assert.Nil(err)
}

func TestMCPRunSuccessUpdate(t *testing.T) {
	assert := assert.New(t)
	setupMcpTestData()
	mockMcpOsReadfile(t, mcpPath, mcpTestData, nil)
	expected := Result{mcpTag: mcpDataBlob}

	collector := MCP{UpdateDataIDs: true}
	result, err := collector.run(ARCHITECTURE_X86_64)

	assert.Equal(expected, result)
	assert.Nil(err)
}

func TestMCPRunSumsMatch(t *testing.T) {
	assert := assert.New(t)
	setupMcpTestData()
	mockMcpOsReadfile(t, mcpPath, mcpTestData, nil)

	// Prime the cache with the expected ID so that the checksum matches
	cacheFilePath := filepath.Join(profiles.GetProfileFilePath(), mcpChecksumFile)
	err := os.WriteFile(cacheFilePath, []byte(mcpDataBlob.Id), 0644)
	assert.Nil(err)

	collector := MCP{UpdateDataIDs: true}
	result, err := collector.run(ARCHITECTURE_X86_64)
	fmt.Println(result)

	var expectedDataBlob profiles.Profile
	expectedDataBlob.Id = mcpDataBlob.Id
	expected := Result{mcpTag: expectedDataBlob}

	assert.Equal(expected, result)
	assert.Nil(err)
}

func TestMCPRunFail(t *testing.T) {
	assert := assert.New(t)

	mockMcpOsReadfile(t, mcpPath, "", fmt.Errorf("forced error"))
	expected := Result{}

	collector := MCP{}
	result, err := collector.run(ARCHITECTURE_X86_64)

	profiles.DeleteProfileCache("*")
	assert.Equal(expected, result)
	assert.ErrorContains(err, "forced error")
}
