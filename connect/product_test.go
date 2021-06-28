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
