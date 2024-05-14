package connect

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/internal/util"
)

var (
	cloudEx = `Version: .*(amazon)|Manufacturer: (Amazon)|Manufacturer: (Google)|Manufacturer: (Microsoft) Corporation`
	cloudRe = regexp.MustCompile(cloudEx)
)

const (
	archX86  = "x86_64"
	archARM  = "aarch64"
	archS390 = "s390x"
	archPPC  = "ppc64le"
)

type hwinfo struct {
	Hostname      string `json:"hostname"`
	Cpus          int    `json:"cpus"`
	Sockets       int    `json:"sockets"`
	Clusters      int    `json:"-"`
	Hypervisor    string `json:"hypervisor"`
	Arch          string `json:"arch"`
	UUID          string `json:"uuid"`
	CloudProvider string `json:"cloud_provider"`
	MemTotal      int    `json:"mem_total,omitempty"`
}

func getHwinfo() (hwinfo, error) {
	var err error
	hw := hwinfo{}

	mandatory := []collectors.Collector{
		collectors.CPU{},
		collectors.Hostname{},
		collectors.Memory{},
		collectors.UUID{},
		collectors.Virtualization{},
	}

	if hw.Arch, err = arch(); err != nil {
		return hwinfo{}, err
	}

	result, err := collectors.CollectInformation(hw.Arch, mandatory)

	if err != nil {
		return hwinfo{}, err
	}

	// TODO: Handle errors when type casting from interface{}
	hw.Cpus = result["cpus"].(int)
	hw.Sockets = result["sockets"].(int)
	hw.Hostname = result["hostname"].(string)

	switch result["hypervisor"].(type) {
	case string:
		hw.Hypervisor = result["hypervisor"].(string)
	default:
		hw.Hypervisor = ""
	}
	hw.MemTotal = result["mem_total"].(int)
	hw.UUID = result["uuid"].(string) // ignore error to match original

	hw.CloudProvider = cloudProvider()

	return hw, nil
}

func cpuinfoS390(hw *hwinfo) error {
	rvsOut, err := readValues("-s")
	if err != nil {
		return err
	}
	rvs := readValues2map(rvsOut)

	if hypervisor, ok := rvs["VM00 Control Program"]; ok {
		// Strip and remove recurring whitespaces e.g. " z/VM    6.1.0" => "z/VM 6.1.0"
		subs := strings.Fields(hypervisor)
		hw.Hypervisor = strings.Join(subs, " ")
	} else {
		util.Debug.Print("Unable to find 'VM00 Control Program'. This system probably runs on an LPAR.")
	}

	return nil
}

func arch() (string, error) {
	output, err := util.Execute([]string{"uname", "-i"}, nil)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

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

func cloudProvider() string {
	output, err := util.Execute([]string{"dmidecode", "-t", "system"}, nil)
	if err != nil {
		return ""
	}
	return findCloudProvider(output)
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
	output, err := util.Execute([]string{"systemd-detect-virt", "-v"}, []int{0, 1})
	if err != nil {
		return "", err
	}
	if bytes.Equal(output, []byte("none")) {
		return "", nil
	}
	return string(output), nil
}

// uuid returns the system uuid on x86 and arm
func uuid() (string, error) {
	if util.FileExists("/sys/hypervisor/uuid") {
		content, err := os.ReadFile("/sys/hypervisor/uuid")
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	output, err := util.Execute([]string{"dmidecode", "-s", "system-uuid"}, nil)
	if err != nil {
		return "", err
	}
	out := string(output)
	if strings.Contains(out, "Not Settable") || strings.Contains(out, "Not Present") {
		return "", nil
	}
	return out, nil
}

// uuidS390 returns the system uuid on S390 or "" if it cannot be found
func uuidS390() string {
	out, err := readValues("-u")
	if err != nil {
		return ""
	}
	uuid := string(out)
	if isUUID(uuid) {
		return uuid
	}
	util.Debug.Print("Not implemented. Unable to determine UUID for s390x. Set to \"\"")
	return ""
}

// isUUID returns true if s is a valid uuid
func isUUID(s string) bool {
	exp := `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`
	uuidRe := regexp.MustCompile(exp)
	return uuidRe.MatchString(s)
}

// getPrivateIPAddr returns the first private IP address on the host
func getPrivateIPAddr() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		ip, _, _ := net.ParseCIDR(addr.String())
		if privateIP(ip) {
			return ip.String(), nil
		}
	}
	return "", nil
}

// privateIP returns true if ip is in a RFC1918 range
func privateIP(ip net.IP) bool {
	for _, block := range []string{"10.0.0.0/8", "192.168.0.0/16", "172.16.0.0/12"} {
		_, ipNet, _ := net.ParseCIDR(block)
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

func hostname() string {
	name, err := os.Hostname()
	// TODO the Ruby version has this "(none)" check - why?
	if err == nil && name != "" && name != "(none)" {
		return name
	}
	util.Debug.Print(err)
	ip, err := getPrivateIPAddr()
	if err != nil {
		util.Debug.Print(err)
		return ""
	}
	return ip
}

// readValues calls read_values from SUSE/s390-tools
func readValues(arg string) ([]byte, error) {
	output, err := util.Execute([]string{"read_values", arg}, nil)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// readValues2map parses the output of "read_values -s" on s390
func readValues2map(b []byte) map[string]string {
	br := bufio.NewScanner(bytes.NewReader(b))
	m := make(map[string]string)
	for br.Scan() {
		line := br.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		m[key] = val
	}
	return m
}

// Returns the parsed value for the given file. The implementation has been
// split from `systemMemory` so testing it is easier, but bear in mind that
// these two are coupled.
func parseMeminfo(file io.Reader) int {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 || fields[0] != "MemTotal:" {
			continue
		}

		val, err := strconv.Atoi(fields[1])
		if err != nil {
			util.Debug.Printf("could not obtain memory information for this system: %v", err)
			return 0
		}
		return val / 1024
	}

	util.Debug.Print("could not obtain memory information for this system")
	return 0
}

// Returns an integer with the amount of megabytes of total memory (i.e.
// `MemTotal` in /proc/meminfo). It will return 0 if this information could not
// be extracted for whatever reason.
func systemMemory() int {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		util.Debug.Print("'/proc/meminfo' could not be read!")
		return 0
	}
	defer file.Close()

	return parseMeminfo(file)
}
