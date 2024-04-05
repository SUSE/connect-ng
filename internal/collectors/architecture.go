package collectors

import (
	"encoding/json"
	"sync"

	"github.com/SUSE/connect-ng/internal/util"
)

type ArchitectureInformation struct {
	_json_key string
	_json_val Result
}

var archCollectorInstance *ArchitectureInformation
var singleton sync.Once

func New() *ArchitectureInformation {
	init := func() {
		archCollectorInstance = &ArchitectureInformation{_json_key: "arch"}
	}
	singleton.Do(init)
	return archCollectorInstance
}

func (a *ArchitectureInformation) get_key() string {
	return a._json_key
}

func (a *ArchitectureInformation) run(arch Architecture) (Result, error) {
	switch arch {
	default:
		return a.arch()
	}
}
func (a *ArchitectureInformation) arch() (Result, error) {
	output, err := util.Execute([]string{"uname", "-i"}, nil)
	if err != nil {
		return nil, err
	}
	m := Result{}
	m[a.get_key()] = output
	return m, nil
}

// write unmarshal function
// func (a *ArchitectureInformation) UnmarshalJSON(data []byte) error {
// 	var m map[string]interface{}
// 	if err := json.Unmarshal(data, &m); err != nil {
// 		return err
// 	}
// 	a._json_key = "arch"
// 	a._json_val = m
// 	return nil
// }

func (a *ArchitectureInformation) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Arch string `json:"arch"`
	}{
		Arch: a._json_val[a.get_key()].(string),
	})
}
