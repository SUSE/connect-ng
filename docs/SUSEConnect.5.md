---
title: SUSECONNECT
section: 5
header: SUSEConnect
date: December 2025
---
# NAME

**SUSEConnect** - SUSE Customer Center registration tool config file

# DESCRIPTION

</etc/SUSEConnect> is the config file for the SUSE registration tool SUSEConnect.  This file allows the registration of the base product that is installed on the system.  NB: using this file, registration of extensions is not supported.

# FORMAT

The file is in [YAML][yaml-spec] format.

Example:

**---**

**url: https://scc.suse.com**

**language: en**

**insecure: false**

**debug: false**

**no_zypper_refs: false**

**auto_agree_with_licenses: false**

**enable_system_uptime_tracking: false**


Each line of the file specifies a single parameter.  The fields are as follows:

  * url: (optional) URL of the registration server.  Corresponds to the --url argument to SUSEConnect. Defaults to https://scc.suse.com
  * language: (optional) Language code to use for error messages
  * insecure: (optional) Do not verify SSL certificates when using https (default: false)
  * debug: (optional) Enable additional debugging output (default: false)
  * no_zypper_refs: (optional) Do not refresh zypper service when registering (default: false)
  * auto_agree_with_licenses: (optional) Automatically agree to extension and module license confirmation prompts (default: false)
  * enable_system_uptime_tracking: (optional) Enable system uptime tracking. The system uptime log will be sent to SCC/RMT as part of keepalive (default: false)


# AUTHOR

SUSE LLC <scc-feedback@suse.de>

# LINKS

[SUSE Customer Center][scc]

# SEE ALSO

SUSEConnect(8)
