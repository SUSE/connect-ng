package connect

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
	"strings"
)

var (
	cloudEx = `Version: .*(amazon)|Manufacturer: (Amazon)|Manufacturer: (Google)|Manufacturer: (Microsoft) Corporation`
	cloudRe = regexp.MustCompile(cloudEx)
)

func arch() (string, error) {
	output, err := execute([]string{"uname", "-i"}, false, nil)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

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

func hypervisor() (string, error) {
	output, err := execute([]string{"systemd-detect-virt", "-v"}, false, []int{0, 1})
	if err != nil {
		return "", err
	}
	if bytes.Equal(output, []byte("none")) {
		return "", nil
	}
	return string(output), nil
}

func uuid() (string, error) {
	if fileExists("/sys/hypervisor/uuid") {
		content, err := os.ReadFile("/sys/hypervisor/uuid")
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	output, err := execute([]string{"dmidecode", "-s", "system-uuid"}, false, nil)
	if err != nil {
		return "", err
	}
	out := string(output)
	if strings.Contains(out, "Not Settable") || strings.Contains(out, "Not Present") {
		return "", nil
	}
	return out, nil
}
