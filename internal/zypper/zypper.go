package zypper

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/SUSE/connect-ng/internal/credentials"
	"github.com/SUSE/connect-ng/internal/util"
)

const (
	zypperPath = "/usr/bin/zypper"
	oemPath    = "/var/lib/suseRegister/OEM"
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

type ZypperProduct struct {
	Name        string `xml:"name,attr"`
	Version     string `xml:"version,attr"`
	Arch        string `xml:"arch,attr"`
	Release     string `xml:"release,attr"`
	Summary     string `xml:"summary,attr"`
	IsBase      bool   `xml:"isbase,attr"`
	ReleaseType string `xml:"registerrelease,attr"`
	ProductLine string `xml:"productline,attr"`

	// these are used by YaST
	Description string `xml:"description"`
}

type ZypperService struct {
	URL  string `xml:"url,attr"`
	Name string `xml:"name,attr"`
}

// FIXME: see how we can do this better
var zypperFilesystemRoot = "/"

func SetFilesystemRoot(arg string) {
	zypperFilesystemRoot = arg
}

func GetFilesystemRoot() string {
	return zypperFilesystemRoot
}

func zypperRun(args []string, validExitCodes []int) ([]byte, error) {
	cmd := []string{zypperPath}
	if zypperFilesystemRoot != "/" {
		cmd = append(cmd, "--root", zypperFilesystemRoot)
	}
	cmd = append(cmd, args...)
	util.QuietOut.Printf("\nExecuting '%s'\n\n", strings.Join(cmd, " "))
	output, err := util.Execute(cmd, validExitCodes)
	if err != nil {
		if ee, ok := err.(util.ExecuteError); ok {
			return nil, ZypperError(ee)
		}
	}
	return output, nil
}

// installedProducts returns installed products
func InstalledProducts() ([]ZypperProduct, error) {
	args := []string{"--disable-repositories", "--xmlout", "--non-interactive", "products", "-i"}
	output, err := zypperRun(args, []int{zypperOK})
	if err != nil {
		return []ZypperProduct{}, err
	}
	return parseProductsXML(output)
}

// get first line of OEM file if present
func oemReleaseType(productLine string) (string, error) {
	if productLine == "" {
		return "", fmt.Errorf("empty productline")
	}
	oemFile := filepath.Join(zypperFilesystemRoot, oemPath, productLine)
	if !util.FileExists(oemFile) {
		return "", fmt.Errorf("OEM file not found: %v", oemFile)
	}
	data, err := os.ReadFile(oemFile)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("empty OEM file: %v", oemFile)
	}
	return strings.TrimSpace(lines[0]), nil
}

// parseProductsXML returns products parsed from zypper XML
func parseProductsXML(xmlDoc []byte) ([]ZypperProduct, error) {
	var products struct {
		Products []ZypperProduct `xml:"product-list>product"`
	}
	if err := xml.Unmarshal(xmlDoc, &products); err != nil {
		return []ZypperProduct{}, err
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
func InstalledServices() ([]ZypperService, error) {
	args := []string{"--xmlout", "--non-interactive", "services", "-d"}
	// Don't fail when zypper exits with 6 (no repositories)
	output, err := zypperRun(args, []int{zypperOK, zypperErrNoRepos})
	if err != nil {
		return []ZypperService{}, err
	}
	return parseServicesXML(output)
}

func parseServicesXML(xmlDoc []byte) ([]ZypperService, error) {
	var services struct {
		Services []ZypperService `xml:"service-list>service"`
	}
	if err := xml.Unmarshal(xmlDoc, &services); err != nil {
		return []ZypperService{}, err
	}
	return services.Services, nil
}

// TODO: memoize?
func BaseProduct() (ZypperProduct, error) {
	products, err := InstalledProducts()
	if err != nil {
		return ZypperProduct{}, err
	}
	for _, product := range products {
		if product.IsBase {
			return product, nil
		}
	}
	return ZypperProduct{}, ErrCannotDetectBaseProduct
}

func DistroTarget() (string, error) {
	output, err := zypperRun([]string{"targetos"}, []int{zypperOK})
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func AddService(serviceURL, serviceName string, refresh bool, insecure bool) error {
	// Remove old service which could be modified by a customer
	if err := RemoveService(serviceName); err != nil {
		return err
	}
	// pass "insecure" setting to zypper via URL
	// https://en.opensuse.org/openSUSE:Libzypp_URIs
	if insecure {
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
	if err = EnableServiceAutorefresh(serviceName); err != nil {
		return err
	}
	sccCredentials, _ := credentials.ReadCredentials(credentials.SystemCredentialsPath(zypperFilesystemRoot))
	if err = credentials.CreateCredentials(sccCredentials.Username,
		sccCredentials.Password,
		sccCredentials.SystemToken,
		credentials.ServiceCredentialsPath(serviceName, zypperFilesystemRoot)); err != nil {
		return err
	}
	if refresh {
		return RefreshService(serviceName)
	}
	return nil
}

func RemoveService(serviceName string) error {
	util.Debug.Println("Removing service: ", serviceName)

	args := []string{"--non-interactive", "removeservice", serviceName}
	_, err := zypperRun(args, []int{zypperOK})
	if err != nil {
		return err
	}
	// return removeServiceCredentials(serviceName)
	return util.RemoveFile(credentials.ServiceCredentialsPath(serviceName, zypperFilesystemRoot))
}

func EnableServiceAutorefresh(serviceName string) error {
	args := []string{"--non-interactive", "modifyservice", "-r", serviceName}
	_, err := zypperRun(args, []int{zypperOK})
	return err
}

func RefreshService(serviceName string) error {
	args := []string{"--non-interactive", "refs", serviceName}
	_, err := zypperRun(args, []int{zypperOK})
	return err
}

func RefreshAllServices() error {
	args := []string{"--non-interactive", "refs"}
	_, err := zypperRun(args, []int{zypperOK})
	return err
}

// InstallReleasePackage ensures the <product-id>-release package is installed.
func InstallReleasePackage(identifier string, autoImportRepoKeys bool) error {
	if identifier == "" {
		return nil
	}
	// return if the rpm is already installed
	args := []string{"rpm", "-q", identifier + "-release"}
	if _, err := util.Execute(args, nil); err == nil {
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

	if autoImportRepoKeys {
		args = append([]string{"--gpg-auto-import-keys"}, args...)
	}

	_, err := zypperRun(args, validExitCodes)
	return err
}

func RemoveReleasePackage(identifier string) error {
	if identifier == "" {
		return nil
	}
	args := []string{"--no-refresh", "--non-interactive", "remove", "-t", "product", identifier}
	_, err := zypperRun(args, []int{zypperOK, zypperInfoCapNotFound})
	return err
}

func SetReleaseVersion(version string) error {
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
func RefreshRepos(version string, force, quiet, verbose, nonInteractive bool, autoImportRepoKeys bool) error {
	args := []string{"ref"}
	flags := zypperFlags(version, quiet, verbose, nonInteractive, false)
	if force {
		args = append(args, "-f")
	}
	if autoImportRepoKeys {
		args = append([]string{"--gpg-auto-import-keys"}, args...)
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

// Package holds package info as returned by `zypper search`
type Package struct {
	Name    string `xml:"name,attr"`
	Edition string `xml:"edition,attr"` // VERSION[-RELEASE]
	Arch    string `xml:"arch,attr"`
	Repo    string `xml:"repository,attr"`
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
	validExitCodes := []int{zypperInfoUpdateNeeded, zypperInfoSecUpdateNeeded}
	if err != nil && slices.Contains(validExitCodes, err.(ZypperError).ExitCode) {
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

// ToTriplet returns <name>/<version>/<arch> string for product
func (zp ZypperProduct) ToTriplet() string {
	return zp.Name + "/" + zp.Version + "/" + zp.Arch
}

// SplitTriplet returns a product from given or error for invalid input
func SplitTriplet(p string) (ZypperProduct, error) {
	if match, _ := regexp.MatchString(`^\S+/\S+/\S+$`, p); !match {
		return ZypperProduct{}, fmt.Errorf("invalid product; <internal name>/<version>/<architecture> format expected")
	}
	parts := strings.Split(p, "/")
	return ZypperProduct{Name: parts[0], Version: parts[1], Arch: parts[2]}, nil
}
