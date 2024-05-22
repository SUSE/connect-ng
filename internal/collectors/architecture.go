package collectors

import "github.com/SUSE/connect-ng/internal/util"

type Architecture struct{}

func (Architecture) run(arch string) (Result, error) {
	// For now just a playback of the already existing value but in future
	// this collector can collect architecture specific information such as SoC information
	return Result{"arch": arch}, nil
}

var DetectArchitecture = func() (string, error) {
	output, err := util.Execute([]string{"uname", "-i"}, nil)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
