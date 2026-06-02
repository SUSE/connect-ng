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

## Basic Configuration

Example:

```yaml
---
url: https://scc.suse.com
language: en
insecure: false
auto_agree_with_licenses: false
enable_system_uptime_tracking: false
no_zypper_refs: false

collectors:
  pci_data:
    state: disabled
  mod_list:
    state: disabled
  installed_pkgs:
    state: disabled
```

The top-level fields are as follows:

  * url: (optional) URL of the registration server. Corresponds to the --url argument to SUSEConnect. Defaults to https://scc.suse.com
  * language: (optional) Language code to use for error messages
  * insecure: (optional) Do not verify SSL certificates when using https (default: false)
  * debug: (optional) Enable additional debugging output (default: false)
  * namespace: (optional) Namespace for the registration proxy
  * email: (optional) Email address for registration
  * no_zypper_refs: (optional) Do not refresh zypper service when registering (default: false)
  * auto_agree_with_licenses: (optional) Automatically agree to extension and module license confirmation prompts (default: false)
  * enable_system_uptime_tracking: (optional) Enable system uptime tracking. The system uptime log will be sent to SCC/RMT as part of keepalive (default: false)

## Collector Configuration

SUSEConnect collects data about your system for registration and support purposes. Certain collectors are mandatory and cannot be disabled, while optional collectors can be configured.

### Mandatory Collectors

The following collectors are always enabled and cannot be disabled:

  * arch: System CPU architecture (x86_64, aarch64, s390x, ppc64le)
  * hypervisor: Hypervisor type (KVM, Xen, VMware, etc.)
  * cloud_provider: Cloud platform (AWS, Azure, GCP, etc.)
  * container_runtime: Container platform (Docker, Podman, etc.)
  * cpus: CPU count, sockets, and thread information
  * mem_total: Total system memory
  * vendor: Hardware vendor (HP, Dell, Lenovo, etc.)
  * uname: Kernel version and OS details
  * hostname: System hostname
  * uuid: System UUID for unique identification
  * sap: SAP workload detection
  * kubernetes_provider: Kubernetes/RKE2/K3s cluster detection
  * ha_active: High-availability cluster detection

### Optional Collectors

The following collectors are enabled by default but can be disabled:

  * pci_data: PCI device information (state: enabled/disabled)
  * mod_list: Loaded kernel modules (state: enabled/disabled)
  * installed_pkgs: System installed packages (state: enabled/disabled)

Valid state values are:
  * `enabled`: Enable the collector
  * `disabled`: Disable the collector

# AUTHOR

SUSE LLC <scc-feedback@suse.de>

# LINKS

[SUSE Customer Center][scc]

# SEE ALSO

SUSEConnect(8)
