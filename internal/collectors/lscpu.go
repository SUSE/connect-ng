package collectors

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

func lscpu() (map[string]string, error) {
	output, err := util.Execute([]string{"lscpu"}, nil)
	if err != nil {
		return nil, err
	}

	return lscpu2map(output), nil
}

func lscpu2map(b []byte) map[string]string {
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

	return m
}
