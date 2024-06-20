package collectors

import "github.com/SUSE/connect-ng/internal/util"

type Uname struct{}

func (Uname) run(arch string) (Result, error) {
	output, err := uname("-r -v")

	if err != nil {
		return NoResult, err
	}
	return Result{"uname": output}, nil
}

var uname = func(flag string) (string, error) {
	output, err := util.Execute([]string{"uname", flag}, nil)
	if err != nil {
		return "", err
	}
	return string(output), err
}
