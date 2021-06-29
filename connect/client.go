package connect

import (
	"fmt"
	"os"
)

// Register announces the system, activates the
// product on SCC and adds the service to the system
func Register() error {
	printInformation("register")
	err := announceOrUpdate()
	if err != nil {
		return err
	}

	// TODO remainder of Register()
	return nil
}

// Deregister deregisters the system
func Deregister() error {
	if fileExists("/usr/sbin/registercloudguest") {
		return fmt.Errorf("SUSE::Connect::UnsupportedOperation: " +
			"De-registration is disabled for on-demand instances. " +
			"Use `registercloudguest --clean` instead.")
	}

	if !IsRegistered() {
		return ErrSystemNotRegistered
	}

	printInformation("deregister")
	if !CFG.Product.isEmpty() {
		return deregisterProduct(CFG.Product)
	}
	baseProd, _ := baseProduct()
	baseProductService, err := upgradeProduct(baseProd)
	if err != nil {
		return err
	}

	tree, err := showProduct(baseProd)
	if err != nil {
		return err
	}
	installed, _ := installedProducts()
	installedIDs := make(map[string]struct{}, 0)
	for _, prod := range installed {
		installedIDs[prod.Name] = struct{}{}
	}

	dependencies := make([]Product, 0)
	for _, e := range tree.toExtensionsList() {
		if _, found := installedIDs[e.Name]; found {
			dependencies = append(dependencies, e)
		}
	}

	// reverse loop over dependencies
	for i := len(dependencies) - 1; i >= 0; i-- {
		if err := deregisterProduct(dependencies[i]); err != nil {
			return err
		}
	}

	if err := deregisterSystem(); err != nil {
		return err
	}

	if err := removeOrRefreshService(baseProductService); err != nil {
		return err
	}
	fmt.Println("\nCleaning up ...")
	if err := Cleanup(); err != nil {
		return err
	}
	fmt.Println(bold(greenText("Successfully deregistered system")))

	return nil
}

func deregisterProduct(product Product) error {
	base, err := baseProduct()
	if err != nil {
		return err
	}
	if product.toTriplet() == base.toTriplet() {
		return ErrBaseProductDeactivation
	}
	fmt.Printf("\nDeactivating %s %s %s ...\n", product.Name, product.Version, product.Arch)
	service, err := deactivateProduct(product)
	if err != nil {
		return err
	}
	if err := removeOrRefreshService(service); err != nil {
		return err
	}
	fmt.Println("-> Removing release package ...")
	return removeReleasePackage(product.Name)
}

// SMT provides one service for all products, removing it would remove all repositories.
// Refreshing the service instead to remove the repos of deregistered product.
func removeOrRefreshService(service Service) error {
	if service.Name == "SMT_DUMMY_NOREMOVE_SERVICE" {
		fmt.Println("-> Refreshing service ...")
		refreshAllServices()
		return nil
	}
	fmt.Println("-> Removing service from system ...")
	return removeService(service.Name)
}

// UpdateSystem resend the system's hardware details on SCC
func UpdateSystem(distroTarget, instanceDataFile string) error {
	fmt.Printf("\nUpdating system details on %s ...\n", CFG.BaseURL)
	var instanceData []byte
	if instanceDataFile != "" {
		var err error
		instanceData, err = os.ReadFile(instanceDataFile)
		if err != nil {
			return err
		}
	}
	sysInfoBody, err := makeSysInfoBody(distroTarget, CFG.Namespace, instanceData)
	if err != nil {
		return err
	}
	return updateSystem(sysInfoBody)
}

// announceOrUpdate Announces the system to the server, receiving and storing its
// credentials. When already announced, sends the current hardware details to the server
func announceOrUpdate() error {
	if IsRegistered() {
		return UpdateSystem("", "")
	}

	// TODO remainder of announceOrUpdate()
	return nil
}

// IsRegistered returns true if there is a valid credentials file
func IsRegistered() bool {
	_, err := getCredentials()
	return err == nil
}

// UpToDate Checks if API endpoint is up-to-date,
// useful when dealing with RegistrationProxy errors
func UpToDate() bool {
	return upToDate()
}

// URLDefault returns true if using https://scc.suse.com
func URLDefault() bool {
	return CFG.BaseURL == defaultBaseURL
}

func printInformation(action string) {
	var server string
	if URLDefault() {
		server = "SUSE Customer Center"
	} else {
		server = "registration proxy " + CFG.BaseURL
	}
	if action == "register" {
		fmt.Printf(bold("Registering system to %s\n"), server)
	} else {
		fmt.Printf(bold("Deregistering system from %s\n"), server)
	}
	if CFG.FsRoot != "" {
		fmt.Println("Rooted at:", CFG.FsRoot)
	}
	if CFG.Email != "" {
		fmt.Println("Using E-Mail:", CFG.Email)
	}
}
