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

func testCreatesNewFileAndAddsAction(t *testing.T, dir string, filePath string) {
	// Override global mcpDir for this test
	oldMcpDir := mcpDir
	mcpDir = dir
	defer func() { mcpDir = oldMcpDir }()

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
	m.osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		// Mock file creation - return a dummy file descriptor
		return &os.File{}, nil
	}

	err := updateActionCount(m, "testAction")
	assert.NoError(t, err)
	assert.True(t, renamed)
	assert.True(t, removedTestFile, "osRemove should be called for the test file")

	var actions []Action
	err = json.Unmarshal(writtenData, &actions)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actions))
	assert.Equal(t, "testAction", actions[0].Name)
	assert.Equal(t, 1, actions[0].Count)
}

func testAddsToExistingFile(t *testing.T, dir string, filePath string) {
	// Override global mcpDir for this test
	oldMcpDir := mcpDir
	mcpDir = dir
	defer func() { mcpDir = oldMcpDir }()

	m := NewMcpActions()
	var writtenData []byte
	renamed := false
	removedTestFile := false
	initialJSON := `[{"name": "existingAction", "count": 5}]`

	m.osStat = func(name string) (os.FileInfo, error) { return nil, nil }
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
		if name == dir+"/.write_test" {
			removedTestFile = true
			return nil
		}
		t.Fatalf("osRemove called with unexpected file: %s", name)
		return nil
	}
	m.osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		// Mock file creation for write test
		return &os.File{}, nil
	}

	err := updateActionCount(m, "newAction")
	assert.NoError(t, err)
	assert.True(t, renamed)
	assert.True(t, removedTestFile, "osRemove should be called for the test file")

	var actions []Action
	err = json.Unmarshal(writtenData, &actions)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(actions))
}

func testHandlesWriteError(t *testing.T, dir string, filePath string) {
	// Override global mcpDir for this test
	oldMcpDir := mcpDir
	mcpDir = dir
	defer func() { mcpDir = oldMcpDir }()

	m := NewMcpActions()
	m.osStat = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	m.osMkdirAll = func(path string, perm os.FileMode) error { return nil }
	m.osReadFile = func(name string) ([]byte, error) { return nil, os.ErrNotExist }
	m.osWriteFile = func(name string, data []byte, perm os.FileMode) error {
		return errors.New("disk full")
	}
	// Mock osRemove for the .write_test file
	m.osRemove = func(name string) error { return nil }
	m.osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		// Mock file creation - return a dummy file descriptor
		return &os.File{}, nil
	}

	err := updateActionCount(m, "testAction")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write to temporary file: disk full")
}

func testHandlesRenameError(t *testing.T, dir string, filePath string) {
	// Override global mcpDir for this test
	oldMcpDir := mcpDir
	mcpDir = dir
	defer func() { mcpDir = oldMcpDir }()

	m := NewMcpActions()
	removed := false
	removedTestFile := false
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
		if name == fmt.Sprintf("%s.%d.tmp", filePath, os.Getpid()) {
			removed = true
			return nil
		}
		if name == dir+"/.write_test" {
			removedTestFile = true
			return nil
		}
		t.Fatalf("osRemove called with unexpected file: %s", name)
		return nil
	}
	m.osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		// Mock file creation for write test
		return &os.File{}, nil
	}

	err := updateActionCount(m, "testAction")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rename failed")
	assert.True(t, removed, "temporary file should be removed on rename failure")
	assert.True(t, removedTestFile, "test file should be removed")
}

func TestUpdateActionCount(t *testing.T) {
	dir, err := os.MkdirTemp("", "suseconnect-mcp-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	filePath := dir + "/suseconnectmcp"

	t.Run("increment count of specific action", func(t *testing.T) {
		testIncCount(t)
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
