package connect

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

const (
	testCaseInstalledProducts       = "testCaseInstalledProducts"
	testCaseInstalledProductsNoBase = "testCaseInstalledProductsNoBase"
)

var zypperExecTestCase = ""

func zypperExecMock(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestZypperHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GO_WANT_HELPER_PROCESS=1")
	cmd.Env = append(cmd.Env, "GO_TESTCASE="+zypperExecTestCase)
	return cmd
}

// This is not a real test. It's used for mocking output of commands.
func TestZypperHelper(t *testing.T) {
	// if ran directly, do nothing
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	testcase := os.Getenv("GO_TESTCASE")
	switch testcase {
	case testCaseInstalledProducts:
		os.Stdout.Write(readTestFile("products.xml", t))
		os.Exit(0)
	case testCaseInstalledProductsNoBase:
		os.Stdout.Write(readTestFile("products-no-base.xml", t))
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Error: unexpected testcase=%s\n", testcase)
		os.Exit(1)
	}
}

func TestParseProductsXML(t *testing.T) {
	products, err := parseProductsXML(readTestFile("products.xml", t))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(products) != 2 {
		t.Errorf("Expected len()==2. Got %d", len(products))
	}
	if products[0].toTriplet() != "SUSE-MicroOS/5.0/x86_64" {
		t.Errorf("Expected SUSE-MicroOS/5.0/x86_64 Got %s", products[0].toTriplet())
	}
}

func TestInstalledProducts(t *testing.T) {
	zypperExecTestCase = testCaseInstalledProducts
	execCommand = zypperExecMock
	defer func() { execCommand = exec.Command }()

	products, err := installedProducts()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(products) != 2 {
		t.Errorf("Expected len()==2. Got %d", len(products))
	}
	if products[0].toTriplet() != "SUSE-MicroOS/5.0/x86_64" {
		t.Errorf("Expected SUSE-MicroOS/5.0/x86_64 Got %s", products[0].toTriplet())
	}
}

func TestBaseProduct(t *testing.T) {
	zypperExecTestCase = testCaseInstalledProducts
	execCommand = zypperExecMock
	defer func() { execCommand = exec.Command }()

	base, err := baseProduct()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if base.toTriplet() != "SUSE-MicroOS/5.0/x86_64" {
		t.Errorf("Expected SUSE-MicroOS/5.0/x86_64 Got %s", base.toTriplet())
	}
}

func TestBaseProductError(t *testing.T) {
	zypperExecTestCase = testCaseInstalledProductsNoBase
	execCommand = zypperExecMock
	defer func() { execCommand = exec.Command }()

	_, err := baseProduct()
	if err != ErrCannotDetectBaseProduct {
		t.Errorf("Unexpected error: %s", err)
	}
}
