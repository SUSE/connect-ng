package connect

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// NOTE: package_search API is not related to connect API
//       the models are different and so are the structures below
// docs: https://scc.suse.com/api/package_search/v4/documentation

func searchPackage(query string, baseProduct Product) ([]SearchPackageResult, error) {
	args := map[string]string{
		"product_id": baseProduct.ToTriplet(),
		"query":      query,
	}
	var packages struct {
		Data []SearchPackageResult `json:"data"`
	}
	resp, err := callHTTP("GET", "/api/package_search/packages", nil, args, authNone)
	if err != nil {
		if ae, ok := err.(APIError); ok && ae.Code == http.StatusNotFound && ae.Message == "" {
			return packages.Data, fmt.Errorf("SUSE::Connect::UnsupportedOperation: " +
				"Package search is not supported by the registration proxy: " +
				"Alternatively, use the web version at https://scc.suse.com/packages/")
		}
		return packages.Data, err
	}
	if err = json.Unmarshal(resp, &packages); err != nil {
		return packages.Data, JSONError{err}
	}
	return packages.Data, nil
}
