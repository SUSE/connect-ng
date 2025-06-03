package connect

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	cred "github.com/SUSE/connect-ng/internal/credentials"
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/internal/zypper"
	"github.com/SUSE/connect-ng/pkg/connection"
	"github.com/SUSE/connect-ng/pkg/registration"
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
func Register(opts *Options) error {
	out := &RegisterOut{}
	api := NewWrappedAPI(opts)

	printInformation(fmt.Sprintf("Registering system to %s", opts.ServerName()), opts)
	if err := api.RegisterOrKeepAlive(opts.Token); err != nil {
		return err
	}

	installReleasePkg := true
	product := opts.Product
	if product.isEmpty() {
		base, err := zypper.BaseProduct()
		if err != nil {
			return err
		}
		product = zypperProductToProduct(base)
		installReleasePkg = false
	}

	if service, err := registerProduct(opts, product, installReleasePkg); err == nil {
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
		// BUG: `out` is then re-written afterwards.
		if err := registerProductTree(opts, p, out); err != nil {
			return err
		}
	}

	switch opts.OutputKind {
	case Text:
		util.Info.Print(util.Bold(util.GreenText("\nSuccessfully registered system")))
	case JSON:
		out.Success = true
		out.Message = "Successfully registered system"
		out, err := json.Marshal(out)
		if err != nil {
			return err
		}
		util.Info.Println(string(out))
	}
	return nil
}

// registerProduct activates the product, adds the service and installs the release package
func registerProduct(opts *Options, product Product, installReleasePkg bool) (Service, error) {
	opts.Print(fmt.Sprintf("\nActivating %s %s %s ...\n", product.Name, product.Version, product.Arch))

	service, err := activateProduct(product, opts.Email)
	if err != nil {
		return Service{}, err
	}

	if !opts.SkipServiceInstall {
		opts.Print("-> Adding service to system ...")

		if err := localAddService(service.URL, service.Name, !opts.NoZypperRefresh, opts.Insecure); err != nil {
			return Service{}, err
		}
	}

	if installReleasePkg && !opts.SkipServiceInstall {
		opts.Print("-> Installing release package ...")

		if err := localInstallReleasePackage(product.Name, opts.AutoImportRepoKeys); err != nil {
			return Service{}, err
		}
	}
	return service, nil
}

// registerProductTree traverses (depth-first search) the product
// tree and registers the recommended and available products
func registerProductTree(opts *Options, product Product, out *RegisterOut) error {
	for _, extension := range product.Extensions {
		if extension.Recommended && extension.Available {
			if service, err := registerProduct(opts, extension, true); err == nil {
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
			if err := registerProductTree(opts, extension, out); err != nil {
				return err
			}
		}
	}
	return nil
}

// Deregister the current system.
func Deregister(opts *Options) error {
	if util.FileExists("/usr/sbin/registercloudguest") && opts.Product.isEmpty() {
		return fmt.Errorf("SUSE::Connect::UnsupportedOperation: " +
			"De-registration via SUSEConnect is disabled by registercloudguest." +
			"Use `registercloudguest --clean` instead.")
	}

	if !IsRegistered() {
		return ErrSystemNotRegistered
	}

	// BUG: this is largely ignored for trees.
	out := &RegisterOut{}

	printInformation(fmt.Sprintf("Deregistering system to %s", opts.ServerName()), opts)
	if !opts.Product.isEmpty() {
		return deregisterProduct(opts.Product, opts, out)
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
		if err := deregisterProduct(dependencies[i], opts, out); err != nil {
			return err
		}
	}

	// remove potential docker and podman configurations for our registry
	creds, err := cred.ReadCredentials(cred.SystemCredentialsPath(opts.FsRoot))
	if err == nil {
		util.Debug.Print("\nRemoving SUSE registry system authentication configuration ...")
		removeRegistryAuthentication(creds.Username, creds.Password)
	}

	api := NewWrappedAPI(opts)
	if err := registration.Deregister(api.GetConnection()); err != nil {
		return err
	}

	if !opts.SkipServiceInstall {
		if err := localRemoveOrRefreshService(baseProductService, opts); err != nil {
			return err
		}
	}

	opts.Print("\nCleaning up ...")
	if err := Cleanup(opts.BaseURL, opts.FsRoot); err != nil {
		return err
	}

	switch opts.OutputKind {
	case Text:
		util.Info.Print(util.Bold(util.GreenText("Successfully deregistered system")))
	case JSON:
		out.Success = true
		out.Message = "Successfully deregistered system"
		out, err := json.Marshal(out)
		if err != nil {
			return err
		}
		util.Info.Println(string(out))
	}

	return nil
}

func deregisterProduct(product Product, opts *Options, out *RegisterOut) error {
	base, err := zypper.BaseProduct()
	if err != nil {
		return err
	}
	if product.ToTriplet() == zypperProductToProduct(base).ToTriplet() {
		return ErrBaseProductDeactivation
	}
	opts.Print(fmt.Sprintf("\nDeactivating %s %s %s ...\n", product.Name, product.Version, product.Arch))
	service, err := deactivateProduct(product)
	if err != nil {
		return err
	}

	if opts.SkipServiceInstall {
		return nil
	}

	if err := localRemoveOrRefreshService(service, opts); err != nil {
		return err
	}

	switch opts.OutputKind {
	case Text:
		util.Info.Print("-> Removing release package ...")
	case JSON:
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
	}
	return zypper.RemoveReleasePackage(product.Name)
}

// SMT provides one service for all products, removing it would remove all repositories.
// Refreshing the service instead to remove the repos of deregistered product.
func removeOrRefreshService(service Service, opts *Options) error {
	if service.Name == "SMT_DUMMY_NOREMOVE_SERVICE" {
		opts.Print("-> Refreshing service ...")
		zypper.RefreshAllServices()
		return nil
	}
	opts.Print("-> Removing service from system ...")
	return zypper.RemoveService(service.Name)
}

// IsRegistered returns true if there is a valid credentials file
func IsRegistered() bool {
	_, err := cred.ReadCredentials(cred.SystemCredentialsPath(CFG.FsRoot))
	return err == nil
}

// Returns true if the current system is targeting an old registration proxy.
func IsOutdatedRegProxy(opts *Options) bool {
	// This is not a registration proxy, bail out.
	if opts.IsScc() {
		return false
	}

	// The trick is to check on an API endpoint which is not supported by SMT.
	// If the endpoint exists it will return 422 since we are omitting required
	// parameters. Then we know we are not dealing with an outdated registration
	// proxy.
	api := NewWrappedAPI(opts)
	conn := api.GetConnection()

	req, err := conn.BuildRequest("GET", "/connect/repositories/installer", nil)
	if err != nil {
		return true
	}

	_, err = conn.Do(req)
	if err == nil {
		return true
	}

	if ae, ok := err.(*connection.ApiError); ok {
		if ae.Code == http.StatusUnprocessableEntity {
			return false
		}
	}
	return true
}

// Print the given message plus some extra registration information that might
// be relevant (i.e. things that have changed from the default behaviour).
func printInformation(msg string, opts *Options) {
	opts.Print(msg)

	if opts.FsRoot != "" {
		opts.Print("Rooted at: " + opts.FsRoot)
	}
	if opts.Email != "" {
		opts.Print("Using E-Mail: " + opts.Email)
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

// Returns the zypper repositories for the installer updates endpoint.
func InstallerUpdates(opts *Options, product Product) ([]zypper.Repository, error) {
	repos := make([]zypper.Repository, 0)

	api := NewWrappedAPI(opts)
	conn := api.GetConnection()

	req, err := conn.BuildRequest("GET", "/connect/repositories/installer", nil)
	if err != nil {
		return repos, err
	}
	req = connection.AddQuery(req, product.toQuery())

	resp, err := conn.Do(req)
	if err != nil {
		return repos, err
	}
	if err = json.Unmarshal(resp, &repos); err != nil {
		return repos, JSONError{err}
	}
	return repos, nil
}

// SyncProducts synchronizes activated system products to the registration server
func SyncProducts(products []Product) ([]Product, error) {
	return syncProducts(products)
}
