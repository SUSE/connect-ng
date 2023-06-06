package connect

import (
	"bytes"
	"flag"
	"net"
	"testing"
)

var testHwinfo = flag.Bool("test-hwinfo", false, "")

func TestLscpu2mapPhysical(t *testing.T) {
	m := lscpu2map(readTestFile("lscpu_phys.txt", t))

	if m["CPU(s)"] != "8" {
		t.Errorf("Found %s CPU(s), expected 8", m["CPU(s)"])
	}
	if m["Socket(s)"] != "1" {
		t.Errorf("Found %s Sockets(s), expected 1", m["Socket(s)"])
	}
	if _, ok := m["Hypervisor vendor"]; ok {
		t.Errorf("Hypervisor vendor should not be set")
	}
}

func TestLscpu2mapVirtual(t *testing.T) {
	m := lscpu2map(readTestFile("lscpu_virt.txt", t))

	if m["CPU(s)"] != "1" {
		t.Errorf("Found %s CPU(s), expected 1", m["CPU(s)"])
	}
	if m["Socket(s)"] != "1" {
		t.Errorf("Found %s Sockets(s), expected 1", m["Socket(s)"])
	}
	if hv, ok := m["Hypervisor vendor"]; !ok || hv != "KVM" {
		t.Errorf("Hypervisor vendor should be KVM")
	}
}

func TestFindCloudProviderAWS(t *testing.T) {
	got := findCloudProvider(readTestFile("dmidecode_aws.txt", t))
	if got != "amazon" {
		t.Errorf("findCloudProvider()==%s, expected amazon", got)
	}
}

func TestFindCloudProviderAWSLarge(t *testing.T) {
	got := findCloudProvider(readTestFile("dmidecode_aws_large.txt", t))
	if got != "Amazon" {
		t.Errorf("findCloudProvider()==%s, expected Amazon", got)
	}
}

func TestFindCloudProviderAzure(t *testing.T) {
	got := findCloudProvider(readTestFile("dmidecode_azure.txt", t))
	if got != "Microsoft" {
		t.Errorf("findCloudProvider()==%s, expected Microsoft", got)
	}
}

func TestFindCloudProviderGoogle(t *testing.T) {
	got := findCloudProvider(readTestFile("dmidecode_google.txt", t))
	if got != "Google" {
		t.Errorf("findCloudProvider()==%s, expected Google", got)
	}
}

func TestFindCloudProviderNoCloud(t *testing.T) {
	got := findCloudProvider(readTestFile("dmidecode_qemu.txt", t))
	if got != "" {
		t.Errorf("findCloudProvider()==%s, expected \"\"", got)
	}
}

func TestIsUUID(t *testing.T) {
	var tests = []struct {
		s      string
		expect bool
	}{
		{"4C4C4544-0059-4810-8034-C2C04F335931", true},
		{"4C4C4544-0059-7777-8034-C2C04F335931", true},
		{"ec293a33-b805-7eef-b7c8-d1238386386f", true},
		{"failed:\n", false},
	}
	for _, test := range tests {
		got := isUUID(test.s)
		if got != test.expect {
			t.Errorf("Got isUUID(%s)==%v, expected %v", test.s, got, test.expect)
		}
	}
}

func TestPrivateIP(t *testing.T) {
	var tests = []struct {
		ip      string
		private bool
	}{
		{"10.0.1.1", true},
		{"192.168.100.10", true},
		{"172.18.10.10", true},
		{"8.8.8.8", false},
		{"172.15.0.1", false},
	}
	for _, test := range tests {
		ip := net.ParseIP(test.ip)
		got := privateIP(ip)
		if got != test.private {
			t.Errorf("Got privateIP(%s)==%v, expected %v", test.ip, got, test.private)
		}
	}
}

func TestReadValues2map(t *testing.T) {
	m := readValues2map(readTestFile("read_values_s.txt", t))
	expect := map[string]string{
		"VM00 CPUs Total":      "1",
		"LPAR CPUs Total":      "6",
		"VM00 IFLs":            "1",
		"LPAR CPUs IFL":        "6",
		"VM00 Control Program": "z/VM    6.1.0",
	}
	for k, v := range expect {
		if m[k] != expect[k] {
			t.Errorf("m[%s]==%s, expected %s", k, m[k], v)
		}
	}
}

func TestGetHwinfo(t *testing.T) {
	if !*testHwinfo {
		t.SkipNow()
	}
	hw, err := getHwinfo()
	t.Logf("HW info: %+v", hw)
	if err != nil {
		t.Fatalf("getHwinfo() failed: %s", err)
	}
	if hw.Hostname == "" {
		t.Error(`Hostname=="", expected not empty`)
	}
	// reading UUID requires root access which is not available in build env
	// if hw.UUID == "" {
	// 	t.Errorf(`UUID=="", expected not empty`)
	// }
	if hw.Cpus == 0 {
		t.Error("Cpus==0, expected>0")
	}
	if hw.Sockets == 0 {
		// on ARM clusters, lscpu can return "Socket(s): -" if DMI is not accessible
		// this parses to hw.Sockets == 0 so we need to skip this test in those cases
		if hw.Arch == archARM && hw.Clusters > 0 {
			t.Log("Reading number of sockets failed on ARM cluster (DMI not accessible?). Check skipped.")
		} else {
			t.Error("Sockets==0, expected>0")
		}
	}
}

func TestParseMeminfo(t *testing.T) {
	var tests = []struct {
		file  string
		value int
	}{
		{"MemTotal:       16297236 kB", 15915},
		{"MemTotal:", 0},
		{"MemSomething:       16297236 kB", 0},
		{"Malformed  16297236 kB", 0},
		{"MemTotal:       notanumber kB", 0},
		{"wubalubadubdub", 0},
		{"", 0},
	}

	for _, v := range tests {
		buff := bytes.NewBufferString(v.file)
		val := parseMeminfo(buff)
		if val != v.value {
			t.Errorf("Expecting '%v', got '%v'", v.value, val)
		}
	}
}


func TestCheckIsContainerProcSelfAttr(t *testing.T) {
	var tests = []struct {
		content  string
		value bool
	}{
		{"docker-default (enforce)", true},
		{"containers-default-0.50.1 (enforce)", true},
		{"unconfined", false},
		{"wubalubadubdub", false},
	}

	for _, v := range tests {
		buff := bytes.NewBufferString(v.content)
		val := parseContainerProcFile(buff)
		if val != v.value {
			t.Errorf("Expecting '%v', got '%v' for '%v'", v.value, val, v.content)
		}
	}
}


func TestCheckIsContainerProcCgroups(t *testing.T) {
	var tests = []struct {
		content []byte
		value bool
	}{
		{readTestFile("container-check-fixtures/buildah-cgroups.txt", t), true},
		{readTestFile("container-check-fixtures/buildkit-cgroups.txt", t), true},
		{readTestFile("container-check-fixtures/docker-cgroups.txt", t), true},
		{readTestFile("container-check-fixtures/no-container-cgroups.txt", t), false},
		{readTestFile("container-check-fixtures/podman-cgroups.txt", t), true},
	}

	for _, v := range tests {
		buff := bytes.NewBuffer(v.content)
		val := parseContainerProcFile(buff)
		if val != v.value {
			t.Errorf("Expecting '%v', got '%v' for '%v'", v.value, val, v.content)
		}
	}
}

