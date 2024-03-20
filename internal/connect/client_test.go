package connect

import "testing"

func mockAddServiceCalled(t *testing.T, expected bool) {
	counter := 0
	localAddService = func(string, string, bool) error {
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
	localInstallReleasePackage = func(string) error {
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

func mockMakeSysInfoBody(t *testing.T, expectedIncludeUptimeLog bool) {
	localMakeSysInfoBody = func(distroTarget, namespace string, instanceData []byte, includeUptimeLog bool) ([]byte, error) {
		if includeUptimeLog != expectedIncludeUptimeLog {
			t.Errorf("Expected includeUptimeLog to be %v\n", expectedIncludeUptimeLog)
		}
		return nil, nil
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

func TestDefaultSystemUptimeTrackingDisable(t *testing.T) {
	localUpdateSystem = func(body []byte) error {
		// we don't care about updating the system for unit test
		return nil
	}
	mockMakeSysInfoBody(t, false)
	UpdateSystem("", "", false, true)
}

func TestSystemUptimeTrackingEnable(t *testing.T) {
	CFG.EnableSystemUptimeTracking = true
	localUpdateSystem = func(body []byte) error {
		// we don't care about updating the system for unit test
		return nil
	}
	mockMakeSysInfoBody(t, true)
	UpdateSystem("", "", false, true)
}

func TestSystemUptimeTrackingEnableNotKeepalive(t *testing.T) {
	CFG.EnableSystemUptimeTracking = true
	localUpdateSystem = func(body []byte) error {
		// we don't care about updating the system for unit test
		return nil
	}
	mockMakeSysInfoBody(t, false)
	UpdateSystem("", "", false, false)
}
