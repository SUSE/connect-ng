package connect

import (
	"testing"
)

func TestParseProductsXML(t *testing.T) {
	products, err := parseProductsXML(readTestFile("products.xml", t))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(products) != 2 {
		t.Errorf("Expected len()==2. Got %d", len(products))
	}
	if products[0].ToTriplet() != "SUSE-MicroOS/5.0/x86_64" {
		t.Errorf("Expected SUSE-MicroOS/5.0/x86_64 Got %s", products[0].ToTriplet())
	}
}

func TestParseServicesXML(t *testing.T) {
	services, err := parseServicesXML(readTestFile("services.xml", t))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(services) != 1 {
		t.Errorf("Expected len()==1. Got %d", len(services))
	}
	if services[0].Name != "SUSE_Linux_Enterprise_Micro_5.0_x86_64" {
		t.Errorf("Expected SUSE_Linux_Enterprise_Micro_5.0_x86_64 Got %s", services[0].Name)
	}
}

func TestParseReposXML(t *testing.T) {
	repos, err := parseReposXML(readTestFile("repos.xml", t))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(repos) != 3 {
		t.Errorf("Expected len()==3. Got %v", len(repos))
	}
	if repos[0].Name != "SLE-Module-Basesystem15-SP2-Pool" {
		t.Errorf("Expected SLE-Module-Basesystem15-SP2-Pool. Got %v", repos[0].Name)
	}
	if repos[0].Priority != 99 {
		t.Errorf("Expected priority 99. Got %v", repos[0].Priority)
	}
	if !repos[0].Enabled {
		t.Errorf("Expected Enabled. Got %v", repos[0].Enabled)
	}
	if repos[1].Priority != 50 {
		t.Errorf("Expected priority 99. Got %v", repos[1].Priority)
	}
	if repos[1].Enabled {
		t.Errorf("Expected not Enabled Got %v", repos[1].Enabled)
	}
}

func TestInstalledProducts(t *testing.T) {
	execute = func(_ []string, _ []int) ([]byte, error) {
		return readTestFile("products.xml", t), nil
	}

	products, err := installedProducts()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(products) != 2 {
		t.Errorf("Expected len()==2. Got %d", len(products))
	}
	if products[0].ToTriplet() != "SUSE-MicroOS/5.0/x86_64" {
		t.Errorf("Expected SUSE-MicroOS/5.0/x86_64 Got %s", products[0].ToTriplet())
	}
}

func TestBaseProduct(t *testing.T) {
	execute = func(_ []string, _ []int) ([]byte, error) {
		return readTestFile("products.xml", t), nil
	}

	base, err := baseProduct()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if base.ToTriplet() != "SUSE-MicroOS/5.0/x86_64" {
		t.Errorf("Expected SUSE-MicroOS/5.0/x86_64 Got %s", base.ToTriplet())
	}
}

func TestBaseProductError(t *testing.T) {
	execute = func(_ []string, _ []int) ([]byte, error) {
		return readTestFile("products-no-base.xml", t), nil
	}
	_, err := baseProduct()
	if err != ErrCannotDetectBaseProduct {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestParseSearchResultXML(t *testing.T) {
	packages, err := parseSearchResultXML(readTestFile("product-search.xml", t))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(packages) != 2 {
		t.Errorf("Expected len()==2. Got %v", len(packages))
	}
	if packages[0].Name != "SLES" {
		t.Errorf("Expected SLES. Got %v", packages[0].Name)
	}
	if packages[0].Edition != "15.2-0" {
		t.Errorf("Expected edition 15.2-0. Got %v", packages[0].Edition)
	}
	if packages[0].Repo != "SLE-Product-SLES15-SP2-Updates" {
		t.Errorf("Expected SLE-Product-SLES15-SP2-Updates. Got %v", packages[0].Repo)
	}
	if packages[1].Edition != "15.2-0" {
		t.Errorf("Expected edition 15.2-0. Got %v", packages[1].Edition)
	}
	if packages[1].Repo != "SLE-Product-SLES15-SP2-Pool" {
		t.Errorf("Expected SLE-Product-SLES15-SP2-Pool. Got %v", packages[1].Repo)
	}
}
