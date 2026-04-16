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

const PkgsChecksumFile = "pkgs.txt"
const PackagesTag = "suse_pkgs"

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

	packagesPayload := formatPackagesPayload(filteredPkgs)

	result, _ := profiles.BuildProfile(p.UpdateDataIDs, PackagesTag, PkgsChecksumFile, packagesPayload)
	return result, nil
}

// formatPackagesPayload converts the raw package strings into the tuple payload format
func formatPackagesPayload(pkgs []string) [][]string {
	payload := make([][]string, 0, len(pkgs))
	for _, p := range pkgs {
		fields := strings.Split(p, "\t")
		payload = append(payload, fields)
	}
	return payload
}

// filterPackages returns a sorted list of packages filtered by vendor
// If filterVendor is empty, returns all packages
func filterPackages(pkgs []byte, filterVendor string) ([]string, error) {
	reader := bytes.NewReader(pkgs)
	sc := bufio.NewScanner(reader)

	var vendorRegex *regexp.Regexp
	if filterVendor != "" {
		// Regex to filter vendors (case insensitive)
		vendorRegex = regexp.MustCompile(`(?i)` + regexp.QuoteMeta(filterVendor))
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
		name := fields[1]
		version := fields[2]
		release := fields[3]
		arch := fields[4]
		setKey := fmt.Sprintf("%s\t%s\t%s\t%s", name, version, release, arch)

		// Apply vendor filter if specified
		if vendorRegex != nil && !vendorRegex.MatchString(vendor) {
			continue
		}

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
