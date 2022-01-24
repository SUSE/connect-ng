---
title: man
section: 8
header: zypper-migration man page
date: January 2022
---
# NAME

**zypper-migration** - perform service pack migration

# SYNOPSIS

zypper migration \[options]

# DESCRIPTION

zypper-migration performs online system migration to new service pack.

# OPTIONS

  **--[no-]allow-vendor-change**
  : Allow vendor change

  **-v**, **--[no-]verbose**
  : Increase verbosity

  **--debug**
  : Enable debug output

  **-q**, **--[no-]quiet**
  : Suppress normal output, print only error messages

  **-n**, **--non-interactive**
  : Do not ask anything, use default answers automatically

  **--query**
  : Query available migration options and exit

  **--disable-repos**
  : Disable obsolete repositories without asking

  **--migration N**
  : Select migration option N

  **--from REPO**
  : Restrict upgrade to specified repository

  **-r**, **--repo REPO**
  : Load only the specified repository

  **-l**, **--auto-agree-with-licenses**
  : Automatically say 'yes' to third party license confirmation prompt

  **--gpg-auto-import-keys**
  : Automatically trust and import new repository signing keys

  **--strict-errors-dist-migration**
  : Handle only breaking distro migration errors

  **--debug-solver**
  : Create solver test case for debugging

  **--recommends**
  : Install also recommended packages

  **--no-recommends**
  : Do not install recommended packages

  **--replacefiles**
  : Install the packages even if they replace files from other packages

  **--details**
  : Show the detailed installation summary

  **--download MODE**
  : Set the download-install mode

  **--download-only**
  : Replace repositories and download the packages, do not install. WARNING: This leaves the system in inconsistent
    state with new repositories and old packages installed. Upgrade with 'zypper
    dist-upgrade' as soon as possible.

  **--no-snapshots**
  : Don't create snapshots during migration

  **-p**, **--product PRODUCT**
  : Do an offline upgrade to PRODUCT. This requires the --root option.

  **--[no-]selfupdate**
  : Specify, if the update stack should update itself at first

  **--root DIRECTORY**
  : Operate on a different root directory.

# SEE ALSO

zypper(8)

# BUGS

No known bugs.

# AUTHOR

SUSE LLC (<scc-feedback@suse.de>)
