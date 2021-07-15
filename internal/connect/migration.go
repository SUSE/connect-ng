package connect

import (
	"fmt"
	"path/filepath"
)

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
	installedIDs := make(map[string]struct{}, 0)
	for _, prod := range installed {
		installedIDs[prod.Name] = struct{}{}
	}

	tree, err := showProduct(base)
	if err != nil {
		return err
	}

	// Get all installed products in right order
	extensions := make([]Product, 0)
	for _, e := range tree.toExtensionsList() {
		if _, found := installedIDs[e.Name]; found {
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

func migrationAddService(URL string, serviceName string) error {
	// don't try to add the service if the plugin with the same name exists (bsc#1128969)
	if fileExists(filepath.Join("/usr/lib/zypp/plugins/services", serviceName)) {
		return nil
	}
	return addService(URL, serviceName, true)
}

func migrationRemoveService(serviceName string) error {
	// don't try to remove the service if the plugin with the same name exists (bsc#1128969)
	if fileExists(filepath.Join("/usr/lib/zypp/plugins/services", serviceName)) {
		return nil
	}
	return removeService(serviceName)
}

func migrationRefreshService(service Service) error {
	// INFO: Remove old and new service because this could be called after filesystem rollback or
	// from inside a failed migration
	if err := migrationRemoveService(service.Name); err != nil {
		return err
	}
	if err := migrationRemoveService(service.ObsoletedName); err != nil {
		return err
	}

	// INFO: Add new service for the same product but with new/valid service url
	migrationAddService(service.URL, service.Name)

	return nil
}
