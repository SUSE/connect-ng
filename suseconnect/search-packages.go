package main

import (
	_ "embed"
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

	fmt.Println("TODO: actual search")
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
