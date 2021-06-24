package connect

import (
	"net"
	"testing"
)

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
