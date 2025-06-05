package connect

import "github.com/SUSE/connect-ng/pkg/registration"

// Service represents an installed service or service information from API
type Service struct {
	ID            int                  `json:"id"`
	URL           string               `xml:"url,attr" json:"url"`
	Name          string               `xml:"name,attr" json:"name"`
	Product       registration.Product `json:"product"`
	ObsoletedName string               `json:"obsoleted_service_name"`
}
