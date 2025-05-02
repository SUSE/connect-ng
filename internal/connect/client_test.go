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

func mockMakeSysInfoBody(t *testing.T, expectedIncludeUptimeLog bool) {
	localMakeSysInfoBody = func(distroTarget, namespace string, instanceData []byte, includeUptimeLog bool) ([]byte, error) {
		if includeUptimeLog != expectedIncludeUptimeLog {
			t.Errorf("Expected includeUptimeLog to be %v\n", expectedIncludeUptimeLog)
		}
		return nil, nil
	}
}

// NOTE: This needs to be reworked.
// The current implementation of logging does not really allow any useful
// testing mechanics and is overly complicated. Refactor this!
func disableOutput() {
	util.Info = log.New(io.Discard, "", 0)
	util.Debug = log.New(io.Discard, "", 0)
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
