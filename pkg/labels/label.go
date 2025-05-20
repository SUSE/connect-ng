package labels

// A label represents a system label in SCC which can be assigned and unassigned from a system using the library
type Label struct {
	Id          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
