package main

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestToolExecutionWithInvalidInputs(t *testing.T) {
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
