package collectors

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

func lscpu() (int, error) {
	output, err := util.Execute([]string{"lscpu"}, nil)
	if err != nil {
		return 0, err
	}
	return getCPUCount(output), nil
}

func getCPUCount(b []byte) int {
	m := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		m[key] = val
	}
	return 0
}
