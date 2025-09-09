package collectors

type Architecture struct{}

func (Architecture) run(arch string) (Result, error) {
	// NOTE: For now, just playback the provided architecture value which
	// is itself gathered by DetectArchitecture method in this module.
	// In the future this will change since we will collect more architecture
	// specific information here
	return Result{"arch": arch}, nil
}

var DetectArchitecture = func() (string, error) {
	output, err := uname([]string{"-i"})
	if err != nil {
		return "", err
	}
	if output == "unknown" {
		return uname([]string{"-m"})
	}
	return output, nil
}
