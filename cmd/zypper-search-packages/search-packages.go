package main

import (
	"bufio"
	_ "embed"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SUSE/connect-ng/internal/connect"
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/pkg/registration"
)

var (
	//go:embed searchPackagesUsage.txt
	searchPackagesUsageText string
)

type searchResult struct {
	Name    string
	Version string
	Release string
	Arch    string
	// package from addon
	ProdIdent   string
	ProdName    string
	ProdEdition string
	ProdArch    string
	ProdFree    bool
	// package from local repo
	Repo      string
	PkgStatus string
}

func (sr searchResult) FullName() string {
	return fmt.Sprintf("%s-%s-%s.%s", sr.Name, sr.Version, sr.Release, sr.Arch)
}

func (sr searchResult) ConnectCmd() string {
	if sr.ProdIdent == "" {
		return ""
	}
	ret := "SUSEConnect --product " + sr.ProdIdent
	if !sr.ProdFree {
		ret += " -r ADDITIONAL REGCODE"
	}
	return ret
}

func (sr searchResult) ModuleOrRepo() string {
	if sr.ProdIdent != "" {
		return sr.ProdIdent
	}
	return sr.Repo
}

func (sr searchResult) ProdNameOrPkgStatus() string {
	if sr.ProdName != "" {
		return sr.ProdName
	}
	return sr.PkgStatus
}

func (sr searchResult) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	p := struct {
		Name   string `xml:"name"`
		Module string `xml:"module"`
	}{sr.FullName(), sr.ModuleOrRepo()}
	return e.EncodeElement(p, start)
}

func main() {
	// Ensure Zypper is installed.
	if err := util.EnsureZypper(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var (
		matchExact    bool
		caseSensitive bool
		sortByName    bool
		sortByRepo    bool
		groupByModule bool
		noLocalRepos  bool
		details       bool
		xmlout        bool

		bNoop bool
		sNoop string
	)

	// flags without variables match defaults and/or will be processed using flag.Visit()
	flag.Bool("match-substrings", false, "")
	flag.Bool("match-words", false, "")
	flag.BoolVar(&matchExact, "match-exact", false, "")
	flag.BoolVar(&matchExact, "x", false, "")
	flag.Bool("provides", false, "")
	flag.Bool("recommends", false, "")
	flag.Bool("requires", false, "")
	flag.Bool("suggests", false, "")
	flag.Bool("supplements", false, "")
	flag.Bool("conflicts", false, "")
	flag.Bool("obsoletes", false, "")
	flag.String("name", "", "")
	flag.String("n", "", "")
	flag.Bool("file-list", false, "")
	flag.Bool("f", false, "")
	flag.Bool("search-descriptions", false, "")
	flag.Bool("d", false, "")
	flag.BoolVar(&caseSensitive, "case-sensitive", false, "")
	flag.BoolVar(&caseSensitive, "C", false, "")
	flag.BoolVar(&bNoop, "installed-only", false, "")
	flag.BoolVar(&bNoop, "i", false, "")
	flag.Bool("not-installed-only", false, "")
	flag.Bool("u", false, "")
	flag.String("type", "", "")
	flag.String("t", "", "")
	flag.StringVar(&sNoop, "repo", "", "")
	flag.StringVar(&sNoop, "r", "", "")
	flag.BoolVar(&sortByName, "sort-by-name", false, "")
	flag.BoolVar(&sortByRepo, "sort-by-repo", false, "")
	flag.BoolVar(&groupByModule, "group-by-module", false, "")
	flag.BoolVar(&groupByModule, "g", false, "")
	flag.BoolVar(&noLocalRepos, "no-query-local", false, "")
	flag.BoolVar(&details, "details", false, "")
	flag.BoolVar(&details, "s", false, "")
	flag.BoolVar(&details, "verbose", false, "")
	flag.BoolVar(&details, "v", false, "")
	flag.BoolVar(&xmlout, "xmlout", false, "")

	// disable default usage handling because flag.failf() displays usage even
	// when ContinueOnError is set. -h/--help is handled explicitly below
	flag.Usage = func() {}
	// switch to continue on parsing errors because of 'zypper search forwarding'
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	// not using flag.Parse() because we need to grab parsing error
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		// print usage text
		if err == flag.ErrHelp {
			fmt.Print(searchPackagesUsageText)
			os.Exit(0)
		}
		fmt.Printf("Could not parse the options: %v\n", err)
	}
	if bNoop || sNoop != "" {
		os.Exit(0)
	}

	if err := checkUnsupportedFlags(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	opts, err := connect.ReadFromConfiguration(connect.DefaultConfigPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// TODO(mssola): to be removed by the end of RR4.
	connect.CFG = opts

	results := searchInModules(flag.Args(), matchExact, caseSensitive)
	if !noLocalRepos {
		repoResults := searchInRepos(flag.Args(), matchExact, caseSensitive)
		results = append(results, repoResults...)
	}

	if len(results) == 0 && !xmlout {
		fmt.Print("No package found\n\n")
		os.Exit(0)
	}

	if xmlout {
		var xmldata struct {
			XMLName xml.Name       `xml:"packages"`
			Results []searchResult `xml:"package"`
		}
		xmldata.Results = results
		xmltext, _ := xml.MarshalIndent(xmldata, "", "  ")
		fmt.Print(xml.Header)
		fmt.Println(string(xmltext))
		fmt.Println()
		os.Exit(0)
	}

	fmt.Print("Following packages were found in following modules:\n\n")

	resultsTable := make([][]string, 0)
	if details {
		for _, r := range results {
			resultsTable = append(resultsTable, []string{r.FullName(), r.ModuleOrRepo(), r.ConnectCmd()})
		}
	} else {
		for _, r := range results {
			resultsTable = append(resultsTable, []string{r.Name, r.ProdNameOrPkgStatus(), r.ConnectCmd()})
		}
	}

	resultsTable = uniqueTable(resultsTable)

	header := []string{"Package", "Module or Repository", "SUSEConnect Activation Command"}
	if groupByModule {
		header = []string{"Module or Repository", "Package"}
		modules := make(map[string][]string, 0)
		for _, row := range resultsTable {
			modules[row[1]] = append(modules[row[1]], row[0])
		}
		resultsTable = nil
		for mod, packages := range modules {
			resultsTable = append(resultsTable, []string{mod, packages[0]})
			for _, p := range packages[1:] {
				resultsTable = append(resultsTable, []string{"", p})
			}
		}
	} else if sortByName {
		sort.Slice(resultsTable, func(i, j int) bool {
			return resultsTable[i][0] < resultsTable[j][0]
		})
	} else if sortByRepo {
		sort.Slice(resultsTable, func(i, j int) bool {
			return resultsTable[i][1] < resultsTable[j][1]
		})
	}

	printTable(header, resultsTable)

	fmt.Print("\nTo activate the respective module or product, use SUSEConnect --product.\nUse SUSEConnect --help for more details.\n\n")
}

func printTable(header []string, table [][]string) {
	// calculate column widths
	maxLengths := make([]int, len(header))
	for _, row := range table {
		for i, e := range row {
			s := len(e)
			if s > maxLengths[i] {
				maxLengths[i] = s
			}
		}
	}
	// generate separators and format strings
	separators := make([]string, len(header))
	formats := make([]string, len(header))
	for i, s := range maxLengths {
		separators[i] = strings.Repeat("-", s)
		formats[i] = fmt.Sprintf("%%-%ds ", s)
	}
	// add header and separator to results table
	table = append([][]string{header, separators}, table...)
	// print table using formats
	for _, row := range table {
		for i, e := range row {
			fmt.Printf(formats[i], e)
		}
		fmt.Println()
	}
}

func readRepoIndex(path string) ([]searchResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return []searchResult{}, err
	}
	defer f.Close()
	return parseRepoIndex(f), nil
}

func parseRepoIndex(r io.Reader) []searchResult {
	ret := make([]searchResult, 0)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}
		// extract release from version
		version := parts[1]
		release := ""
		vr := strings.Split(version, "-")
		if len(vr) > 1 {
			release = vr[len(vr)-1]
			version = strings.Join(vr[:len(vr)-1], "-")
		}
		sr := searchResult{
			Name:    parts[0],
			Version: version,
			Release: release,
			Arch:    parts[2],
		}
		ret = append(ret, sr)
	}
	return ret
}

func packageWanted(name, query string, matchExact, caseSensitive bool) bool {
	if !caseSensitive {
		name = strings.ToLower(name)
		query = strings.ToLower(query)
	}
	if matchExact {
		return name == query
	}
	return strings.Contains(name, query)
}

func searchInRepos(patterns []string, matchExact, caseSensitive bool) []searchResult {
	ret := make([]searchResult, 0)
	reposPath := "/var/cache/zypp/solv"
	repos, _ := os.ReadDir(reposPath)
	for _, repo := range repos {
		if !repo.IsDir() {
			continue
		}
		reponame := repo.Name()
		pkgStatus := "Available in repo " + reponame

		if reponame == "@System" {
			reponame = "Installed"
			pkgStatus = "Installed"
		}
		repoPackages, err := readRepoIndex(filepath.Join(reposPath, repo.Name(), "solv.idx"))
		if err != nil {
			fmt.Printf("Cannot read index for repository %v.\n", reponame)
		}
		for _, p := range repoPackages {
			p.Repo = reponame
			p.PkgStatus = pkgStatus
			for _, query := range patterns {
				if packageWanted(p.Name, query, matchExact, caseSensitive) {
					ret = append(ret, p)
					break
				}
			}
		}
	}
	return ret
}

func searchInModules(patterns []string, matchExact, caseSensitive bool) []searchResult {
	ret := make([]searchResult, 0)
	for _, query := range patterns {
		found, err := connect.SearchPackage(query, registration.Product{})
		if err != nil {
			fmt.Printf("Could not search for the package: %v", err)
		}
		for _, pkg := range found {
			if !packageWanted(pkg.Name, query, matchExact, caseSensitive) {
				continue
			}
			for _, p := range pkg.Products {
				sr := searchResult{
					Name:        pkg.Name,
					Version:     pkg.Version,
					Release:     pkg.Release,
					Arch:        pkg.Arch,
					ProdName:    fmt.Sprintf("%s (%s)", p.Name, p.Ident), // TODO: anything better? do we need raw p.Name?
					ProdIdent:   p.Ident,
					ProdEdition: p.Edition,
					ProdArch:    p.Arch,
					ProdFree:    p.Free,
				}
				ret = append(ret, sr)
			}
		}
	}
	return ret
}

func checkUnsupportedFlags() error {
	// flag -> reason details
	unsupported := map[string]string{
		"match-words":         "by whole words",
		"provides":            "by dependencies",
		"recommends":          "by dependencies",
		"requires":            "by dependencies",
		"suggests":            "by dependencies",
		"supplements":         "by dependencies",
		"conflicts":           "by dependencies",
		"obsoletes":           "by dependencies",
		"f":                   "in file list",
		"file-list":           "in file list",
		"d":                   "in summaries and descriptions",
		"search-descriptions": "in summaries and descriptions",
	}
	reasons := connect.NewStringSet()
	flag.Visit(func(f *flag.Flag) {
		if details, found := unsupported[f.Name]; found {
			reasons.Add(fmt.Sprintf("Extended search does not support search %s.", details))
			return
		}
		// special case: --type <TYPE> argument
		if f.Name == "type" && f.Value.String() != "package" {
			reasons.Add("Extended package search can only search for the resolvable type 'package'.")
		}
	})

	if reasons.Len() != 0 {
		return fmt.Errorf("Cannot perform extended package search:\n\n%v", strings.Join(reasons.Strings(), "\n"))
	}
	return nil
}

func uniqueTable(table [][]string) [][]string {
	ret := make([][]string, 0)
	// key is row cells joined with '|'
	present := make(map[string]struct{}, 0)
	for _, row := range table {
		key := strings.Join(row, "|")
		if _, found := present[key]; !found {
			present[key] = struct{}{}
			ret = append(ret, row)
		}
	}
	return ret
}
