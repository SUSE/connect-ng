package connect

import (
	"encoding/xml"
	"os"
	"path/filepath"
)

const (
	productPath = "/etc/products.d/"
)

// Product represents an installed product
type Product struct {
	Name    string `xml:"name"`
	Version string `xml:"version"`
	Arch    string `xml:"arch"`
	Summary string `xml:"summary"`
	IsBase  bool
}

func (p Product) ToTriplet() string {
	return p.Name + "/" + p.Version + "/" + p.Arch
}

func ProductsEqual(p1, p2 Product) bool {
	return p1.Name == p2.Name && p1.Version == p2.Version && p1.Arch == p2.Arch
}

// GetInstalledProducts returns products installed on the system
func GetInstalledProducts() []Product {
	return getProducts(productPath)
}

func getProducts(path string) []Product {
	baseProdSymLink := filepath.Join(path, "baseproduct")
	baseProd, err := filepath.EvalSymlinks(baseProdSymLink)
	if err != nil {
		Error.Fatal(err)
	}

	prodFiles, err := filepath.Glob(filepath.Join(path, "*.prod"))
	if err != nil {
		Error.Fatal(err)
	}
	var products []Product
	for _, prodFile := range prodFiles {
		p := productFromXMLFile(prodFile)
		if prodFile == baseProd {
			p.IsBase = true
		}
		products = append(products, p)
	}
	return products
}

func productFromXMLFile(path string) Product {
	xmlStr, err := os.ReadFile(path)
	if err != nil {
		Error.Fatal(err)
	}
	var p Product
	if err = xml.Unmarshal(xmlStr, &p); err != nil {
		Error.Fatal(err)
	}
	return p
}
