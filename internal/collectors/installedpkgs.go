package collectors

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/pkg/profiles"
)

type InstalledPackages struct {
	UpdateDataIDs bool
}

const (
	pkgsChecksumFile = "pkgs.txt"
	packagesTag      = "installed_pkgs"
	filterVendor     = "SUSE"
)

func (p InstalledPackages) run(arch string) (Result, error) {
	util.Debug.Print("InstalledPackages.UpdateDataIds", p.UpdateDataIDs)

	queryFormat := "%{VENDOR}\\t%{NAME}\\t%{VERSION}\\t%{RELEASE}\\t%{ARCH}\n"
	pkgs, err := util.Execute([]string{"rpm", "-qa", "--queryformat", queryFormat}, nil)
	if err != nil {
		return Result{}, err
	}

	filteredPkgs, err := filterPackages(pkgs)
	if err != nil {
		return Result{}, err
	}

	packagesPayload := formatPackagesPayload(filteredPkgs)

	result, _ := profiles.BuildProfile(p.UpdateDataIDs, packagesTag, pkgsChecksumFile, packagesPayload)
	return result, nil
}

// formatPackagesPayload converts the raw package strings into the positional array payload format
func formatPackagesPayload(pkgs []string) [][]string {
	payload := make([][]string, 0, len(pkgs))
	for _, p := range pkgs {
		fields := strings.Split(p, "\t")
		payload = append(payload, fields)
	}
	return payload
}

// filterPackages returns a sorted list of packages filtered by vendor
func filterPackages(pkgs []byte) ([]string, error) {
	reader := bytes.NewReader(pkgs)
	sc := bufio.NewScanner(reader)

	vendorRegex, err := regexp.Compile(`(?i)` + regexp.QuoteMeta(filterVendor))
	if err != nil {
		return nil, fmt.Errorf("error compiling vendorRegex: %v", err)
	}

	pkgSet := make(map[string]struct{})
	var pkgList []string

	for sc.Scan() {
		pkg := sc.Text()

		fields := strings.Split(pkg, "\t")
		if len(fields) < 5 {
			continue
		}

		vendor := fields[0]

		// Apply vendor filter
		if !vendorRegex.MatchString(vendor) {
			continue
		}

		name := fields[1]
		version := fields[2]
		release := fields[3]
		arch := fields[4]
		setKey := fmt.Sprintf("%s\t%s\t%s\t%s", name, version, release, arch)

		// Remove duplicates
		if _, ok := pkgSet[setKey]; !ok {
			pkgSet[setKey] = struct{}{}
			pkgList = append(pkgList, setKey)
		}
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("error reading rpm query output: %v", err)
	}

	sort.Strings(pkgList)

	return pkgList, nil
}
