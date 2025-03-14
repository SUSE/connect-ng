---
title: SUSECONNECT
section: 8
header: SUSEConnect
date: January 2022
---
# NAME
**SUSEConnect** - SUSE Customer Center registration tool

# SYNOPSIS

**SUSEConnect [<optional>...] -p PRODUCT**

# DESCRIPTION

Register SUSE Linux Enterprise installations with the SUSE Customer Center.
Registration allows access to software repositories (including updates)
and allows online management of subscriptions and organizations.

By default, SUSEConnect registers the base SUSE Linux Enterprise product
installed on a system. It can also be used to register extensions and modules.

To register an extension or a module, use the **--product <PRODUCT-IDENTIFIER>**
option together with the product identifier of the extension or module.
You can see a list of all available extensions and modules for your system by
using the **--list-extensions** option.

Manage subscriptions at the SUSE Customer Center: https://scc.suse.com

# OPTIONS

  **-p**, **--product <PRODUCT>**
  : Specify a product for activation/deactivation. Only one product can be
    processed at a time. Defaults to the base SUSE Linux Enterprise product on
    this system. Product identifiers can be obtained with **--list-extensions**.
    Format: <name>/<version>/<architecture>

  **-r**, **--regcode <REGCODE>**
  : Subscription registration code for the product to be registered.
    Relates that product to the specified subscription and enables software
    repositories for that product. It can also be used with `-r -` to pass the registration code via stdin or `-r @/file/path` to specify a file containing the code.

  **-d**, **--de-register**
  : De-registers the system and base product, or in conjunction with
    --product, a single extension, and removes all its services installed by
    SUSEConnect. After de-registration, the system no longer consumes a
    subscription slot in SCC.

  **-l**, **--list-extensions**
  : List all extensions and modules available for installation on this system.

  **--instance-data <path to file>**
  : Path to the XML file holding the public key and instance data
    for cloud registration with SMT.

  **-e**, **--email <email>**
  : Email address for product registration.

  **--url <URL>**
  : URL of registration server (e.g. https://scc.suse.com).

  **--namespace <NAMESPACE>**
  : Namespace option for use with SMT staging environments.

  **-s**, **--status**
  : Get current system registration status in json format.

  **--status-text**
  : Get current system registration status in text format.

  **--keepalive**
  : Send a keepalive call to the registration server, so it can detect which
    systems are still running.

  **--write-config**
  : Write options to config file at /etc/SUSEConnect.

  **--cleanup**
  : Remove old system credentials and all zypper services installed by
    SUSEConnect.

  **--rollback**
  : Revert the registration state in case of a failed migration.

  **--root <PATH>**
  : Path to the root folder, uses the same parameter for zypper.

  **--gpg-auto-import-keys**
  : Automatically trust and import new repository signing keys.

  **--version**
  : Print program version.

  **--debug**
  : Provide debug output.

  **--json**
  : Print output in JSON format. This flag is only supported for registering, de-registering and list-extensions.

  **-h**, **--help**
  : Show help message.

# EXIT CODES

  SUSEConnect sets the following exit codes:

  * 0:  Registration successful
  * 64: Connection refused
  * 65: Access error, e.g. files not readable
  * 66: Parser error: Server JSON response was not parseable
  * 67: Server responded with error: see log output

# COMPARED TO SUSE_REGISTER
## BEFORE
  **suse_register -a email=<email> -a regcode-sles=<regcode> -L <logfile>**

## AFTER
  **SUSEConnect --url <registration-server-url> -r <regcode> >> <logfile>**

# USE WITH REGISTRATION PROXY

  SUSEConnect can also be used to register systems with a local SUSE
  registration proxy (RMT/SMT) instead of the SUSE Customer Center.
  Use **SUSEConnect --url <registration-proxy-server-url>** to register systems with RMT/SMT.

# IMPLEMENTATION

  SUSEConnect is implemented in Golang. It communicates with the registration
  server using a RESTful JSON API over HTTP using TLS encryption.

# ENVIRONMENT

  SUSEConnect respects the HTTP_PROXY environment variable.
  See https://www.suse.com/support/kb/doc/?id=000017441 for more details
  on how to manually configure proxy usage.

# FILES

  **/etc/SUSEConnect**
  : Configuration file containing server URL, regcode and language for
    registration.

# AUTHOR

SUSE LLC (<scc-feedback@suse.de>)

# LINKS

SUSE Customer Center: https://scc.suse.com

SUSEConnect on GitHub: https://github.com/SUSE/connect-ng

# SEE ALSO

SUSEConnect(5)
