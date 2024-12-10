package registration

// Product as defined from SCC'S API.
type Product struct {
	Name    string `json:"identifier"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
	Summary string `json:"summary,omitempty"`
	IsBase  bool   `json:"isbase"`

	FriendlyName string `json:"friendly_name,omitempty"`
	ReleaseType  string `json:"release_type,omitempty"`
	Available    bool   `json:"available"`
	Free         bool   `json:"free"`
	Recommended  bool   `json:"recommended"`

	// optional extension products
	Extensions []Product `json:"extensions,omitempty"`

	Description  string       `json:"description,omitempty"`
	EULAURL      string       `json:"eula_url,omitempty"`
	FormerName   string       `json:"former_identifier,omitempty"`
	ProductType  string       `json:"product_type,omitempty"`
	ShortName    string       `json:"shortname,omitempty"`
	LongName     string       `json:"name,omitempty"`
	ReleaseStage string       `json:"release_stage,omitempty"`
	Repositories []Repository `json:"repositories,omitempty"`
}
