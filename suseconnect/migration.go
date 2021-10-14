package main

import (
	"bufio"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/SUSE/connect-ng/internal/connect"
)

var (
	//go:embed migrationUsage.txt
	migrationUsageText string
	// flag indicating interuption by INT/TERM signal
	interrupted bool

	// ErrInterrupted is returned when execution was interrupted by INT/TERM signal
	ErrInterrupted = errors.New("Interrupted")
)

// logger shortcuts
var (
	Debug      *log.Logger = connect.Debug
	QuietOut   *log.Logger = connect.QuietOut
	VerboseOut             = log.New(io.Discard, "", 0)
)

// implements flag.Value interface used to hold values of args which could be
// passed multiple times e.g. $ command -x 1 -x 2 -x 3
type multiArg []string

func (a *multiArg) String() string {
	return strings.Join(*a, "|")
}
func (a *multiArg) Set(v string) error {
	*a = append(*a, v)
	return nil
}

func migrationMain() {
	var (
		debug          bool
		verbose        bool
		quiet          bool
		nonInteractive bool
		noSnapshots    bool
		noSelfUpdate   bool
		breakMySystem  bool
		query          bool
		disableRepos   bool
		migrationNum   int
		fsRoot         string
		toProduct      string
		from           multiArg
		repo           multiArg
		download       multiArg // using multiArg here to make flags simpler to visit
	)

	flag.Usage = func() {
		fmt.Print(migrationUsageText)
	}
	// flags without variables match defaults
	flag.BoolVar(&debug, "debug", false, "")
	flag.Bool("no-verbose", false, "")
	flag.BoolVar(&verbose, "verbose", false, "")
	flag.BoolVar(&verbose, "v", false, "")
	flag.Bool("no-quiet", false, "")
	flag.BoolVar(&quiet, "quiet", false, "")
	flag.BoolVar(&quiet, "q", false, "")
	flag.BoolVar(&nonInteractive, "non-interactive", false, "")
	flag.BoolVar(&nonInteractive, "n", false, "")
	flag.BoolVar(&noSnapshots, "no-snapshots", false, "")
	flag.Bool("selfupdate", false, "")
	flag.BoolVar(&noSelfUpdate, "no-selfupdate", false, "")
	flag.BoolVar(&breakMySystem, "break-my-system", false, "")
	flag.BoolVar(&query, "query", false, "")
	flag.BoolVar(&disableRepos, "disable-repos", false, "")
	flag.IntVar(&migrationNum, "migration", 0, "")
	flag.StringVar(&fsRoot, "root", "", "")
	flag.StringVar(&toProduct, "product", "", "")
	// zypper dup passthrough args
	// bool flags don't need variables as these will be processed using flag.Visit()
	flag.Bool("auto-agree-with-licenses", false, "")
	flag.Bool("l", false, "")
	flag.Bool("allow-vendor-change", false, "")
	flag.Bool("no-allow-vendor-change", false, "")
	flag.Bool("debug-solver", false, "")
	flag.Bool("recommends", false, "")
	flag.Bool("no-recommends", false, "")
	flag.Bool("replacefiles", false, "")
	flag.Bool("details", false, "")
	flag.Bool("download-only", false, "")
	flag.Var(&download, "download", "")
	flag.Var(&from, "from", "")
	flag.Var(&repo, "r", "")
	flag.Var(&repo, "repo", "")

	flag.Parse()
	if err := checkFlagContradictions(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// this is only to keep the flag parsing logic simple and avoid double
	// negations/negatives below
	selfUpdate := !noSelfUpdate

	if verbose {
		VerboseOut.SetOutput(os.Stdout)
	}

	if debug {
		connect.EnableDebug()
	}

	if !quiet {
		QuietOut.SetOutput(os.Stdout)
	}

	connect.CFG.Load()

	// pass root to connect config
	if fsRoot != "" {
		connect.CFG.FsRoot = fsRoot
		// if we update a chroot system, we cannot create snapshots of it
		noSnapshots = true
	}

	if !connect.IsSnapperConfigured() {
		noSnapshots = true
		VerboseOut.Println("Snapper not configured")
	}

	if toProduct != "" && fsRoot == "" && !breakMySystem {
		fmt.Println("The --product option can only be used together with the --root option")
		os.Exit(1)
	}

	if selfUpdate {
		// reset root (if set) as the update stack can be outside of
		// the to be updated system
		connect.CFG.FsRoot = ""
		echo := connect.SetSystemEcho(true)
		if pending, err := connect.PatchCheck(true, quiet, verbose, nonInteractive, false); err != nil {
			fmt.Printf("patch pre-check failed: %v\n", err)
			os.Exit(1)
		} else if pending {
			// install pending updates and restart
			if err := connect.Patch(true, quiet, verbose, nonInteractive, true); err != nil {
				fmt.Printf("patch failed: %v\n", err)
				os.Exit(1)
			}
			// stop infinite restarting
			// check that the patches were really installed
			if pending, err := connect.PatchCheck(true, true, false, true, true); pending || err != nil {
				if pending {
					fmt.Println("there are still some patches pending")
				}
				if err != nil {
					fmt.Printf("patch check returned error: %v\n", err)
				}
				fmt.Println("patch failed, exiting.")
				os.Exit(1)
			}
			QuietOut.Print("\nRestarting the migration script...\n")
			// this should replace current process with a new one but stop on error
			// just in case
			if err := syscall.Exec(os.Args[0], os.Args, []string{}); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		connect.SetSystemEcho(echo)
		// restore root if needed
		if fsRoot != "" {
			connect.CFG.FsRoot = fsRoot
		}
	}
	QuietOut.Print("\n")

	// This is only necessary, if we run with --root option
	echo := connect.SetSystemEcho(true)
	if err := connect.RefreshRepos("", false, quiet, verbose, nonInteractive); err != nil {
		fmt.Println("repository refresh failed, exiting")
		os.Exit(1)
	}
	connect.SetSystemEcho(echo)

	systemProducts, err := checkSystemProducts(true)
	if err != nil {
		fmt.Printf("Can't determine the list of installed products: %v\n", err)
		os.Exit(1)
	}

	printProducts(systemProducts)

	if len(systemProducts) == 0 {
		fmt.Println("No products found, migration is not possible.")
		os.Exit(1)
	}

	// This is not fully correct name as installedIDs could also contain some products
	// which are activated but not installed. This matches the original implementation.
	installedIDs := connect.NewStringSet()
	for _, prod := range systemProducts {
		installedIDs.Add(prod.ToTriplet())
	}

	allMigrations, err := fetchAllMigrations(systemProducts, toProduct)
	if err != nil {
		fmt.Printf("Can't get available migrations from server: %v\n", err)
		os.Exit(1)
	}

	// preprocess the migrations lists
	migrations := make([]connect.MigrationPath, 0)
	unavailableMigrations := make([]connect.MigrationPath, 0)
	for _, m := range allMigrations {
		mAvailable := true
		for _, p := range m {
			mAvailable = mAvailable && p.Available
		}

		sortMigrationProducts(m, installedIDs)

		if mAvailable {
			migrations = append(migrations, m)
		} else {
			unavailableMigrations = append(unavailableMigrations, m)
		}
	}

	if len(unavailableMigrations) > 0 && !quiet {
		printMigrations(unavailableMigrations,
			installedIDs,
			"Unavailable migrations (product is not mirrored):",
			false)
	}

	if len(migrations) == 0 {
		QuietOut.Print("No migration available.\n\n")
		if len(unavailableMigrations) > 0 {
			// no need to print a msg - unavailable migrations are listed above
			os.Exit(1)
		}
		os.Exit(0)
	}

	if nonInteractive && migrationNum == 0 {
		// select the first option
		migrationNum = 1
	}

	// this part is only used in interactive mode
	for migrationNum <= 0 || migrationNum > len(migrations) {
		printMigrations(migrations, installedIDs, "Available migrations:", true)
		if query {
			os.Exit(0)
		}
		fmt.Print("[num/q]: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			QuietOut.Print("\nStandard input seems to be closed, please use '--non-interactive' option\n")
			os.Exit(1)
		}
		choice := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if choice == "q" {
			os.Exit(0)
		}
		if n, err := strconv.Atoi(choice); err == nil {
			migrationNum = n
		}
	}

	migration := migrations[migrationNum-1]

	var preSnapshotNum int
	if !noSnapshots {
		snapshotNum, err := connect.CreatePreSnapshot()
		if err != nil {
			// NOTE: original version ignored all errors here.
			// Snapshot number was usually left at 0 in those cases.
			fmt.Printf("Snapshot creation failed: %v", err)
		}
		preSnapshotNum = snapshotNum
	}

	// do not create extra snapshots (bsc#947270)
	os.Setenv("DISABLE_SNAPPER_ZYPP_PLUGIN", "1")

	// allow interrupt only at specified points
	// we have to check zypper exitstatus == 8 even after interrupt
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		Debug.Printf("Signal received: %v", sig)
		interrupted = true
	}()

	dupArgs := zypperDupArgs()
	fsInconsistent, err := applyMigration(migration, systemProducts,
		quiet, verbose, nonInteractive, disableRepos, dupArgs)

	if err != nil {
		fmt.Println(err)
		QuietOut.Print("\nMigration failed.\n\n")
	}

	if fsInconsistent {
		fmt.Println("The migration to the new service pack has failed. The system is most")
		fmt.Println("likely in an inconsistent state.")
		fmt.Print("\n")
		fmt.Println("We strongly recommend to rollback to a snapshot created before the")
		fmt.Println("migration was started (via selecting the snapshot in the boot menu")
		fmt.Println("if you use snapper) or restore the system from a backup.")
		os.Exit(2)
	}

	if !noSnapshots && preSnapshotNum > 0 {
		_, err := connect.CreatePostSnapshot(preSnapshotNum)
		if err != nil {
			// NOTE: original version ignored all errors here.
			fmt.Printf("Snapshot creation failed: %v", err)
		}
		// NOTE: original code contains disabled part of code titled:
		// "Filesystem rollback - considered too dangerous" here
		// it used `snapper undochange` to restore system to previous state
		// on filesystem level.
	}

	// make sure all release packages are installed (bsc#1171652)
	if err == nil {
		_, err := checkSystemProducts(false)
		if err != nil {
			fmt.Printf("Can't determine the list of installed products after migration: %v\n", err)
			// the system has been sucessfully upgraded, zypper reported no error so
			// the only way to get here is a scc problem - it is better to just exit
			os.Exit(1)
		}
	}

	if err != nil {
		QuietOut.Print("\nPerforming repository rollback...\n")

		// restore repo configuration from backup file
		if err := connect.ZypperRestore(); err != nil {
			// NOTE: original ignores failures of this command
			fmt.Printf("Zypper restore failed: %v\n", err)
		}

		if err := connect.Rollback(); err == nil {
			QuietOut.Println("Rollback successful.")
		} else {
			fmt.Printf("Rollback failed: %v\n", err)
		}
		os.Exit(1)
	}
}

func checkSystemProducts(rollbackOnFailure bool) ([]connect.Product, error) {
	systemProducts, err := connect.SystemProducts()
	if err != nil {
		return systemProducts, err
	}

	releasePackageMissing := false
	for _, p := range systemProducts {
		// if a release package for registered product is missing -> try install it
		err := connect.InstallReleasePackage(p.Name)
		if err != nil {
			releasePackageMissing = true
			QuietOut.Printf("Can't install release package for registered product %s\n", p.Name)
			QuietOut.Printf("%v\n", err)
		}
	}

	if releasePackageMissing && rollbackOnFailure {
		// some release packages are missing and can't be installed
		QuietOut.Println("Calling SUSEConnect rollback to make sure SCC is synchronized with the system state.")
		if err := connect.Rollback(); err != nil {
			return systemProducts, err
		}
		// re-read the list of products
		systemProducts, err := connect.SystemProducts()
		if err != nil {
			return systemProducts, err
		}
	}

	return systemProducts, nil
}

func printProducts(products []connect.Product) {
	VerboseOut.Println("Installed products:")
	for _, p := range products {
		VerboseOut.Printf("  %-25s %s\n", p.ToTriplet(), p.Summary)
	}
	VerboseOut.Print("\n")
}

func printMigrations(migrations []connect.MigrationPath,
	installedIDs connect.StringSet,
	header string,
	withIndex bool) {
	fmt.Printf("%s\n\n", header)
	for idx, m := range migrations {
		for pidx, p := range m {
			prefix := "       "
			suffix := ""
			// print index only in first product row
			if withIndex && pidx == 0 {
				prefix = fmt.Sprintf("   %2d |", idx+1)
			}
			if !p.Available {
				suffix = suffix + " (not available)"
			}
			if installedIDs.Contains(p.ToTriplet()) {
				suffix = suffix + " (already installed)"
			}
			fmt.Printf("%s%s%s\n", prefix, p.FriendlyName, suffix)
		}
		fmt.Print("\n")
	}
	fmt.Print("\n")
}

// sort migrations to put already installed products last and base products first
func sortMigrationProducts(m connect.MigrationPath, installedIDs connect.StringSet) {
	sort.SliceStable(m, func(i, j int) bool {
		// first check installation status
		firstInstalled := installedIDs.Contains(m[i].ToTriplet())
		secondInstalled := installedIDs.Contains(m[j].ToTriplet())
		if firstInstalled != secondInstalled {
			return !firstInstalled
		}
		// if installation status is the same, check 'base' field
		firstBase := m[i].IsBase
		secondBase := m[j].IsBase
		return firstBase && !secondBase
	})
}

// three-way comparison of editions (EDITION=VERSION[-RELEASE])
// release part is ignored
func compareEditions(left, right string) int {
	// cut off (optional) release part
	leftParts := strings.Split(left, "-")
	rightParts := strings.Split(right, "-")
	// split version into parts
	leftParts = strings.Split(leftParts[0], ".")
	rightParts = strings.Split(rightParts[0], ".")

	// right-pad parts with zeros to match length
	for len(leftParts) < len(rightParts) {
		leftParts = append(leftParts, "0")
	}
	for len(rightParts) < len(leftParts) {
		rightParts = append(rightParts, "0")
	}

	// lenghts are equal so we can use one index
	for i := range leftParts {
		var l, r int
		// take and convert i-th part of left and right
		// NOTE: fmt.Sscan() is used over strconv.Atoi() to better match ruby behavior
		// for strings with non-digit characters like "123abc" or "123.456".
		fmt.Sscan(leftParts[i], &l)
		fmt.Sscan(rightParts[i], &r)

		if l < r {
			return -1
		}
		if l > r {
			return 1
		}
	}
	return 0
}

func cleanupProductRepos(p connect.Product, force bool) error {
	productPackages, err := connect.FindProductPackages(p.Name)
	if err != nil {
		return err
	}
	repos, err := connect.Repos()
	if err != nil {
		return err
	}
	for _, availableProduct := range productPackages {
		// skip non-obsolete products
		if compareEditions(availableProduct.Edition, p.Edition()) >= 0 {
			continue
		}
		// filter out "(System Packages)" and already disabled repos
		found := false
		for _, r := range repos {
			if r.Name == availableProduct.Repo && r.Enabled {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		QuietOut.Printf("Found obsolete repository %s", availableProduct.Repo)
		if force {
			QuietOut.Println("... disabling.")
			connect.DisableRepo(availableProduct.Repo)
		} else {
			for {
				fmt.Printf("\nDisable obsolete repository %s [y/n] (y): ", availableProduct.Repo)
				scanner := bufio.NewScanner(os.Stdin)
				if !scanner.Scan() {
					QuietOut.Print("\nStandard input seems to be closed, please use '--non-interactive' option\n")
					os.Exit(1)
				}
				choice := strings.ToLower(strings.TrimSpace(scanner.Text()))
				if interrupted {
					return ErrInterrupted
				}
				if choice == "n" {
					fmt.Print("\n")
					break
				} else if choice == "y" || choice == "" {
					fmt.Println("... disabling.")
					connect.DisableRepo(availableProduct.Repo)
					break
				}
			}
		}
	}
	return nil
}

// updates system records in SCC/SMT
// adds/removes services to match target state
// disables obsolete repos
// returns base product version string
func migrateSystem(migration connect.MigrationPath, forceDisableRepos bool) (string, error) {
	var baseProductVersion string

	for _, p := range migration {
		msg := "Upgrading product " + p.FriendlyName
		QuietOut.Println(msg)
		service, err := connect.UpgradeProduct(p)
		if err != nil {
			return baseProductVersion, fmt.Errorf("%s: %v", msg, err)
		}

		if service.ObsoletedName != "" {
			msg := "Removing service " + service.ObsoletedName
			VerboseOut.Println(msg)
			err = connect.MigrationRemoveService(service.ObsoletedName)
			if err != nil {
				return baseProductVersion, err
			}
		}

		if err := cleanupProductRepos(p, forceDisableRepos); err != nil {
			return baseProductVersion, err
		}

		msg = "Adding service " + service.Name
		VerboseOut.Println(msg)
		err = connect.MigrationAddService(service.URL, service.Name)
		if err != nil {
			return baseProductVersion, err
		}

		// store the base product version
		if p.IsBase {
			baseProductVersion = p.Version
		}
		if interrupted {
			return baseProductVersion, fmt.Errorf("%s: %v", msg, ErrInterrupted)
		}
	}
	return baseProductVersion, nil
}

// containsProduct returns true if given slice of products contains one with given name.
func containsProduct(products []connect.Product, name string) bool {
	for _, p := range products {
		if p.Name == name {
			return true
		}
	}
	return false
}

// returns fs_inconsistent flag
func applyMigration(migration connect.MigrationPath, systemProducts []connect.Product,
	quiet, verbose, nonInteractive, forceDisableRepos bool,
	dupArgs []string) (bool, error) {

	fsInconsistent := false

	if err := connect.ZypperBackup(); err != nil {
		// NOTE: original ignores failures of this command
		fmt.Printf("Zypper backup failed: %v\n", err)
	}

	if interrupted {
		return fsInconsistent, fmt.Errorf("Preparing migration: %v", ErrInterrupted)
	}

	// Disable all old repos in case of Leap -> SLES migration (bsc#1184237)
	if containsProduct(systemProducts, "Leap") && containsProduct(migration, "SLES") {
		QuietOut.Println("Migration from Leap to SLES - disabling old repositories")
		repos, err := connect.Repos()
		if err != nil {
			return fsInconsistent, err
		}
		for _, r := range repos {
			if r.Enabled {
				connect.DisableRepo(r.Name)
			}
		}
	}

	baseProductVersion, err := migrateSystem(migration, nonInteractive || forceDisableRepos)
	if err != nil {
		return fsInconsistent, err
	}

	echo := connect.SetSystemEcho(true)
	if err := connect.RefreshRepos(baseProductVersion, true, false, false, false); err != nil {
		return fsInconsistent, fmt.Errorf("Refresh of repositories failed: %v", err)
	}
	if interrupted {
		return fsInconsistent, ErrInterrupted
	}

	err = connect.DistUpgrade(baseProductVersion, quiet, verbose, nonInteractive, dupArgs)
	connect.SetSystemEcho(echo)
	// TODO: export connect.zypperErrCommit (8)?
	if err != nil && err.(connect.ZypperError).ExitCode == 8 {
		fsInconsistent = true
	}
	if interrupted {
		return fsInconsistent, ErrInterrupted
	}

	return fsInconsistent, err
}

// checkFlagContradictions returns an error if a flag and its negative are both provided.
// e.g. "cmd --quiet --no-quiet ..."
func checkFlagContradictions() error {
	var err error
	seen := connect.NewStringSet()

	flag.Visit(func(f *flag.Flag) {
		if strings.HasPrefix(f.Name, "no-") {
			if seen.Contains(f.Name[3:]) {
				err = fmt.Errorf("Flags contradict: --%s and --%s", f.Name[3:], f.Name)
			}
		} else {
			if seen.Contains("no-" + f.Name) {
				err = fmt.Errorf("Flags contradict: --%s and --%s", f.Name, "no-"+f.Name)
			}
		}
		seen.Add(f.Name)
	})

	return err
}

func zypperDupArgs() []string {
	// NOTE: "r" is not listed here as it shares values list with "repo"
	wanted := connect.NewStringSet("auto-agree-with-licenses", "l",
		"allow-vendor-change", "no-allow-vendor-change",
		"debug-solver", "recommends", "no-recommends",
		"replacefiles:", "details", "download",
		"download-only", "from", "repo")

	args := []string{}

	// special case (from original)
	// pass `no-allow-vendor-change` to `zypper dup` unless
	// `allow-vendor-change` was used with `zypper migration`.
	// this means that the default from /etc/zypp.conf is always
	// ignored
	avc := "no-allow-vendor-change"
	flag.Visit(func(f *flag.Flag) {
		// skip not wanted
		if !wanted.Contains(f.Name) {
			return
		}
		// special case (update var and skip flag)
		if strings.Contains(f.Name, "allow-vendor-change") {
			avc = f.Name
			return
		}
		// multiArg? loop over values
		if val, ok := f.Value.(*multiArg); ok {
			for _, v := range *val {
				args = append(args, "--"+f.Name, v)
			}
		} else { // flag arg
			args = append(args, "--"+f.Name)
		}
	})
	// special case (add avc flag)
	args = append(args, "--"+avc)
	return args
}

func fetchAllMigrations(installed []connect.Product, target string) ([]connect.MigrationPath, error) {
	// offline migrations to given product
	if target != "" {
		newProduct, err := connect.SplitTriplet(target)
		if err != nil {
			return []connect.MigrationPath{}, err
		}
		return connect.OfflineProductMigrations(installed, newProduct)
	}
	// online migrations
	return connect.ProductMigrations(installed)
}
