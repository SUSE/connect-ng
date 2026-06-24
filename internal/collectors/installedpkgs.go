package collectors

import (
	"bufio"
	"bytes"
	"fmt"
	"slices"
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

	result, _ := profiles.BuildProfile(p.UpdateDataIDs, packagesTag, pkgsChecksumFile, filteredPkgs)
	return result, nil
}

// filterPackages returns a sorted list of packages filtered by vendor
func filterPackages(pkgs []byte) ([][]string, error) {
	reader := bytes.NewReader(pkgs)
	sc := bufio.NewScanner(reader)

	pkgSet := make(map[string]struct{})
	var pkgList [][]string

	for sc.Scan() {
		pkg := sc.Text()

		fields := strings.Split(pkg, "\t")
		if len(fields) < 5 {
			continue
		}

		vendor := fields[0]

		// Apply vendor filter
		if !strings.Contains(strings.ToUpper(vendor), filterVendor) {
			continue
		}

		// fields[1:5] includes Name, Version, Release, Arch
		packageFields := fields[1:5]
		setKey := strings.Join(packageFields, "\t")

		// Remove duplicates
		if _, ok := pkgSet[setKey]; !ok {
			pkgSet[setKey] = struct{}{}
			pkgList = append(pkgList, packageFields)
		}
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("error reading rpm query output: %v", err)
	}

	slices.SortFunc(pkgList, func(a, b []string) int {
		return slices.Compare(a, b)
	})

	return pkgList, nil
}
