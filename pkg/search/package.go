package search

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/registration"
)

// The result as given from the /package_search API.
type SearchPackageResult struct {
	ID       int                    `json:"id"`
	Name     string                 `json:"name"`
	Arch     string                 `json:"arch"`
	Version  string                 `json:"version"`
	Release  string                 `json:"release"`
	Products []registration.Product `json:"products"`
}

// Performs a package search request for the given search `query`. It also
// expects the triplet identifier for the product to be used as a `base`.
func Package(conn connection.Connection, query, base string) ([]SearchPackageResult, error) {
	args := map[string]string{"product_id": base, "query": query}
	var packages struct {
		Data []SearchPackageResult `json:"data"`
	}

	request, err := conn.BuildRequest("GET", "/api/package_search/packages", args)
	if err != nil {
		return packages.Data, err
	}

	response, doErr := conn.Do(request)
	if doErr != nil {
		if ae, ok := doErr.(*connection.ApiError); ok && ae.Code == http.StatusNotFound && ae.Message == "" {
			return packages.Data, fmt.Errorf("SUSE::Connect::UnsupportedOperation: " +
				"Package search is not supported by the registration proxy: " +
				"Alternatively, use the web version at https://scc.suse.com/packages/")
		}
		return packages.Data, doErr
	}

	if err = json.Unmarshal(response, &packages); err != nil {
		return packages.Data, err
	}
	return packages.Data, nil
}
