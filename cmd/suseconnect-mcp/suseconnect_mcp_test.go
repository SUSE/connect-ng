package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestToolExecutionWithInvalidInputs(t *testing.T) {
	oldMcpDir := mcpDir
	mcpDir = t.TempDir()
	defer func() { mcpDir = oldMcpDir }()

	assert := assert.New(t)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name     string
		testFunc func() (*mcp.CallToolResult, JSONOutput, error)
		errorMsg string
	}{
		{
			name: "ActivateProduct with invalid triplet format",
			testFunc: func() (*mcp.CallToolResult, JSONOutput, error) {
				return ActivateProduct(ctx, req, ActivateInput{Product: "invalid-format", Regcode: "test"})
			},
			errorMsg: "Please provide the product identifier in this format",
		},
		{
			name: "ActivateProduct with empty product",
			testFunc: func() (*mcp.CallToolResult, JSONOutput, error) {
				return ActivateProduct(ctx, req, ActivateInput{Product: "", Regcode: "test"})
			},
			errorMsg: "Please provide the product identifier in this format",
		},
		{
			name: "ActivateProduct with too many parts",
			testFunc: func() (*mcp.CallToolResult, JSONOutput, error) {
				return ActivateProduct(ctx, req, ActivateInput{Product: "product/version/arch/extra", Regcode: "test"})
			},
			errorMsg: "Please provide the product identifier in this format",
		},
		{
			name: "DeactivateProduct with invalid triplet format",
			testFunc: func() (*mcp.CallToolResult, JSONOutput, error) {
				return DeactivateProduct(ctx, req, DeactivateInput{Product: "invalid-format"})
			},
			errorMsg: "Please provide the product identifier in this format",
		},
		{
			name: "DeactivateProduct with empty product",
			testFunc: func() (*mcp.CallToolResult, JSONOutput, error) {
				return DeactivateProduct(ctx, req, DeactivateInput{Product: ""})
			},
			errorMsg: "Please provide the product identifier in this format",
		},
		{
			name: "DeactivateProduct with only two parts",
			testFunc: func() (*mcp.CallToolResult, JSONOutput, error) {
				return DeactivateProduct(ctx, req, DeactivateInput{Product: "product/version"})
			},
			errorMsg: "Please provide the product identifier in this format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, output, err := tt.testFunc()

			// Validate response structure
			assert.Nil(result, "Result should always be nil per MCP convention")
			assert.IsType(JSONOutput{}, output, "Output should be JSONOutput type")
			assert.NotNil(err, "Should return error for invalid input")
			assert.NotEmpty(output.Error, "Error field should be populated")
			assert.Empty(output.Response, "Response should be empty on error")
			assert.Contains(output.Error, tt.errorMsg, "Should contain expected error message")
		})
	}
}

func updateActionCountHelper(actionName, dir, filePath string) error {
	m := NewMcpActions()
	return updateActionCountIn(m, actionName, dir, filePath)
}

func updateActionCountIn(m *McpActions, actionName, dir, filePath string) error {
	if _, err := m.osStat(dir); os.IsNotExist(err) {
		err = m.osMkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Ensure that the process has write access to the directory, otherwise fail early
		testFile := dir + "/.write_test"
		f, openErr := os.OpenFile(testFile, os.O_WRONLY|os.O_CREATE, 0644)
		if openErr != nil {
			return fmt.Errorf("no write access to %s: %w", dir, openErr)
		}
		f.Close()

		err = m.osRemove(testFile)
		if err != nil {
			// This mimics the production code where this error is a warning and not fatal
		}
	}

	file, err := m.osReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read file: %w", err)
		}
	} else {
		err = json.Unmarshal(file, &m.actions)
		if err != nil {
			// Can be empty file
		}
	}

	m.incCount(actionName)

	data, err := json.Marshal(m.actions)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	tempFilePath := fmt.Sprintf("%s.%d.tmp", filePath, os.Getpid())
	err = m.osWriteFile(tempFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	err = m.osRename(tempFilePath, filePath)
	if err != nil {
		m.osRemove(tempFilePath)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

func testIncCount(t *testing.T) {
	m := NewMcpActions()

	m.incCount("action1")
	assert.Equal(t, 1, len(m.actions))
	assert.Equal(t, "action1", m.actions[0].Name)
	assert.Equal(t, 1, m.actions[0].Count)

	m.incCount("action2")
	assert.Equal(t, 2, len(m.actions))

	m.incCount("action1")
	assert.Equal(t, 2, len(m.actions))
	for _, action := range m.actions {
		if action.Name == "action1" {
			assert.Equal(t, 2, action.Count)
		}
	}
}

func testUpdateActionCountCreatesDir(t *testing.T) {
	assert := assert.New(t)
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "test-suseconnect-mcp")
	assert.NoError(err)
	defer os.RemoveAll(tempDir)

	// Set mcpDir to a non-existent subdirectory
	newMcpDir := tempDir + "/new-mcp-dir"
	oldMcpDir := mcpDir
	mcpDir = newMcpDir
	defer func() { mcpDir = oldMcpDir }()

	// Call updateActionCount
	err = updateActionCount("testAction")
	assert.NoError(err)

	// Check if the directory was created
	_, err = os.Stat(newMcpDir)
	assert.NoError(err, "The new mcpDir should have been created")
}

func testCreatesNewFileAndAddsAction(t *testing.T, dir string, filePath string) {
	m := NewMcpActions()
	var writtenData []byte
	renamed := false
	removedTestFile := false

	m.osStat = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	m.osMkdirAll = func(path string, perm os.FileMode) error { return nil }
	m.osReadFile = func(name string) ([]byte, error) { return nil, os.ErrNotExist }
	m.osWriteFile = func(name string, data []byte, perm os.FileMode) error {
		assert.Equal(t, fmt.Sprintf("%s.%d.tmp", filePath, os.Getpid()), name)
		writtenData = data
		return nil
	}
	m.osRename = func(oldpath, newpath string) error {
		assert.Equal(t, fmt.Sprintf("%s.%d.tmp", filePath, os.Getpid()), oldpath)
		assert.Equal(t, filePath, newpath)
		renamed = true
		return nil
	}
	m.osRemove = func(name string) error {
		if name == dir+"/.write_test" {
			removedTestFile = true
			return nil
		}
		t.Fatalf("osRemove called with unexpected file: %s", name)
		return nil
	}

	err := updateActionCountIn(m, "testAction", dir, filePath)
	assert.NoError(t, err)
	assert.True(t, renamed)
	assert.True(t, removedTestFile, "osRemove should be called for the test file")

	var actions []Action
	err = json.Unmarshal(writtenData, &actions)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actions))
	assert.Equal(t, "testAction", actions[0].Name)
	assert.Equal(t, 1, actions[0].Count)

	// Clean up the dummy file created by os.OpenFile
	os.Remove(dir + "/.write_test")
}

func testAddsToExistingFile(t *testing.T, dir string, filePath string) {
	m := NewMcpActions()
	var writtenData []byte
	renamed := false
	initialJSON := `[{"name": "existingAction", "count": 5}]`

	// This test doesn't create a directory, so no write check is performed
	m.osStat = func(name string) (os.FileInfo, error) { return nil, nil } // File exists
	m.osMkdirAll = func(path string, perm os.FileMode) error {
		t.Fatalf("osMkdirAll should not be called when dir exists")
		return nil
	}
	m.osReadFile = func(name string) ([]byte, error) { return []byte(initialJSON), nil }
	m.osWriteFile = func(name string, data []byte, perm os.FileMode) error {
		writtenData = data
		return nil
	}
	m.osRename = func(oldpath, newpath string) error {
		renamed = true
		return nil
	}
	m.osRemove = func(name string) error {
		t.Fatalf("osRemove should not be called on success when dir exists")
		return nil
	}

	err := updateActionCountIn(m, "newAction", dir, filePath)
	assert.NoError(t, err)
	assert.True(t, renamed)

	var actions []Action
	err = json.Unmarshal(writtenData, &actions)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(actions))
}

func testHandlesWriteError(t *testing.T, dir string, filePath string) {
	m := NewMcpActions()
	m.osStat = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	m.osMkdirAll = func(path string, perm os.FileMode) error { return nil }
	m.osReadFile = func(name string) ([]byte, error) { return nil, os.ErrNotExist }
	m.osWriteFile = func(name string, data []byte, perm os.FileMode) error {
		return errors.New("disk full")
	}
	// Mock osRemove for the .write_test file
	m.osRemove = func(name string) error { return nil }

	err := updateActionCountIn(m, "testAction", dir, filePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write to temporary file: disk full")
	// Clean up the dummy file created by os.OpenFile
	os.Remove(dir + "/.write_test")
}

func testHandlesRenameError(t *testing.T, dir string, filePath string) {
	m := NewMcpActions()
	removed := false
	m.osStat = func(name string) (os.FileInfo, error) { return nil, nil }
	m.osMkdirAll = func(path string, perm os.FileMode) error { return nil }
	m.osReadFile = func(name string) ([]byte, error) { return nil, nil }
	m.osWriteFile = func(name string, data []byte, perm os.FileMode) error {
		return nil
	}
	m.osRename = func(oldpath, newpath string) error {
		return errors.New("rename failed")
	}
	m.osRemove = func(name string) error {
		assert.Equal(t, fmt.Sprintf("%s.%d.tmp", filePath, os.Getpid()), name)
		removed = true
		return nil
	}

	err := updateActionCountIn(m, "testAction", dir, filePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rename failed")
	assert.True(t, removed, "temporary file should be removed on rename failure")
}

func TestUpdateActionCount(t *testing.T) {
	dir, err := os.MkdirTemp("", "suseconnect-mcp-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	filePath := dir + "/suseconnectmcp"

	t.Run("increment count of specific action", func(t *testing.T) {
		testIncCount(t)
	})

	t.Run("creates new Dir and Updates action count", func(t *testing.T) {
		testUpdateActionCountCreatesDir(t)
	})

	t.Run("creates new file and adds action", func(t *testing.T) {
		testCreatesNewFileAndAddsAction(t, dir, filePath)
	})

	t.Run("adds to existing file", func(t *testing.T) {
		testAddsToExistingFile(t, dir, filePath)
	})

	t.Run("handles write error", func(t *testing.T) {
		testHandlesWriteError(t, dir, filePath)
	})

	t.Run("handles rename error", func(t *testing.T) {
		testHandlesRenameError(t, dir, filePath)
	})
}
