package models

type Extension struct {
	Name         string       `json:"identifier"`
	Version      string       `json:"version"`
	Arch         string       `json:"arch"`
	FriendlyName string       `json:"name"`
	Activated    bool         `json:"activated"`
	Available    bool         `json:"available"`
	Free         bool         `json:"free"`
	Extensions   []*Extension `json:"extensions"`
}
