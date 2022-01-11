package main

import (
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/SUSE/connect-ng/internal/connect"
)

var (
	//go:embed connectUsage.txt
	connectUsageText string
)

// multi-call entry points
func main() {
	switch filepath.Base(os.Args[0]) {
	case "zypper-migration":
		migrationMain()
	case "zypper-search-packages":
		searchPackagesMain()
	default:
		connectMain()
	}
}

func connectMain() {
	var (
		status           bool
		statusText       bool
		debug            bool
		writeConfig      bool
		deRegister       bool
		cleanup          bool
		rollback         bool
		baseURL          string
		fsRoot           string
		namespace        string
		token            string
		product          string
		instanceDataFile string
		listExtensions   bool
		email            string
		version          bool
	)

	// display help like the ruby SUSEConnect
	flag.Usage = func() {
		fmt.Print(connectUsageText)
	}

	flag.BoolVar(&status, "status", false, "")
	flag.BoolVar(&status, "s", false, "")
	flag.BoolVar(&statusText, "status-text", false, "")
	flag.BoolVar(&debug, "debug", false, "")
	flag.BoolVar(&writeConfig, "write-config", false, "")
	flag.BoolVar(&deRegister, "de-register", false, "")
	flag.BoolVar(&deRegister, "d", false, "")
	flag.BoolVar(&cleanup, "cleanup", false, "")
	flag.BoolVar(&listExtensions, "list-extensions", false, "")
	flag.BoolVar(&rollback, "rollback", false, "")
	flag.BoolVar(&version, "version", false, "")
	flag.BoolVar(&connect.CFG.AutoImportRepoKeys, "gpg-auto-import-keys", false, "")
	flag.StringVar(&baseURL, "url", "", "")
	flag.StringVar(&fsRoot, "root", "", "")
	flag.StringVar(&namespace, "namespace", "", "")
	flag.StringVar(&token, "regcode", "", "")
	flag.StringVar(&token, "r", "", "")
	flag.StringVar(&product, "product", "", "")
	flag.StringVar(&product, "p", "", "")
	flag.StringVar(&instanceDataFile, "instance-data", "", "")
	flag.StringVar(&email, "email", "", "")
	flag.StringVar(&email, "e", "", "")

	flag.Parse()
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Root privileges are required to register products and change software repositories.")
		os.Exit(1)
	}
	if debug {
		connect.EnableDebug()
	}
	connect.Debug.Println("cmd line:", os.Args)
	connect.Debug.Println("For http debug use: GODEBUG=http2debug=2", strings.Join(os.Args, " "))
	connect.CFG.Load()
	if baseURL != "" {
		if err := validateURL(baseURL); err != nil {
			fmt.Printf("URL \"%s\" not valid: %s\n", baseURL, err)
			os.Exit(1)
		}
		connect.CFG.BaseURL = baseURL
		writeConfig = true
	}
	if fsRoot != "" {
		if fsRoot[0] != '/' {
			fmt.Println("The path specified in the --root option must be absolute.")
			os.Exit(1)
		}
		connect.CFG.FsRoot = fsRoot
	}
	if namespace != "" {
		connect.CFG.Namespace = namespace
		writeConfig = true
	}
	if token != "" {
		connect.CFG.Token = token
	}
	if product != "" {
		if p, err := connect.SplitTriplet(product); err != nil {
			fmt.Print("Please provide the product identifier in this format: ")
			fmt.Print("<internal name>/<version>/<architecture>. You can find ")
			fmt.Print("these values by calling: 'SUSEConnect --list-extensions'\n")
			os.Exit(1)
		} else {
			connect.CFG.Product = p
		}
	}
	if instanceDataFile != "" {
		connect.CFG.InstanceDataFile = instanceDataFile
	}
	if email != "" {
		connect.CFG.Email = email
	}
	if lang, ok := os.LookupEnv("LANG"); ok {
		if lang != "" {
			connect.CFG.Language = lang
		}
	}
	if status {
		output, err := connect.GetProductStatuses("json")
		exitOnError(err)
		fmt.Println(output)
	} else if statusText {
		output, err := connect.GetProductStatuses("text")
		exitOnError(err)
		fmt.Print(output)
	} else if listExtensions {
		output, err := connect.GetExtensionsList()
		exitOnError(err)
		fmt.Print(output)
	} else if deRegister {
		err := connect.Deregister()
		exitOnError(err)
	} else if cleanup {
		err := connect.Cleanup()
		exitOnError(err)
	} else if rollback {
		err := connect.Rollback()
		exitOnError(err)
	} else if version {
		fmt.Println(connect.GetShortenedVersion())
		os.Exit(0)
	} else {
		if instanceDataFile != "" && connect.URLDefault() {
			fmt.Print("Please use --instance-data only in combination ")
			fmt.Print("with --url pointing to your RMT or SMT server\n")
			os.Exit(1)
		} else if connect.URLDefault() && token == "" && product == "" {
			flag.Usage()
			os.Exit(1)
		} else if fileExists("/etc/sysconfig/rhn/systemid") {
			fmt.Println("This system is managed by SUSE Manager / Uyuni, do not use SUSEconnect.")
			os.Exit(1)
		} else {
			err := connect.Register()
			exitOnError(err)
		}
	}
	if writeConfig {
		if err := connect.CFG.Save(); err != nil {
			fmt.Printf("Problem writing configuration: %s\n", err)
			os.Exit(1)
		}
	}
}

func maybeBrokenSMTError() error {
	if !connect.URLDefault() && !connect.UpToDate() {
		return fmt.Errorf("Your Registration Proxy server doesn't support this function. " +
			"Please update it and try again.")
	}
	return nil
}

func exitOnError(err error) {
	if err == nil {
		return
	}
	if ze, ok := err.(connect.ZypperError); ok {
		fmt.Println(ze)
		os.Exit(ze.ExitCode)
	}
	if je, ok := err.(connect.JSONError); ok {
		if err := maybeBrokenSMTError(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Print("Error: Cannot parse response from server\n")
			fmt.Println(je)
		}
		os.Exit(66)
	}
	if ae, ok := err.(connect.APIError); ok {
		if ae.Code == http.StatusUnauthorized && connect.IsRegistered() {
			fmt.Print("Error: Invalid system credentials, probably because the ")
			fmt.Print("registered system was deleted in SUSE Customer Center. ")
			fmt.Print("Check ", connect.CFG.BaseURL, " whether your system appears there. ")
			fmt.Print("If it does not, please call SUSEConnect --cleanup and re-register this system.\n")
		} else if err := maybeBrokenSMTError(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(ae)
		}
		os.Exit(67)
	}
	switch err {
	case connect.ErrSystemNotRegistered:
		fmt.Print("Deregistration failed. Check if the system has been ")
		fmt.Print("registered using the --status-text option or use the ")
		fmt.Print("--regcode parameter to register it.\n")
		os.Exit(69)
	case connect.ErrListExtensionsUnregistered:
		fmt.Print("To list extensions, you must first register the base product, ")
		fmt.Print("using: SUSEConnect -r <registration code>\n")
		os.Exit(1)
	case connect.ErrBaseProductDeactivation:
		fmt.Print("Can not deregister base product. Use SUSEConnect -d to deactivate ")
		fmt.Print("the whole system.\n")
		os.Exit(70)
	default:
		fmt.Printf("SUSEConnect error: %s\n", err)
		os.Exit(1)
	}
}

func validateURL(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("Missing scheme or host")
	}
	return nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
