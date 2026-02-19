package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"

	"github.com/SUSE/connect-ng/internal/connect"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ToolInput struct{}

type JSONOutput struct {
	Response string `json:"response" jsonschema:"the response from the tool"`
}

func RegistrationStatus(ctx context.Context, req *mcp.CallToolRequest, input ToolInput ) ( 
	*mcp.CallToolResult, JSONOutput, error) {
	slog.Info("RegistrationStatus tool called")
	opts, err := connect.ReadFromConfiguration(connect.DefaultConfigPath)
  if err != nil {
		return nil, JSONOutput{}, err
	}
	statuses, err := connect.GetProductStatuses(opts, connect.StatusJSON)
	if err != nil {
		return nil, JSONOutput{}, err
	}
	return nil, JSONOutput{Response: statuses}, nil
}

func ListExtensions(ctx context.Context, req *mcp.CallToolRequest, input ToolInput) (
	*mcp.CallToolResult, JSONOutput, error) {
	slog.Info("ListExtensions tool called")
	opts, err := connect.ReadFromConfiguration(connect.DefaultConfigPath)
	if err != nil {
		return nil, JSONOutput{}, err
	}
  api := connect.NewWrappedAPI(opts)
	tree, err := connect.RenderExtensionTree(api, true)
	if err != nil {
		return nil, JSONOutput{}, err
	}
	return nil, JSONOutput{Response: tree}, nil
}

func main() {
	listenAddr := flag.String("http", "", "address for http transport, defaults to stdio")
	flag.Parse()

	server := mcp.NewServer(&mcp.Implementation{Name: "suseconnect", Version: "v0.0.1"}, nil)
	
	mcp.AddTool(server, &mcp.Tool{
		Name:        "RegistrationStatus",
		Description: "Tool to output the registration status of the system and activated/non-activated installed products",
	}, RegistrationStatus)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ListExtensions",
		Description: "List available extension products for your SUSE system. Your system's base product must be activated first.",
	}, ListExtensions)

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