package main

import (
	"testing"

	"github.com/SUSE/connect-ng/internal/connect"
)

func TestSortMigrationProducts(t *testing.T) {
	migration := connect.MigrationPath{
		{Name: "ruby", IsBase: false, Version: "2.5", Arch: "x86_64"},
		{Name: "gcc", IsBase: false, Version: "10.1", Arch: "x86_64"},
		{Name: "base-system", IsBase: true, Version: "15.2", Arch: "x86_64"},
		{Name: "awk", IsBase: false, Version: "4.2", Arch: "x86_64"},
		{Name: "python", IsBase: false, Version: "3.6", Arch: "x86_64"},
		{Name: "SLES", IsBase: true, Version: "15.2", Arch: "x86_64"},
	}
	installedIDs := connect.NewStringSet("ruby/2.5/x86_64", "python/3.6/x86_64")

	sortMigrationProducts(migration, installedIDs)

	expected := []string{
		// base products first (keep original order)
		"base-system/15.2/x86_64",
		"SLES/15.2/x86_64",
		// non-base, non-installed products (keep original order)
		"gcc/10.1/x86_64",
		"awk/4.2/x86_64",
		// installed products last (keep original order)
		"ruby/2.5/x86_64",
		"python/3.6/x86_64",
	}

	for i, p := range migration {
		if p.ToTriplet() != expected[i] {
			t.Fatalf("Got: %s expected: %s", p.ToTriplet(), expected[i])
		}
	}
}

func TestCompareEditions(t *testing.T) {
	samples := []struct {
		Left  string
		Right string
		Out   int
	}{
		{"1.2.3", "2.2.3", -1},
		{"12.3-4", "12.5", -1},
		{"1.0", "1.0", 0},
		{"9-1", "8-0", 1},
		{"0-3", "0-1", 0},
		{"1.0.1", "1.0", 1},
		{"2.0", "2", 0},
		{"10.9", "9.10.1", 1},
		{"1", "2", -1},
	}

	for _, s := range samples {
		if out := compareEditions(s.Left, s.Right); out != s.Out {
			t.Fatalf("Compared %v and %v. Got: %v expected: %v", s.Left, s.Right, out, s.Out)
		}
	}
}
