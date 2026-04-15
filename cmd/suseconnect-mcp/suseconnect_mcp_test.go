package main

import (
	"context"
	"testing"

	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestProductTripletParsing(t *testing.T) {
	assert := assert.New(t)

	t.Run("valid triplets with component validation", func(t *testing.T) {
		tests := []struct {
			triplet      string
			expectedName string
			expectedVer  string
			expectedArch string
		}{
			{
				triplet:      "sle-module-basesystem/15.5/x86_64",
				expectedName: "sle-module-basesystem",
				expectedVer:  "15.5",
				expectedArch: "x86_64",
			},
			{
				triplet:      "SLES/15/aarch64",
				expectedName: "SLES",
				expectedVer:  "15",
				expectedArch: "aarch64",
			},
			{
				triplet:      "sle-module-python3/15.6/ppc64le",
				expectedName: "sle-module-python3",
				expectedVer:  "15.6",
				expectedArch: "ppc64le",
			},
			{
				triplet:      "sle-module-server-applications/15.4/s390x",
				expectedName: "sle-module-server-applications",
				expectedVer:  "15.4",
				expectedArch: "s390x",
			},
			{
				triplet:      "sle-module-containers/15.3/x86_64",
				expectedName: "sle-module-containers",
				expectedVer:  "15.3",
				expectedArch: "x86_64",
			},
		}

		for _, tt := range tests {
			t.Run(tt.triplet, func(t *testing.T) {
				product, err := registration.FromTriplet(tt.triplet)
				assert.NoError(err)
				assert.NotNil(product)
				assert.Equal(tt.expectedName, product.Identifier)
				assert.Equal(tt.expectedVer, product.Version)
				assert.Equal(tt.expectedArch, product.Arch)
			})
		}
	})

	t.Run("invalid triplets", func(t *testing.T) {
		tests := []struct {
			triplet string
			name    string
		}{
			{triplet: "", name: "empty string"},
			{triplet: "single", name: "single part"},
			{triplet: "two/parts", name: "two parts"},
			{triplet: "four/parts/too/many", name: "four parts"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := registration.FromTriplet(tt.triplet)
				assert.Error(err, "Expected error for: %s (%s)", tt.triplet, tt.name)
			})
		}
	})
}

func TestInputStructures(t *testing.T) {
	assert := assert.New(t)

	t.Run("ToolInput", func(t *testing.T) {
		input := ToolInput{}
		assert.NotNil(input)
	})

	t.Run("RegisterInput", func(t *testing.T) {
		input := RegisterInput{
			Regcode: "test-code",
			Email:   "test@example.com",
		}
		assert.Equal("test-code", input.Regcode)
		assert.Equal("test@example.com", input.Email)
	})

	t.Run("RegisterInput with empty email", func(t *testing.T) {
		input := RegisterInput{
			Regcode: "test-code",
			Email:   "",
		}
		assert.Equal("test-code", input.Regcode)
		assert.Empty(input.Email)
	})

	t.Run("ActivateInput", func(t *testing.T) {
		input := ActivateInput{
			Regcode: "test-code",
			Product: "sle-module/15.5/x86_64",
			Email:   "test@example.com",
		}
		assert.Equal("test-code", input.Regcode)
		assert.Equal("sle-module/15.5/x86_64", input.Product)
		assert.Equal("test@example.com", input.Email)
	})

	t.Run("ActivateInput with optional fields", func(t *testing.T) {
		input := ActivateInput{
			Product: "sle-module/15.5/x86_64",
		}
		assert.Equal("sle-module/15.5/x86_64", input.Product)
		assert.Empty(input.Regcode)
		assert.Empty(input.Email)
	})

	t.Run("DeactivateInput", func(t *testing.T) {
		input := DeactivateInput{
			Product: "sle-module/15.5/x86_64",
		}
		assert.Equal("sle-module/15.5/x86_64", input.Product)
	})
}

func TestJSONOutputStructure(t *testing.T) {
	assert := assert.New(t)

	t.Run("success output", func(t *testing.T) {
		output := JSONOutput{
			Response: "Operation successful",
			Error:    "",
		}
		assert.NotEmpty(output.Response)
		assert.Empty(output.Error)
	})

	t.Run("error output", func(t *testing.T) {
		output := JSONOutput{
			Response: "",
			Error:    "Operation failed",
		}
		assert.Empty(output.Response)
		assert.NotEmpty(output.Error)
	})

	t.Run("both fields can be populated", func(t *testing.T) {
		output := JSONOutput{
			Response: "Partial response",
			Error:    "Warning message",
		}
		assert.NotEmpty(output.Response)
		assert.NotEmpty(output.Error)
	})
}

func TestToolFunctionSignatures(t *testing.T) {
	t.Run("all function signatures match MCP requirements", func(t *testing.T) {
		// These assignments will fail at compile time if signatures are wrong
		var _ func(context.Context, *mcp.CallToolRequest, ToolInput) (*mcp.CallToolResult, JSONOutput, error) = RegistrationStatus
		var _ func(context.Context, *mcp.CallToolRequest, ToolInput) (*mcp.CallToolResult, JSONOutput, error) = ListExtensions
		var _ func(context.Context, *mcp.CallToolRequest, RegisterInput) (*mcp.CallToolResult, JSONOutput, error) = RegisterSystem
		var _ func(context.Context, *mcp.CallToolRequest, ToolInput) (*mcp.CallToolResult, JSONOutput, error) = DeregisterSystem
		var _ func(context.Context, *mcp.CallToolRequest, ActivateInput) (*mcp.CallToolResult, JSONOutput, error) = ActivateProduct
		var _ func(context.Context, *mcp.CallToolRequest, DeactivateInput) (*mcp.CallToolResult, JSONOutput, error) = DeactivateProduct
	})
}

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

func TestValidInputParsing(t *testing.T) {
	assert := assert.New(t)

	t.Run("valid product triplets for ActivateProduct", func(t *testing.T) {
		validTriplets := []struct {
			triplet      string
			expectedName string
			expectedVer  string
			expectedArch string
		}{
			{
				triplet:      "sle-module-basesystem/15.5/x86_64",
				expectedName: "sle-module-basesystem",
				expectedVer:  "15.5",
				expectedArch: "x86_64",
			},
			{
				triplet:      "sle-module-server-applications/15.4/aarch64",
				expectedName: "sle-module-server-applications",
				expectedVer:  "15.4",
				expectedArch: "aarch64",
			},
			{
				triplet:      "SLES/15/x86_64",
				expectedName: "SLES",
				expectedVer:  "15",
				expectedArch: "x86_64",
			},
			{
				triplet:      "sle-module-python3/15.6/ppc64le",
				expectedName: "sle-module-python3",
				expectedVer:  "15.6",
				expectedArch: "ppc64le",
			},
		}

		for _, tt := range validTriplets {
			t.Run(tt.triplet, func(t *testing.T) {
				product, err := registration.FromTriplet(tt.triplet)
				assert.NoError(err, "Triplet should be valid: %s", tt.triplet)
				assert.NotNil(product)
				assert.Equal(tt.expectedName, product.Identifier)
				assert.Equal(tt.expectedVer, product.Version)
				assert.Equal(tt.expectedArch, product.Arch)
			})
		}
	})

	t.Run("valid product triplets for DeactivateProduct", func(t *testing.T) {
		validTriplets := []struct {
			triplet      string
			expectedName string
			expectedVer  string
			expectedArch string
		}{
			{
				triplet:      "sle-module-basesystem/15.5/x86_64",
				expectedName: "sle-module-basesystem",
				expectedVer:  "15.5",
				expectedArch: "x86_64",
			},
			{
				triplet:      "SLES/15/aarch64",
				expectedName: "SLES",
				expectedVer:  "15",
				expectedArch: "aarch64",
			},
			{
				triplet:      "sle-module-containers/15.3/s390x",
				expectedName: "sle-module-containers",
				expectedVer:  "15.3",
				expectedArch: "s390x",
			},
		}

		for _, tt := range validTriplets {
			t.Run(tt.triplet, func(t *testing.T) {
				product, err := registration.FromTriplet(tt.triplet)
				assert.NoError(err, "Triplet should be valid: %s", tt.triplet)
				assert.NotNil(product)
				assert.Equal(tt.expectedName, product.Identifier)
				assert.Equal(tt.expectedVer, product.Version)
				assert.Equal(tt.expectedArch, product.Arch)
			})
		}
	})

	t.Run("RegisterSystem with all fields", func(t *testing.T) {
		input := RegisterInput{
			Regcode: "SUSE-TEST-CODE-12345",
			Email:   "test@suse.com",
		}

		assert.NotEmpty(input.Regcode)
		assert.NotEmpty(input.Email)
	})

	t.Run("RegisterSystem with minimal fields", func(t *testing.T) {
		input := RegisterInput{
			Regcode: "SUSE-TEST-CODE",
		}

		assert.NotEmpty(input.Regcode)
		assert.Empty(input.Email)
	})
}
