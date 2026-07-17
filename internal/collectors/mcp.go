package collectors

import (
	"encoding/json"
	"os"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/pkg/profiles"
)

const mcpChecksumFile = "mcp-profile-id"
const mcpTag = "mcp_stats"
const mcpPath = "/var/lib/suseconnect-mcp/suseconnectmcp"

var mcpOsReadfile = os.ReadFile

type Action struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type MCP struct {
	UpdateDataIDs bool
}

func getSuseconnectmcpData(inputStream []byte) ([]Action, error) {
	var actions []Action
	err := json.Unmarshal(inputStream, &actions)
	if err != nil {
		return nil, err
	}
	return actions, nil
}

func (mcp MCP) run(arch string) (Result, error) {
	util.Debug.Print("mcp.UpdateDataIDs: ", mcp.UpdateDataIDs)
	output, err := mcpOsReadfile(mcpPath)
	if err != nil {
		return Result{}, err
	}
	mcpData, _ := getSuseconnectmcpData(output)
	result, err := profiles.BuildProfile(mcp.UpdateDataIDs, mcpTag, mcpChecksumFile, mcpData)

	return result, err
}
