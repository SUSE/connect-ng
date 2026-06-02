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

const PkgsChecksumFile = "installed-pkgs-profile-id"
const PackagesTag = "installed_pkgs"

func (p InstalledPackages) run(arch string) (Result, error) {
	util.Debug.Print("InstalledPackages.UpdateDataIds", p.UpdateDataIDs)

	// This could come from CollectorOptions by adding a Params field to installed_pkgs collector
	filterVendor := "SUSE"

	qf := "%{VENDOR}\\t%{NAME}\\t%{VERSION}\\t%{RELEASE}\\t%{ARCH}\n"
	pkgs, err := util.Execute([]string{"rpm", "-qa", "--qf", qf}, nil)
	if err != nil {
		return Result{}, err
	}

	filteredPkgs, err := filterPackages(pkgs, filterVendor)
	if err != nil {
		return Result{}, err
	}

	result, _ := profiles.BuildProfile(p.UpdateDataIDs, PackagesTag, PkgsChecksumFile, filteredPkgs)
	return result, nil
}

// filterPackages returns a sorted 2D array of packages filtered by vendor
// Each package is represented as [name, version, release, architecture]
// If filterVendor is empty, returns all packages
func filterPackages(pkgs []byte, filterVendor string) ([][]string, error) {
	reader := bytes.NewReader(pkgs)
	sc := bufio.NewScanner(reader)

	var vendorRegex *regexp.Regexp
	if filterVendor != "" {
		// Regex to filter vendors (case insensitive)
		vendorRegex = regexp.MustCompile(`(?i)` + regexp.QuoteMeta(filterVendor))
	}

	pkgSet := make(map[string]struct{})
	pkgList := [][]string{}

	for sc.Scan() {
		pkg := sc.Text()
		fields := strings.Split(pkg, "\t")
		if len(fields) < 5 {
			continue
		}

		vendor := fields[0]
		name := fields[1]
		version := fields[2]
		release := fields[3]
		arch := fields[4]
		setKey := fmt.Sprintf("%s %s %s %s", name, version, release, arch)

		// Apply vendor filter if specified
		if vendorRegex != nil && !vendorRegex.MatchString(vendor) {
			continue
		}

		// Remove duplicates
		if _, ok := pkgSet[setKey]; !ok {
			pkgSet[setKey] = struct{}{}
			pkgList = append(pkgList, []string{name, version, release, arch})
		}
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("error reading rpm query output: %v", err)
	}

	// Sort by name, then version, then release, then arch
	sort.Slice(pkgList, func(i, j int) bool {
		if pkgList[i][0] != pkgList[j][0] {
			return pkgList[i][0] < pkgList[j][0]
		}
		if pkgList[i][1] != pkgList[j][1] {
			return pkgList[i][1] < pkgList[j][1]
		}
		if pkgList[i][2] != pkgList[j][2] {
			return pkgList[i][2] < pkgList[j][2]
		}
		return pkgList[i][3] < pkgList[j][3]
	})

	return pkgList, nil
}
