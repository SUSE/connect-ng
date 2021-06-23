package connect

import (
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
