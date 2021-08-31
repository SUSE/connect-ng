package main

// TODO LIST
// * zypp_backup/zypp_restore functions
// * zypper dup wrapper with options (these are mostly pass-through from plugin args)
// *   passthrough zypper dup options:
// *     --allow-vendor-change
// *     --from
// *     --repo
// *     --debug-solver
// *     --recommends
// *     --no-recommends
// *     --replacefiles
// *     --details
// *     --download (including --download-only)
// * --selfupdate option
// * --query plugin option
// * --break-my-system option
// * --product option (offline migration)
// * obsolete repo disabling (including --disable-repos option)
// * interactive migration mode
// * snapshots (snapper wrapper)
// * Leap -> SLES migration case
// * system.execute() function with pipe output from executed program to stdout

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
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
	Debug    *log.Logger = connect.Debug
	QuietOut *log.Logger = connect.QuietOut
)

func migrationMain() {
	var (
		dummy             bool
		verbose           bool
		quiet             bool
		nonInteractive    bool
		autoAgreeLicenses bool
		noSnapshots       bool
		noSelfUpdate      bool
		migrationNum      int
		fsRoot            string
	)

	flag.Usage = func() {
		fmt.Print(migrationUsageText)
	}
	// dummy flag to keep default but accept cli arg
	flag.BoolVar(&dummy, "no-verbose", false, "")
	flag.BoolVar(&verbose, "verbose", false, "")
	flag.BoolVar(&verbose, "v", false, "")
	// dummy flag to keep default but accept cli arg
	flag.BoolVar(&dummy, "no-quiet", false, "")
	flag.BoolVar(&quiet, "quiet", false, "")
	flag.BoolVar(&quiet, "q", false, "")
	flag.BoolVar(&nonInteractive, "non-interactive", false, "")
	flag.BoolVar(&nonInteractive, "n", false, "")
	flag.BoolVar(&autoAgreeLicenses, "auto-agree-with-licenses", false, "")
	flag.BoolVar(&autoAgreeLicenses, "l", false, "")
	flag.BoolVar(&noSnapshots, "no-snapshots", false, "")
	// dummy flag to keep default but accept cli arg
	flag.BoolVar(&dummy, "selfupdate", false, "")
	flag.BoolVar(&noSelfUpdate, "no-selfupdate", false, "")
	flag.IntVar(&migrationNum, "migration", 0, "")
	flag.StringVar(&fsRoot, "root", "", "")

	flag.Parse()
	// this is only to keep the flag parsing logic simple and avoid double
	// negations/negatives below
	selfUpdate := !noSelfUpdate

	if verbose {
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

	if !isSnapperConfigured() {
		noSnapshots = true
		Debug.Println("Snapper not configured")
	}

	// if options[:to_product] && !options[:root] && !options[:break_my_system]
	//   print "The --product option can only be used together with the --root option\n"
	//   exit 1
	// end

	if selfUpdate {
		//   # Update stack can be outside of the to be updated system
		//   cmd = "zypper " +
		//         (options[:non_interactive] ? "--non-interactive " : "") +
		//         (options[:verbose] ? "--verbose " : "") +
		//         (options[:quiet] ? "--quiet " : "") +
		//         "patch-check --updatestack-only"
		//   print "\nExecuting '#{cmd}'\n\n" unless options[:quiet]
		//   if !system cmd
		//     if $?.exitstatus >= 100
		//       # install pending updates and restart
		//       cmd = "zypper " +
		//             (options[:non_interactive] ? "--non-interactive " : "") +
		//             (options[:verbose] ? "--verbose " : "") +
		//             (options[:quiet] ? "--quiet " : "") +
		//             "--no-refresh patch --updatestack-only"
		//       print "\nExecuting '#{cmd}'\n\n" unless options[:quiet]
		//       system cmd

		//       # stop infinite restarting
		//       # check that the patches were really installed
		//       cmd = "zypper " +
		//         (options[:root] ? "--root #{options[:root]} " : "") +
		//         "--non-interactive --quiet --no-refresh patch-check --updatestack-only > /dev/null"
		//       if ! system cmd
		//         print "patch failed, exiting.\n"
		//         exit 1
		//       end

		//       print "\nRestarting the migration script...\n" unless options[:quiet]
		//       exec $0, *save_argv
		//     end
		//     exit 1
		//   end
	}
	QuietOut.Print("\n")

	// This is only necessary, if we run with --root option
	if err := connect.RefreshRepos("", false, quiet, verbose, nonInteractive); err != nil {
		fmt.Println("repository refresh failed, exiting")
		os.Exit(1)
	}

	systemProducts, err := checkSystemProducts(true, true)
	if err != nil {
		fmt.Printf("Can't determine the list of installed products: %v\n", err)
		os.Exit(1)
	}

	if len(systemProducts) == 0 {
		fmt.Println("No products found, migration is not possible.")
		os.Exit(1)
	}

	// This is not fully correct name as installedIDs could also contain some products
	// which are activated but not installed. This matches the original implementation.
	installedIDs := make(map[string]struct{}, 0)
	for _, prod := range systemProducts {
		installedIDs[prod.ToTriplet()] = struct{}{}
	}

	allMigrations := make([]connect.MigrationPath, 0)
	// if options[:to_product]
	//   begin
	//     identifier, version, arch = options[:to_product].split('/')
	//     new_product = OpenStruct.new(
	//                                  identifier: identifier,
	//                                  version:   version,
	//                                  arch:       arch
	//                                  )
	//     migrations_all = SUSE::Connect::YaST.system_offline_migrations(system_products, new_product)
	//   rescue => e
	//     print "Can't get available migrations from server: #{e.class}: #{e.message}\n"
	//     exit 1
	//   end
	// else
	{
		allMigrations, err = connect.ProductMigrations(systemProducts)
		if err != nil {
			fmt.Printf("Can't get available migrations from server: %v\n", err)
			os.Exit(1)
		}
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
			"Unavailable migrations (product is not mirrored):")
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

	// TODO: this part is only used in interactive mode
	for migrationNum <= 0 || migrationNum > len(migrations) {
		//   print "Available migrations:\n\n"
		//   migrations.each_with_index do |migration, index|
		//     printf "   %2d |", index + 1
		//     migration.each do |p|
		//       print " #{p.friendly_name}" + (p.already_installed ? " (already installed)" : "") + "\n       "
		//     end
		//     print "\n"
		//   end
		//   print "\n"
		//   if options[:query]
		//     exit 0
		//   end
		//   while migration_num <= 0 || migration_num > migrations.length do
		//     print "[num/q]: "
		//     choice = gets
		//     if !choice
		//       print "\nStandard input seems to be closed, please use '--non-interactive' option\n" unless options[:quiet]
		//       exit 1
		//     end
		//     choice.chomp!
		//     exit 0 if choice.eql?("q") || choice.eql?("Q")
		//     migration_num = choice.to_i
		//   end
	}

	migration := migrations[migrationNum-1]

	if !noSnapshots {
		//   cmd = "snapper create --type pre --cleanup-algorithm=number --print-number --userdata important=yes --description 'before online migration'"
		//   print "\nExecuting '#{cmd}'\n\n" unless options[:quiet]
		//   pre_snapshot_num = `#{cmd}`.to_i
	}

	// TODO: make sure to set this in snapper wrapper function when it's implemented
	// ENV['DISABLE_SNAPPER_ZYPP_PLUGIN'] = '1'

	// allow interrupt only at specified points
	// we have to check zypper exitstatus == 8 even after interrupt
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		Debug.Printf("Signal received: %v", sig)
		interrupted = true
	}()

	fsInconsistent, err := applyMigration(migration)

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

	// if !options[:no_snapshots] && pre_snapshot_num > 0
	//   cmd = "snapper create --type post --pre-number #{pre_snapshot_num} --cleanup-algorithm=number --print-number --userdata important=yes --description 'after online migration'"
	//   print "\nExecuting '#{cmd}'\n\n" unless options[:quiet]
	//   post_snapshot_num = `#{cmd}`.to_i
	// #  Filesystem rollback - considered too dangerous
	// #  if !result && post_snapshot_num > 0 && fs_inconsistent
	// #    cmd = "snapper undochange #{pre_snapshot_num}..#{post_snapshot_num}"
	// #    unless options[:non_interactive]
	// #      while true
	// #        print "Perform filesystem rollback with '#{cmd}' [y/n] (y): "
	// #        choice = gets.chomp
	// #        print "\n"
	// #        if choice.eql?('n') || choice.eql?('N')
	// #          fs_inconsistent = false
	// #          break
	// #        end
	// #        if choice.eql?('y') || choice.eql?('Y')|| choice.eql?('')
	// #          break
	// #        end
	// #      end
	// #    end
	// #    if fs_inconsistent
	// #      print "\nExecuting '#{cmd}'\n\n" unless options[:quiet]
	// #      system cmd
	// #    end
	// #  end
	// end

	// make sure all release packages are installed (bsc#1171652)
	if err == nil {
		_, err := checkSystemProducts(false, false)
		if err != nil {
			fmt.Printf("Can't determine the list of installed products after migration: %v\n", err)
			// the system has been sucessfully upgraded, zypper reported no error so
			// the only way to get here is a scc problem - it is better to just exit
			os.Exit(1)
		}
	}

	if err != nil {
		QuietOut.Print("\nPerforming repository rollback...\n")

		// TODO
		// restore repo configuration from backup file
		//     zypp_restore
		if err := connect.Rollback(); err == nil {
			QuietOut.Println("Rollback successful.")
		} else {
			fmt.Printf("Rollback failed: %v\n", err)
		}
		os.Exit(1)
	}
}

func isSnapperConfigured() bool {
	// TODO
	// system "/usr/bin/snapper --no-dbus list-configs 2>/dev/null | grep -q \"^root \""
	return false
}

func checkSystemProducts(rollbackOnFailure bool, printInstalledProducts bool) ([]connect.Product, error) {
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

	if printInstalledProducts {
		Debug.Println("Installed products:")
		for _, p := range systemProducts {
			Debug.Printf("  %-25s %s\n", p.ToTriplet(), p.Summary)
		}
		Debug.Print("\n")
	}
	return systemProducts, nil
}

func printMigrations(migrations []connect.MigrationPath,
	installedIDs map[string]struct{},
	header string) {
	fmt.Printf("%s\n\n", header)
	for _, m := range migrations {
		for _, p := range m {
			suffix := ""
			if !p.Available {
				suffix = suffix + " (not available)"
			}
			if _, installed := installedIDs[p.ToTriplet()]; installed {
				suffix = suffix + " (already installed)"
			}
			fmt.Printf("        %s%s\n", p.FriendlyName, suffix)
		}
		fmt.Print("\n")
	}
	fmt.Print("\n")
}

// sort migrations to put already installed products last and base products first
func sortMigrationProducts(m connect.MigrationPath, installedIDs map[string]struct{}) {
	sort.SliceStable(m, func(i, j int) bool {
		// first check installation status
		_, firstInstalled := installedIDs[m[i].ToTriplet()]
		_, secondInstalled := installedIDs[m[j].ToTriplet()]
		if firstInstalled != secondInstalled {
			return !firstInstalled
		}
		// if installation status is the same, check 'base' field
		firstBase := m[i].IsBase
		secondBase := m[j].IsBase
		return firstBase && !secondBase
	})
}

// updates system records in SCC/SMT
// adds/removes services to match target state
// disables obsolete repos
// returns base product version string
func migrateSystem(migration connect.MigrationPath) (string, error) {
	var baseProductVersion string

	for _, p := range migration {
		msg := "Upgrading product " + p.FriendlyName
		connect.QuietOut.Println(msg)
		service, err := connect.UpgradeProduct(p)
		if err != nil {
			return baseProductVersion, fmt.Errorf("%s: %v", msg, err)
		}

		if service.ObsoletedName != "" {
			msg := "Removing service " + service.ObsoletedName
			Debug.Println(msg)
			err = connect.MigrationRemoveService(service.ObsoletedName)
			if err != nil {
				return baseProductVersion, err
			}
		}

		//     SUSE::Connect::Migration::find_products(p.identifier).each do |available_product|
		//       # filter out "(System Packages)" and already disabled repos
		//       next unless SUSE::Connect::Migration::repositories.detect { |r| r[:name].eql?(available_product[:repository]) && r[:enabled] != 0 }
		//       if ProductVersion.new(available_product[:edition]) < ProductVersion.new(p.version)
		//         print "Found obsolete repository #{available_product[:repository]}" unless options[:quiet]
		//         if options[:non_interactive] || options[:disable_repos]
		//           print "... disabling.\n" unless options[:quiet]
		//           SUSE::Connect::Migration::disable_repository available_product[:repository]
		//         else
		//           while true
		//             print "\nDisable obsolete repository #{available_product[:repository]} [y/n] (y): "
		//             choice = gets.chomp
		//             if interrupted {
		//	             return baseProductVersion, ErrInterrupted
		//             }
		//             if choice.eql?('n') || choice.eql?('N')
		//               print "\n"
		//               break
		//             end
		//             if  choice.eql?('y') || choice.eql?('Y')|| choice.eql?('')
		//               print "... disabling.\n"
		//               SUSE::Connect::Migration::disable_repository available_product[:repository]
		//               break
		//             end
		//           end
		//         end
		//       end
		//     end

		msg = "Adding service " + service.Name
		Debug.Println(msg)
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

// returns fs_inconsistent flag
func applyMigration(migration connect.MigrationPath) (bool, error) {
	fsInconsistent := false
	// TODO
	//   zypp_backup(options[:root] ? options[:root]: "/")

	if interrupted {
		return fsInconsistent, fmt.Errorf("Preparing migration: %v", ErrInterrupted)
	}

	//   if system_products.detect { |p| p.identifier == "Leap" } &&
	//      migration.detect { |p| p.identifier == "SLES" }
	//      # bsc#1184237
	//      print "Migration from Leap to SLES - disabling old repositories\n" unless options[:quiet]
	//      SUSE::Connect::Migration::repositories.each do |repo|
	//          SUSE::Connect::Migration::disable_repository repo[:name] if repo[:enabled] != 0
	//      end
	//   end

	baseProductVersion, err := migrateSystem(migration)
	if err != nil {
		return fsInconsistent, err
	}

	if err := connect.RefreshRepos(baseProductVersion, true, false, false, false); err != nil {
		return fsInconsistent, fmt.Errorf("Refresh of repositories failed: %v", err)
	}
	if interrupted {
		return fsInconsistent, ErrInterrupted
	}

	// TODO
	//   cmd = "zypper " +
	//         (options[:root] ? "--root #{options[:root]} " : "") +
	//         (base_product_version ? "--releasever #{base_product_version} " : "") +
	//         (options[:non_interactive] ? "--non-interactive " : "") +
	//         (options[:verbose] ? "--verbose " : "") +
	//         (options[:quiet] ? "--quiet " : "") +
	//         " --no-refresh " +
	//         " dist-upgrade " +
	//         (options[:allow_vendor_change] ? "--allow-vendor-change " : "--no-allow-vendor-change ") +
	//         (options[:auto_agree] ? "--auto-agree-with-licenses " : "") +
	//         (options[:debug_solver] ? "--debug-solver " : "") +
	//         (options[:recommends] ? "--recommends " : "") +
	//         (options[:no_recommends] ? "--no-recommends " : "") +
	//         (options[:replacefiles] ? "--replacefiles " : "") +
	//         (options[:details] ? "--details " : "") +
	//         (options[:download] ? "--download #{options[:download]} " : "") +
	//         (options[:repo].map { |r| "--repo #{r}" }.join(" ")) +
	//         (options[:from].map { |r| "--from #{r}" }.join(" "))
	//   msg = "Executing '#{cmd}'"
	//   print "\n#{msg}\n\n" unless options[:quiet]
	//   result = system cmd
	// TODO:
	//if $?.exitstatus == 8
	fsInconsistent = true
	if interrupted {
		return fsInconsistent, ErrInterrupted
	}

	return fsInconsistent, nil
}
