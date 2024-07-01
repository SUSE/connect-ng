NAME          = suseconnect-ng
VERSION       = $(shell bash -c "cat build/packaging/suseconnect-ng.spec | sed -n 's/^Version:\s*\(.*\)/\1/p'")
DIST          = $(NAME)-$(VERSION)/
PWD           = $(shell pwd)
GO            = go
OUT           = -o out/
GOFLAGS       = -v -mod=vendor
BINFLAGS      = -buildmode=pie
SOFLAGS       = -buildmode=c-shared

CONTAINER     = registry.suse.com/bci/golang:1.21-openssl
CRM           = docker run --rm -it --privileged
ENVFILE       = .env
WORKDIR       = /usr/src/connect-ng
MOUNT         = -v $(PWD):$(WORKDIR)

.PHONY: dist build clean ci-env build-rpm feature-tests format vendor

all: clean build test

dist: clean internal/connect/version.txt vendor
	@mkdir -p $(DIST)/build/packaging
	@cp -r internal $(DIST)
	@cp -r third_party $(DIST)
	@cp -r cmd $(DIST)
	@cp -r docs $(DIST)
	@cp go.mod $(DIST)
	@cp go.sum $(DIST)
	@cp LICENSE LICENSE.LGPL README.md $(DIST)
	@cp SUSEConnect.example $(DIST)
	@cp build/packaging/suseconnect-keepalive* $(DIST)/build/packaging
	@cp -r build/packaging/suseconnect-ng* $(DIST)

	@tar cfvj vendor.tar.xz vendor
	@tar cfvj $(NAME)-$(VERSION).tar.xz $(NAME)-$(VERSION)/

	@rm -r $(NAME)-$(VERSION)


vendor:
	@$(GO) mod download
	@$(GO) mod verify
	@$(GO) mod vendor

out:
	mkdir -p out

internal/connect/version.txt:
	@echo "$(VERSION)" > internal/connect/version.txt

build: clean out internal/connect/version.txt
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/zypper-migration
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/zypper-search-packages
	$(GO) build $(GOFLAGS) $(SOFLAGS) $(OUT) github.com/SUSE/connect-ng/third_party/libsuseconnect

# This "arm" means ARM64v8 little endian, the one being delivered currently on
# OBS.
build-arm: clean out internal/connect/version.txt
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect

build-s390: clean out internal/connect/version.txt
	GOOS=linux GOARCH=s390x $(GO) build $(GOFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect

build-ppc64le: clean out internal/connect/version.txt
	GOOS=linux GOARCH=ppc64le $(GO) build $(GOFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect

test: internal/connect/version.txt
	$(GO) test ./internal/* ./cmd/suseconnect

ci-env:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(CONTAINER) bash

check-format:
	@gofmt -l internal/* cmd/*

build-rpm:
	$(CRM) $(MOUNT) -w $(WORKDIR) $(CONTAINER) bash -c 'build/ci/build-rpm'

feature-tests:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(CONTAINER) bash -c 'build/ci/build-rpm && build/ci/configure && build/ci/run-feature-tests'

test-yast: build-so
	docker build -t go-connect-test-yast -f third_party/Dockerfile.yast . && docker run -t go-connect-test-yast

clean:
	go clean
	@rm -f internal/connect/version.txt
	@rm -rf connect-ng-$(NAME)-$(VERSION)/
	@rm -rf vendor.tar.xz
	@rm -f connect-ng-*.tar.xz
