package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/SUSE/connect-ng/internal/connect"
)

var (
	//go:embed migrationUsage.txt
	migrationUsageText string
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

	// # This is only necessary, if we run with --root option
	// cmd = "zypper " +
	//       (options[:root] ? "--root #{options[:root]} " : "") +
	//       (options[:non_interactive] ? "--non-interactive " : "") +
	//       (options[:verbose] ? "--verbose " : "") +
	//       (options[:quiet] ? "--quiet " : "") +
	//       " refresh"
	// print "\nExecuting '#{cmd}'\n\n" unless options[:quiet]
	// if !system cmd
	//   print print "repository refresh failed, exiting\n"
	//   exit 1
	// end

	// begin
	//   system_products = SUSE::Connect::Migration::system_products

	//   release_package_missing = false
	//   system_products.each do |ident|
	//     begin
	//       # if a release package for registered product is missing -> try install it
	//       SUSE::Connect::Migration.install_release_package(ident.identifier)
	//     rescue => e
	//       release_package_missing = true
	//       print "Can't install release package for registered product #{ident.identifier}\n" unless options[:quiet]
	//       print "#{e.class}: #{e.message}\n" unless options[:quiet]
	//     end
	//   end

	//   if release_package_missing
	//     # some release packages are missing and can't be installed
	//     print "Calling SUSEConnect rollback to make sure SCC is synchronized with the system state.\n" unless options[:quiet]
	//     SUSE::Connect::Migration.rollback

	//     # re-read the list of products
	//     system_products = SUSE::Connect::Migration::system_products
	//   end

	//   if options[:verbose]
	//     print "Installed products:\n"
	//     system_products.each {|p|
	//       printf "  %-25s %s\n", "#{p.identifier}/#{p.version}/#{p.arch}", p.summary
	//     }
	//     print "\n"
	//   end
	// rescue => e
	//   print "Can't determine the list of installed products: #{e.class}: #{e.message}\n"
	//   exit 1
	// end

	// if system_products.length == 0
	//   print "No products found, migration is not possible.\n"
	//   exit 1
	// end

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
	//   begin
	//     migrations_all = SUSE::Connect::YaST.system_migrations system_products
	//   rescue => e
	//     print "Can't get available migrations from server: #{e.class}: #{e.message}\n"
	//     exit 1
	//   end
	// end

	// #preprocess the migrations lists
	// migrations = Array.new
	// migrations_unavailable = Array.new
	// migrations_all.each do |migration|
	//   migr_available = true
	//   migration.each do |p|
	//     p.available = !defined?(p.available) || p.available
	//     p.already_installed = !! system_products.detect { |ip| ip.identifier.eql?(p.identifier) && ip.version.eql?(p.version) && ip.arch.eql?(p.arch) }
	//     if !p.available
	//       migr_available = false
	//     end
	//   end
	//   # put already_installed products last and base products first
	//   migration = migration.sort_by.with_index { |p, idx| [p.already_installed ? 1 : 0, p.base ? 0 : 1, idx] }
	//   if migr_available
	//     migrations << migration
	//   else
	//     migrations_unavailable << migration
	//   end
	// end

	// if migrations_unavailable.length > 0 && !options[:quiet]
	//   print "Unavailable migrations (product is not mirrored):\n\n"
	//   migrations_unavailable.each do |migration|
	//     migration.each do |p|
	//       print "        #{p.friendly_name}" + (p.available ? "" : " (not available)") + (p.already_installed ? " (already installed)" : "") + "\n"
	//     end
	//     print "\n"
	//   end
	//   print "\n"
	// end

	// if migrations.length == 0
	//   print "No migration available.\n\n" unless options[:quiet]
	//   if migrations_unavailable.length > 0
	//     # no need to print a msg - unavailable migrations are listed above
	//     exit 1
	//   end
	//   exit 0
	// end

	// migration_num = options[:migration]
	// if options[:non_interactive] && migration_num == 0
	//   # select the first option
	//   migration_num = 1
	// end

	// while migration_num <= 0 || migration_num > migrations.length do
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
	// end

	// migration = migrations[migration_num - 1]

	// if !options[:no_snapshots]
	//   cmd = "snapper create --type pre --cleanup-algorithm=number --print-number --userdata important=yes --description 'before online migration'"
	//   print "\nExecuting '#{cmd}'\n\n" unless options[:quiet]
	//   pre_snapshot_num = `#{cmd}`.to_i
	// end

	// ENV['DISABLE_SNAPPER_ZYPP_PLUGIN'] = '1'

	// base_product_version = nil # unknown yet

	// result = false
	// fs_inconsistent = false
	// msg = "Preparing migration"

	// begin
	//   # allow interrupt only at specified points
	//   # we have to check zypper exitstatus == 8 even after interrupt
	//   interrupted = false
	//   trap('INT') { interrupted = true }
	//   trap('TERM') { interrupted = true }

	//   zypp_backup(options[:root] ? options[:root]: "/")

	//   raise "Interrupted." if interrupted

	//   if system_products.detect { |p| p.identifier == "Leap" } &&
	//      migration.detect { |p| p.identifier == "SLES" }
	//      # bsc#1184237
	//      print "Migration from Leap to SLES - disabling old repositories\n" unless options[:quiet]
	//      SUSE::Connect::Migration::repositories.each do |repo|
	//          SUSE::Connect::Migration::disable_repository repo[:name] if repo[:enabled] != 0
	//      end
	//   end

	//   migration.each do |p|
	//     msg = "Upgrading product #{p.friendly_name}"
	//     print "#{msg}.\n" unless options[:quiet]
	//     service = SUSE::Connect::YaST.upgrade_product p

	//     unless service[:obsoleted_service_name].empty?
	//       msg = "Removing service #{service[:obsoleted_service_name]}"
	//       print "#{msg}.\n" if options[:verbose]
	//       SUSE::Connect::Migration::remove_service service[:obsoleted_service_name]
	//     end

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
	//             raise "Interrupted." if interrupted
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

	//     msg = "Adding service #{service[:name]}"
	//     print "#{msg}.\n" if options[:verbose]
	//     SUSE::Connect::Migration::add_service service[:url], service[:name]

	//     # store the base product version
	//     if p.base
	//       base_product_version = p.version
	//     end
	//     raise "Interrupted." if interrupted
	//   end

	//   cmd = "zypper " +
	//         (options[:root] ? "--root #{options[:root]} " : "") +
	//         (base_product_version ? "--releasever #{base_product_version} " : "") +
	//         "ref -f"
	//   msg = "Executing '#{cmd}'"
	//   print "\n#{msg}\n\n" unless options[:quiet]
	//   result = system cmd

	//   raise "Refresh of repositories failed." unless result
	//   raise "Interrupted." if interrupted

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
	//   fs_inconsistent = true if $?.exitstatus == 8
	//   raise "Interrupted." if interrupted

	// rescue => e
	//   print "#{msg}: #{e.class}: #{e.message}\n"
	//   result = false
	// end

	// print "\nMigration failed.\n\n" unless result || options[:quiet]

	// if fs_inconsistent
	//   print "The migration to the new service pack has failed. The system is most\n"
	//   print "likely in an inconsistent state.\n"
	//   print "\n"
	//   print "We strongly recommend to rollback to a snapshot created before the\n"
	//   print "migration was started (via selecting the snapshot in the boot menu\n"
	//   print "if you use snapper) or restore the system from a backup.\n"
	//   exit 2
	// end

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

	// # make sure all release packages are installed (bsc#1171652)
	// if result
	//   begin
	//     system_products = SUSE::Connect::Migration::system_products

	//     system_products.each do |ident|
	//       begin
	//         # if a release package for registered product is missing -> try install it
	//         SUSE::Connect::Migration.install_release_package(ident.identifier)
	//       rescue => e
	//         print "Can't install release package for registered product #{ident.identifier}\n" unless options[:quiet]
	//         print "#{e.class}: #{e.message}\n" unless options[:quiet]
	//       end
	//     end
	//   rescue => e
	//     print "Can't determine the list of products installed after migration: #{e.class}: #{e.message}\n"
	//     # the system has been sucessfully upgraded, zypper reported no error so
	//     # the only way to get here is a scc problem - it is better to just exit
	//     #
	//     exit 1
	//   end
	// end

	// if !result
	//   print "\nPerforming repository rollback...\n" unless options[:quiet]
	//   begin
	//     # restore repo configuration from backup file
	//     zypp_restore
	//     ret = SUSE::Connect::Migration.rollback
	//     print "Rollback successful.\n" unless options[:quiet]
	//   rescue => e
	//     print "Rollback failed: #{e.class}: #{e.message}\n"
	//   end
	//   exit 1
	// end
}

func isSnapperConfigured() bool {
	// TODO
	// system "/usr/bin/snapper --no-dbus list-configs 2>/dev/null | grep -q \"^root \""
	return false
}
