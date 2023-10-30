package connect

import (
	"fmt"
	"os"
	"path/filepath"
)

// Register announces the system, activates the
// product on SCC and adds the service to the system
func Register() error {
	printInformation("register")
	err := announceOrUpdate()
	if err != nil {
		return err
	}

	installReleasePkg := true
	product := CFG.Product
	if product.isEmpty() {
		product, err = baseProduct()
		if err != nil {
			return err
		}
		installReleasePkg = false
	}

	if err = registerProduct(product, installReleasePkg); err != nil {
		return err
	}

	if product.IsBase {
		p, err := showProduct(product)
		if err != nil {
			return err
		}
		if err := registerProductTree(p); err != nil {
			return err
		}
	}
	Info.Print(bold(greenText("\nSuccessfully registered system")))
	return nil
}

// registerProduct activates the product, adds the service and installs the release package
func registerProduct(product Product, installReleasePkg bool) error {
	Info.Printf("\nActivating %s %s %s ...\n", product.Name, product.Version, product.Arch)
	service, err := activateProduct(product, CFG.Email)
	if err != nil {
		return err
	}
	Info.Print("-> Adding service to system ...")
	if err := addService(service.URL, service.Name, !CFG.NoZypperRefresh); err != nil {
		return err
	}
	if installReleasePkg {
		Info.Print("-> Installing release package ...")
		if err := InstallReleasePackage(product.Name); err != nil {
			return err
		}
	}
	return nil
}

// registerProductTree traverses (depth-first search) the product
// tree and registers the recommended and available products
func registerProductTree(product Product) error {
	for _, extension := range product.Extensions {
		if extension.Recommended && extension.Available {
			if err := registerProduct(extension, true); err != nil {
				return err
			}
			if err := registerProductTree(extension); err != nil {
				return err
			}
		}
	}
	return nil
}

// Deregister deregisters the system
func Deregister() error {
	if fileExists("/usr/sbin/registercloudguest") && CFG.Product.isEmpty() {
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
	installedIDs := NewStringSet()
	for _, prod := range installed {
		installedIDs.Add(prod.Name)
	}

	dependencies := make([]Product, 0)
	for _, e := range tree.toExtensionsList() {
		if installedIDs.Contains(e.Name) {
			dependencies = append(dependencies, e)
		}
	}

	// reverse loop over dependencies
	for i := len(dependencies) - 1; i >= 0; i-- {
		if err := deregisterProduct(dependencies[i]); err != nil {
			return err
		}
	}

	// remove potential docker and podman configurations for our registry
	creds, err := getCredentials()
	if err == nil {
		removeRegistryAuthentication(creds.Username, creds.Password)
	}

	if err := deregisterSystem(); err != nil {
		return err
	}

	if err := removeOrRefreshService(baseProductService); err != nil {
		return err
	}
	Info.Print("\nCleaning up ...")
	if err := Cleanup(); err != nil {
		return err
	}
	Info.Print(bold(greenText("Successfully deregistered system")))

	return nil
}

func deregisterProduct(product Product) error {
	base, err := baseProduct()
	if err != nil {
		return err
	}
	if product.ToTriplet() == base.ToTriplet() {
		return ErrBaseProductDeactivation
	}
	Info.Printf("\nDeactivating %s %s %s ...\n", product.Name, product.Version, product.Arch)
	service, err := deactivateProduct(product)
	if err != nil {
		return err
	}
	if err := removeOrRefreshService(service); err != nil {
		return err
	}
	Info.Print("-> Removing release package ...")
	return removeReleasePackage(product.Name)
}

// SMT provides one service for all products, removing it would remove all repositories.
// Refreshing the service instead to remove the repos of deregistered product.
func removeOrRefreshService(service Service) error {
	if service.Name == "SMT_DUMMY_NOREMOVE_SERVICE" {
		Info.Print("-> Refreshing service ...")
		refreshAllServices()
		return nil
	}
	Info.Print("-> Removing service from system ...")
	return removeService(service.Name)
}

// AnnounceSystem announce system via SCC/Registration Proxy
func AnnounceSystem(distroTgt string, instanceDataFile string) (string, string, error) {
	Info.Printf(bold("\nAnnouncing system to %s ..."), CFG.BaseURL)
	instanceData, err := readInstanceData(instanceDataFile)
	if err != nil {
		return "", "", err
	}
	sysInfoBody, err := makeSysInfoBody(distroTgt, CFG.Namespace, instanceData)
	if err != nil {
		return "", "", err
	}
	return announceSystem(sysInfoBody)
}

// UpdateSystem resend the system's hardware details on SCC
func UpdateSystem(distroTarget, instanceDataFile string) error {
	Info.Printf(bold("\nUpdating system details on %s ..."), CFG.BaseURL)
	instanceData, err := readInstanceData(instanceDataFile)
	if err != nil {
		return err
	}
	sysInfoBody, err := makeSysInfoBody(distroTarget, CFG.Namespace, instanceData)
	if err != nil {
		return err
	}
	return updateSystem(sysInfoBody)
}

// SendKeepAlivePing updates the system information on the server
func SendKeepAlivePing() error {
	if !IsRegistered() {
		return ErrPingFromUnregistered
	}
	err := UpdateSystem("", "")
	if err == nil {
		Info.Print(bold(greenText("\nSuccessfully updated system")))
	}
	return err
}

// announceOrUpdate Announces the system to the server, receiving and storing its
// credentials. When already announced, sends the current hardware details to the server
func announceOrUpdate() error {
	if IsRegistered() {
		return UpdateSystem("", "")
	}

	distroTgt := ""
	if !CFG.Product.isEmpty() {
		distroTgt = CFG.Product.distroTarget()
	}
	login, password, err := AnnounceSystem(distroTgt, CFG.InstanceDataFile)
	if err != nil {
		return err
	}

	if err = writeSystemCredentials(login, password, ""); err == nil {
		setupRegistryAuthentication(login, password)
	}
	return err
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
		Info.Printf(bold("Registering system to %s"), server)
	} else {
		Info.Printf(bold("Deregistering system from %s"), server)
	}
	if CFG.FsRoot != "" {
		Info.Print("Rooted at:", CFG.FsRoot)
	}
	if CFG.Email != "" {
		Info.Print("Using E-Mail:", CFG.Email)
	}
}

func readInstanceData(instanceDataFile string) ([]byte, error) {
	if instanceDataFile == "" {
		return nil, nil
	}
	path := filepath.Join(CFG.FsRoot, instanceDataFile)
	Debug.Print("Reading file from: ", path)
	instanceData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return instanceData, nil
}

// ProductMigrations returns the online migration paths for the installed products
func ProductMigrations(installed []Product) ([]MigrationPath, error) {
	return productMigrations(installed)
}

// OfflineProductMigrations returns the offline migration paths for the installed products and target
func OfflineProductMigrations(installed []Product, targetBaseProduct Product) ([]MigrationPath, error) {
	return offlineProductMigrations(installed, targetBaseProduct)
}

// UpgradeProduct upgades the records for given product in SCC/SMT
// The service record for new product is returned
func UpgradeProduct(product Product) (Service, error) {
	return upgradeProduct(product)
}

// SearchPackage returns packages which are available in the extensions tree for given base product
func SearchPackage(query string, baseProd Product) ([]SearchPackageResult, error) {
	// default to system base product if empty product passed
	if baseProd.isEmpty() {
		var err error
		baseProd, err = baseProduct()
		if err != nil {
			return []SearchPackageResult{}, err
		}
	}
	return searchPackage(query, baseProd)
}

// ShowProduct fetches product details from SCC/SMT
func ShowProduct(productQuery Product) (Product, error) {
	return showProduct(productQuery)
}

// ActivatedProducts returns list of products activated in SCC/SMT
func ActivatedProducts() ([]Product, error) {
	var products []Product
	activations, err := systemActivations()
	if err != nil {
		return products, err
	}
	for _, a := range activations {
		products = append(products, a.Service.Product)
	}
	return products, nil
}

// ActivateProduct activates given product in SMT/SCC
// returns Service to be added to zypper
func ActivateProduct(product Product, email string) (Service, error) {
	return activateProduct(product, email)
}

// SystemActivations returns a map keyed by "Identifier/Version/Arch"
func SystemActivations() (map[string]Activation, error) {
	return systemActivations()
}

// DeactivateProduct deactivates given product in SMT/SCC
// returns Service to be removed from zypper
func DeactivateProduct(product Product) (Service, error) {
	return deactivateProduct(product)
}

// DeregisterSystem deletes current system in SMT/SCC
func DeregisterSystem() error {
	return deregisterSystem()
}

// InstallerUpdates returns an array of Installer-Updates repositories for the given product
func InstallerUpdates(product Product) ([]Repo, error) {
	return installerUpdates(product)
}

// SyncProducts synchronizes activated system products to the registration server
func SyncProducts(products []Product) ([]Product, error) {
	return syncProducts(products)
}
