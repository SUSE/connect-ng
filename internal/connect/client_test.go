package connect

import (
	"io"
	"log"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
)

func mockAddServiceCalled(t *testing.T, expected bool) {
	counter := 0
	localAddService = func(string, string, bool, bool) error {
		counter += 1
		return nil
	}

	if !expected {
		if counter > 1 {
			t.Errorf("Expected addService not to be called.")
		}
	}
}

func mockInstallReleasePackage(t *testing.T, expected bool) {
	counter := 0
	localInstallReleasePackage = func(string, bool) error {
		counter += 1
		return nil
	}

	if !expected {
		if counter > 1 {
			t.Errorf("Expected InstallReleasePackage not to be called.")
		}
	}
}

func mockRemoveOrRefreshService(t *testing.T, expected bool) {
	counter := 0
	localRemoveOrRefreshService = func(Service, bool) error {
		counter += 1
		return nil
	}

	if !expected {
		if counter > 1 {
			t.Errorf("Expected removeOrRefreshService not to be called.")
		}
	}

}

// NOTE: This needs to be reworked.
// The current implementation of logging does not really allow any useful
// testing mechanics and is overly complicated. Refactor this!
func disableOutput() {
	util.Info = log.New(io.Discard, "", 0)
	util.Debug = log.New(io.Discard, "", 0)
}

func TestClientRegisterWithServiceInstallSkipSuccessful(t *testing.T) {
	disableOutput()

	CFG.Product = Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
	CFG.SkipServiceInstall = true
	mockAddServiceCalled(t, false)
	mockInstallReleasePackage(t, false)
	Register(false)
}

func TestClientRegisterWithoutServiceInstallSkipSuccessful(t *testing.T) {
	disableOutput()

	CFG.Product = Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
	CFG.SkipServiceInstall = false
	mockAddServiceCalled(t, true)
	mockInstallReleasePackage(t, true)
	Register(false)
}

func TestClientDeregistrationWithServiceInstallSkipSuccessful(t *testing.T) {
	disableOutput()

	CFG.Product = Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
	CFG.SkipServiceInstall = true
	mockAddServiceCalled(t, false)
	mockInstallReleasePackage(t, false)
	Deregister(false)
}

func TestClientDeregistrationWithoutServiceInstallSkipSuccessful(t *testing.T) {
	disableOutput()

	CFG.Product = Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
	CFG.SkipServiceInstall = false
	mockAddServiceCalled(t, true)
	mockInstallReleasePackage(t, true)
	Deregister(false)
}
