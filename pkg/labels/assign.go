package labels

import (
	"encoding/json"

	"github.com/SUSE/connect-ng/pkg/connection"
)

type assignLabelsRequestResponse struct {
	Labels []Label `json:"labels"`
}

// Assign manages labels in SCC. It is possible to set labels and alter them.
// Supplying already existing labels will not recreate them. Labels not
// existing in SCC will get created automatically.
func AssignLabels(conn connection.Connection, labels []Label) ([]Label, error) {
	updated := []Label{}
	payload := assignLabelsRequestResponse{
		Labels: labels,
	}

	request, buildErr := conn.BuildRequest("POST", "/connect/systems/labels", payload)
	if buildErr != nil {
		return []Label{}, buildErr
	}

	login, password, credErr := conn.GetCredentials().Login()
	if credErr != nil {
		return []Label{}, credErr
	}

	connection.AddSystemAuth(request, login, password)

	_, response, doErr := conn.Do(request)
	if doErr != nil {
		return []Label{}, doErr
	}

	if err := json.Unmarshal(response, &updated); err != nil {
		return []Label{}, err
	}

	return updated, nil
}
