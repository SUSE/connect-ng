package labels

import (
	"encoding/json"
	"fmt"

	"github.com/SUSE/connect-ng/pkg/connection"
)

// UnassignLabel removes a label from a system in SCC. The label itself will not get
// deleted from the account.
// If the label should be deleted, have a look into the organization API of SCC.
// see: https://scc.suse.com/connect/v4/documentation#/organizations/delete_organizations_labels__id_
func UnassignLabel(conn connection.Connection, labelId int) ([]Label, error) {
	labels := []Label{}
	url := fmt.Sprintf("/connect/systems/labels/%d", labelId)

	request, buildErr := conn.BuildRequest("DELETE", url, nil)
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

	if err := json.Unmarshal(response, &labels); err != nil {
		return []Label{}, err
	}

	return labels, nil
}
