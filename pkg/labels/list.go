package labels

import (
	"encoding/json"

	"github.com/SUSE/connect-ng/pkg/connection"
)

// ListLabels fetches the currently assigned labels for this system in SCC.
func ListLabels(conn connection.Connection) ([]Label, error) {
	labels := []Label{}

	request, buildErr := conn.BuildRequest("GET", "/connect/systems/labels", nil)
	if buildErr != nil {
		return []Label{}, buildErr
	}

	login, password, credErr := conn.GetCredentials().Login()
	if credErr != nil {
		return []Label{}, credErr
	}

	connection.AddSystemAuth(request, login, password)

	response, doErr := conn.Do(request)
	if doErr != nil {
		return []Label{}, doErr
	}

	if err := json.Unmarshal(response, &labels); err != nil {
		return []Label{}, err
	}

	return labels, nil
}
