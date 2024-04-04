package collectors

import (
	"github.com/SUSE/connect-ng/internal/util"
	"sync"
)

type ArchitectureInformation struct {
	_json_key string
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
