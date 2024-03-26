package models

// Service represents an installed service or service information from API
// Stage 2 suggestion : Create two separate service structs - one for SCC_Service and Zypper_Sevice
// so the source is explicit
type Service struct {
	ID            int     `json:"id"`
	URL           string  `xml:"url,attr" json:"url"`
	Name          string  `xml:"name,attr" json:"name"`
	Product       Product `json:"product"`
	ObsoletedName string  `json:"obsoleted_service_name"`
}

// from client.go
type ServiceOut struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}
