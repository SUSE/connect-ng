package connect

import (
	"testing"
)

func TestDistroTarget(t *testing.T) {
	p := Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
	got := p.distroTarget()
	expect := "sle-15-x86_64"
	if got != expect {
		t.Errorf("got: %s, expected: %s", got, expect)
	}
}

func TestEmptyProduct(t *testing.T) {
	p1 := Product{}
	if !p1.isEmpty() {
		t.Errorf("expected %v to be empty", p1)
	}

	p2 := Product{Name: "Dummy"}
	if !p2.isEmpty() {
		t.Errorf("expected %v to be empty", p2)
	}

	p3 := Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
	if p3.isEmpty() {
		t.Errorf("expected %v not to be empty", p3)
	}
}
