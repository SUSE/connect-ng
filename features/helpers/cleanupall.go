package helpers

import (
	"github.com/SUSE/connect-ng/internal/zypper"
)

func CleanupAll() {
	TrySUSEConnectDeregister()
	TrySUSEConnectCleanup()
	CleanupPolutedFilesystem()

	zypper.SetFilesystemRoot("/")
}
