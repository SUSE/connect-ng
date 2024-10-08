package connect

import (
	"github.com/SUSE/connect-ng/internal/util"
	"strings"
)

var (
	localSetLabels = setLabels
)

type Label struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func AssignAndCrateLabels(labels []string) error {
	collection := []Label{}

	for _, name := range labels {
		name = strings.TrimSpace(name)
		collection = append(collection, Label{Name: name})
	}

	util.Debug.Printf(util.Bold("Setting Labels %s"), strings.Join(labels, ","))
	return localSetLabels(collection)
}
