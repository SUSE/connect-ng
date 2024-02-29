package connect

import (
	"bufio"
	"fmt"
	"encoding/json"
	"bytes"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)




var (
	cloudEx = `Version: .*(amazon)|Manufacturer: (Amazon)|Manufacturer: (Google)|Manufacturer: (Microsoft) Corporation`
	cloudRe = regexp.MustCompile(cloudEx)

	containerEx = `.*(docker|runc|buildah|buildkit|nerdctl|libpod).*`
	containerRe = regexp.MustCompile(containerEx)
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
	hw := hwinfo{}
	var err error
	if hw.Arch, err = arch(); err != nil {
		return hwinfo{}, err
	}
	hw.Hostname = hostname()
	hw.CloudProvider = cloudProvider()

	// Include memory information if possible.
	hw.MemTotal = systemMemory()

	var lscpuM map[string]string
	if lscpuM, err = lscpu(); err != nil {
		return hw, err
	}

	hw.Cpus, _ = strconv.Atoi(lscpuM["CPU(s)"])
	hw.Sockets, _ = strconv.Atoi(lscpuM["Socket(s)"])
	hw.UUID, _ = uuid()
	
	// Try to use systemd to detect the hypervisor.
	// see `systemd-detect-virt --list` for the full list of possible
	// outputs. Do not bail yet here.

	var hypervisorIdent []string

	if hypervisor, err := hypervisor(); err == nil && hypervisor != "" {
		hypervisorIdent = append(hypervisorIdent, hypervisor)
	}

	// attempt to determine if we're running on a container without systemd
	if runtime, err := containerRuntime(); err == nil && runtime != "" {
		hypervisorIdent = append(hypervisorIdent, runtime)
	}

	// additionally append the hypervisor vendor reported by lscpu.
	// e.g. KVM, XEN, IBM, pHyp
	if hypervisorVendor, ok := lscpuM["Hypervisor vendor"]; ok {
		hypervisorIdent = append(hypervisorIdent, hypervisorVendor)
	}

	// additionally append the hypervisor vendor reported by lscpu.
	// e.g. KVM, XEN, IBM, pHyp
	if hypvervisor, ok := lscpuM["Hypervisor"]; ok {
		// remove z/VM from the beginning of the string
		if hw.Arch == archS390 {
			hypvervisor = strings.SplitN(hypvervisor, " ", 2)[1]
		}

		hypervisorIdent = append(hypervisorIdent, hypvervisor)
	}
	
	hw.Hypervisor = strings.Join(hypervisorIdent, "/")

	hw.Clusters, _ = strconv.Atoi(lscpuM["Cluster(s)"])

	if hw.Arch == archS390 {
		// enrich data for s390x, uuid in s390 is returned
		// by read_values -u, but it's not yet released.
		//
		// as a fallback, it'll be constructed by the same components
		// that are returned by read_values -s
		cpuinfoS390(&hw)
	}

	return hw, nil
}

func cpuinfoS390(hw *hwinfo) error {
	rvsOut, err := readValues("-s")
	if err != nil {
		return err
	}
	rvs := readValues2map(rvsOut)

	if cpus, ok := rvs["VM00 CPUs Total"]; ok {
		hw.Cpus, _ = strconv.Atoi(cpus)
	} else if cpus, ok := rvs["LPAR CPUs Total"]; ok {
		hw.Cpus, _ = strconv.Atoi(cpus)
	}

	if sockets, ok := rvs["VM00 IFLs"]; ok {
		hw.Sockets, _ = strconv.Atoi(sockets)
	} else if sockets, ok := rvs["LPAR CPUs IFL"]; ok {
		hw.Sockets, _ = strconv.Atoi(sockets)
	}

	// when read_values finally ships -u, we can remove these blocks
	hw.UUID = fmt.Sprintf("%s-%s", rvs["Sequence Code"], rvs["LPAR Name"])
	if vm_name, ok := rvs["VM00 Name"]; ok {
		hw.UUID = fmt.Sprintf("%s-%s", hw.UUID, vm_name)
	}

	return nil
}

func arch() (string, error) {
	output, err := execute([]string{"uname", "-m"}, nil)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func lscpu() (map[string]string, error) {
	output, err := execute([]string{"lscpu", "-J"}, nil)
	if err != nil {
		return nil, err
	}
	return lscpu2map(output), nil
}

func lscpu2map(b []byte) map[string]string {

	type lscpuItem struct {
		Key string `json:"field"`
		Value string `json:"data"`
	}

	type lscpuOutput struct {
		LSCPU []lscpuItem `json:"lscpu"`
	}

	var lscpu lscpuOutput;

	m := make(map[string]string)

	if err := json.Unmarshal(b, &lscpu); err != nil {
		return m
	}
	
	for _, item := range lscpu.LSCPU {
		fieldname := strings.Split(item.Key, ":")
		m[fieldname[0]] = item.Value
	}

	fmt.Printf("%#v", m)
	return m
}

func cloudProvider() string {
	output, err := execute([]string{"dmidecode", "-t", "system"}, nil)
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
	output, err := execute([]string{"systemd-detect-virt", "-v"}, []int{0, 1})
	if err != nil {
		return "", err
	}
	if bytes.Equal(output, []byte("none")) {
		return "", nil
	}
	return string(output), nil
}

func containerRuntime() (string, error) {
	content, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return "", err
	}
	matches := containerRe.FindSubmatch(content)

	if len(matches) == 0 {
		return "", nil
	}

	return string(matches[1]), nil
}

// uuid returns the system uuid on x86 and arm
func uuid() (string, error) {
	if fileExists("/sys/hypervisor/uuid") {
		content, err := os.ReadFile("/sys/hypervisor/uuid")
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	output, err := execute([]string{"dmidecode", "-s", "system-uuid"}, nil)
	if err != nil {
		return "", err
	}
	out := string(output)
	if strings.Contains(out, "Not Settable") || strings.Contains(out, "Not Present") {
		return "", nil
	}
	return out, nil
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
	Debug.Print(err)
	ip, err := getPrivateIPAddr()
	if err != nil {
		Debug.Print(err)
		return ""
	}
	return ip
}

// readValues calls read_values from SUSE/s390-tools
func readValues(arg string) ([]byte, error) {
	output, err := execute([]string{"read_values", arg}, nil)
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

// Returns an integer with the amount of megabytes of total memory (i.e.
// `MemTotal` in /proc/meminfo). It will return 0 if this information could not
// be extracted for whatever reason.

func systemMemory() int {
	output, err := execute([]string{"lsmem", "-J", "-b"}, nil)

	if err != nil {
		return 0
	}

	return parseMeminfo(output)
}

func parseMeminfo(b []byte) int {
	type lsmemItem struct {
		Size int `json:"size"`
	}

	type lsmemOutput struct {
		Blocks []lsmemItem `json:"memory"`
	}

	var lsmem lsmemOutput;

	ram := 0

	if err := json.Unmarshal(b, &lsmem); err != nil {
		return ram
	}
	
	for _, block := range lsmem.Blocks {
		ram += block.Size
	}

	return ram / (1024*1024) // memory is reported in bytes
}
