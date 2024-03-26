package zypper

import "encoding/xml"

// Repository holds repository data as returned by `zypper repos` or "show_product" API
type Repository struct {
	// SCC docs say that "id" should be integer but SMT returns string sometimes
	// not mapping to struct field as it doesn't seem to be used by Connect
	Name     string `xml:"name,attr" json:"name"`
	Alias    string `xml:"alias,attr" json:"-"`
	Type     string `xml:"type,attr" json:"-"`
	Priority int    `xml:"priority,attr" json:"-"`
	Enabled  bool   `xml:"enabled,attr" json:"enabled"`
	URL      string `xml:"url" json:"url"`

	DistroTarget     string   `json:"distro_target,omitempty"`
	Description      string   `json:"description,omitempty"`
	AutoRefresh      bool     `json:"autorefresh"`
	InstallerUpdates bool     `json:"installer_updates"`
	Arch             []string `json:"arch,omitempty"`
}

func parseReposXML(xmlDoc []byte) ([]Repository, error) {
	var repos struct {
		Repos []Repository `xml:"repo-list>repo"`
	}
	if err := xml.Unmarshal(xmlDoc, &repos); err != nil {
		return []Repository{}, err
	}
	return repos.Repos, nil
}

// Repos returns repositories configured on the system
func Repos() ([]Repository, error) {
	args := []string{"--xmlout", "--non-interactive", "repos", "-d"}
	// Don't fail when zypper exits with 6 (no repositories)
	output, err := zypperRun(args, []int{zypperOK, zypperErrNoRepos})
	if err != nil {
		return []Repository{}, err
	}
	return parseReposXML(output)
}
