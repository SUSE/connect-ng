package helpers

import (
	"github.com/SUSE/connect-ng/internal/zypper"
)

func CleanupAll() {
	TryConfigCleanup()
	TrySUSEConnectDeregister()
	TrySUSEConnectCleanup()
	CleanupPolutedFilesystem()

	zypper.SetFilesystemRoot("/")
}
