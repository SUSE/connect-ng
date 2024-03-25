package connect

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
    "github.com/SUSE/connect-ng/internal/connect/models"
)


// installedProducts returns installed products
func installedProducts() ([]Product, error) {
	args := []string{"--disable-repositories", "--xmlout", "--non-interactive", "products", "-i"}
	output, err := zypperRun(args, []int{zypperOK})
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
	// override ProductType with OEM value if defined
	for i, p := range products.Products {
		if oemValue, err := oemReleaseType(p.ProductLine); err == nil {
			products.Products[i].ReleaseType = oemValue
		}
	}
	return products.Products, nil
}

// InstalledServices returns list of services installed on the system
func InstalledServices() ([]Service, error) {
	args := []string{"--xmlout", "--non-interactive", "services", "-d"}
	// Don't fail when zypper exits with 6 (no repositories)
	output, err := zypperRun(args, []int{zypperOK, zypperErrNoRepos})
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
	output, err := zypperRun([]string{"targetos"}, []int{zypperOK})
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
	// pass "insecure" setting to zypper via URL
	// https://en.opensuse.org/openSUSE:Libzypp_URIs
	if CFG.Insecure {
		u, err := url.Parse(serviceURL)
		if err != nil {
			return err
		}
		q := u.Query()
		q.Set("ssl_verify", "no")
		u.RawQuery = q.Encode()
		serviceURL = u.String()
	}
	args := []string{"--non-interactive", "addservice", "-t", "ris", serviceURL, serviceName}
	_, err := zypperRun(args, []int{zypperOK})
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

	args := []string{"--non-interactive", "removeservice", serviceName}
	_, err := zypperRun(args, []int{zypperOK})
	if err != nil {
		return err
	}
	return removeServiceCredentials(serviceName)
}

func enableServiceAutorefresh(serviceName string) error {
	args := []string{"--non-interactive", "modifyservice", "-r", serviceName}
	_, err := zypperRun(args, []int{zypperOK})
	return err
}

func refreshService(serviceName string) error {
	args := []string{"--non-interactive", "refs", serviceName}
	_, err := zypperRun(args, []int{zypperOK})
	return err
}

func refreshAllServices() error {
	args := []string{"--non-interactive", "refs"}
	_, err := zypperRun(args, []int{zypperOK})
	return err
}

// InstallReleasePackage ensures the <product-id>-release package is installed.
func InstallReleasePackage(identifier string) error {
	if identifier == "" {
		return nil
	}
	// return if the rpm is already installed
	args := []string{"rpm", "-q", identifier + "-release"}
	if _, err := execute(args, nil); err == nil {
		return nil
	}

	// In the case of packagehub we accept some repos to fail the initial refresh,
	// because the signing key is not yet imported. It is part of the -release package,
	// so the repos will be trusted after the release package is installed.
	validExitCodes := []int{zypperOK}
	if identifier == "PackageHub" {
		validExitCodes = append(validExitCodes, zypperInfoReposSkipped)
	}

	args = []string{"--no-refresh", "--non-interactive", "install", "--no-recommends",
		"--auto-agree-with-product-licenses", "-t", "product", identifier}

	if CFG.AutoImportRepoKeys {
		args = append([]string{"--gpg-auto-import-keys"}, args...)
	}

	_, err := zypperRun(args, validExitCodes)
	return err
}

func removeReleasePackage(identifier string) error {
	if identifier == "" {
		return nil
	}
	args := []string{"--no-refresh", "--non-interactive", "remove", "-t", "product", identifier}
	_, err := zypperRun(args, []int{zypperOK, zypperInfoCapNotFound})
	return err
}

func setReleaseVersion(version string) error {
	args := []string{"--non-interactive", "--releasever", version, "ref", "-f"}
	_, err := zypperRun(args, []int{zypperOK})
	return err
}

func zypperFlags(version string, quiet bool, verbose bool,
	nonInteractive bool, noRefresh bool) []string {
	flags := []string{}
	if nonInteractive {
		flags = append(flags, "--non-interactive")
	}
	if verbose {
		flags = append(flags, "--verbose")
	}
	if quiet {
		flags = append(flags, "--quiet")
	}
	if version != "" {
		flags = append(flags, "--releasever", version)
	}
	if noRefresh {
		flags = append(flags, "--no-refresh")
	}
	return flags
}

// RefreshRepos runs zypper to refresh all repositories
func RefreshRepos(version string, force, quiet, verbose, nonInteractive bool) error {
	args := []string{"ref"}
	flags := zypperFlags(version, quiet, verbose, nonInteractive, false)
	if force {
		args = append(args, "-f")
	}
	if CFG.AutoImportRepoKeys {
		args = append(args, "--gpg-auto-import-keys")
	}
	args = append(flags, args...)
	_, err := zypperRun(args, []int{zypperOK})
	return err
}

// DistUpgrade runs zypper dist-upgrade with given flags and extra args
func DistUpgrade(version string, quiet, verbose, AutoAgreeLicenses, nonInteractive bool, extraArgs []string) error {
	flags := zypperFlags(version, quiet, verbose, nonInteractive, true)
	args := append(flags, "dist-upgrade")
	if AutoAgreeLicenses {
		args = append(args, "--auto-agree-with-licenses")
	}
	args = append(args, extraArgs...)
	_, err := zypperRun(args, []int{zypperOK})
	return err
}

func parseReposXML(xmlDoc []byte) ([]Repo, error) {
	var repos struct {
		Repos []Repo `xml:"repo-list>repo"`
	}
	if err := xml.Unmarshal(xmlDoc, &repos); err != nil {
		return []Repo{}, err
	}
	return repos.Repos, nil
}

// Repos returns repositories configured on the system
func Repos() ([]Repo, error) {
	args := []string{"--xmlout", "--non-interactive", "repos", "-d"}
	// Don't fail when zypper exits with 6 (no repositories)
	output, err := zypperRun(args, []int{zypperOK, zypperErrNoRepos})
	if err != nil {
		return []Repo{}, err
	}
	return parseReposXML(output)
}

func parseSearchResultXML(xmlDoc []byte) ([]Package, error) {
	var packages struct {
		Packages []Package `xml:"search-result>solvable-list>solvable"`
	}
	if err := xml.Unmarshal(xmlDoc, &packages); err != nil {
		return []Package{}, err
	}
	return packages.Packages, nil
}

// FindProductPackages returns list of product packages for given product
func FindProductPackages(identifier string) ([]Package, error) {
	args := []string{"--xmlout", "--no-refresh", "--non-interactive", "search", "-s",
		"--match-exact", "-t", "product", identifier}
	// Don't fail when zypper exits with 104 (no product found) or 6 (no repositories)
	output, err := zypperRun(args, []int{zypperOK, zypperErrNoRepos, zypperInfoCapNotFound})
	if err != nil {
		return []Package{}, err
	}
	return parseSearchResultXML(output)
}

// DisableRepo disables zypper repo by name
func DisableRepo(name string) error {
	args := []string{"--non-interactive", "modifyrepo", "-d", name}
	_, err := zypperRun(args, []int{zypperOK})
	return err
}

// PatchCheck returns true if there are any patches pending to be installed.
func PatchCheck(updateStackOnly, quiet, verbose, nonInteractive, noRefresh bool) (bool, error) {
	flags := zypperFlags("", quiet, verbose, nonInteractive, noRefresh)
	args := append(flags, "patch-check")
	if updateStackOnly {
		args = append(args, "--updatestack-only")
	}
	_, err := zypperRun(args, []int{zypperOK})
	// zypperInfoUpdateNeeded or zypperInfoSecUpdateNeeded exit codes indicate
	// pending patches. return clear error
	if err != nil && containsInt(
		[]int{zypperInfoUpdateNeeded, zypperInfoSecUpdateNeeded},
		err.(ZypperError).ExitCode) {
		return true, nil
	}
	return false, err
}

// Patch installs all available needed patches.
func Patch(updateStackOnly, quiet, verbose, nonInteractive, noRefresh bool) error {
	flags := zypperFlags("", quiet, verbose, nonInteractive, noRefresh)
	args := append(flags, "patch")
	if updateStackOnly {
		args = append(args, "--updatestack-only")
	}
	_, err := zypperRun(args, []int{zypperOK})
	return err
}
