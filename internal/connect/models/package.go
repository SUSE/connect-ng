package models

// Package holds package info as returned by `zypper search`
type Package struct {
	Name    string `xml:"name,attr"`
	Edition string `xml:"edition,attr"` // VERSION[-RELEASE]
	Arch    string `xml:"arch,attr"`
	Repo    string `xml:"repository,attr"`
}

// from package_search.go
// SearchPackageResult represents package search result
type SearchPackageResult struct {
	ID       int                    `json:"id"`
	Name     string                 `json:"name"`
	Arch     string                 `json:"arch"`
	Version  string                 `json:"version"`
	Release  string                 `json:"release"`
	Products []SearchPackageProduct `json:"products"`
}
