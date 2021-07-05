package connect

import (
	"encoding/xml"
)

const (
	zypperPath = "/usr/bin/zypper"
)

const (
	zypperOK = 0

	// Single-digit codes denote errors
	zypperErrBug         = 1 // Unexpected situation occurred, probably caused by a bug
	zypperErrSyntax      = 2 // zypper was invoked with an invalid command or option, or a bad syntax
	zypperErrInvalidArgs = 3 // Some of provided arguments were invalid. E.g. an invalid URI was provided to the addrepo command
	zypperErrZypp        = 4 // A problem is reported by ZYPP library
	zypperErrPrivileges  = 5 // User invoking zypper has insufficient privileges for specified operation
	zypperErrNoRepos     = 6 // No repositories are defined
	zypperErrZyppLocked  = 7 // The ZYPP library is locked, e.g. packagekit is running
	zypperErrCommit      = 8 // An error occurred during installation or removal of packages. You may run zypper verify to repair any dependency problems

	// Codes from 100 and above denote additional information passing
	zypperInfoUpdateNeeded    = 100 // Returned by the patch-check command if there are patches available for installation
	zypperInfoSecUpdateNeeded = 101 // Returned by the patch-check command if there are security patches available for installation
	zypperInfoRebootNeeded    = 102 // Returned after a successful installation of a patch which requires reboot of computer
	zypperInfoRestartNeeded   = 103 // Returned after a successful installation of a patch which requires restart of the package manager itself
	zypperInfoCapNotFound     = 104 // install or remove command encountered arguments matching no of the available package names or capabilities
	zypperInfoOnSignal        = 105 // Returned upon exiting after receiving a SIGINT or SIGTERM
	zypperInfoReposSkipped    = 106 // Some repository had to be disabled temporarily because it failed to refresh
)

func zypperRun(args []string, quiet bool, validExitCodes []int) ([]byte, error) {
	cmd := []string{zypperPath, "--non-interactive"}
	if CFG.FsRoot != "" {
		cmd = append(cmd, "--root", CFG.FsRoot)
	}
	cmd = append(cmd, args...)
	output, err := execute(cmd, quiet, validExitCodes)
	if err != nil {
		if ee, ok := err.(ExecuteError); ok {
			return nil, ZypperError{ExitCode: ee.ExitCode, Output: ee.Output}
		}
	}
	return output, nil
}

// installedProducts returns installed products
func installedProducts() ([]Product, error) {
	args := []string{"--disable-repositories", "--xmlout", "products", "-i"}
	output, err := zypperRun(args, false, []int{zypperOK})
	if err != nil {
		return []Product{}, err
	}
	return parseProductsXML(output)
}

// parseProductsXML returns products parsed from zypper XML
func parseProductsXML(xmlDoc []byte) ([]Product, error) {
	var products struct {
		Products []Product `xml:"product-list>product"`
	}
	if err := xml.Unmarshal(xmlDoc, &products); err != nil {
		return []Product{}, err
	}
	return products.Products, nil
}

func installedServices() ([]Service, error) {
	args := []string{"--xmlout", "services", "-d"}
	// Don't fail when zypper exits with 6 (no repositories)
	output, err := zypperRun(args, false, []int{zypperOK, zypperErrNoRepos})
	if err != nil {
		return []Service{}, err
	}
	return parseServicesXML(output)
}

func parseServicesXML(xmlDoc []byte) ([]Service, error) {
	var services struct {
		Services []Service `xml:"service-list>service"`
	}
	if err := xml.Unmarshal(xmlDoc, &services); err != nil {
		return []Service{}, err
	}
	return services.Services, nil
}

// TODO: memoize?
func baseProduct() (Product, error) {
	products, err := installedProducts()
	if err != nil {
		return Product{}, err
	}
	for _, product := range products {
		if product.IsBase {
			return product, nil
		}
	}
	return Product{}, ErrCannotDetectBaseProduct
}

func zypperDistroTarget() (string, error) {
	output, err := zypperRun([]string{"targetos"}, false, []int{zypperOK})
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func addService(serviceURL, serviceName string, refresh bool) error {
	// Remove old service which could be modified by a customer
	if err := removeService(serviceName); err != nil {
		return err
	}
	args := []string{"addservice", "-t", "ris", serviceURL, serviceName}
	_, err := zypperRun(args, true, []int{zypperOK})
	if err != nil {
		return err
	}
	if err = enableServiceAutorefresh(serviceName); err != nil {
		return err
	}
	if err = writeServiceCredentials(serviceName); err != nil {
		return err
	}
	if refresh {
		return refreshService(serviceName)
	}
	return nil
}

func removeService(serviceName string) error {
	Debug.Println("Removing service: ", serviceName)

	args := []string{"removeservice", serviceName}
	_, err := zypperRun(args, true, []int{zypperOK})
	if err != nil {
		return err
	}
	return removeServiceCredentials(serviceName)
}

func enableServiceAutorefresh(serviceName string) error {
	args := []string{"modifyservice", "-r", serviceName}
	_, err := zypperRun(args, true, []int{zypperOK})
	return err
}

func refreshService(serviceName string) error {
	args := []string{"refs", serviceName}
	_, err := zypperRun(args, true, []int{zypperOK})
	return err
}

func refreshAllServices() error {
	args := []string{"refs"}
	_, err := zypperRun(args, true, []int{zypperOK})
	return err
}

func installReleasePackage(identifier string) error {
	if identifier == "" {
		return nil
	}
	// return if the rpm is already installed
	args := []string{"rpm", "-q", identifier + "-release"}
	if _, err := execute(args, false, nil); err == nil {
		return nil
	}

	// In the case of packagehub we accept some repos to fail the initial refresh,
	// because the signing key is not yet imported. It is part of the -release package,
	// so the repos will be trusted after the release package is installed.
	validExitCodes := []int{zypperOK}
	if identifier == "PackageHub" {
		validExitCodes = append(validExitCodes, zypperInfoReposSkipped)
	}

	args = []string{"--no-refresh", "install", "--no-recommends",
		"--auto-agree-with-product-licenses", "-t", "product", identifier}

	_, err := zypperRun(args, false, validExitCodes)
	return err
}

func removeReleasePackage(identifier string) error {
	if identifier == "" {
		return nil
	}
	args := []string{"--no-refresh", "remove", "-t", "product", identifier}
	_, err := zypperRun(args, true, []int{zypperOK, zypperInfoCapNotFound})
	return err
}

func setReleaseVersion(version string) error {
	args := []string{"--releasever", version, "ref", "-f"}
	_, err := zypperRun(args, true, []int{zypperOK})
	return err
}
