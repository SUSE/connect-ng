# Building the package
## TL;DR

The package is available at `obs://systemsmanagement:SCC/suseconnect-ng` and it
can be fetched with `osc` in the usual way. This retrieves the last tagged
release version.

## Step 1. Version management

The version of `connect-ng` is a dynamically derived from the version specified
in the RPM package [suseconnect-ng.spec](build/packaging/suseconnect-ng.spec),
so to update the version, just need update the value on the `Version:` line in the
spec file.

## Step 2. Ensure the change log is up to date

Ensure that the [suseconnect-ng.changes](build/packaging/suseconnect-ng.changes)
file is up to date. If necessary create a new entry for the target version, and
add appropriate changes entries to it.

**Note**: Remember to include any relevant Bugzilla (bsc#xxxxxxx) and Jira
(jsc#xxxxxxx) references in the changelog entries.

## Step 3. Verify the package builds locally

To verify that the package builds locally you can use the `build-rpm` Makefile
target to run a build locally in an appropriate container environment and fix
any build issues encountered.

See below for details on testing the locally built package.

## Step 4. Tagging the release

When preparing a release the version specifed in the
[suseconnect-ng.spec](build/packaging/suseconnect-ng.spec)
should be used to create a `v<Version>` tag, which will be the basis for building
an updated package in OBS.

**Note**: Remember to push the tag to the repository if you are creating the tag
locally.

## Step 5. Updating the package in OBS

The RPM is built via the OBS [systemsmanagement:SCC/suseconnect-ng](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng)
package.

**Note**: To modify the package there please ensure that you have been assigned
the required access.

First you need to checkout the package:

```bash
$ mkdir obs
$ cd obs
$ osc co systemsmanagement:SCC suseconnect-ng -o .
```

To build the new candidate version of the package you will first need to update
the `revision` parameter in the _service` file, specifying the new version tag
created previously.

Once the `_service` file has been updated you can then `osc service manualrun`
to update the package. Note that `manualrun` can be abbrediated to `mr`.

Remember to delete the tarball associated with the previous version.

```bash
$ osc service manualrun
```

This retrieve the sources associated with the specified tag, generates a compressed
tarball, extracts the spec, changes and rpmlintrc files and generates a compressed
Go vendor tarball to be used to build the package.

Once you have successfully run the service to update the package, you can build it
as usual with:

```bash
$ osc build Leap_16.0 x86_64
$ osc build SLE_15_SP7 x86_64
```

**Note**: You can optionally specify `--no-verify` to skip PGP signature checks if your
local builds fail for that reason.

Once you are happy that the updated package is building correctly, make sure to
add any new files and remove any that are no longer required, either using the
`osc add ...`, `osc remove ...` or `osc addremove` commands.

Then use `osc commit` to submit the updated changes to the OBS project.

## Step 6. Verify updated package builds in OBS for relevant code streams

Either check the status of the
[suseconnect-ng package in OBS](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng)
or using the `osc results` command from the command line and verify that it
builds successfully for the relevant code streams or that any build failures
are for known reasons, e.g. go1.xx-openssl packages in OBS are not available
for older streams such as SLE 12 and SLE 15 prior to SP3.

## Step 7. Submit the updated package to Factory

If not triggered automatically, because
[systemsmanagement:SCC/suseconnect-ng](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng)
is the devel project for the `suseconnect-ng` package in openSUSE:Factory, submit
the updated package to openSUSE:Factory.

```bash
$ osc submitrequest systemsmanagement:SCC suseconnect-ng openSUSE:Factory
```

### Monitor your OBS submissions

You can check the status of your OBS requests
[here](https://build.opensuse.org/package/requests/systemsmanagement:SCC/suseconnect-ng).

Monitor the submission and address any issues identified until the submission is
accepted.

**Note**: If any patches need to be applied, please ensure that any associated changes
to the suseconnect-ng.changes file are replicated to the
[SUSE/connect-ng/build/packaging/suseconnect-ng.changes](build/packaging/suseconnect-ng.changes)
file.

## Step 8. Verify updated package in the Internal Build Service

The IBS equivalent of the OBS
[systemsmanagement:SCC/suseconnect-ng](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng)
package is
[Devel:SCC:suseconnect/suseconnect-ng](https://build.suse.de/package/Devel:SCC:suseconnect/suseconnect-ng).

Once the OBS submission to openSUSE:Factory has been accepted the IBS package
should be automatically updated.

Verify that it has been updated, and that the updated package builds successfully
for all relevant code streams, including the older SLE 12 and SLE 15 streams.

## Step 9. Submit maintenance updates to SLES

### Determine target codestreams where to submit maintenance updates

To checkout in which codestreams the package is currently maintaned, run:

```bash
osc -A https://api.suse.de maintained suseconnect-ng
```

Some code streams may be missing in the previous command, for a more detailed
view which target codestreams are in which state, find the `suseconnect-ng`
package on[maintenance.suse.de](https://smelt.suse.de/maintained/)

### Submit maintenance updates

For each maintained codestream you need to create a new maintance or submit request:

```bash
osc -A https://api.suse.de mr Devel:SCC:suseconnect suseconnect-ng SUSE:SLE-15-SP6:Update
```

**Note:**

* When asked whether or not to supersede a request, the answer is usually "no".
  Saying "yes" would overwrite the previous request made, cancelling the release
  process for its codestream.

**Note**: The codestreams of SLE-15-SP1, SLE-15-SP2 and SLE-15-SP3 are connected, that means we only need to submit for SP1 and it will get released on SP2 and SP3 also.

**Note**: The codestreams that are not yet on final codefreeze (alphas or betas) work with submit requests rather than maintenance requests.

**Note**: SLE Micro <= 5.5 inherits MicroOS and SLE Releases.

**Note**: `submitrequest` (`sr`)'s should auto-translate to `maintenancerequest` (`mr`)'s!

**Note**: In case the `sr` (submit request) command is not working properly, try
  `mr` (maintenance request) command. If a maintenance request is not
  applicable, the maintainers will notify you in the request.

### RES8 Submissions

For RES8 (SUSE Linux Enterprise Server with Expanded Support), package updates
need to get done by EPAM (contact is: res-coord@suse.de). We agreed to only push
critical security updates there.

### Monitor your IBS submissions

You can check the status of your IBS requests
[here](https://build.suse.de/package/requests/Devel:SCC:suseconnect/suseconnect-ng).

Monitor the submission and address any issues identified until the submission is
accepted.

**Note**: If any patches need to be applied, please ensure that any associated changes
to the suseconnect-ng.changes file are replicated to the
[SUSE/connect-ng/build/packaging/suseconnect-ng.changes](build/packaging/suseconnect-ng.changes)
file.

# Testing the package locally

You can perform basic validating tests of the RPM package using some of the
available [Makefile](Makefile) targets, such as `build-rpm` or `feature-tests`.
Alternatively you can build the packages localy with `osc build ...` or retrieve
the built packages from OBS with `osc getbinaries ...` and test them out inside
a container environment.

## Performing basic suseconnect feature validation

You can verify basic operation of the built RPM packages using the feature
tests by running `make feature-tests` which will build the associated RPM
packages in a local container environment, which are then used to install
the `suseconnect-ng` package and it's dependencies, so that the feature
tests can exercise the installed `suseconnect` binary.

## Testing built packages in a local container

If you have built the `suseconnect-ng` packages locally using `osc build ...`
you can copy them to a directory such as `/tmp/suseconnect-rpms`. Alternatively
you can download the packages built in OBS for the
[systemsmanagement:SCC/suseconnect-ng](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng)
package.

### Setup using locally built RPMs

Build the RPMs locally using `osc build ...` and copy them to a directory, e.g.
`/tmp/suseconnect-rpms`.

```bash
$ osc build SLE_15_SP7 x86_64
$ cp /var/tmp/build-root/SLE_15_SP7-x86_64/home/abuild/rpmbuild/RPMS/*suseconnect*.rpm /tmp/suseconnect-rpms
```

### Setup using packages built in OBS

Download the packages built in OBS for the
[systemsmanagement:SCC/suseconnect-ng](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng)
package to `/tmp/suseconnect-rpms`.

```bash
$ osc getbinaries systemsmanagement:SCC suseconnect-ng SLE_15_SP7 x86_64 -d /tmp/suseconnect-rpms
```

### Start a local container
You can use either `podman` or `docker` to start a local container matching distro
release and architecture that the candidate packages were built for. The important
thing is that the `/tmp/suseconnect-rpms` directory needs to be mounted into the
container runtime environment.

```bash
$ podman run --rm -it --privileged -v /tmp/suseconnect-rpms:/rpms registry.suse.com/bci/bci-base:15.7
a1ed05dd1bbb:/ # zypper --no-gpg-checks in /rpms/*.rpm
a1ed05dd1bbb:/ # suseconnect --version
1.21.1
```

# The CI

This project makes use of the feature from OBS in which you can build a package
from a given PR. This way we check that your changes don't break the packaging
side of `connect-ng`.

You can see the configuration on this on
[.obs/workflows.yml](.obs/workflows.yml). Right now this is taking place inside
of a personal project, but this is to be changed (see [this Jira
ticket](https://jira.suse.com/browse/CSD-79)).
