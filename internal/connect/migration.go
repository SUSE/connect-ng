package connect

import (
	"path/filepath"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/internal/zypper"
	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/registration"
)

// MigrationPath holds a list of products
type MigrationPath []registration.Product

// Rollback restores system state to before failed migration
func Rollback(conn connection.Connection, opts *Options) error {
	util.Info.Print("Starting to sync system product activations to the server. This can take some time...")

	base, err := zypper.BaseProduct()
	if err != nil {
		return err
	}

	// First rollback the base_product
	meta, _, err := registration.Upgrade(conn, base.Identifier, base.Version, base.Arch)
	if err != nil {
		return err
	}
	if err = migrationRefreshService(meta, opts.Insecure); err != nil {
		return err
	}

	// Fetch the product tree
	installed, err := zypper.InstalledProducts()
	if err != nil {
		return err
	}
	installedIDs := NewStringSet()
	for _, prod := range installed {
		installedIDs.Add(prod.Identifier)
	}

	tree, err := registration.FetchProductInfo(conn, base.Identifier, base.Version, base.Arch)
	if err != nil {
		return err
	}

	// Get all installed products in right order
	extensions := make([]registration.Product, 0)
	for _, e := range tree.ToExtensionsList() {
		if installedIDs.Contains(e.Identifier) {
			extensions = append(extensions, e)
		}
	}

	// Rollback all extensions
	for _, e := range extensions {
		meta, _, err := registration.Upgrade(conn, e.Identifier, e.Version, e.Arch)
		if err != nil {
			return err
		}
		if err := migrationRefreshService(meta, opts.Insecure); err != nil {
			return err
		}
	}

	// Synchronize installed products with SCC activations (removes obsolete
	// activations)
	if _, err := SyncProducts(conn, installed); err != nil {
		return err
	}

	// Set releasever to the new baseproduct version
	return zypper.SetReleaseVersion(base.Version)
}

// MigrationAddService adds zypper service in migration context
func MigrationAddService(URL string, serviceName string, insecure bool) error {
	// don't try to add the service if the plugin with the same name exists (bsc#1128969)
	if util.FileExists(filepath.Join("/usr/lib/zypp/plugins/services", serviceName)) {
		return nil
	}
	return zypper.AddService(URL, serviceName, true, insecure)
}

// MigrationRemoveService removes zypper service in migration context
func MigrationRemoveService(serviceName string) error {
	// don't try to remove the service if the plugin with the same name exists (bsc#1128969)
	if util.FileExists(filepath.Join("/usr/lib/zypp/plugins/services", serviceName)) {
		return nil
	}
	return zypper.RemoveService(serviceName)
}

func migrationRefreshService(meta *registration.Metadata, insecure bool) error {
	// INFO: Remove old and new service because this could be called after filesystem rollback or
	// from inside a failed migration
	if err := MigrationRemoveService(meta.Name); err != nil {
		return err
	}
	if err := MigrationRemoveService(meta.ObsoletedName); err != nil {
		return err
	}

	// INFO: Add new service for the same product but with new/valid service url
	MigrationAddService(meta.URL, meta.Name, insecure)

	return nil
}
