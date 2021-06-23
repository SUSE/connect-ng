package connect

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
)

var (
	cloudEx = `Version: .*(amazon)|Manufacturer: (Amazon)|Manufacturer: (Google)|Manufacturer: (Microsoft) Corporation`
	cloudRe = regexp.MustCompile(cloudEx)
)

func lscpu() (map[string]string, error) {
	output, err := execute([]string{"lscpu"}, false, nil)
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

func cloudProvider() (string, error) {
	output, err := execute([]string{"dmidecode", "-t", "system"}, false, nil)
	if err != nil {
		return "", err
	}
	return findCloudProvider(output), nil
}

// findCloudProvider returns the cloud provider from "dmidecode -t system" output
func findCloudProvider(b []byte) string {
	match := cloudRe.FindSubmatch(b)
	for i, m := range match {
		if i != 0 && len(m) > 0 {
			return string(m)
		}
	}
	return ""
}
