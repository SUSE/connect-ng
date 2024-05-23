package collectors

import "github.com/SUSE/connect-ng/internal/util"

type Architecture struct{}

func (Architecture) run(arch string) (Result, error) {
	// NOTE: For now, just playback the provided architecture value which
	// is itself gathered by DetectArchitecture method in this module.
	// In the future this will change since we will collect more architecture
	// specific information here
	return Result{"arch": arch}, nil
}

var DetectArchitecture = func() (string, error) {
	output, err := util.Execute([]string{"uname", "-i"}, nil)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
