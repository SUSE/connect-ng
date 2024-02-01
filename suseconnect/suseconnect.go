package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"

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

// singleStringFlag cannot be set more than once.
// e.g. `cmd -p abc -p def` will give a parse error.
type singleStringFlag struct {
	value string
	isSet bool
}

func (p *singleStringFlag) String() string {
	return p.value
}

func (p *singleStringFlag) Set(value string) error {
	if p.isSet {
		return fmt.Errorf("this flag can only be specified once\n")
	}
	p.value = value
	p.isSet = true
	return nil
}

func connectMain() {
	var (
		status                bool
		keepAlive             bool
		statusText            bool
		debug                 bool
		writeConfig           bool
		deRegister            bool
		cleanup               bool
		rollback              bool
		baseURL               string
		fsRoot                string
		namespace             string
		token                 string
		product               singleStringFlag
		instanceDataFile      string
		listExtensions        bool
		autoAgreeWithLicenses bool
		email                 string
		version               bool
		jsonFlag              bool
	)

	// display help like the ruby SUSEConnect
	flag.Usage = func() {
		fmt.Print(connectUsageText)
	}

	flag.BoolVar(&status, "status", false, "")
	flag.BoolVar(&status, "s", false, "")
	flag.BoolVar(&statusText, "status-text", false, "")
	flag.BoolVar(&keepAlive, "keepalive", false, "")
	flag.BoolVar(&debug, "debug", false, "")
	flag.BoolVar(&writeConfig, "write-config", false, "")
	flag.BoolVar(&deRegister, "de-register", false, "")
	flag.BoolVar(&deRegister, "d", false, "")
	flag.BoolVar(&cleanup, "cleanup", false, "")
	flag.BoolVar(&cleanup, "clean", false, "")
	flag.BoolVar(&listExtensions, "list-extensions", false, "")
	flag.BoolVar(&listExtensions, "l", false, "")
	flag.BoolVar(&rollback, "rollback", false, "")
	flag.BoolVar(&version, "version", false, "")
	flag.BoolVar(&connect.CFG.AutoImportRepoKeys, "gpg-auto-import-keys", false, "")
	flag.BoolVar(&autoAgreeWithLicenses, "auto-agree-with-licenses", false, "")
	flag.StringVar(&baseURL, "url", "", "")
	flag.StringVar(&fsRoot, "root", "", "")
	flag.StringVar(&namespace, "namespace", "", "")
	flag.StringVar(&token, "regcode", "", "")
	flag.StringVar(&token, "r", "", "")
	flag.StringVar(&instanceDataFile, "instance-data", "", "")
	flag.StringVar(&email, "email", "", "")
	flag.StringVar(&email, "e", "", "")
	flag.Var(&product, "product", "")
	flag.Var(&product, "p", "")
	flag.BoolVar(&jsonFlag, "json", false, "")

	flag.Parse()
	if version {
		fmt.Println(connect.GetShortenedVersion())
		os.Exit(0)
	}
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
	if product.isSet {
		if p, err := connect.SplitTriplet(product.value); err != nil {
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
	if _, ok := os.LookupEnv("SKIP_SERVICE_INSTALL"); ok {
		connect.CFG.SkipServiceInstall = true
	}
	if autoAgreeWithLicenses {
		connect.CFG.AutoAgreeEULA = true
	} else {
		// check for "SUSEConnect --auto-agree-with-licenses=false ..."
		// which should take precedence over setting in /etc/SUSEConnect
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "auto-agree-with-licenses" {
				connect.CFG.AutoAgreeEULA = false
			}
		})
	}

	// Reading the configuration/flags is done, now let's check if the
	// filesystem can handle operations from SUSEConnect.
	if err := connect.ReadOnlyFilesystem(connect.CFG.FsRoot); err != nil {
		exitOnError(err)
	}

	if status {
		if jsonFlag {
			exitOnError(errors.New("cannot use the json option with the 'status' command"))
		}
		output, err := connect.GetProductStatuses("json")
		exitOnError(err)
		fmt.Println(output)
	} else if keepAlive {
		if isSumaManaged() {
			os.Exit(0)
		}
		if jsonFlag {
			exitOnError(errors.New("cannot use the json option with the 'keepalive' command"))
		}
		err := connect.SendKeepAlivePing()
		exitOnError(err)
	} else if statusText {
		if jsonFlag {
			exitOnError(errors.New("cannot use the json option with the 'status-text' command"))
		}
		output, err := connect.GetProductStatuses("text")
		exitOnError(err)
		fmt.Print(output)
	} else if listExtensions {
		output, err := connect.RenderExtensionTree(jsonFlag)
		exitOnError(err)
		fmt.Println(output)
		os.Exit(0)
	} else if deRegister {
		err := connect.Deregister(jsonFlag)
		if jsonFlag && err != nil {
			out := connect.RegisterOut{Success: false, Message: err.Error()}
			str, _ := json.Marshal(&out)
			fmt.Println(string(str))
			os.Exit(1)
		} else {
			exitOnError(err)
		}
	} else if cleanup {
		if jsonFlag {
			exitOnError(errors.New("cannot use the json option with the 'cleanup' command"))
		}
		err := connect.Cleanup()
		exitOnError(err)
	} else if rollback {
		if jsonFlag {
			exitOnError(errors.New("cannot use the json option with the 'rollback' command"))
		}
		err := connect.Rollback()
		exitOnError(err)
	} else {
		if instanceDataFile != "" && connect.URLDefault() {
			fmt.Print("Please use --instance-data only in combination ")
			fmt.Print("with --url pointing to your RMT or SMT server\n")
			os.Exit(1)
		} else if connect.URLDefault() && token == "" && product.value == "" {
			flag.Usage()
			os.Exit(1)
		} else if isSumaManaged() {
			fmt.Println("This system is managed by SUSE Manager / Uyuni, do not use SUSEconnect.")
			os.Exit(1)
		} else {

			// If the base system/extensions have EULAs, we need to make sure
			// that they are accepted before proceeding on the registering. If
			// they don't have EULA's, then this is a no-op.

			// disabling the license dialog feature for now due to bsc#1218878, bsc#1218649
			// err := connect.AcceptEULA()
			// exitOnError(err)

			err := connect.Register(jsonFlag)
			if jsonFlag && err != nil {
				out := connect.RegisterOut{Success: false, Message: err.Error()}
				str, _ := json.Marshal(&out)
				fmt.Println(string(str))
				os.Exit(1)
			} else {
				exitOnError(err)
			}
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
	if ue, ok := err.(*url.Error); ok && errors.Is(ue, syscall.ECONNREFUSED) {
		fmt.Println("Error:", err)
		os.Exit(64)
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
	case connect.ErrPingFromUnregistered:
		fmt.Print("Error sending keepalive: ")
		fmt.Print("System is not registered. Use the --regcode parameter to register it.\n")
		os.Exit(71)
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

func isSumaManaged() bool {
	return fileExists("/etc/sysconfig/rhn/systemid")
}
