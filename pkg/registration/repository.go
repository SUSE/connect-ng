package registration

// Repository as defined by SCC's API.
type Repository struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`

	DistroTarget     string   `json:"distro_target,omitempty"`
	Description      string   `json:"description,omitempty"`
	AutoRefresh      bool     `json:"autorefresh"`
	InstallerUpdates bool     `json:"installer_updates"`
	Arch             []string `json:"arch,omitempty"`
}
