---
title: man
section: 8
header: zypper-search-packages man page
date: January 2022
---
# NAME

**zypper-search-packages** - extended search for packages

# SYNOPSIS

zypper search-packages \[options] package1 [package2 [...]]

# DESCRIPTION

zypper search-packages performs extended search for packages covering
all potential SLE modules by querying RMT/SCC.
This operation needs access to a network.

Same as for the normal search operation the search string can be a part of a package
name unless the option --match-exact is used.

# OPTIONS

  **--match-substrings**
  : Search for a match to partial words (default).

  **-x**, **--match-exact**
  : Search for an exact match of the search strings

  **-C**, **--case-sensitive**
  : Perform case-sensitive search.

  **--sort-by-name**
  : Sort packages by name (default).

  **--sort-by-repo**
  : Sort packages by repository or module.

  **-g**, **--group-by-module**
  : Group the results by module (default: group by package)

  **--no-query-local**
  : Do not search installed packages and packages in available repositories.

  **-s**, **--details**
  : Display more detailed information about found packages

  **--xmlout**
  : Switch to XML output

# SEE ALSO

zypper(8)

# BUGS

No known bugs.

# AUTHOR

SUSE LLC (<scc-feedback@suse.de>)
