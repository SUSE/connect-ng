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
need to checkout the package and run the service that fetches the code (change
the `revision` parameter in `_service` to build the tar from another branch/sha
than `main`):

```bash
$ mkdir obs; cd obs; osc co systemsmanagement:SCC suseconnect-ng -o .
$ osc service manualrun
```

This will create a `connect-ng` directory with the latest changes from git's
`main` branch. After this, you can apply whatever changes you want inside of
the `connect-ng` and commit it.

If you have the source code somewhere else, you have two ways to go. First of
all, you can produce a patch by using `git diff` on your locally cloned
repository and then apply it into the source code that was downloaded when you
first run `osc service manualrun`. Then, you can run `osc service manualrun` in
order to produce the tar again and build it.

Otherwise, you can simply push your changes into a remote development branch.
Then, inside of your local copy of the OBS project, change the `revision`
parameter inside of the `_service` file to match the branch name or the commit
sha you are testing. After that, you can continuously run `osc service
manualrun` and `osc build ...` in order to test multiple iterations of your
development branch.

Either way you will give get an RPM that you can install locally.

## Step 3. Update package in OBS devel project

As with any other OBS project, whenever you are done with testing a package, you
can update the package as usual:

1. Make sure that the `revision` on the `_service` file is back on `main`
   (upstream should always point to `main`).
2. Make sure to run `osc service manualrun` and `osc build ...` again and that
   everything is working.
3. Add/remove files with the `add` and `remove` commands (tip: you can also use
   the `addremove` command to automate this!).
4. Update the changelog with the `vc` command and call `commit`.

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
