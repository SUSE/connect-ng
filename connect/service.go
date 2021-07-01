package connect

// Service represents an installed service or service information from API
type Service struct {
	ID            int     `json:"id"`
	URL           string  `json:"url"`
	Name          string  `json:"name"`
	Product       Product `json:"product"`
	ObsoletedName string  `json:"obsoleted_service_name"`
}
