package connect

import (
	"fmt"
	"path/filepath"
)

// MigrationPath holds a list of products
type MigrationPath []Product

// Rollback restores system state to before failed migration
func Rollback() error {
	fmt.Println("Starting to sync system product activations to the server. This can take some time...")

	base, err := baseProduct()
	if err != nil {
		return err
	}

	// First rollback the base_product
	service, err := downgradeProduct(base)
	if err != nil {
		return err
	}
	if err = migrationRefreshService(service); err != nil {
		return err
	}

	// Fetch the product tree
	installed, err := installedProducts()
	if err != nil {
		return err
	}
	installedIDs := NewStringSet()
	for _, prod := range installed {
		installedIDs.Add(prod.Name)
	}

	tree, err := showProduct(base)
	if err != nil {
		return err
	}

	// Get all installed products in right order
	extensions := make([]Product, 0)
	for _, e := range tree.toExtensionsList() {
		if installedIDs.Contains(e.Name) {
			extensions = append(extensions, e)
		}
	}

	// Rollback all extensions
	for _, e := range extensions {
		service, err := downgradeProduct(e)
		if err != nil {
			return err
		}
		if err := migrationRefreshService(service); err != nil {
			return err
		}
	}

	// Synchronize installed products with SCC activations (removes obsolete activations)
	if _, err := syncProducts(installed); err != nil {
		return err
	}

	// Set releasever to the new baseproduct version
	return setReleaseVersion(base.Version)
}

// MigrationAddService adds zypper service in migration context
func MigrationAddService(URL string, serviceName string) error {
	// don't try to add the service if the plugin with the same name exists (bsc#1128969)
	if fileExists(filepath.Join("/usr/lib/zypp/plugins/services", serviceName)) {
		return nil
	}
	return addService(URL, serviceName, true)
}

// MigrationRemoveService removes zypper service in migration context
func MigrationRemoveService(serviceName string) error {
	// don't try to remove the service if the plugin with the same name exists (bsc#1128969)
	if fileExists(filepath.Join("/usr/lib/zypp/plugins/services", serviceName)) {
		return nil
	}
	return removeService(serviceName)
}

func migrationRefreshService(service Service) error {
	// INFO: Remove old and new service because this could be called after filesystem rollback or
	// from inside a failed migration
	if err := MigrationRemoveService(service.Name); err != nil {
		return err
	}
	if err := MigrationRemoveService(service.ObsoletedName); err != nil {
		return err
	}

	// INFO: Add new service for the same product but with new/valid service url
	MigrationAddService(service.URL, service.Name)

	return nil
}
