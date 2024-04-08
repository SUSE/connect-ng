package main

import (
	"strings"
	"testing"
)

var idx1 = "gcc\t7-3.6.1\tx86_64\n" +
	"gcc-c++\t7-3.6.1\tx86_64\n" +
	"gcc7\t7.5.0+r278197-lp153.188.1\tx86_64\n" +
	"gcc7-c++\t7.5.0+r278197-lp153.188.1\tx86_64\n" +
	"bad line\n" +
	"should\tbe\tskipped\ttoo\n" +
	"patch:openSUSE-2021-230\t1\tnoarch\n"

func TestPackageWanted(t *testing.T) {
	samples := []struct {
		Name  string
		Query string
		Exact bool
		CSens bool
		Out   bool
	}{
		{"abc", "abc", true, true, true},
		{"abc2", "abc", true, true, false},
		{"abc2", "abc", false, true, true},
		{"Abc", "Abc", true, true, true},
		{"Abc", "abc", true, true, false},
		{"Abc", "abc", true, false, true},
		{"Abc2", "Abc", true, true, false},
		{"Abc2", "Abc", false, true, true},
		{"Abc2", "abc", false, true, false},
		{"Abc2", "abc", false, false, true},
		{"abc", "abc2", false, false, false},
	}

	for _, s := range samples {
		if out := packageWanted(s.Name, s.Query, s.Exact, s.CSens); out != s.Out {
			t.Errorf("Checked name=%v with query=%v (Exact:%v, CSens:%v). Got: %v expected: %v", s.Name, s.Query, s.Exact, s.CSens, out, s.Out)
		}
	}
}

func TestParseRepoIndex(t *testing.T) {
	r := strings.NewReader(idx1)
	p := parseRepoIndex(r)
	if l := len(p); l != 5 {
		t.Errorf("got %v packages, expected 5", l)
	}
	// release extracted from version
	exp0 := searchResult{Name: "gcc", Version: "7", Release: "3.6.1", Arch: "x86_64"}
	if p[0] != exp0 {
		t.Errorf("Got %+v expected %+v", p[0], exp0)
	}
	// version without release
	exp4 := searchResult{Name: "patch:openSUSE-2021-230", Version: "1", Arch: "noarch"}
	if p[4] != exp4 {
		t.Errorf("Got %+v expected %+v", p[4], exp4)
	}
}
