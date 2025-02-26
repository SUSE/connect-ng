package registration

import "time"

// A subscription used in the offline registration workflow
type Subscription struct {
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
