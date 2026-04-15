package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/SUSE/connect-ng/internal/connect"
	"github.com/SUSE/connect-ng/pkg/registration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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

func RegistrationStatus(ctx context.Context, req *mcp.CallToolRequest, input ToolInput) (
	*mcp.CallToolResult, JSONOutput, error) {
	slog.Info("RegistrationStatus tool called")

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
	listenAddr := flag.String("http", "", "address for http transport, defaults to stdio")
	flag.Parse()

	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Root privileges are required to run the MCP server.")
		os.Exit(1)
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "suseconnect", Version: "v0.0.1"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "RegistrationStatus",
		Description: "Tool to output the registration status of the system and activated/non-activated installed products",
	}, RegistrationStatus)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ListExtensions",
		Description: "List available extension products for your SUSE system. Your system's base product must be activated first.",
	}, ListExtensions)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "RegisterSystem",
		Description: "Registers and activates your SUSE system. This will enable access to online repositories and additional extensions and modules.",
	}, RegisterSystem)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ActivateProduct",
		Description: "Activates and additional extension product or module on your SUSE system. Available extensions can get queried with the ListExtensions tool.",
	}, ActivateProduct)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "DeregisterSystem",
		Description: "Deregisters your SUSE system. This will remove the system's registration and disable access to online repositories.",
	}, DeregisterSystem)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "DeactivateProduct",
		Description: "Deactivates an extension product or module on your SUSE system.",
	}, DeactivateProduct)

	if *listenAddr == "" {
		// Run the server on the stdio transport.
		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			slog.Error("Server failed", "error", err)
		}
	} else {
		// Create a streamable HTTP handler.
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil)

		// Run the server on the HTTP transport.
		slog.Info("Server listening", "address", *listenAddr)
		if err := http.ListenAndServe(*listenAddr, handler); err != nil {
			slog.Error("Server failed", "error", err)
		}
	}
}
