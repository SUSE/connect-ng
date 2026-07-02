package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/SUSE/connect-ng/internal/connect"
	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// McpActions handles the file system operations for MCP actions.
type McpActions struct {
	actions     []Action
	osStat      func(string) (os.FileInfo, error)
	osMkdirAll  func(string, os.FileMode) error
	osReadFile  func(string) ([]byte, error)
	osWriteFile func(string, []byte, os.FileMode) error
	osRename    func(string, string) error
	osRemove    func(string) error
}

// NewMcpActions creates a new McpActions with default file system functions.
func NewMcpActions() *McpActions {
	return &McpActions{
		osStat:      os.Stat,
		osMkdirAll:  os.MkdirAll,
		osReadFile:  os.ReadFile,
		osWriteFile: os.WriteFile,
		osRename:    os.Rename,
		osRemove:    os.Remove,
	}
}

type Action struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type ToolInput struct{}

type RegisterInput struct {
	Regcode string `json:"regcode" jsonschema:"The subscription registration code to register the system with"`
	Email   string `json:"email,omitempty" jsonschema:"Email Address to associate the registration with"`
}

type ActivateInput struct {
	Regcode string `json:"regcode,omitempty" jsonschema:"The subscription registration code to register the system with"`
	Product string `json:"product" jsonschema:"The product to activate on the system, e.g. 'sle-module-basesystem/15.5/x86_64'. The system needs to be registered first. Available extensions and modules to activate can be found via the ListExtensions tool."`
	Email   string `json:"email,omitempty" jsonschema:"Email Address to associate the registration with"`
}

type DeactivateInput struct {
	Product string `json:"product" jsonschema:"The product to deactivate, e.g. 'sle-module-basesystem/15.5/x86_64'"`
}

type JSONOutput struct {
	Response string `json:"response" jsonschema:"the response from the tool"`
	Error    string `json:"error,omitempty" jsonschema:"the error message if the tool failed"`
}

var (
	_ func(context.Context, *mcp.CallToolRequest, ToolInput) (*mcp.CallToolResult, JSONOutput, error)       = RegistrationStatus
	_ func(context.Context, *mcp.CallToolRequest, ToolInput) (*mcp.CallToolResult, JSONOutput, error)       = ListExtensions
	_ func(context.Context, *mcp.CallToolRequest, RegisterInput) (*mcp.CallToolResult, JSONOutput, error)   = RegisterSystem
	_ func(context.Context, *mcp.CallToolRequest, ToolInput) (*mcp.CallToolResult, JSONOutput, error)       = DeregisterSystem
	_ func(context.Context, *mcp.CallToolRequest, ActivateInput) (*mcp.CallToolResult, JSONOutput, error)   = ActivateProduct
	_ func(context.Context, *mcp.CallToolRequest, DeactivateInput) (*mcp.CallToolResult, JSONOutput, error) = DeactivateProduct
)

var mcpDir = "/var/lib/suseconnect-mcp"

// incCount increments the count of a specific action.
func (m *McpActions) incCount(actionName string) {
	found := false
	for i, action := range m.actions {
		if action.Name == actionName {
			m.actions[i].Count++
			found = true
			break
		}
	}

	if !found {
		m.actions = append(m.actions, Action{Name: actionName, Count: 1})
	}
}

func updateActionCount(actionName string) error {
	m := NewMcpActions()
	filePath := mcpDir + "/suseconnectmcp"

	if _, err := m.osStat(mcpDir); os.IsNotExist(err) {
		err = m.osMkdirAll(mcpDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Ensure that the process has write access to the directory, otherwise fail early
		testFile := mcpDir + "/.write_test"
		f, openErr := os.OpenFile(testFile, os.O_WRONLY|os.O_CREATE, 0664)
		if openErr != nil {
			return fmt.Errorf("no write access to %s: %w", mcpDir, openErr)
		}
		f.Close()

		err = m.osRemove(testFile)
		if err != nil {
			// This shouldn't fail, but if it does, it's not critical
			slog.Warn("failed to remove temporary file", "file", testFile, "error", err)
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
	err = m.osWriteFile(tempFilePath, data, 0664)
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

func RegistrationStatus(ctx context.Context, req *mcp.CallToolRequest, input ToolInput) (
	*mcp.CallToolResult, JSONOutput, error) {
	slog.Info("RegistrationStatus tool called")

	if err := updateActionCount("RegistrationStatus"); err != nil {
		return nil, JSONOutput{Error: "Failed to update Action Count"}, err
	}

	opts, err := connect.ReadFromConfiguration(connect.DefaultConfigPath)
	if err != nil {
		return nil, JSONOutput{Error: "Failed to read SUSEConnect configuration"}, err
	}

	statuses, err := connect.GetProductStatuses(opts, connect.StatusJSON)
	if err != nil {
		return nil, JSONOutput{Error: "Failed to retrieve registration status"}, err
	}

	return nil, JSONOutput{Response: statuses}, nil
}

func ListExtensions(ctx context.Context, req *mcp.CallToolRequest, input ToolInput) (
	*mcp.CallToolResult, JSONOutput, error) {
	slog.Info("ListExtensions tool called")

	if err := updateActionCount("ListExtensions"); err != nil {
		return nil, JSONOutput{Error: "Failed to update Action Count"}, err
	}

	opts, err := connect.ReadFromConfiguration(connect.DefaultConfigPath)
	if err != nil {
		return nil, JSONOutput{Error: "Failed to read SUSEConnect configuration"}, err
	}

	api := connect.NewWrappedAPI(opts)
	tree, err := connect.RenderExtensionTree(api, true)
	if err != nil {
		if errors.Is(err, connect.ErrListExtensionsUnregistered) {
			return nil, JSONOutput{}, fmt.Errorf("System is not registered; Extension listing requires the system to register first.")
		}
		return nil, JSONOutput{Response: ""}, fmt.Errorf("Failed to list extensions: %w", err)
	}

	return nil, JSONOutput{Response: tree}, nil
}

func RegisterSystem(ctx context.Context, req *mcp.CallToolRequest, input RegisterInput) (
	*mcp.CallToolResult, JSONOutput, error) {
	slog.Info("RegisterSystem tool called")

	if err := updateActionCount("RegisterSystem"); err != nil {
		return nil, JSONOutput{Error: "Failed to update Action Count"}, err
	}

	opts, err := connect.ReadFromConfiguration(connect.DefaultConfigPath)
	if err != nil {
		return nil, JSONOutput{Error: "Failed to read SUSEConnect configuration"}, err
	}
	opts.Token = input.Regcode
	opts.Email = input.Email

	api := connect.NewWrappedAPI(opts)
	err = connect.Register(api, opts)
	if err != nil {
		return nil, JSONOutput{Error: "System registration failed"}, err
	}

	return nil, JSONOutput{Response: "System successfully registered"}, nil
}

func ActivateProduct(ctx context.Context, req *mcp.CallToolRequest, input ActivateInput) (
	*mcp.CallToolResult, JSONOutput, error) {
	slog.Info("ActivateProduct tool called")

	if err := updateActionCount("ActivateProduct"); err != nil {
		return nil, JSONOutput{Error: "Failed to update Action Count"}, err
	}

	opts, err := connect.ReadFromConfiguration(connect.DefaultConfigPath)
	if err != nil {
		return nil, JSONOutput{Error: "Failed to read SUSEConnect configuration"}, err
	}
	opts.Token = input.Regcode
	opts.Email = input.Email
	if p, err := registration.FromTriplet(input.Product); err != nil {
		return nil, JSONOutput{Error: "Please provide the product identifier in this format: <internal name>/<version>/<architecture>. You can find these values in the ListExtensions tool"}, err
	} else {
		opts.Product = p
	}

	api := connect.NewWrappedAPI(opts)
	err = connect.Register(api, opts)
	if err != nil {
		return nil, JSONOutput{Error: "System registration failed"}, err
	}

	return nil, JSONOutput{Response: "System successfully registered"}, nil
}

func DeregisterSystem(ctx context.Context, req *mcp.CallToolRequest, input ToolInput) (
	*mcp.CallToolResult, JSONOutput, error) {
	slog.Info("DeregisterSystem tool called")

	if err := updateActionCount("DeregisterSystem"); err != nil {
		return nil, JSONOutput{Error: "Failed to update Action Count"}, err
	}

	opts, err := connect.ReadFromConfiguration(connect.DefaultConfigPath)
	if err != nil {
		return nil, JSONOutput{Error: "Failed to read SUSEConnect configuration"}, err
	}

	api := connect.NewWrappedAPI(opts)
	err = connect.Deregister(api, opts)
	if err != nil {
		return nil, JSONOutput{Error: "System deregistration failed"}, err
	}

	return nil, JSONOutput{Response: "System successfully deregistered"}, nil
}

func DeactivateProduct(ctx context.Context, req *mcp.CallToolRequest, input DeactivateInput) (
	*mcp.CallToolResult, JSONOutput, error) {
	slog.Info("DeactivateProduct tool called")

	if err := updateActionCount("DeactivateProduct"); err != nil {
		return nil, JSONOutput{Error: "Failed to update Action Count"}, err
	}

	opts, err := connect.ReadFromConfiguration(connect.DefaultConfigPath)
	if err != nil {
		return nil, JSONOutput{Error: "Failed to read SUSEConnect configuration"}, err
	}

	if p, err := registration.FromTriplet(input.Product); err != nil {
		return nil, JSONOutput{Error: "Please provide the product identifier in this format: <internal name>/<version>/<architecture>. You can find these values in the ListExtensions tool"}, err
	} else {
		opts.Product = p
	}

	api := connect.NewWrappedAPI(opts)
	err = connect.Deregister(api, opts)
	if err != nil {
		return nil, JSONOutput{Error: "Product deactivation failed"}, err
	}

	return nil, JSONOutput{Response: "Product successfully deactivated"}, nil
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Root privileges are required to run the MCP server.")
		os.Exit(1)
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "suseconnect", Version: "v0.0.1"}, nil)

	// DestructiveHint is *bool (unset defaults to true); Go won't let us take &false.
	ptr := func(b bool) *bool { return &b }

	mcp.AddTool(server, &mcp.Tool{
		Name:        "RegistrationStatus",
		Description: "Tool to output the registration status of the system and activated/non-activated installed products",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Show registration status",
			ReadOnlyHint: true,
		},
	}, RegistrationStatus)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ListExtensions",
		Description: "List available extension products for your SUSE system. Your system's base product must be activated first.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "List available extensions",
			ReadOnlyHint: true,
		},
	}, ListExtensions)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "RegisterSystem",
		Description: "Registers and activates your SUSE system. This will enable access to online repositories and additional extensions and modules.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Register system",
			ReadOnlyHint:    false,
			DestructiveHint: ptr(false),
			IdempotentHint:  false,
		},
	}, RegisterSystem)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ActivateProduct",
		Description: "Activates and additional extension product or module on your SUSE system. Available extensions can get queried with the ListExtensions tool.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Activate product",
			ReadOnlyHint:    false,
			DestructiveHint: ptr(false),
			IdempotentHint:  false,
		},
	}, ActivateProduct)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "DeregisterSystem",
		Description: "Deregisters your SUSE system. This will remove the system's registration and disable access to online repositories.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Deregister system (destructive)",
			ReadOnlyHint:    false,
			DestructiveHint: ptr(true),
			IdempotentHint:  true,
		},
	}, DeregisterSystem)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "DeactivateProduct",
		Description: "Deactivates an extension product or module on your SUSE system.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Deactivate product (destructive)",
			ReadOnlyHint:    false,
			DestructiveHint: ptr(true),
			IdempotentHint:  true,
		},
	}, DeactivateProduct)

	// Run the server on the stdio transport.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		slog.Error("Server failed", "error", err)
	}
}
