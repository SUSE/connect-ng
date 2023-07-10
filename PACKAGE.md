# Building the package
## TL;DR

The package is available at `obs://systemsmanagement:SCC/suseconnect-ng` and it
can be fetched with `osc` in the usual way (or you can also copy the `_service`
file from git and recreate the package with `osc service manualrun`).

## Step 1. Update version

The version of `connect-ng` is a mix of a version specified in the `.spec` file
and the git commit sha. The version is thus updated automatically on every git
commit. On a service run, the latest git tag gets copied as version to the .spec file.

## Step 2. The OBS package

The RPM is built from OBS, and you need to access the main repository from
there:
[systemsmanagement:SCC/suseconnect-ng](https://build.opensuse.org/package/show/systemsmanagement:SCC/suseconnect-ng).
From this repository, you can build the package as usual with:

```bash
$ osc build openSUSE_Leap_15.4 x86_64 --no-verify
$ osc build SLE_15_SP3 x86_64 --no-verify
```

### Testing the package locally

First of all, you need to create an updated tar file. In order to do this, you
need to checkout the package and run the service that fetches the code (change the
`revision` parameter in `_service` to build the tar from another branch than `main`):

```bash
$ mkdir obs; cd obs; osc co systemsmanagement:SCC suseconnect-ng -o .
$ osc service manualrun
```

This will create a `connect-ng` directory with the latest changes from git's
default branch. After this, you can apply whatever changes you want inside of
the `connect-ng` and commit it. If you have the source code somewhere else, you
can simply create a patch file from there, apply it and commit the changes
locally on this newly cloned repository. After all that, just run the service
again and you will get an updated tar file. That is:

```bash
# Optional: produce the patch file with `git diff` or `git show` from your
# development repository if it's located somewhere else and apply it on the
# `suseconnect-ng` directory you have from the previous `service manualrun`
# command.

# On the local copy of the package.
$ osc service manualrun
$ osc build openSUSE_Leap_15.4 x86_64 --no-verify
```

This will give you an RPM you can install locally.

## Step 3. Update package in OBS devel project

The package is updated manually and it relies on the Git repository from
`connect-ng` to contain the latest changes. Whenever you want to update the
package on OBS, simply run the service and commit the changes like so:

```bash
$ osc service manualrun
# review the changes
$ osc commit
```

## Submit maintenance updates

To get a maintenance request accepted, each changelog entry needs to have at
least one reference to a bug or feature request like `bsc#123` or `fate#123`.

**Note**: If you want to disable automatic changes made by osc (e.g. License
string) use the `--no-cleanup` switch. Can be used for commands like `osc mr`,
`osc sr` and `osc ci`.

### Submit maintenance updates for SLES to the Internal Build Service

#### Get target codestreams where to submit

To checkout in which codestreams the package is currently maintaned, run:

```bash
osc -A https://api.suse.de maintained suseconnect-ng
```

For a more detailed view which target codestreams are in which state, find the
`suseconnect-ng` package on
[maintenance.suse.de](https://maintenance.suse.de/maintained/)

#### Submit updates

For each maintained codestream you need to create a new maintenance request:

```bash
osc -A https://api.suse.de mr openSUSE.org:systemsmanagement:SCC suseconnect-ng SUSE:SLE-15-SP4:GA
```

**Note**: In case the `mr` (maintenance request) command is not working
properly, try `sr` (submit request) command.

**Note:**

* When asked whether or not to supersede a request, the answer is usually "no".
  Saying "yes" would overwrite the previous request made, cancelling the release
  process for its codestream.

For RES8 (SUSE Linux Enterprise Server with Expanded Support), package updates
need to get done by EPAM (contact is: res-coord@suse.de). We agreed to only push
critical security updates there.

You can check the status of your requests
[here](https://build.opensuse.org/package/requests/systemsmanagement:SCC/suseconnect-ng).

Whenever your requests get accepted, they still have to pass maintenance testing
before they get released to customers. You can check their progress by searching
for the package on
[maintenance.suse.de](https://maintenance.suse.de/maintained/). If you still
need help, the maintenance team can be reached at
[maint-coord@suse.de](maint-coord@suse.de) or `#discuss-maintenance` on Slack.

# The CI

This project makes use of the feature from OBS in which you can build a package
from a given PR. This way we check that your changes don't break the packaging
side of `connect-ng`.

You can see the configuration on this on
[.obs/workflows.yml](.obs/workflows.yml). Right now this is taking place inside
of a personal project, but this is to be changed (see [this Jira
ticket](https://jira.suse.com/browse/CSD-79)).
