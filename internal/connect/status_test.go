package connect

import (
	"testing"
)

func TestBuildStatuses(t *testing.T) {
	products := []Product{
		{Name: "p0", Version: "v0", Arch: "a0"},
		{Name: "p1", Version: "v1", Arch: "a1"},
		{Name: "p2", Version: "v2", Arch: "a2"},
	}

	a0 := Activation{RegCode: "regcode0"}
	a1 := Activation{RegCode: "regcode1"}
	a0.Service.Product.Free = false
	a1.Service.Product.Free = true
	activations := map[string]Activation{
		"p0/v0/a0": a0,
		"p1/v1/a1": a1,
	}

	statuses := buildStatuses(products, activations)

	if statuses[0].Status != registered {
		t.Errorf("Expected statuses[0].Status==%s, got \"%s\"", registered, notRegistered)
	}
	if statuses[1].Status != registered {
		t.Errorf("Expected statuses[1].Status==%s, got \"%s\"", registered, notRegistered)
	}
	if statuses[2].Status != notRegistered {
		t.Errorf("Expected statuses[2].Status==%s, got \"%s\"", notRegistered, registered)
	}
	if statuses[0].RegCode != "regcode0" {
		t.Errorf("Expected statuses[0].RegCode==regcode0, got \"%s\"", statuses[0].RegCode)
	}
	if statuses[1].RegCode != "regcode1" {
		t.Errorf("Expected statuses[1].RegCode==regcode1, got \"%s\"", statuses[1].RegCode)
	}
	if statuses[2].RegCode != "" {
		t.Errorf("Expected statuses[2].RegCode==\"\", got \"%s\"", statuses[2].RegCode)
	}
}
