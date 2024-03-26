package scc

import (
	"encoding/json"
	"net/http"

	"github.com/SUSE/connect-ng/internal/config"
	"github.com/SUSE/connect-ng/internal/connect/models"
	"github.com/SUSE/connect-ng/internal/utils"
)

// announceSystem announces a system to SCC
// https://scc.suse.com/connect/v4/documentation#/subscriptions/post_subscriptions_systems
// The body parameter is produced by makeSysInfoBody()
func announceSystem(body []byte) (string, string, error) {
	resp, err := callHTTP("POST", "/connect/subscriptions/systems", body, nil, authToken)
	if err != nil {
		return "", "", err
	}
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err = json.Unmarshal(resp, &creds); err != nil {
		return "", "", utils.JSONError{err}
	}
	return creds.Login, creds.Password, nil
}

func upToDate() bool {
	// REVIST 404 case - see original
	// Should fail in any case. 422 error means that the endpoint is there and working right
	_, err := callHTTP("GET", "/connect/repositories/installer", nil, nil, authNone)
	if err == nil {
		return false
	}
	if ae, ok := err.(utils.APIError); ok {
		if ae.Code == http.StatusUnprocessableEntity {
			return true
		}
	}
	return false
}

// systemActivations returns a map keyed by "Identifier/Version/Arch"
func SystemActivations() (map[string]models.Activation, error) {
	activeMap := make(map[string]models.Activation)
	resp, err := callHTTP("GET", "/connect/systems/activations", nil, nil, authSystem)
	if err != nil {
		return activeMap, err
	}
	var activations []models.Activation
	if err = json.Unmarshal(resp, &activations); err != nil {
		return activeMap, utils.JSONError{err}
	}
	for _, activation := range activations {
		activeMap[activation.ToTriplet()] = activation
	}
	return activeMap, nil
}

func ShowProduct(productQuery models.Product) (models.Product, error) {
	resp, err := callHTTP("GET", "/connect/systems/products", nil, productQuery.ToQuery(), authSystem)
	remoteProduct := models.Product{}
	if err != nil {
		return remoteProduct, err
	}
	if err = json.Unmarshal(resp, &remoteProduct); err != nil {
		return remoteProduct, utils.JSONError{err}
	}
	return remoteProduct, nil
}

func upgradeProduct(product models.Product) (models.Service, error) {
	// NOTE: this can add some extra attributes to json payload which
	//       seem to be safely ignored by the API.
	payload, err := json.Marshal(product)
	remoteService := models.Service{}
	if err != nil {
		return remoteService, err
	}
	resp, err := callHTTP("PUT", "/connect/systems/products", payload, nil, authSystem)
	if err != nil {
		return remoteService, err
	}
	if err = json.Unmarshal(resp, &remoteService); err != nil {
		return remoteService, utils.JSONError{err}
	}
	return remoteService, nil
}

func downgradeProduct(product models.Product) (models.Service, error) {
	return upgradeProduct(product)
}

func activateProduct(product models.Product, email string) (models.Service, error) {
	var payload = struct {
		Indentifier string `json:"identifier"`
		Version     string `json:"version"`
		Arch        string `json:"arch"`
		ReleaseType string `json:"release_type"`
		Token       string `json:"token"`
		Email       string `json:"email"`
	}{
		product.Name,
		product.Version,
		product.Arch,
		product.ReleaseType,
		config.CFG.Token,
		email,
	}

	service := models.Service{}
	body, err := json.Marshal(payload)
	if err != nil {
		return service, err
	}
	resp, err := callHTTP("POST", "/connect/systems/products", body, nil, authSystem)
	if err != nil {
		return service, err
	}
	err = json.Unmarshal(resp, &service)
	if err != nil {
		return service, utils.JSONError{err}
	}
	return service, nil
}

func deactivateProduct(product models.Product) (models.Service, error) {
	// NOTE: this can add some extra attributes to json payload which
	//       seem to be safely ignored by the API.
	payload, err := json.Marshal(product)
	remoteService := models.Service{}
	if err != nil {
		return remoteService, err
	}
	resp, err := callHTTP("DELETE", "/connect/systems/products", payload, nil, authSystem)
	if err != nil {
		return remoteService, err
	}
	if err = json.Unmarshal(resp, &remoteService); err != nil {
		return remoteService, utils.JSONError{err}
	}
	return remoteService, nil
}

func deregisterSystem() error {
	_, err := callHTTP("DELETE", "/connect/systems", nil, nil, authSystem)
	return err
}

func syncProducts(products []models.Product) ([]models.Product, error) {
	remoteProducts := make([]models.Product, 0)
	var payload struct {
		Products []models.Product `json:"products"`
	}
	payload.Products = products
	body, err := json.Marshal(payload)
	if err != nil {
		return remoteProducts, err
	}
	resp, err := callHTTP("POST", "/connect/systems/products/synchronize", body, nil, authSystem)
	if err != nil {
		return remoteProducts, err
	}
	err = json.Unmarshal(resp, &remoteProducts)
	if err != nil {
		return remoteProducts, utils.JSONError{err}
	}
	return remoteProducts, nil
}

// updateSystem updates the system's hardware information on SCC
// https://scc.suse.com/connect/v4/documentation#/systems/put_systems
// The body parameter is produced by makeSysInfoBody()
func updateSystem(body []byte) error {
	_, err := callHTTP("PUT", "/connect/systems", body, nil, authSystem)
	return err
}

// makeSysInfoBody returns the JSON payload needed for the announce/update system calls
func makeSysInfoBody(distroTarget, namespace string, instanceData []byte) ([]byte, error) {
	var payload struct {
		Hostname     string `json:"hostname"`
		DistroTarget string `json:"distro_target"`
		InstanceData string `json:"instance_data,omitempty"`
		Namespace    string `json:"namespace,omitempty"`
		Hwinfo       hwinfo `json:"hwinfo"`
	}
	if distroTarget != "" {
		payload.DistroTarget = distroTarget
	} else {
		var err error
		payload.DistroTarget, err = zypperDistroTarget()
		if err != nil {
			return nil, err
		}
	}
	payload.InstanceData = string(instanceData)
	payload.Namespace = namespace

	hw, err := getHwinfo()
	if err != nil {
		return nil, err
	}
	payload.Hostname = hw.Hostname
	payload.Hwinfo = hw

	return json.Marshal(payload)
}

func productMigrations(installed []models.Product) ([]MigrationPath, error) {
	migrations := make([]MigrationPath, 0)
	var payload struct {
		InstalledProducts []models.Product `json:"installed_products"`
	}
	payload.InstalledProducts = installed
	body, err := json.Marshal(payload)
	if err != nil {
		return migrations, err
	}
	resp, err := callHTTP("POST", "/connect/systems/products/migrations", body, nil, authSystem)
	if err != nil {
		return migrations, err
	}
	if err = json.Unmarshal(resp, &migrations); err != nil {
		return migrations, utils.JSONError{err}
	}
	return migrations, nil
}

func offlineProductMigrations(installed []models.Product, target models.Product) ([]MigrationPath, error) {
	migrations := make([]MigrationPath, 0)
	var payload struct {
		InstalledProducts []models.Product `json:"installed_products"`
		TargetBaseProduct models.Product   `json:"target_base_product"`
	}
	payload.InstalledProducts = installed
	payload.TargetBaseProduct = target
	body, err := json.Marshal(payload)
	if err != nil {
		return migrations, err
	}
	resp, err := callHTTP("POST", "/connect/systems/products/offline_migrations", body, nil, authSystem)
	if err != nil {
		return migrations, err
	}
	if err = json.Unmarshal(resp, &migrations); err != nil {
		return migrations, utils.JSONError{err}
	}
	return migrations, nil
}

func installerUpdates(product models.Product) ([]models.Repository, error) {
	repos := make([]models.Repository, 0)
	resp, err := callHTTP("GET", "/connect/repositories/installer", nil, product.ToQuery(), authNone)
	if err != nil {
		return repos, err
	}
	if err = json.Unmarshal(resp, &repos); err != nil {
		return repos, utils.JSONError{err}
	}
	return repos, nil
}

//moved from status.go

// SystemProducts returns sum of installed and activated products
// Products from zypper have priority over products from
// activations as they have summary field which is missing
// in the latter.
func SystemProducts() ([]models.Product, error) {
	products, err := models.InstalledProducts()
	if err != nil {
		return products, err
	}
	installedIDs := models.NewStringSet()
	for _, prod := range products {
		installedIDs.Add(prod.ToTriplet())
	}
	if !IsRegistered() {
		return products, nil
	}
	activations, err := SystemActivations()
	if err != nil {
		return products, err
	}
	for _, a := range activations {
		if !installedIDs.Contains(a.Service.Product.ToTriplet()) {
			products = append(products, a.Service.Product)
		}
	}

	return products, nil
}
