package connect

import (
	"path/filepath"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/internal/zypper"
	"github.com/SUSE/connect-ng/pkg/registration"
)

// MigrationPath holds a list of products
type MigrationPath []registration.Product

// Rollback restores system state to before failed migration
func Rollback(opts *Options) error {
	util.Info.Print("Starting to sync system product activations to the server. This can take some time...")

	base, err := zypper.BaseProduct()
	if err != nil {
		return err
	}

	wrapper := NewWrappedAPI(opts)

	// First rollback the base_product
	service, err := registration.UpdateProduct(wrapper.GetConnection(), base)
	if err != nil {
		return err
	}
	if err = migrationRefreshService(service, opts.Insecure); err != nil {
		return err
	}

	// Fetch the product tree
	installed, err := zypper.InstalledProducts()
	if err != nil {
		return err
	}
	installedIDs := NewStringSet()
	for _, prod := range installed {
		installedIDs.Add(prod.Name)
	}

	tree, err := registration.FetchProductInfo(wrapper.GetConnection(), base.Identifier, base.Version, base.Arch)
	if err != nil {
		return err
	}

	// Get all installed products in right order
	extensions := make([]registration.Product, 0)
	for _, e := range tree.ToExtensionsList() {
		if installedIDs.Contains(e.Name) {
			extensions = append(extensions, e)
		}
	}

	// Rollback all extensions
	for _, e := range extensions {
		service, err := registration.UpdateProduct(wrapper.GetConnection(), e)
		if err != nil {
			return err
		}
		if err := migrationRefreshService(service, opts.Insecure); err != nil {
			return err
		}
	}

	// Synchronize installed products with SCC activations (removes obsolete
	// activations)
	if _, err := syncProducts(installed); err != nil {
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

func migrationRefreshService(service registration.Service, insecure bool) error {
	// INFO: Remove old and new service because this could be called after filesystem rollback or
	// from inside a failed migration
	if err := MigrationRemoveService(service.Name); err != nil {
		return err
	}
	if err := MigrationRemoveService(service.ObsoletedName); err != nil {
		return err
	}

	// INFO: Add new service for the same product but with new/valid service url
	MigrationAddService(service.URL, service.Name, insecure)

	return nil
}
