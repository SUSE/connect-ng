#!BuildTag: rmt-ci-container
#!UseOBSRepositories
FROM registry.suse.com/suse/sle15:15.5

RUN zypper --non-interactive install go1.21-openssl ruby2.5 ruby2.5-devel ruby2.5-rubygem-bundler \
                                     ruby-devel vim curl wget gzip tar xz rpmbuild nodejs dmidecode

# Set default bundler to the required version. This way we can easily update ruby in case required
RUN update-alternatives --install /usr/bin/bundle bundle /usr/bin/bundle.ruby2.5 5 && \
    update-alternatives --install /usr/bin/bundler bundler /usr/bin/bundler.ruby2.5 5

RUN mkdir /usr/src/connect-ng
WORKDIR /usr/src/connect-ng
