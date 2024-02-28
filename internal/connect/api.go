package connect

import (
	"encoding/json"
	"net/http"
)

type SystemInformation struct {
	Hostname     string `json:"hostname"`
	DistroTarget string `json:"distro_target"`
	InstanceData string `json:"instance_data,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
	Hwinfo       hwinfo `json:"hwinfo"`
}

type SystemCredentialsResponse struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type SystemActivationRequest struct {
	Indentifier string `json:"identifier"`
	Version     string `json:"version"`
	Arch        string `json:"arch"`
	ReleaseType string `json:"release_type"`
	Token       string `json:"token"`
	Email       string `json:"email"`
}

type SystemSynchronizeProductsRequest struct {
	Products []Product `json:"products"`
}

type SystemMigrationsRequest struct {
	InstalledProducts []Product `json:"installed_products"`
	TargetBaseProduct Product   `json:"target_base_product"`
}

// announceSystem announces a system to SCC
// https://scc.suse.com/connect/v4/documentation#/subscriptions/post_subscriptions_systems
// The body parameter is produced by makeSysInfoBody()
func announceSystem(body []byte) (string, string, error) {
	resp, err := callHTTP("POST", "/connect/subscriptions/systems", body, nil, authToken)
	if err != nil {
		return "", "", err
	}
	creds := SystemCredentialsResponse{}
	if err = json.Unmarshal(resp, &creds); err != nil {
		return "", "", JSONError{err}
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
	if ae, ok := err.(APIError); ok {
		if ae.Code == http.StatusUnprocessableEntity {
			return true
		}
	}
	return false
}

// systemActivations returns a map keyed by "Identifier/Version/Arch"
func systemActivations() (map[string]Activation, error) {
	activeMap := make(map[string]Activation)
	resp, err := callHTTP("GET", "/connect/systems/activations", nil, nil, authSystem)
	if err != nil {
		return activeMap, err
	}
	var activations []Activation
	if err = json.Unmarshal(resp, &activations); err != nil {
		return activeMap, JSONError{err}
	}
	for _, activation := range activations {
		activeMap[activation.toTriplet()] = activation
	}
	return activeMap, nil
}

func showProduct(productQuery Product) (Product, error) {
	resp, err := callHTTP("GET", "/connect/systems/products", nil, productQuery.toQuery(), authSystem)
	remoteProduct := Product{}
	if err != nil {
		return remoteProduct, err
	}
	if err = json.Unmarshal(resp, &remoteProduct); err != nil {
		return remoteProduct, JSONError{err}
	}
	return remoteProduct, nil
}

func upgradeProduct(product Product) (Service, error) {
	// NOTE: this can add some extra attributes to json payload which
	//       seem to be safely ignored by the API.
	payload, err := json.Marshal(product)
	remoteService := Service{}
	if err != nil {
		return remoteService, err
	}
	resp, err := callHTTP("PUT", "/connect/systems/products", payload, nil, authSystem)
	if err != nil {
		return remoteService, err
	}
	if err = json.Unmarshal(resp, &remoteService); err != nil {
		return remoteService, JSONError{err}
	}
	return remoteService, nil
}

func downgradeProduct(product Product) (Service, error) {
	return upgradeProduct(product)
}

func activateProduct(product Product, email string) (Service, error) {
	payload := SystemActivationRequest{
		Indentifier: product.Name,
		Version:     product.Version,
		Arch:        product.Arch,
		ReleaseType: product.ReleaseType,
		Token:       CFG.Token,
		Email:       email,
	}

	service := Service{}
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
		return service, JSONError{err}
	}
	return service, nil
}

func deactivateProduct(product Product) (Service, error) {
	// NOTE: this can add some extra attributes to json payload which
	//       seem to be safely ignored by the API.
	payload, err := json.Marshal(product)
	remoteService := Service{}
	if err != nil {
		return remoteService, err
	}
	resp, err := callHTTP("DELETE", "/connect/systems/products", payload, nil, authSystem)
	if err != nil {
		return remoteService, err
	}
	if err = json.Unmarshal(resp, &remoteService); err != nil {
		return remoteService, JSONError{err}
	}
	return remoteService, nil
}

func deregisterSystem() error {
	_, err := callHTTP("DELETE", "/connect/systems", nil, nil, authSystem)
	return err
}

func syncProducts(products []Product) ([]Product, error) {
	remoteProducts := make([]Product, 0)
	payload := SystemSynchronizeProductsRequest{
		Products: products,
	}
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
		return remoteProducts, JSONError{err}
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
	payload := SystemInformation{}

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

func productMigrations(installed []Product) ([]MigrationPath, error) {
	migrations := make([]MigrationPath, 0)
	payload := SystemMigrationsRequest{
		InstalledProducts: installed,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return migrations, err
	}
	resp, err := callHTTP("POST", "/connect/systems/products/migrations", body, nil, authSystem)
	if err != nil {
		return migrations, err
	}
	if err = json.Unmarshal(resp, &migrations); err != nil {
		return migrations, JSONError{err}
	}
	return migrations, nil
}

func offlineProductMigrations(installed []Product, target Product) ([]MigrationPath, error) {
	migrations := make([]MigrationPath, 0)
	payload := SystemMigrationsRequest{
		InstalledProducts: installed,
		TargetBaseProduct: target,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return migrations, err
	}
	resp, err := callHTTP("POST", "/connect/systems/products/offline_migrations", body, nil, authSystem)
	if err != nil {
		return migrations, err
	}
	if err = json.Unmarshal(resp, &migrations); err != nil {
		return migrations, JSONError{err}
	}
	return migrations, nil
}

func installerUpdates(product Product) ([]Repo, error) {
	repos := make([]Repo, 0)
	resp, err := callHTTP("GET", "/connect/repositories/installer", nil, product.toQuery(), authNone)
	if err != nil {
		return repos, err
	}
	if err = json.Unmarshal(resp, &repos); err != nil {
		return repos, JSONError{err}
	}
	return repos, nil
}
