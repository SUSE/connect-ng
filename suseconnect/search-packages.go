package main

import (
	_ "embed"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/SUSE/connect-ng/internal/connect"
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

func searchPackagesMain() {
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

	results := searchInModules(flag.Args(), matchExact, caseSensitive)
	// TODO
	// repo_results = search_pkgs_in_repos(options, params)
	// results.concat repo_results

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
	// TODO
	// results.uniq!

	header := []string{"Package", "Module or Repository", "SUSEConnect Activation Command"}
	if groupByModule {
		header = []string{"Module or Repository", "Package"}
		// TODO
		//   modules = {}
		//   results.each do | entry |
		//     modules[entry[1]] ||= [];
		//     modules[entry[1]].push entry[0];
		//   end
		//   results = []
		//   modules.each do | mod, packages |
		//     pkg = packages.shift
		//     results.push [ mod, pkg ]
		//     packages.each do | pkg |
		//       results.push [ "", pkg ]
		//     end
		//   end
	} else if sortByName {
		// TODO
		//   results.sort! { | a, b |
		//     a[0] <=> b[0]
		//   }
	} else if sortByRepo {
		// TODO
		//   results.sort! { | a, b |
		//     a[1] <=> b[1]
		//   }
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

func searchInModules(patterns []string, matchExact, caseSensitive bool) []searchResult {
	ret := make([]searchResult, 0)
	for _, query := range patterns {
		found, err := connect.SearchPackage(query, connect.Product{})
		if err != nil {
			fmt.Printf("Could not search for the package: %v", err)
		}
		for _, pkg := range found {
			// skip unwanted packages depending on the flags
			if matchExact {
				if caseSensitive && pkg.Name != query ||
					strings.ToLower(pkg.Name) != strings.ToLower(query) {
					continue
				}
			} else if caseSensitive && !strings.Contains(pkg.Name, query) {
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
