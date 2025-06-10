package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/SUSE/connect-ng/internal/connect"
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/internal/zypper"
	"github.com/SUSE/connect-ng/pkg/registration"
)

var (
	//go:embed connectUsage.txt
	connectUsageText string
)

const (
	outdatedRegProxy = "Your Registration Proxy server doesn't support this function."
)

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

func main() {
	// SUSEConnect only works on Linux.
	if err := util.EnsureLinux(); err != nil {
		exitOnError(err, nil)
	}

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
		labels                string
		product               singleStringFlag
		instanceDataFile      string
		listExtensions        bool
		autoImportRepoKeys    bool
		autoAgreeWithLicenses bool
		email                 string
		version               bool
		jsonFlag              bool
		info                  bool
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
	flag.BoolVar(&autoImportRepoKeys, "gpg-auto-import-keys", false, "")
	flag.BoolVar(&autoAgreeWithLicenses, "auto-agree-with-licenses", false, "")
	flag.StringVar(&baseURL, "url", "", "")
	flag.StringVar(&fsRoot, "root", "", "")
	flag.StringVar(&namespace, "namespace", "", "")
	flag.StringVar(&token, "regcode", "", "")
	flag.StringVar(&token, "r", "", "")
	flag.StringVar(&labels, "set-labels", "", "")
	flag.StringVar(&instanceDataFile, "instance-data", "", "")
	flag.StringVar(&email, "email", "", "")
	flag.StringVar(&email, "e", "", "")
	flag.Var(&product, "product", "")
	flag.Var(&product, "p", "")
	flag.BoolVar(&jsonFlag, "json", false, "")
	flag.BoolVar(&info, "info", false, "")
	flag.BoolVar(&info, "i", false, "")

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
		util.EnableDebug()
	}
	util.Debug.Println("cmd line:", os.Args)
	util.Debug.Println("For http debug use: GODEBUG=http2debug=2", strings.Join(os.Args, " "))

	// Fetch the options to be passed to the internal/connect library by reading
	// at the default configuration. If that default configuration is not there,
	// then it will simply default to scc.suse.com with no proxy in between.
	opts, err := connect.ReadFromConfiguration(connect.DefaultConfigPath)
	exitOnError(err, nil)

	if baseURL != "" {
		if err := validateURL(baseURL); err != nil {
			fmt.Printf("SUSEConnect error: URL \"%s\" not valid: %s\n", baseURL, err)
			os.Exit(1)
		}
		opts.ChangeBaseURL(baseURL)
		writeConfig = true
	}
	if fsRoot != "" {
		if fsRoot[0] != '/' {
			fmt.Println("SUSEConnect error: the path specified in the --root option must be absolute.")
			os.Exit(1)
		}
		opts.FsRoot = fsRoot
		zypper.SetFilesystemRoot(fsRoot)
	}
	if namespace != "" {
		opts.Namespace = namespace
		writeConfig = true
	}
	if token != "" {
		opts.Token = token
		processedToken, processTokenErr := processToken(token)
		if processTokenErr != nil {
			fmt.Printf("SUSEConnect error: %v", processTokenErr)
			os.Exit(1)
		}
		opts.Token = processedToken
	}
	if product.isSet {
		if p, err := registration.FromTriplet(product.value); err != nil {
			fmt.Print("Please provide the product identifier in this format: ")
			fmt.Print("<internal name>/<version>/<architecture>. You can find ")
			fmt.Print("these values by calling: 'SUSEConnect --list-extensions'\n")
			os.Exit(1)
		} else {
			opts.Product = p
		}
	}
	if instanceDataFile != "" {
		opts.InstanceDataFile = instanceDataFile
	}
	if email != "" {
		opts.Email = email
	}
	if lang, ok := os.LookupEnv("LANG"); ok {
		if lang != "" {
			opts.Language = lang
		}
	}
	if _, ok := os.LookupEnv("SKIP_SERVICE_INSTALL"); ok {
		opts.SkipServiceInstall = true
	}
	if autoImportRepoKeys {
		opts.AutoImportRepoKeys = true
	}
	if autoAgreeWithLicenses {
		opts.AutoAgreeEULA = true
	} else {
		// check for "SUSEConnect --auto-agree-with-licenses=false ..."
		// which should take precedence over setting in /etc/SUSEConnect
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "auto-agree-with-licenses" {
				opts.AutoAgreeEULA = false
			}
		})
	}

	// Set the JSON output specifier early on so all later usages can access this information
	if jsonFlag {
		opts.OutputKind = connect.JSON
	}

	// Reading the configuration/flags is done, now let's check if the
	// filesystem can handle operations from SUSEConnect for specific actions
	// which require filesystem to be read write (aka writing outside of /etc)
	// /etc is writable at any time, system token roation works just fine.
	//
	// Rollback *must* be allowed because is used as a synchonization mechanism
	// in the transactional-update toolkit.
	if deRegister || cleanup {
		if err := util.ReadOnlyFilesystem(opts.FsRoot); err != nil {
			exitOnError(err, opts)
		}
	}

	// TODO(mssola): to be removed by the end of RR4.
	connect.CFG = opts

	if status || statusText {
		if jsonFlag {
			exitOnError(errors.New("cannot use the json option with this command"), opts)
		}
		if status {
			err = connect.PrintProductStatuses(opts, connect.StatusJSON)
		} else {
			err = connect.PrintProductStatuses(opts, connect.StatusText)
		}
		exitOnError(err, opts)
	} else if keepAlive {
		if isSumaManaged() {
			os.Exit(0)
		}
		if jsonFlag {
			exitOnError(errors.New("cannot use the json option with the 'keepalive' command"), opts)
		}
		api := connect.NewWrappedAPI(opts)
		err = api.KeepAlive()
		exitOnError(err, opts)
		util.Info.Print(util.Bold(util.GreenText("\nSuccessfully updated system")))
	} else if listExtensions {
		output, err := connect.RenderExtensionTree(opts, jsonFlag)
		exitOnError(err, opts)
		fmt.Println(output)
		os.Exit(0)
	} else if deRegister {
		err := connect.Deregister(opts)
		if jsonFlag && err != nil {
			out := connect.RegisterOut{Success: false, Message: err.Error()}
			str, _ := json.Marshal(&out)
			fmt.Println(string(str))
			os.Exit(1)
		} else {
			exitOnError(err, opts)
		}
	} else if cleanup {
		if jsonFlag {
			exitOnError(errors.New("cannot use the json option with the 'cleanup' command"), opts)
		}
		err := connect.Cleanup(opts.BaseURL, opts.FsRoot)
		exitOnError(err, opts)
	} else if rollback {
		if jsonFlag {
			exitOnError(errors.New("cannot use the json option with the 'rollback' command"), opts)
		}
		err := connect.Rollback(opts)
		exitOnError(err, opts)
	} else if info {
		sysInfo, err := connect.FetchSystemInformation()
		exitOnError(err, opts)

		out, err := json.Marshal(sysInfo)
		exitOnError(err, opts)

		fmt.Print(string(out))
	} else {
		if instanceDataFile != "" && opts.IsScc() {
			fmt.Print("Please use --instance-data only in combination ")
			fmt.Print("with --url pointing to your RMT or SMT server\n")
			os.Exit(1)
		} else if opts.IsScc() && token == "" && product.value == "" {
			flag.Usage()
			os.Exit(1)
		} else if isSumaManaged() {
			fmt.Println("This system is managed by SUSE Manager / Uyuni, do not use SUSEconnect.")
			os.Exit(1)
		} else {
			// NOTE: if the base system/extensions have EULAs we need to make
			// sure that they are accepted before proceeding on the registering.
			// But this has been disabled for now due to bsc#1218878,
			// bsc#1218649.

			// We need a read-write filesystem to install release packages.
			if err := util.ReadOnlyFilesystem(opts.FsRoot); err != nil {
				exitOnError(err, opts)
			}

			err := connect.Register(opts)
			if err != nil {
				if jsonFlag {
					out := connect.RegisterOut{Success: false, Message: err.Error()}
					str, _ := json.Marshal(&out)
					fmt.Println(string(str))
					os.Exit(1)
				} else {
					exitOnError(err, opts)
				}
			}

			// After successful registration we try to set labels if we are
			// targetting SCC.
			//
			// TODO(mssola): to be removed once we sort out the token callback
			// for the `internal/connect` library.
			if connect.CFG.IsScc() && len(labels) > 0 {
				wrapper := connect.NewWrappedAPI(opts)
				_, err := wrapper.AssignLabels(strings.Split(labels, ","))
				if err != nil && !jsonFlag {
					fmt.Printf("Problem setting labels for this system: %s\n", err)
				}
			}
		}
	}
	if writeConfig {
		if err := opts.SaveAsConfiguration(); err != nil {
			fmt.Printf("SUSEConnect error: cannot save configuration: %s\n", err)
			os.Exit(1)
		}
	}
}

func exitOnError(err error, opts *connect.Options) {
	if err == nil {
		return
	}
	if ze, ok := err.(zypper.ZypperError); ok {
		fmt.Println(ze)
		os.Exit(ze.ExitCode)
	}
	if ue, ok := err.(*url.Error); ok && errors.Is(ue, syscall.ECONNREFUSED) {
		fmt.Println("Error:", err)
		os.Exit(64)
	}
	if je, ok := err.(connect.JSONError); ok {
		if connect.IsOutdatedRegProxy(opts) {
			fmt.Println(outdatedRegProxy)
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
		} else if connect.IsOutdatedRegProxy(opts) {
			fmt.Println(outdatedRegProxy)
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
		return fmt.Errorf("missing scheme or host")
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

func processToken(token string) (string, error) {
	if strings.HasPrefix(token, "@") {
		tokenFilePath := strings.TrimPrefix(token, "@")
		file, err := os.Open(tokenFilePath)
		if err != nil {
			return "", fmt.Errorf("failed to open token file '%s': %w", tokenFilePath, err)
		}
		defer file.Close()
		return readTokenFromReader(file)
	} else if token == "-" {
		return readTokenFromReader(os.Stdin)
	} else {
		return token, nil
	}
}

func readTokenFromReader(reader io.Reader) (string, error) {
	bufReader := bufio.NewReader(reader)
	tokenBytes, err := bufReader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read token from reader: %w", err)
	}
	token := strings.TrimSpace(tokenBytes)
	if token == "" {
		return "", fmt.Errorf("token cannot be empty after reading")
	}
	return token, nil
}
