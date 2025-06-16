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
	"github.com/SUSE/connect-ng/pkg/search"
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
)

// Register announces the system, activates the
// product on SCC and adds the service to the system
func Register(opts *Options) error {
	out := &RegisterOut{}
	api := NewWrappedAPI(opts)

	if opts.OutputKind != JSON {
		printInformation(fmt.Sprintf("Registering system to %s", opts.ServerName()), opts)
	}

	if err := api.RegisterOrKeepAlive(opts.Token); err != nil {
		return err
	}

	installReleasePkg := true
	product := opts.Product
	if product.IsEmpty() {
		base, err := zypper.BaseProduct()
		if err != nil {
			return err
		}
		product = base
		installReleasePkg = false
	}

	if service, err := registerProduct(opts, product, installReleasePkg); err == nil {
		out.Products = append(out.Products, ProductService{
			Product: ProductOut{
				Name:       product.Name,
				Identifier: product.Identifier,
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
		// TODO(mssola): credentials should be a pointer of sorts, so it can be
		// re-used instead of having to re-reload them every time we need it.
		api = NewWrappedAPI(opts)
		p, err := registration.FetchProductInfo(api.GetConnection(), product.Identifier, product.Version, product.Arch)
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

// registerProduct activates the product, adds the service and installs the
// release package
func registerProduct(opts *Options, product registration.Product, installReleasePkg bool) (registration.Service, error) {
	opts.Print(fmt.Sprintf("\nActivating %s %s %s ...\n", product.Identifier, product.Version, product.Arch))

	service, err := ActivateProduct(product, opts)
	if err != nil {
		return registration.Service{}, err
	}

	if !opts.SkipServiceInstall {
		opts.Print("-> Adding service to system ...")

		if err := localAddService(service.URL, service.Name, !opts.NoZypperRefresh, opts.Insecure); err != nil {
			return registration.Service{}, err
		}
	}

	if installReleasePkg && !opts.SkipServiceInstall {
		opts.Print("-> Installing release package ...")

		if err := localInstallReleasePackage(product.Identifier, opts.AutoImportRepoKeys); err != nil {
			return registration.Service{}, err
		}
	}
	return service, nil
}

// registerProductTree traverses (depth-first search) the product
// tree and registers the recommended and available products
func registerProductTree(opts *Options, product *registration.Product, out *RegisterOut) error {
	for _, extension := range product.Extensions {
		if extension.Recommended && extension.Available {
			if service, err := registerProduct(opts, extension, true); err == nil {
				out.Products = append(out.Products, ProductService{
					Product: ProductOut{
						Name:       product.Name,
						Identifier: product.Identifier,
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
			if err := registerProductTree(opts, &extension, out); err != nil {
				return err
			}
		}
	}
	return nil
}

// Deregister the current system.
func Deregister(opts *Options) error {
	if util.FileExists("/usr/sbin/registercloudguest") && opts.Product.IsEmpty() {
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
	if !opts.Product.IsEmpty() {
		return deregisterProduct(opts.Product, opts, out)
	}
	base, err := zypper.BaseProduct()
	if err != nil {
		return err
	}

	api := NewWrappedAPI(opts)
	baseMeta, _, err := registration.Upgrade(api.GetConnection(), base.Identifier, base.Version, base.Arch)
	if err != nil {
		return err
	}

	api = NewWrappedAPI(opts)
	tree, err := registration.FetchProductInfo(api.GetConnection(), base.Identifier, base.Version, base.Arch)
	if err != nil {
		return err
	}
	installed, _ := zypper.InstalledProducts()
	installedIDs := NewStringSet()
	for _, prod := range installed {
		installedIDs.Add(prod.Name)
	}

	dependencies := make([]registration.Product, 0)
	for _, e := range tree.ToExtensionsList() {
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

	if err := registration.Deregister(api.GetConnection()); err != nil {
		return err
	}

	if !opts.SkipServiceInstall {
		if err := localRemoveOrRefreshService(baseMeta.Name, opts); err != nil {
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

func deregisterProduct(product registration.Product, opts *Options, out *RegisterOut) error {
	base, err := zypper.BaseProduct()
	if err != nil {
		return err
	}
	if product.ToTriplet() == base.ToTriplet() {
		return ErrBaseProductDeactivation
	}

	opts.Print(fmt.Sprintf("\nDeactivating %s %s %s ...\n", product.Name, product.Version, product.Arch))
	wrapper := NewWrappedAPI(opts)
	metadata, _, err := registration.Deactivate(wrapper.GetConnection(), product.Identifier, product.Version, product.Arch)
	if err != nil {
		return err
	}

	if opts.SkipServiceInstall {
		return nil
	}

	if err := localRemoveOrRefreshService(metadata.Name, opts); err != nil {
		return err
	}

	switch opts.OutputKind {
	case Text:
		util.Info.Print("-> Removing release package ...")
	case JSON:
		out.Products = append(out.Products, ProductService{
			Product: ProductOut{
				Name:       product.Name,
				Identifier: product.Identifier,
				Version:    product.Version,
				Arch:       product.Arch,
			},
			Service: ServiceOut{
				Id:   metadata.ID,
				Name: metadata.Name,
				Url:  metadata.URL,
			},
		})
	}
	return zypper.RemoveReleasePackage(product.Name)
}

// SMT provides one service for all products, removing it would remove all
// repositories. Refreshing the service instead to remove the repos of
// deregistered product.
func removeOrRefreshService(serviceName string, opts *Options) error {
	if serviceName == "SMT_DUMMY_NOREMOVE_SERVICE" {
		opts.Print("-> Refreshing service ...")
		zypper.RefreshAllServices()
		return nil
	}
	opts.Print("-> Removing service from system ...")
	return zypper.RemoveService(serviceName)
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

// SearchPackage returns all the packages which are available in the extensions
// tree for the given base product.
func SearchPackage(opts *Options, query string) ([]search.SearchPackageResult, error) {
	// The base product from which the search will occur is the system's base
	// product.
	var err error
	base, err := zypper.BaseProduct()
	if err != nil {
		return []search.SearchPackageResult{}, err
	}

	api := NewWrappedAPI(opts)
	return search.Package(api.GetConnection(), query, base.ToTriplet())
}

// ActivatedProducts returns list of products activated in SCC/SMT
func ActivatedProducts(opts *Options) ([]*registration.Product, error) {
	var products []*registration.Product

	wrapper := NewWrappedAPI(opts)
	activations, err := registration.FetchActivations(wrapper.GetConnection())
	if err != nil {
		return products, err
	}
	for _, a := range activations {
		products = append(products, a.Product)
	}
	return products, nil
}

// ActivateProduct activates given product in SMT/SCC
// returns Service to be added to zypper
func ActivateProduct(product registration.Product, opts *Options) (registration.Service, error) {
	wrapper := NewWrappedAPI(opts)
	meta, pr, err := registration.Activate(wrapper.GetConnection(), product.Identifier, product.Version, product.Arch, opts.Token)
	if err != nil {
		return registration.Service{}, err
	}

	return registration.Service{
		ID:            meta.ID,
		URL:           meta.URL,
		Name:          meta.Name,
		ObsoletedName: meta.ObsoletedName,
		Product:       *pr,
	}, nil
}

// Returns the zypper repositories for the installer updates endpoint.
func InstallerUpdates(opts *Options, product registration.Product) ([]zypper.Repository, error) {
	repos := make([]zypper.Repository, 0)

	api := NewWrappedAPI(opts)
	conn := api.GetConnection()

	req, err := conn.BuildRequest("GET", "/connect/repositories/installer", nil)
	if err != nil {
		return repos, err
	}
	req = connection.AddQuery(req, product.ToQuery())

	resp, err := conn.Do(req)
	if err != nil {
		return repos, err
	}
	if err = json.Unmarshal(resp, &repos); err != nil {
		return repos, JSONError{err}
	}
	return repos, nil
}

// SyncProducts syncronizes the products from the current system with the SCC
// server.
func SyncProducts(opts *Options, products []registration.Product) ([]registration.Product, error) {
	remoteProducts := make([]registration.Product, 0)

	api := NewWrappedAPI(opts)
	conn := api.GetConnection()

	creds := conn.GetCredentials()
	login, password, credErr := creds.Login()
	if credErr != nil {
		return remoteProducts, credErr
	}

	var payload struct {
		Products []registration.Product `json:"products"`
	}
	payload.Products = products

	request, buildErr := conn.BuildRequest("POST", "/connect/systems/products/synchronize", payload)
	if buildErr != nil {
		return remoteProducts, buildErr
	}

	connection.AddSystemAuth(request, login, password)

	response, doErr := conn.Do(request)
	if doErr != nil {
		return remoteProducts, doErr
	}

	err := json.Unmarshal(response, &remoteProducts)
	return remoteProducts, err
}

// Call `updateMigrations` for online migrations.
func ProductMigrations(opts *Options, installed []registration.Product) ([]MigrationPath, error) {
	var payload struct {
		InstalledProducts []registration.Product `json:"installed_products"`
	}
	payload.InstalledProducts = installed

	return updateMigrations(opts, "/connect/systems/products/migrations", payload)
}

// Call `updateMigrations` for offline migrations.
func OfflineProductMigrations(opts *Options, installed []registration.Product, target registration.Product) ([]MigrationPath, error) {
	var payload struct {
		InstalledProducts []registration.Product `json:"installed_products"`
		TargetBaseProduct registration.Product   `json:"target_base_product"`
	}
	payload.InstalledProducts = installed
	payload.TargetBaseProduct = target

	return updateMigrations(opts, "/connect/systems/products/offline_migrations", payload)
}

// Post on a product migrations endpoint and get back the list of MigrationPath
// related to this operation.
func updateMigrations(opts *Options, url string, payload any) ([]MigrationPath, error) {
	migrations := make([]MigrationPath, 0)

	api := NewWrappedAPI(opts)
	conn := api.GetConnection()

	creds := conn.GetCredentials()
	login, password, credErr := creds.Login()
	if credErr != nil {
		return migrations, credErr
	}

	request, buildErr := conn.BuildRequest("POST", url, payload)
	if buildErr != nil {
		return migrations, buildErr
	}

	connection.AddSystemAuth(request, login, password)

	response, doErr := conn.Do(request)
	if doErr != nil {
		return migrations, doErr
	}

	err := json.Unmarshal(response, &migrations)
	return migrations, err
}
