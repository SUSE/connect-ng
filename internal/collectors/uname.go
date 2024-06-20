package collectors

import "github.com/SUSE/connect-ng/internal/util"

type Uname struct{}

func (Uname) run(arch string) (Result, error) {
	output, err := uname([]string{"-r", "-v"})

	if err != nil {
		return NoResult, err
	}
	return Result{"uname": output}, nil
}

var uname = func(flags []string) (string, error) {
	call := append([]string{"uname"}, flags...)
	output, err := util.Execute(call, nil)
	if err != nil {
		return "", err
	}
	return string(output), err
}
