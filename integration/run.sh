#!/bin/sh -xe
if [ -z $PRODUCT ]
then
  echo "PRODUCT env var for integration testing not set!"
  exit 1
fi

shortversion=$(git describe --tags --abbrev=0 --match='v*' | sed 's:^v::' )

if true; then

remote=$(pwd)
head=$(git rev-list -n1 HEAD)
mkdir osc-package
cd osc-package
osc -A https://api.opensuse.org co -o . 'systemsmanagement:SCC/suseconnect-ng'
cp ../_service ./
sed -i "s|https://github.com/SUSE/connect-ng.git|file://$remote|g" _service
sed -i "s|main|$head|g" _service
TAR_SCM_TESTMODE=1 osc service manualrun
osc build $PRODUCT x86_64 --trust-all-projects --clean
cd ..
zypper --non-interactive install -- -SUSEConnect -zypper-migration-plugin
zypper --non-interactive --no-gpg-checks install /oscbuild/$PRODUCT-x86_64/home/abuild/rpmbuild/RPMS/x86_64/*

else

# this could be used to test an alread built version
zypper ar https://download.opensuse.org/repositories/systemsmanagement:/SCC/SLE_15_SP3/ devel
zypper --non-interactive --gpg-auto-import-keys ref
zypper --non-interactive install suseconnect-ng -SUSEConnect -zypper-migration-plugin

fi

cd /tmp/connect
# cucumber picks up /tmp/connect/bin/SUSEConnect from the earlier bundle install if it is not removed here
rm bin/SUSEConnect
# the tests match the exact version defined in the ruby code, replace it with ours
sed -i "s|VERSION = '[^']*'|VERSION = '$shortversion'|g" lib/suse/connect/version.rb
# load test regcodes into env vars
set -a
. ~/.regcodes
set +a
cucumber
