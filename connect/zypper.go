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
	cmd := []string{zypperPath}
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
	args := []string{"--disable-repositories", "--xmlout", "--non-interactive", "products", "-i"}
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
