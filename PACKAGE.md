# TL;DR

The package is available at `obs://systemsmanagement:SCC/suseconnect-ng` and it
can be fetched with `osc` in the usual way.

This currently corresponds to the most recently tagged release version.

# Table of Contents

* [Minimym Requirements](#minimum-requirements)
* [Building and Releasing suseconnect-ng](#building-and-releasing-suseconnect-ng)
* [Testing the package locally](#testing-the-package-locally)
* [The CI](#the-ci)
* [Optional Release or Patch/Hotfix Branching Strategy](#optional-release-or-patchhotfix-branching-strategy)

# Minimum Requirements

The following tools should be available and verified as working:

  * A Golang v1.24 or later environment
    * needed to run certain go commands locally for testing and
      validation.
  * `git`
  * `osc`
    * ensure that the following service support packages, and any
      dependencies, are also available:
      * `obs-service-tar_scm`
      * `obs-service-go_modules`
      * `obs-service-extract_file`
      * `obs-service-recompress`
  * `obs-git` (optional)
    *  used by `src.suse.de` (see below) SLFO 1.2 & later submissions as
       an alternative to manually cloning repos with `git`
  * `docker`
    * `podman` can be used in some cases but `docker` is required
      for the [SUSE/connect-ng](github.com/SUSE/connect-ng) repo
      tooling.

Ensure the following accounts are accessible and working correctly:
  * [github.com](https://github.com)
    * ensure your account is able to:
      * access the [SUSE/connect-ng](github.com/SUSE/connect-ng) repo
        with permissions to push branches that can run CI tests, push
        tags, and create GitHub releases.
  * [build.opensuse.org](https://build.opensuse.org)
    * ensure your account is a member of the
      [systemsmanagement:SCC/suseconnect-ng users](https://build.opensuse.org/package/users/systemsmanagement:SCC/suseconnect-ng)
      either directly, or indirectly via a group, with maintainer
      permissions.
    * ensure your local `osc` credentials are setup correctly
  * [build.suse.de](https://build.suse.de)
    * ensure your account is a member of the
      [Devel:SCC:suseconnect/suseconnect-ng users](https://build.suse.de/package/users/Devel:SCC:suseconnect/suseconnect-ng)
      either directly, or indirectly via a group, with maintainer
      and possibly bugowner permissions.
    * ensure your local `osc -A https://api.suse.de` credentials are
      setup correctly, with a valid SSH Key (RSA 4K or ED25519) enrolled
      with [idp-mfs.suse.de](https://idp-mfa.suse.de/#!/login).
  * [src.suse.de](https://src.suse.de)
    * ensure your account is setup with a valid SSH key for authenticating
      git operations; use the same one as enrolled with `idp-mfa.suse.de`.
    * review the [Git Workflow Documentation](https://src.suse.de/SUSE/git-workflow-documentation),
      in particular the sections related to package submissions and
      maintenance.

# Building and Releasing suseconnect-ng

## Step 1. Version management

The version of `connect-ng` is dynamically derived from the version specified
in the RPM package [suseconnect-ng.spec](build/packaging/suseconnect-ng.spec),
so to update the version, just update the value on the `Version:` line in the
spec file.

The Makefile target `show-version` will report the currently specified version.

## Step 2. Ensure the change log is up to date

Ensure that the [suseconnect-ng.changes](build/packaging/suseconnect-ng.changes)
file is up to date. If necessary create a new entry for the target version, and
add appropriate changes entries to it.
Make sure that "(pre)" is removed from the changes log.

**Note**: Remember to include any relevant Bugzilla (bsc#xxxxxxx) and Jira
(jsc#xxxxxxx) references in the changelog entries.

## Step 3. Verify the package builds locally

Verify that the RPM package builds using the Makefile's `build-rpm` target; this
builds it locally inside an appropriate container environment.  Fix any build
issues encountered, repeating as necessary.

See [below](#Testing-the-package-locally) for details on how to test the locally
built package.

## Step 4. Tagging the release

Use the version specifed in the [suseconnect-ng.spec](build/packaging/suseconnect-ng.spec),
which can be retrieved using `make show-version`, to create a `v<Version>` tag
that will become the basis for building an updated package in OBS.

A tag can be created locally against the latest version of the main branch as
follows, using v1.22.0 as an example:

```bash
$ git fetch origin
$ git checkout origin/main
$ TAG=v$(make show-version)
$ echo $TAG # make sure this is the one you want
$ git tag $TAG
$ git push origin $TAG
```

Alternatively tags can be created in the GitHub web interface as part of creating
a draft release, but please don't create the release in GitHub until later. This is done in
[Step 10 below](#step-10-create-the-github-release-for-the-version-tag).

**Note**: Remember to push the tag to the repository if creating the tag locally.

## Step 5. Updating the package in OBS

The RPM is built in
[OBS systemsmanagement:SCC/suseconnect-ng](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng).

**Notes**:
  * To be able to modify the package there please ensure that your
    account has been assigned the
    [required permissions](https://build.opensuse.org/package/users/systemsmanagement:SCC/suseconnect-ng)
    either directly or as a member of an authorized group.
  * The `maintainer` and `bugowner` roles are recommended to be able to submit
    packages without issues.

First checkout the package:

```bash
$ mkdir obs
$ cd obs
$ osc co systemsmanagement:SCC suseconnect-ng -o .
```

Update the `revision` parameter in the `tar_scm` service definition in the
`_service` file to specify the new version tag created previously.

*Remember to delete the tarball associated with the previous version.*

Then run `osc service manualrun` to update the package. Note that `manualrun` can
be abbrediated to `mr`. There should be only one suseconnect-ng-*.tar.xz file present.
So the existing suseconnect-ng-*.tar.xz must be removed and the new one added.

```bash
$ osc rm suseconnect-ng-*.tar.xz # remove the previous tar.xz
$ osc service manualrun
$ osc add suseconnect-ng-*.tar.xz
```

This will perform the following:
* retrieve the git sources associated with the specified tag.
* generate an `xz` compressed tarball of the sources.
* extract the spec, changes and rpmlintrc files from the tarball.
* generate an `xz` compressed Go vendor tarball to be used when building the package.

After successfully updating the package, test build it locally as usual with:

```bash
$ osc build Leap_16.0 x86_64
$ osc build SLE_15_SP7 x86_64
```

**Notes**:
  * If the local build fails due to PGP signature verification issues:
    * try running again with the `--clean` option to recreate the build area
      which should retrieve the latest versions of relevant signing keys.
    * ensure that the gpg tool is installed and available on your system.
    * the `--no-verify` option can be used to temporarily disable that check
      if all else fails. Should only be use if absolutly needed, since it also
      disables rpmlint verification checks at the end of the local build run.

The `osc build` command may complain about unreferenced files (such as dropped
patches) being found; remove them and repeat as necessary until the package
builds successfully.

Remember to add any new files and remove any that are no longer required,
and update the package to reflect these changes using the `osc add <filename>`,
`osc remove <filename>` or `osc addremove` commands.

Then submit the updated package to the OBS project like so:
```bash
$ osc commit
```

## Step 6. Verify updated package builds in OBS for relevant code streams

Check the status of the OBS package build with one of the following:
* Web UI [suseconnect-ng package in OBS](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng#build)
* or from command line:
```bash
$ osc results
```

Verify that the package builds successfuly for the relevant code streams or
that any build failures are for known reasons, e.g. go1.xx-openssl packages
are not available in OBS for older streams such as SLE 12 and SLE 15 prior
to SP3.

Consider also downloading and verifying that the built RPMs perform as
expected for a subset of the relevant code streams. See below for details
on how to test downloaded package RPMs locally.

## Step 7. Submit the updated package to Factory

Determine if there are any outstanding requests with
```bash
$ osc request list
```
If there is a recent request for the target stream that includes the intended
updated package content, this may indicate an automatic devel project/branch
submission was triggered, and there is no need to submit a new request; please
review the active request carefully to confirm it does in fact include the
intended content.

Otherwise, if there is no active request containing the desired updated package
content, a new request needs to be submitted for the target stream, it is OK to
supersede any existing requests if prompted to confirm doing so.
```bash
$ osc submitrequest systemsmanagement:SCC suseconnect-ng openSUSE:Factory
```

### Monitor OBS submissions

Check the status of any OBS submit requests using
[here](https://build.opensuse.org/package/requests/systemsmanagement:SCC/suseconnect-ng).

Monitor the submission and address any issues identified until the submission is
accepted.

**Note**: If any patches need to be applied, please ensure that any changes
to the suseconnect-ng.changes file are replicated to the
[SUSE/connect-ng/build/packaging/suseconnect-ng.changes](build/packaging/suseconnect-ng.changes)
file.

## Step 8. Verify updated package in the Internal Build Service

Once the OBS openSUSE:Factory submission has been accepted the focus switches
to IBS, where the
[Devel:SCC:suseconnect/suseconnect-ng](https://build.suse.de/package/show/Devel:SCC:suseconnect/suseconnect-ng)
package is a remote link to the
[OBS systemsmanagement:SCC/suseconnect-ng](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng).

**Note**: Do not submit directly to the IBS package; updates should
always be submitted to the OBS package and will get mirrored to IBS.

Verify that the IBS package has been updated to match OBS, and that
the package builds successfully for all relevant code streams,
including the older SLE 12 and SLE 15 streams; required go1.xx-openssl
packages are available in IBS so packages for all active code streams
are expected to build successfully.

**Note**: If the IBS package hasn't been updated, verify that it still
shows as being a remote link to `openSUSE.org:systemsmanagement:SCC/suseconnect-ng`
in the upper right section of the
[Devel:SCC:suseconnect/suseconnect-ng page](https://build.suse.de/package/Devel:SCC:suseconnect/suseconnect-ng).
If not, this needs to be rectified; ask for help from the rest of the
team if not sure what to do in order to restore the remote link.

Consider also downloading and verifying that the IBS built RPMs perform
as expected for a subset of the relevant code streams. See
[below](#testing-built-packages-in-a-local-container) for details,
in particular the sections on
[downloading IBS built packages](#setup-using-packages-built-in-ibs)
to [test them in a local container](#start-a-local-container).

## Step 9. Submit maintenance updates to SLES

### Determine target code streams to submit maintenance updates against

Consult [smelt's suseconnect-ng maintained page](https://smelt.suse.de/maintained/?q=suseconnect-ng)
to determine the set of code streams that suseconnect-ng is actively
maintained for, as determined by active SLE releases and inheritance
of built packages in later releases from earlier releases.

Alternatively, from the CLI, the following command can be used:

```bash
osc -A https://api.suse.de maintained suseconnect-ng
```

**Notes**:
  * The CLI command currently reports SLE 15 SP1 as still being maintained,
    whereas the smelt page doesn't, and doesn't list SLFO 1.2, so it is
    recommended to use the smelt page.
  * If in doubt, submit against the code stream; if it is no longer active
    the submission will be rejected.
  * SLE 12 and 15 code streams may be active even if that specific SLE
    release has reached [end of life](https://www.suse.com/lifecycle/) if
    a later active stream inherits a package from that build; for example
    the SLE 12 SP5 release inherits the suseconnect-ng package from the
    SLE 12 SP2 code stream, so SLE 12 SP2 is the active code stream.

### Submit maintenance updates

For each maintained code stream create a new IBS maintenance or submit request,
or an equivalent src.suse.de submission for SLE 16 streams.

#### SLE 12 and SLE 15 update requests

The following should be the complete list of requests done via osc,
Please update this list as releases are added or removed:

```bash
osc -A https://api.suse.de mr Devel:SCC:suseconnect suseconnect-ng SUSE:SLE-12-SP2:Update # for SLE 12 SP5
osc -A https://api.suse.de mr Devel:SCC:suseconnect suseconnect-ng SUSE:SLE-15-SP4:Update # and SLE Micro 5.3 & 5.4
osc -A https://api.suse.de mr Devel:SCC:suseconnect suseconnect-ng SUSE:SLE-15-SP5:Update # and SLE Micro 5.5
osc -A https://api.suse.de mr Devel:SCC:suseconnect suseconnect-ng SUSE:SLE-15-SP6:Update # and SLE 15 SP7

osc -A https://api.suse.de sr Devel:SCC:suseconnect suseconnect-ng SUSE:ALP:Source:Standard:1.0 # for SL Micro 6.0
osc -A https://api.suse.de sr Devel:SCC:suseconnect suseconnect-ng SUSE:SLFO:1.1 # for SL Micro 6.1
```

**Notes:**

* When asked whether or not to supersede a request, the answer is usually "no".
  Saying "yes" would overwrite the previous request made, cancelling the release
  process for its code stream.

* The codes streams of SLE-15-SP1, SLE-15-SP2 and SLE-15-SP3 are connected,
  meaning that a submission is only needed for SP1 and it will get released
  on SP2 and SP3 also.

* The codes streams that are not yet on final codefreeze (alphas or betas) work
  with submit requests rather than maintenance requests.

* SLE Micro <= 5.5 inherits MicroOS and SLE Releases.

* `submitrequest` (`sr`)'s should auto-translate to `maintenancerequest` (`mr`)'s!

* In case the `sr` (submit request) command is not working properly, try
  `mr` (maintenance request) command. If a maintenance request is not
  applicable, the maintainers will notify you in the request.

#### SLFO 1.0 (ALP) and 1.1 update requests

For these code streams use a `submitrequest` rather than a `maintenancerequest`.

#### SLE 16.0 (SLFO 1.2) and later maintenance updates

For SLE 16.0 and later code streams, follow the
[Git Workflow documenmtation](https://src.suse.de/SUSE/git-workflow-documentation)
to submit a PR to add an updated suseconnect-ng package to the relevant code stream
branch, e.g. `slfo-1.2`.

See [suseconnect-ng PR#1](https://src.suse.de/pool/suseconnect-ng/pulls/1) for an
example of a previous submission.

#### RES8 Submissions

For RES8 (SUSE Linux Enterprise Server with Expanded Support), package updates
need to get done by EPAM (contact is: res-coord@suse.de). We agreed to only push
critical security updates there.

### Monitoring code stream submissions

Check the status of IBS requests
[here](https://build.suse.de/package/requests/Devel:SCC:suseconnect/suseconnect-ng).

Similarly check that status of SLE 16 submissions via the PRs that were submitted,

Monitor the submissions and address any issues identified for a submission until it
is accepted.

**Note**: If any patches need to be applied, please ensure that any changes
to the suseconnect-ng.changes file are replicated to the
[SUSE/connect-ng/build/packaging/suseconnect-ng.changes](build/packaging/suseconnect-ng.changes)
file.

## Step 10. Create the GitHub Release for the version tag
This not be confused with [Step 4. Tagging the release](#step-4-tagging-the-release). This step creates the release
artifact in [SUSE/connect-ng](github.com/SUSE/connect-ng) repo on GitHub.


Once the new suseconnect-ng release has been successfully submitted, and any feedback
issues have been addressed, create a GitHub release matching the version tag used for
the maintenance updates.

This done via the GitHub web UI. Navigate to the [SUSE/connect-ng/releases area](https://github.com/SUSE/connect-ng/releases) and
look at one of the existing release tags such as:
[v1.22.1](https://github.com/SUSE/connect-ng/releases#release-v1.22.1)
to see the desired format and content.

The release is created via the "Draft a new release" button or directly from:
["Draft a new release" button](https://github.com/SUSE/connect-ng/releases/new)
From "Tag: Select tag" pull down select the git tag created in Step 4 (i.e. 'v1.22.1')

For Release Title: This is a text entry box and should be filled in with the git tag value for this release (i.e. 'v1.22.1')

The contents of the "Release Notes" text box are created by copying the changes list from the [suseconnect-ng.changes entry associated with the tag](https://github.com/SUSE/connect-ng/blob/main/build/packaging/suseconnect-ng.changes)
start the copy from "- Update version to.." and end before next release seperator in the changelog file.

Change "- Update version" to "Version"

Click "Preview" to view what the Release Tag will look like. If all looks good, click "Publish Release"

## Step 11. Update the version on the main branch

Once the tagged version of suseconnect-ng has been successfully released, the
version in the [suseconnect-ng.spec](build/packaging/suseconnect-ng.spec) file
on the main branch should be updated to the next minor release, e.g. `X.Y.Z` =>
`X.Y+1.0`.

A matching placeholder entry should also be added to the top of the
[suseconnect-ng.changes](build/packaging/suseconnect-ng.changes) file as
follows, replacing `X.Y` with the appropriate next version:

```
-------------------------------------------------------------------
Wed May 13 18:22:48 UTC 2026 - Fergal Mc Carthy <fmccarthy@suse.com>

- Update version to X.Y (pre):
  -

```

The `osc vc` cammand can help with this as follows:

```bash
$ cd /path/to/clone/of/connect-ng
$ osc vc build/packaging/suseconnect-ng.changes
# add placeholder version text and save the file
```

Further updates to the changelog can handled via normal file editing; only
use `osc vc` to create the initial placeholder entry for a given version.

## Step 12. Tracking submissions

The maintenance update requests created in Step 9, will result in a number of URLs that can be
used for tacking the state of the requests. For example, for `1.22.1`:

| URL                                            | Stream                      |
| -----------------------------------------------| ----------------------------|
| https://src.suse.de/pool/suseconnect-ng/pulls/4| slfo-1.2 (SLE 16, Micro 6.2)|
| https://src.suse.de/pool/suseconnect-ng/pulls/3| slfo-main|
| https://build.suse.de/request/show/414112| SUSE:SLFO:1.1 (Micro 6.1)|
| https://build.suse.de/request/show/414111| SUSE:ALP:Source:Standard:1.0 (ALP, Micro 6.0)|
| https://build.suse.de/request/show/414100| SUSE:SLE-15-SP6:Update|
| https://build.suse.de/request/show/414110| SUSE:SLE-15-SP5:Update|
| https://build.suse.de/request/show/414109| SUSE:SLE-15-SP4:Update|
| https://build.suse.de/request/show/414107| SUSE:SLE-12-SP2:Update|

Requests to src.suse.de will show as "Manually merged" and build.suse.de requests will show as "accepted".

## Step 12. Released Status

Once package updates have been accepted/merged, the [SCC](https://scc.suse.com) can be used to determine update package is available to users.
| URL                                            | Stream                      |
| --------------------------------------------------------------------------------------------------------------------------| -----|
| [SLES 16.1](https://scc.suse.com/packages?name=SUSE%20Linux%20Enterprise%20Server&version=16.1&arch=x86_64&query=suseconnect-ng&module=)| 16.1|
| [SLES 16.0](https://scc.suse.com/packages?name=SUSE%20Linux%20Enterprise%20Server&version=16.0&arch=x86_64&query=suseconnect-ng&module=)| 16.0|
| [SLES 15.7](https://scc.suse.com/packages?name=SUSE%20Linux%20Enterprise%20Server&version=15.7&arch=x86_64&query=suseconnect-ng&module=)| 15.7|
| [SLES 15.6](https://scc.suse.com/packages?name=SUSE%20Linux%20Enterprise%20Server&version=15.6&arch=x86_64&query=suseconnect-ng&module=)| 15.6|
| [SLES 15.5](https://scc.suse.com/packages?name=SUSE%20Linux%20Enterprise%20Server&version=15.5&arch=x86_64&query=suseconnect-ng&module=)| 15.5|
| [SLES 15.4](https://scc.suse.com/packages?name=SUSE%20Linux%20Enterprise%20Server&version=15.4&arch=x86_64&query=suseconnect-ng&module=)| 15.4|
| [SLES 12.5](https://scc.suse.com/packages?name=SUSE%20Linux%20Enterprise%20Server&version=12.5&arch=x86_64&query=suseconnect-ng&module=)| 12.5|

If [scc.suse.com package search](https://scc.suse.com/packages) does not show the desired version as released, then [search suseconnect-ng in smelt](https://smelt.suse.de/search/?q=suseconnect-ng) to find the
expected release date. Click on "Created" until the "suseconnect-ng" submissions are in decending order by creation date. This
should show a number of "active" state suseconnect-ng packages. Clicking on an
entry's `ID` link will bring up a page showing "Planned release date" and
"Scheduled release date" fields providing useful information as to when the
package update will be released.

# Testing the package locally

As mentioned in [Step 3 above](#step-3-verify-the-package-builds-locally)
there are Makefile targets available, such as `build-rpm`,
that can be used to verify that the RPM package builds, and basic verification
testing is possible using the `feature-tests` target.

Perform basic validation tests of the RPM package building process, and the
built RPMs using some of the available [Makefile](Makefile) targets, such as
`build-rpm` or `feature-tests`.

## Building locally with osc

RPMs can also be built locally by checking out
[OBS systemsmanagement:SCC/suseconnect-ng](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng),
or [IBS Devel:SCC:suseconnect/suseconnect-ng](https://build.suse.de/package/Devel:SCC:suseconnect/suseconnect-ng),
and then building with the appropriate `osc build` command.

### Preparing local build inputs

The following inputs are needed when building the candidate codebase locally
using `osc`:

* suseconnect-ng-X.Y.Z.tar.xz
  * Generated by running `make dist`
* vendor.tar.xz
  * Also generated by running `make dist`
* [suseconnect-ng.spec](build/packaging/suseconnect-ng.spec)
* [suseconnect-ng.changes](build/packaging/suseconnect-ng.changes)
* [suseconnect-ng-rpmlintrc](build/packaging/suseconnect-ng-rpmlintrc)

These will need to be copied into the checked out package before building it.

### Checking out and building locally with OBS package

To just test build locally with the OBS package, use the following commands:

```bash
$ osc co systemsmanagement:SCC/suseconnect-ng
$ cd systemsmanagement:SCC/suseconnect-ng
$ rm suseconnect-ng*.tar.xz  # remove old suseconnect-ng tarball
$ cp /path/to/clone/of/connect-ng/suseconnect-ng-X.Y.Z.tar.xz .
$ cp /path/to/clone/of/connect-ng/vendor.tar.xz .
$ cp /path/to/clone/of/connect-ng/build/packaging/suseconnect-ng* .
$ osc build SLE_15_SP6  # or whichever code stream you want to test
```

The build RPMs will be available in the
`/var/tmp/build-root/${CODE_STREAM}-${ARCH}/home/abuild/rpmbuild/RPMS/${ARCH}/`
directory.

### Checking out and building locally with IBS package

To just test build locally with the OBS package, use the following commands:

```bash
$ osc -A https://api.suse.de co Devel:SCC:suseconnect/suseconnect-ng
$ cd Devel:SCC:suseconnect/suseconnect-ng
$ rm suseconnect-ng*.tar.xz  # remove old suseconnect-ng tarball
$ cp /path/to/clone/of/connect-ng/suseconnect-ng-X.Y.Z.tar.xz .
$ cp /path/to/clone/of/connect-ng/vendor.tar.xz .
$ cp /path/to/clone/of/connect-ng/build/packaging/suseconnect-ng* .
$ osc -A https://api.suse.de build SLE_15_SP6_Update  # or whichever code stream you want to test
```

The build RPMs will be available in the
`/var/tmp/build-root/${CODE_STREAM}-${ARCH}/home/abuild/rpmbuild/RPMS/${ARCH}/`
directory.

### Branching, Commiting and Building in OBS

To test building the package in OBS it is recommended to create a personal
branch of
[systemsmanagement:SCC/suseconnect-ng](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng)
and then commit the test build inputs to that area:

```bash
$ osc bco systemsmanagement:SCC/suseconnect-ng
$ cd home:${OBS_USER}:branches:systemsmanagement:SCC/suseconnect-ng
$ rm suseconnect-ng*.tar.xz  # remove old suseconnect-ng tarball
$ cp /path/to/clone/of/connect-ng/suseconnect-ng-X.Y.Z.tar.xz .
$ cp /path/to/clone/of/connect-ng/vendor.tar.xz .
$ cp /path/to/clone/of/connect-ng/build/packaging/suseconnect-ng* .
$ osc build SLE_15_SP6_Update  # or whichever code stream you want to test
$ osc addremove  # to update the index of files to be committed
$ osc commit
$ osc results  # repeat to check the status of builds, or use --watch to wait
```

Once the builds have completed the built packages can be downloaded for the
desired code stream (`${CODE_STREAM}`) and architecture (${ARCH}) using
`osc getbinaries` within the checked out package directory as follows:

```bash
$ mkdir /tmp/obs_suseconnect_rpms
$ osc results ${CODE_STREAM} ${ARCH}  # confirm that the packages have built
$ osc getbinaries -d /tmp/obs_suseconnect_rpms ${CODE_STREAM} ${ARCH}
```

### Branching, Commiting and Building in IBS

To test building the package in IBS it is recommended to create a personal
branch of
[IBS Devel:SCC:suseconnect/suseconnect-ng](https://build.suse.de/package/Devel:SCC:suseconnect/suseconnect-ng)
and then commit the test build inputs to that area:

```bash
$ osc -A https://api.suse.de bco Devel:SCC:suseconnect/suseconnect-ng
$ cd home:${IBS_USER}:branches:Devel:SCC:suseconnect/suseconnect-ng
$ rm suseconnect-ng*.tar.xz  # remove old suseconnect-ng tarball
$ cp /path/to/clone/of/connect-ng/suseconnect-ng-X.Y.Z.tar.xz .
$ cp /path/to/clone/of/connect-ng/vendor.tar.xz .
$ cp /path/to/clone/of/connect-ng/build/packaging/suseconnect-ng* .
$ osc -A https://api.suse.de build SLE_15_SP6_Update  # or whichever code stream you want to test
$ osc -A https://api.suse.de addremove  # to update the index of files to be committed
$ osc -A https://api.suse.de commit
$ osc -A https://api.suse.de results  # repeat to check the status of builds, or use --watch to wait
```

Once the builds have completed the built packages can be downloaded for the
desired code stream (`${CODE_STREAM}`) and architecture (${ARCH}) using
`osc getbinaries` within the checked out package directory as follows:

```bash
$ mkdir /tmp/obs_suseconnect_rpms
$ osc -A https://api.suse.de results ${CODE_STREAM} ${ARCH}  # confirm that the packages have built
$ osc -A https://api.suse.de getbinaries -d /tmp/obs_suseconnect_rpms ${CODE_STREAM} ${ARCH}
```

## Performing basic suseconnect feature validations

Verify basic operation of the built RPM packages using the feature tests
by running `make feature-tests` which will build the RPM packages in a
local container environment, then install them, along with any dependencies,
within that container, and then run the feature tests specified in the
[features/suseconnect/](features/suseconnect/) directory, which will
exercise the installed `suseconnect` binary.

## Adhoc testing of RPMs with the CI env

To perform adhoc testing with built RPMs the Makefile `ci-env` target can be
used to launch an appropriate local container environment that mounts the
development repo as /usr/src/connect-ng, and makes it the active working
directory.

The RPMs can be built using the `build/ci/build-rpm` script.

The built RPMs can be installed using the `build/ci/configure` script.

The feature-tests can be run using the `build/ci/run-tests` script or
manually using `go test -v features/suseconnect/*` or a subset of the
feature tests can be run by specifying just the specify file under
[features/suseconnect/](features/suseconnect/) or using normal `go test`
selection methods.

Manual testing of the `suseconnect` command can be performed using the
installed command.

## Testing built packages in a local container

For locally built `suseconnect-ng` packages using `osc build`,
copy them to a directory such as `/tmp/suseconnect-rpms`. See local
[OBS](#checking-out-and-building-locally-with-obs-package) and
[IBS](#checking-out-and-building-locally-with-ibs-package) build
instructions above.

For RPMs built in the build services, see the above sections on
building in [OBS](#branching-commiting-and-building-in-obs) and
[IBS](#branching-commiting-and-building-in-ibs), and then download
the rpms to the `/tmp/suseconnect-rpms` directory.

### Setup using locally built RPMs

Build the RPMs locally using `osc build` locally and copy them to a directory, e.g.
`/tmp/suseconnect-rpms`.

For OBS use:

```bash
$ cd /path/to/systemsmanagement:SCC/suseconnect-ng
$ osc build SLE_15_SP6 x86_64
$ cp /var/tmp/build-root/SLE_15_SP6-x86_64/home/abuild/rpmbuild/RPMS/*suseconnect*.rpm /tmp/suseconnect-rpms
```

For IBS use:

```bash
$ cd /path/to/Devel:SCC:suseconnect/suseconnect-ng
$ osc build SLE_15_SP6_Update x86_64
$ cp /var/tmp/build-root/SLE_15_SP6_Update-x86_64/home/abuild/rpmbuild/RPMS/*suseconnect*.rpm /tmp/suseconnect-rpms
```

### Setup using packages built in OBS

Download the packages built in
[OBS systemsmanagement:SCC/suseconnect-ng](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng)
to `/tmp/suseconnect-rpms`:

```bash
$ osc results systemsmanagement:SCC suseconnect-ng SLE_15_SP6 x86_64  # confirm packages have built successfully
$ osc getbinaries -d /tmp/suseconnect-rpms systemsmanagement:SCC suseconnect-ng SLE_15_SP6 x86_64
```

Or to use RPMs built in a personal branch:

```bash
$ osc results home:${OBS_USER}:branches:systemsmanagement:SCC suseconnect-ng SLE_15_SP6 x86_64  # confirm packages have built successfully
$ osc getbinaries -d /tmp/suseconnect-rpms home:${OBS_USER}:branches:systemsmanagement:SCC suseconnect-ng SLE_15_SP6 x86_64
```

### Setup using packages built in IBS

Download the packages built in
[IBS Devel:SCC:suseconnect/suseconnect-ng](https://build.suse.de/package/Devel:SCC:suseconnect/suseconnect-ng)
package to `/tmp/suseconnect-rpms`.

```bash
$ osc -A https://api.suse.de results Devel:SCC:suseconnect suseconnect-ng SLE_15_SP6_Update x86_64  # confirm packages have built successfully
$ osc -A https://api.suse.de getbinaries -d /tmp/suseconnect-rpms Devel:SCC:suseconnect suseconnect-ng SLE_15_SP6_Update x86_64
```

Or to use RPMs built in a personal branch:

```bash
$ osc -A https://api.suse.de getbinaries -d /tmp/suseconnect-rpms home:${IBS_USER}:branches:Devel:SCC:suseconnect suseconnect-ng SLE_15_SP6_Update x86_64
```

### Start a local container

Use `podman` or `docker` to start a SLE BCI or similar container, matching the distro
release and architecture that the candidate packages were built for. The important
thing is that the `/tmp/suseconnect-rpms` directory needs to be mounted into the
container runtime environment.

```bash
$ podman run --rm -it --privileged -v /tmp/suseconnect-rpms:/rpms registry.suse.com/bci/bci-base:15.6
a1ed05dd1bbb:/ # zypper --no-gpg-checks in /rpms/*.rpm
a1ed05dd1bbb:/ # suseconnect --version
1.22.0
```

# The CI

This project makes use of the feature from OBS in which you can build a package
from a given PR. This way we check that your changes don't break the packaging
side of `connect-ng`.

You can see the configuration on this on
[.obs/workflows.yml](.obs/workflows.yml). Right now this is taking place inside
of a personal project, but this is to be changed (see [this Jira
ticket](https://jira.suse.com/browse/CSD-79)).

# Optional Release or Patch/Hotfix Branching Strategy

By using a release branching strategy similar to that outlined in this section
it should be possible to avoid needing to pause the normal development workflow
while preparing a release for submission, in case there is a need to fix issues
that come up during the openSUSE:Factory or SLE code streams submission process,
or during any of the associated QA testing.

This strategy will also support creating patch or hotfix updates for the currently
active release, without having to impact the main branch development workflow.

## Create a release branch based on release tag

If an issue arises that requires changes to the release being submitted,
a release branch named for the major and minor parts of the tag version,
e.g. `release/X.Y`, can be created from the associated release tag and
updates can be developed on that release branch, as shown in this
illustration:

```
                             vX.Y.0 tag
          vX.Y-1.0 tag       |
          |                  |
main -----+------------------+--------------------
                             \
                              \
                   release/X.Y +---+-------+
                                   |       |
                                   |       vX.Y.2 tag
                                   |
                                   vX.Y.1 tag
```

In the case of an ongoing vX.Y.0 tagged release submission for vX.Y.0, where
a bug is reported by QA, the release branch can be created as follows:

```bash
$ cd /path/to/clone/of/connect-ng
$ git fetch --all  # fetch latest repo updates
$ git checkout -b release/X.Y vX.Y.0
```

## Setup the release branch for the next patch level

Once the release branch has been created the version in the
[suseconnect-ng.spec](build/packaging/suseconnect-ng.spec)
on that branch should be updated to next patch release version, e.g.
`X.Y.0` => `X.Y.1`.

A matching placeholder entry for the new version should be added to
the top of the
[suseconnect-ng.changes](build/packaging/suseconnect-ng.changes) file.
See [Step 11](#step-11-update-the-version-on-the-main-branch) above
for how the `osc vc` command can be used to assist in creating the
release branch placeholder changelog entry.

## Develop and Test changes as normal on the release branch

Whether developing new changes on the release branch, or backporting fixes
from the main branch, the process will generally be the same as the normal
development workflow:

* Propose one or more PRs targeting the release branch with desired changes
  * For newly developed changes remember to also propose equivalent versions
    on the main branch if relevant.
  * Updates should include change log entries in the
    [suseconnect-ng.changes](build/packaging/suseconnect-ng.changes) file.
* PRs should be reviewed and tested for correctness before merging.

## Tag the updated release branch with the new version when ready

Once the updates to the release branch are ready, a new tag, e.e. `vX.Y.1`,
can be created for the release branch's current version, and the normal
package submission process can be initiated for that new tag.

For example:

```bash
$ cd /path/to/clone/of/connect-ng
$ git fetch --all  # fetch latest repo updates
$ git checkout release/X.Y
$ git status   # check if branch is out of date
$ git merge --ff  # merge in pending updates if needed
$ git tag vX.Y.1  # or whatever the desired patch level version is
$ git push origin vX.Y.1  # push the tag up to GitHub
```

## Repeat as needed for subsequent release updates

If further issues arise during the submission process, or additional critical
updates are identified, then repeat the process of:

  * bumping the version patch level in the
    [suseconnect-ng.spec](build/packaging/suseconnect-ng.spec) file
  * adding a matching entry in the
    [suseconnect-ng.changes](build/packaging/suseconnect-ng.changes) file.
  * developing changes on the release branch or backportingt them from `main`
  * creating the new release tag when ready
  * submitting the updated release.

## IMPORTANT: Ensure release branch changes are propagated to the main branch

It is critcally important that any changes made on the `release/X.Y` branch
are matched by equivalent changes on the main branch. The recommendation is
to do this promptly as each new release tag is created, rather than waiting
until the release submissions are accepted or published, but that may depend
on circumstances.

Running `git log vX.Y.0..release/X.Y` will show the log of the changes
that have been made on the release branch. Include a `-p` option to see the
detailed code base modification for each log entry.

For each of the changes shown ensure that they are reflected appropriately on
the `main` branch:

* For each additional release version created on the `release/X.Y` branch,
  ensure that a matching changelog entry for the release branch's
  [suseconnect-ng.changes](build/packaging/suseconnect-ng.changes) file
  is added to below the current top of file placeholder entry on the `main`
  branch, **exactly** matching the text of the entry on the `release/X.Y`
  branch. Remember that newer version entries should be added above older
  version entries.

* If a change has been backported from the `main` branch then the associated
  code changes, or an equivalent implementation of them, will already exist
  on `main`. However, any changelog entry for those changes on `main` will
  also need to be moved from the placeholder changelog entry at the top of
  the [suseconnect-ng.changes](build/packaging/suseconnect-ng.changes) file
  to the corresponding release branch version changelog entry that is ported
  over from the `release/X.Y` branch.

* If a change has been developed on the `release/X.Y` branch then it will
  need to be ported to the `main` branch, either as a direct cherry-pick
  or as an equivalent code change if the underlyng code has changed.
  Ensure that any changelog entry for it is reflected in the corresponding
  changelog entry ported from the `release/X.Y` branch to the
  [suseconnect-ng.changes](build/packaging/suseconnect-ng.changes) file
  on the `main` branch.

## IMPORTANT: Ensure consistency of changelog on main branch

Once all the changes on the `release/X.Y` branch have been incorportated
into the main branch ensure that there are no discrepancies between the
[suseconnect-ng.changes](build/packaging/suseconnect-ng.changes) file on
the `main` and `release/X.Y` branches.

Running `git diff release/X.Y main -- build/packaging/suseconnect-ng.changes`
should show a single set of differences at the top of the file, matching
the placeholder entry for the next release. If additional differences are
shown, please correct them on the `main` branch.

**Notes**:

* Even a minor difference in whitespacing within the historical
  part of the [suseconnect-ng.changes](build/packaging/suseconnect-ng.changes)
  file can lead to a rejection of a submission to the SLE code streams

* If a divergence is somehow released within the historical parts of the
  .changes file between different code streams this will require **manual**
  effort to address for __*every subsequent release submission to that code
  stream*__.

## Create GitHub Releases for any additional release branch tags created

For any additional releases versions that are tagged on the release branch,
create corresponding GitHub releases for them, using the change log from
that version's entry in the
[suseconnect-ng.changes](build/packaging/suseconnect-ng.changes) file.

