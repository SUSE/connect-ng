package connect

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchPackage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(readTestFile("package_search.json", t))
	}))
	defer ts.Close()

	CFG.BaseURL = ts.URL
	baseProduct := Product{Name: "SLES", Version: "15.2", Arch: "x86_64"}
	packages, err := searchPackage("gcc", baseProduct)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if l := len(packages); l != 2 {
		t.Fatalf("len(packages) == %d, expected 2", l)
	}
	pack0 := packages[0]
	if pack0.ID != 21156745 {
		t.Errorf("pack0.ID == %v, expected 21156745", pack0.ID)
	}
	if pack0.Name != "gcc10" {
		t.Errorf("pack0.Name == '%v', expected 'gcc10'", pack0.Name)
	}
	if pack0.Arch != "x86_64" {
		t.Errorf("pack0.Arch == '%v', expected 'x86_64'", pack0.Arch)
	}
	if pack0.Version != "10.3.0+git1587" {
		t.Errorf("pack0.Version == '%v', expected '10.3.0+git1587'", pack0.Version)
	}
	if pack0.Release != "1.6.4" {
		t.Errorf("pack0.Release == '%v', expected '1.6.4'", pack0.Release)
	}

	if l := len(packages[1].Products); l != 1 {
		t.Fatalf("len(packages[1].Products) == %v, expected 1", l)
	}
	prod10 := packages[1].Products[0]
	if prod10.Name != "Basesystem Module" {
		t.Errorf("prod1.Name == '%v', expected 'Basesystem Module'", prod10.Name)
	}
	if prod10.Ident != "sle-module-basesystem/15.2/x86_64" {
		t.Errorf("prod1.Ident == '%v', expected 'sle-module-basesystem/15.2/x86_64'", prod10.Ident)
	}
	if prod10.Type != "module" {
		t.Errorf("prod1.Type == '%v', expected 'module'", prod10.Type)
	}
	if !prod10.Free {
		t.Error("prod1 not free, expected free")
	}
	if prod10.Edition != "15 SP2" {
		t.Errorf("prod1.Edition == '%v', expected '15 SP2'", prod10.Edition)
	}
	if prod10.Arch != "x86_64" {
		t.Errorf("prod1.Arch == '%v', expected 'x86_64'", prod10.Arch)
	}
}
