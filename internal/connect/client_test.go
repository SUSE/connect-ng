package connect

import "testing"

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

func TestClientRegisterWithServiceInstallSkipSuccessful(t *testing.T) {
	CFG.Product = Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
	CFG.SkipServiceInstall = true
	mockAddServiceCalled(t, false)
	mockInstallReleasePackage(t, false)
	Register(false)
}

func TestClientRegisterWithoutServiceInstallSkipSuccessful(t *testing.T) {
	CFG.Product = Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
	CFG.SkipServiceInstall = false
	mockAddServiceCalled(t, true)
	mockInstallReleasePackage(t, true)
	Register(false)
}

func TestClientDeregistrationWithServiceInstallSkipSuccessful(t *testing.T) {
	CFG.Product = Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
	CFG.SkipServiceInstall = true
	mockAddServiceCalled(t, false)
	mockInstallReleasePackage(t, false)
	Deregister(false)
}

func TestClientDeregistrationWithoutServiceInstallSkipSuccessful(t *testing.T) {
	CFG.Product = Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
	CFG.SkipServiceInstall = false
	mockAddServiceCalled(t, true)
	mockInstallReleasePackage(t, true)
	Deregister(false)
}
