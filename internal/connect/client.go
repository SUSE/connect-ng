package connect

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	cred "github.com/SUSE/connect-ng/internal/credentials"
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/internal/zypper"
)

type RegisterOut struct {
	Success  bool             `json:"success"`
	Products []ProductService `json:"products"`
	Message  string           `json:"message"`
}

type ProductService struct {
	Product ProductOut `json:"product"`
	Service ServiceOut `json:"service"`
}

type ProductOut struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
}

type ServiceOut struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

var (
	localAddService             = zypper.AddService
	localInstallReleasePackage  = zypper.InstallReleasePackage
	localRemoveOrRefreshService = removeOrRefreshService
	localMakeSysInfoBody        = makeSysInfoBody
	localUpdateSystem           = updateSystem
)

// Register announces the system, activates the
// product on SCC and adds the service to the system
func Register(jsonOutput bool) error {
	out := &RegisterOut{}

	printInformation("register", jsonOutput)
	err := announceOrUpdate(jsonOutput)
	if err != nil {
		return err
	}

	installReleasePkg := true
	product := CFG.Product
	if product.isEmpty() {
		base, err := zypper.BaseProduct()
		if err != nil {
			return err
		}
		product = zypperProductToProduct(base)
		installReleasePkg = false
	}

	if service, err := registerProduct(product, installReleasePkg, jsonOutput); err == nil {
		out.Products = append(out.Products, ProductService{
			Product: ProductOut{
				Name:       product.LongName,
				Identifier: product.Name,
				Version:    product.Version,
				Arch:       product.Arch,
			},
			Service: ServiceOut{
				Id:   service.ID,
				Name: service.Name,
				Url:  service.URL,
			},
		})
	} else {
		return err
	}

	if product.IsBase {
		p, err := showProduct(product)
		if err != nil {
			return err
		}
		if err := registerProductTree(p, jsonOutput, out); err != nil {
			return err
		}
	}
	if jsonOutput {
		out.Success = true
		out.Message = "Successfully registered system"
		out, err := json.Marshal(out)
		if err != nil {
			return err
		}
		util.Info.Println(string(out))
	} else {
		util.Info.Print(util.Bold(util.GreenText("\nSuccessfully registered system")))
	}
	return nil
}

// registerProduct activates the product, adds the service and installs the release package
func registerProduct(product Product, installReleasePkg bool, jsonOutput bool) (Service, error) {
	if jsonOutput {
		util.Debug.Printf("\nActivating %s %s %s ...\n", product.Name, product.Version, product.Arch)
	} else {
		util.Info.Printf("\nActivating %s %s %s ...\n", product.Name, product.Version, product.Arch)
	}

	service, err := activateProduct(product, CFG.Email)
	if err != nil {
		return Service{}, err
	}

	if !CFG.SkipServiceInstall {
		if jsonOutput {
			util.Debug.Print("-> Adding service to system ...")
		} else {
			util.Info.Print("-> Adding service to system ...")
		}

		if err := localAddService(service.URL, service.Name, !CFG.NoZypperRefresh, CFG.Insecure); err != nil {
			return Service{}, err
		}
	}

	if installReleasePkg && !CFG.SkipServiceInstall {
		if jsonOutput {
			util.Debug.Print("-> Installing release package ...")
		} else {
			util.Info.Print("-> Installing release package ...")
		}

		if err := localInstallReleasePackage(product.Name, CFG.AutoImportRepoKeys); err != nil {
			return Service{}, err
		}
	}
	return service, nil
}

// registerProductTree traverses (depth-first search) the product
// tree and registers the recommended and available products
func registerProductTree(product Product, jsonOutput bool, out *RegisterOut) error {
	for _, extension := range product.Extensions {
		if extension.Recommended && extension.Available {
			if service, err := registerProduct(extension, true, jsonOutput); err == nil {
				out.Products = append(out.Products, ProductService{
					Product: ProductOut{
						Name:       product.LongName,
						Identifier: product.Name,
						Version:    product.Version,
						Arch:       product.Arch,
					},
					Service: ServiceOut{
						Id:   service.ID,
						Name: service.Name,
						Url:  service.URL,
					},
				})
			} else {
				return err
			}
			if err := registerProductTree(extension, jsonOutput, out); err != nil {
				return err
			}
		}
	}
	return nil
}

// Deregister deregisters the system
func Deregister(jsonOutput bool) error {
	if util.FileExists("/usr/sbin/registercloudguest") && CFG.Product.isEmpty() {
		return fmt.Errorf("SUSE::Connect::UnsupportedOperation: " +
			"De-registration is disabled for on-demand instances. " +
			"Use `registercloudguest --clean` instead.")
	}

	if !IsRegistered() {
		return ErrSystemNotRegistered
	}

	out := &RegisterOut{}

	printInformation("deregister", jsonOutput)
	if !CFG.Product.isEmpty() {
		return deregisterProduct(CFG.Product, jsonOutput, out)
	}
	base, err := zypper.BaseProduct()
	if err != nil {
		return err
	}
	baseProd := zypperProductToProduct(base)
	baseProductService, err := upgradeProduct(baseProd)
	if err != nil {
		return err
	}

	tree, err := showProduct(baseProd)
	if err != nil {
		return err
	}
	installed, _ := zypper.InstalledProducts()
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
		if err := deregisterProduct(dependencies[i], jsonOutput, out); err != nil {
			return err
		}
	}

	// remove potential docker and podman configurations for our registry
	creds, err := cred.ReadCredentials(cred.SystemCredentialsPath(CFG.FsRoot))
	if err == nil {
		util.Debug.Print("\nRemoving SUSE registry system authentication configuration ...")
		removeRegistryAuthentication(creds.Username, creds.Password)
	}

	if err := deregisterSystem(); err != nil {
		return err
	}

	if !CFG.SkipServiceInstall {
		if err := localRemoveOrRefreshService(baseProductService, jsonOutput); err != nil {
			return err
		}
	}
	if !jsonOutput {
		util.Info.Print("\nCleaning up ...")
	}
	if err := Cleanup(); err != nil {
		return err
	}
	if jsonOutput {
		out.Success = true
		out.Message = "Successfully deregistered system"
		out, err := json.Marshal(out)
		if err != nil {
			return err
		}
		util.Info.Println(string(out))
	} else {
		util.Info.Print(util.Bold(util.GreenText("Successfully deregistered system")))
	}

	return nil
}

func deregisterProduct(product Product, jsonOutput bool, out *RegisterOut) error {
	base, err := zypper.BaseProduct()
	if err != nil {
		return err
	}
	if product.ToTriplet() == zypperProductToProduct(base).ToTriplet() {
		return ErrBaseProductDeactivation
	}
	if !jsonOutput {
		util.Info.Printf("\nDeactivating %s %s %s ...\n", product.Name, product.Version, product.Arch)
	}
	service, err := deactivateProduct(product)
	if err != nil {
		return err
	}

	if CFG.SkipServiceInstall {
		return nil
	}

	if err := localRemoveOrRefreshService(service, jsonOutput); err != nil {
		return err
	}
	if jsonOutput {
		out.Products = append(out.Products, ProductService{
			Product: ProductOut{
				Name:       product.LongName,
				Identifier: product.Name,
				Version:    product.Version,
				Arch:       product.Arch,
			},
			Service: ServiceOut{
				Id:   service.ID,
				Name: service.Name,
				Url:  service.URL,
			},
		})
	} else {
		util.Info.Print("-> Removing release package ...")
	}
	return zypper.RemoveReleasePackage(product.Name)
}

// SMT provides one service for all products, removing it would remove all repositories.
// Refreshing the service instead to remove the repos of deregistered product.
func removeOrRefreshService(service Service, jsonOutput bool) error {
	if service.Name == "SMT_DUMMY_NOREMOVE_SERVICE" {
		if !jsonOutput {
			util.Info.Print("-> Refreshing service ...")
		}
		zypper.RefreshAllServices()
		return nil
	}
	if !jsonOutput {
		util.Info.Print("-> Removing service from system ...")
	}
	return zypper.RemoveService(service.Name)
}

// AnnounceSystem announce system via SCC/Registration Proxy
func AnnounceSystem(distroTgt string, instanceDataFile string, quiet bool) (string, string, error) {
	if !quiet {
		util.Info.Printf(util.Bold("\nAnnouncing system to %s ..."), CFG.BaseURL)
	}

	instanceData, err := readInstanceData(instanceDataFile)
	if err != nil {
		return "", "", err
	}
	sysInfoBody, err := localMakeSysInfoBody(distroTgt, CFG.Namespace, instanceData, false)
	if err != nil {
		return "", "", err
	}
	return announceSystem(sysInfoBody)
}

// UpdateSystem resend the system's hardware details on SCC
func UpdateSystem(distroTarget, instanceDataFile string, quiet bool, keepalive bool) error {
	if !quiet {
		util.Info.Printf(util.Bold("\nUpdating system details on %s ..."), CFG.BaseURL)
	}
	instanceData, err := readInstanceData(instanceDataFile)
	if err != nil {
		return err
	}
	includeUptimeLog := keepalive && CFG.EnableSystemUptimeTracking
	sysInfoBody, err := localMakeSysInfoBody(distroTarget, CFG.Namespace, instanceData, includeUptimeLog)
	if err != nil {
		return err
	}
	return localUpdateSystem(sysInfoBody)
}

// SendKeepAlivePing updates the system information on the server
func SendKeepAlivePing() error {
	if !IsRegistered() {
		return ErrPingFromUnregistered
	}
	err := UpdateSystem("", "", false, true)
	if err == nil {
		util.Info.Print(util.Bold(util.GreenText("\nSuccessfully updated system")))
	}
	return err
}

// announceOrUpdate Announces the system to the server, receiving and storing
// its credentials. When already announced, sends the current hardware details
// to the server. The output is not shown on stdout if `quiet` is set to true.
func announceOrUpdate(quiet bool) error {
	if IsRegistered() {
		return UpdateSystem("", "", quiet, false)
	}

	distroTgt := ""
	if !CFG.Product.isEmpty() {
		distroTgt = CFG.Product.distroTarget()
	}
	login, password, err := AnnounceSystem(distroTgt, CFG.InstanceDataFile, quiet)
	if err != nil {
		return err
	}

	if err = cred.CreateCredentials(login, password, "", cred.SystemCredentialsPath(CFG.FsRoot)); err == nil {
		// If the user is authenticated against the SCC, then setup the Docker
		// Registry configuration for the system. Otherwise, if the system is
		// behind a proxy (e.g. RMT), this step might fail and it's best to
		// avoid it (see bsc#1231185).
		if CFG.IsScc() {
			util.Debug.Print("\nAdding SUSE registry system authentication configuration ...")
			setupRegistryAuthentication(login, password)
		}
	}
	return err
}

// IsRegistered returns true if there is a valid credentials file
func IsRegistered() bool {
	_, err := cred.ReadCredentials(cred.SystemCredentialsPath(CFG.FsRoot))
	return err == nil
}

// UpToDate Checks if API endpoint is up-to-date,
// useful when dealing with RegistrationProxy errors
func UpToDate() bool {
	return upToDate()
}

func printInformation(action string, jsonOutput bool) {
	var server string
	if CFG.IsScc() {
		server = "SUSE Customer Center"
	} else {
		server = "registration proxy " + CFG.BaseURL
	}
	if action == "register" {
		if jsonOutput {
			util.Debug.Printf(util.Bold("Registering system to %s"), server)
		} else {
			util.Info.Printf(util.Bold("Registering system to %s"), server)
		}
	} else {
		if jsonOutput {
			util.Debug.Printf(util.Bold("Deregistering system from %s"), server)
		} else {
			util.Info.Printf(util.Bold("Deregistering system from %s"), server)
		}
	}
	if CFG.FsRoot != "" {
		if jsonOutput {
			util.Debug.Print("Rooted at:", CFG.FsRoot)
		} else {
			util.Info.Print("Rooted at:", CFG.FsRoot)
		}
	}
	if CFG.Email != "" {
		if jsonOutput {
			util.Debug.Print("Using E-Mail:", CFG.Email)
		} else {
			util.Info.Print("Using E-Mail:", CFG.Email)
		}
	}
}

func readInstanceData(instanceDataFile string) ([]byte, error) {
	if instanceDataFile == "" {
		return nil, nil
	}
	path := filepath.Join(CFG.FsRoot, instanceDataFile)
	util.Debug.Print("Reading file from: ", path)
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
		base, err := zypper.BaseProduct()
		if err != nil {
			return []SearchPackageResult{}, err
		}
		baseProd = zypperProductToProduct(base)
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
func InstallerUpdates(product Product) ([]zypper.Repository, error) {
	return installerUpdates(product)
}

// SyncProducts synchronizes activated system products to the registration server
func SyncProducts(products []Product) ([]Product, error) {
	return syncProducts(products)
}
