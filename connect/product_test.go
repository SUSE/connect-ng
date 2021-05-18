package connect

import (
	"testing"
)

var (
	p1 = Product{Name: "SLES", Version: "15.1", Arch: "x86_64", IsBase: true}
	p2 = Product{Name: "sle-module-basesystem", Version: "15.1", Arch: "x86_64", IsBase: false}
	p3 = Product{Name: "sle-module-containers", Version: "15.1", Arch: "x86_64", IsBase: false}
	p4 = Product{Name: "sle-module-desktop-applications", Version: "15.1", Arch: "x86_64", IsBase: false}
	p5 = Product{Name: "sle-module-development-tools", Version: "15.1", Arch: "x86_64", IsBase: false}
	p6 = Product{Name: "sle-module-server-applications", Version: "15.1", Arch: "x86_64", IsBase: false}
)

func TestGetProducts(t *testing.T) {
	want := []Product{p1, p2, p3, p4, p5, p6}
	got, _ := getProducts("../testdata/products.d/")
	if len(got) != len(want) {
		t.Errorf("getProducts() = %+v, want %v", got, want)
	}
	for i := range want {
		if !ProductsEqual(got[i], want[i]) {
			t.Errorf("getProducts() = \n%v, want= \n%v", got, want)
		}
	}
}
