package util

var (
	//
	// systemctl settings
	//

	// path to the systemctl command
	SystemctlBin = "/usr/bin/systemctl"

	// extra options to pass to the
	SystemctlBaseCmd = []string{
		SystemctlBin,
		"--no-pager",
		"--nolegend",
	}
)

func Systemctl(cmd string, args ...string) ([]byte, error) {
	// construct systemctl command line
	cmdLine := append([]string{}, SystemctlBaseCmd...)
	cmdLine = append(cmdLine, cmd)
	cmdLine = append(cmdLine, args...)

	exitCodes := []int{0}

	// execute systemctl command
	output, err := Execute(cmdLine, exitCodes)
	if err != nil {
		return nil, err
	}

	return output, nil
}
