package connect

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func mockIsRegistered(isRegistered bool) {
	localIsRegistered = func() bool {
		return isRegistered
	}
}

func mockProductSetup(t *testing.T) {
	extensions := []Product{}

	data := readTestFile("extensions.json", t)
	if err := json.Unmarshal(data, &extensions); err != nil {
		t.Fatalf("Could not read extensions.json: '%s'", err)
	}

	localBaseProduct = func() (Product, error) {
		return Product{
			Name:    "SLES",
			Version: "15.4",
			Release: "0",
			Arch:    "x86_64",
			IsBase:  true,
		}, nil
	}

	localShowProduct = func(base Product) (Product, error) {
		base.Extensions = extensions
		return base, nil
	}
}

func mockSystemActivations() {
	activations := map[string]Activation{}
	activations["SLES/15.4/x86_64"] = Activation{}
	activations["sle-module-basesystem/15.4/x86_64"] = Activation{}
	activations["sle-module-server-applications/15.4/x86_64"] = Activation{}

	localSystemActivations = func() (map[string]Activation, error) {
		return activations, nil
	}
}

func mockRootWritable(isWritable bool) {
	localRootWritable = func() bool {
		return isWritable
	}
}

func expectStringMatches(t *testing.T, input string, match string) {
	if !strings.Contains(input, match) {
		message := "Expect input to match '%s' but did not!\n The input was: %s"
		t.Errorf(fmt.Sprintf(message, match, input))
	}
}

func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Expected no error but got: '%'", err)
	}
}

func TestPrintExtensionsNotRegistered(t *testing.T) {
	mockIsRegistered(false)

	_, err := RenderExtensionTree(false)

	if err != ErrListExtensionsUnregistered {
		t.Errorf("Expected not registered list-extensions to show specific error but did not")
	}
}

func TestPrintExtensionsAsText(t *testing.T) {
	mockIsRegistered(true)
	mockProductSetup(t)
	mockSystemActivations()
	mockRootWritable(true)

	result, err := RenderExtensionTree(false)

	expectNoError(t, err)
	expectStringMatches(t, result, "AVAILABLE EXTENSIONS AND MODULES")
	expectStringMatches(t, result, "Python 3 Module 15 SP4 x86_64")
	expectStringMatches(t, result, "You can find more information about available modules here:")
}

func TestPrintExtensionsAsJSON(t *testing.T) {
	mockIsRegistered(true)
	mockProductSetup(t)
	mockSystemActivations()
	mockRootWritable(true)

	result, err := RenderExtensionTree(true)
	json := "{\"identifier\":\"PackageHub\",\"version\":\"15.4\",\"arch\":\"x86_64\",\"name\":\"SUSE Package Hub 15 SP4 x86_64\",\"activated\":false,\"available\":false,\"free\":true,\"extensions\":[]}"

	expectNoError(t, err)
	expectStringMatches(t, result, json)
}

func TestPrintExtensionsProductActivated(t *testing.T) {
	mockIsRegistered(true)
	mockProductSetup(t)
	mockSystemActivations()
	mockRootWritable(true)

	result, err := RenderExtensionTree(false)
	activated := "\x1b[1mBasesystem Module 15 SP4 x86_64\x1b[0m \x1b[33m(Activated)\x1b[0m"

	expectNoError(t, err)
	expectStringMatches(t, result, activated)
}

func TestPrintExtensionsNotAvailable(t *testing.T) {
	mockIsRegistered(true)
	mockProductSetup(t)
	mockSystemActivations()
	mockRootWritable(true)

	result, err := RenderExtensionTree(false)
	unavailable := "\x1b[1mLegacy Module 15 SP4 x86_64\x1b[0m \x1b[31m(Not available)\x1b[0m"

	expectNoError(t, err)
	expectStringMatches(t, result, unavailable)

}

func TestPrintExtensionsFilesystemNotWritable(t *testing.T) {
	mockIsRegistered(true)
	mockProductSetup(t)
	mockSystemActivations()
	mockRootWritable(false)

	result, err := RenderExtensionTree(false)

	expectNoError(t, err)
	expectStringMatches(t, result, "transactional-update register")
}
