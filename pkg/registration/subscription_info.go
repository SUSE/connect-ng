package registration

import (
	"encoding/json"
	"time"

	"github.com/SUSE/connect-ng/pkg/connection"
)

// A subscription used in the offline registration workflow
type SubscriptionInfo struct {
	Kind           string         `json:"kind"`
	Name           string         `json:"name"`
	StartsAt       time.Time      `json:"starts_at"`
	ExpiresAt      time.Time      `json:"expires_at"`
	Limit          int            `json:"limit"`
	Notifications  string         `json:"notifications"`
	ProductClasses []ProductClass `json:"product_classes"`
}

// Describes a whole product class like Rancher Manager or SUSE Enterprise Linux
// A detailed product would include the version.
type ProductClass struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// FetchSubscriptionInfo queries /connect/subscriptions/info to retrieve
// comprehensive subscription metadata including start/expire times and product classes.
// Returns a single SubscriptionInfo object with the subscription details.
func FetchSubscriptionInfo(conn connection.Connection, regcode string) (*SubscriptionInfo, error) {
	request, buildErr := conn.BuildRequest("GET", "/connect/subscriptions/info", nil)
	if buildErr != nil {
		return nil, buildErr
	}

	connection.AddRegcodeAuth(request, regcode)

	response, doErr := conn.Do(request)
	if doErr != nil {
		return nil, doErr
	}

	var info SubscriptionInfo
	if err := json.Unmarshal(response, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// FetchSubscriptionProducts queries /connect/subscriptions/products to retrieve
// the full product list covered by the subscription. Returns an array of Product objects
// with complete details (repositories, extensions, etc).
func FetchSubscriptionProducts(conn connection.Connection, regcode string) ([]Product, error) {
	request, buildErr := conn.BuildRequest("GET", "/connect/subscriptions/products", nil)
	if buildErr != nil {
		return nil, buildErr
	}

	connection.AddRegcodeAuth(request, regcode)

	response, doErr := conn.Do(request)
	if doErr != nil {
		return nil, doErr
	}

	var products []Product
	if err := json.Unmarshal(response, &products); err != nil {
		return nil, err
	}

	return products, nil
}
