NAME          = suseconnect-ng
VERSION       = $(shell bash -c "cat build/packaging/suseconnect-ng.spec | sed -n 's/^Version:\s*\(.*\)/\1/p'")
DIST          = $(NAME)-$(VERSION)/
PWD           = $(shell pwd)
GO            = go
OUT           = -o out/
# coverage testing enabled by default
COVERAGE     ?= true
COVERAGE_DIR  = $(PWD)/coverage
COVERAGE_UNIT = $(COVERAGE_DIR)/unit
GOFLAGS       = -v -mod=vendor
BINFLAGS      = -buildmode=pie
SOFLAGS       = -buildmode=c-shared

CONTAINER     = registry.suse.com/bci/golang:1.24-openssl
SLE15SP6CONTAINER = registry.suse.com/bci/bci-base:15.6
SLE16CONTAINER = registry.suse.com/bci/bci-base:16.0
CRM           = docker run --rm -it --privileged
ENVFILE       = .env
WORKDIR       = /usr/src/connect-ng
MOUNT         = -v $(PWD):$(WORKDIR)

define go-tool-covdata
	@if [ -z "$(strip $(1))" ]; then \
		echo "ERROR: no go tool covdata action specified."; \
		exit 1; \
	fi
	@if $(if $(filter true,$(strip $(COVERAGE))),true,false); then \
		$(GO) tool covdata $(1) -i=$(COVERAGE_UNIT); \
	fi
endef

define cover-test-flags
	$(if $(filter true,$(strip $(COVERAGE))),-cover -args -test.gocoverdir=$(COVERAGE_UNIT))
endef

.PHONY: all build build-arm build-ppc64le build-rpm build-s390 check-format
.PHONY: ci-env clean dist feature-tests out show-version test test-yast vendor vet
.PHONY: coverage coverage-check-enabled coverage-dirs coverage-func coverage-percent

all: clean build test

dist: clean internal/connect/version.txt vendor
	@mkdir -p $(DIST)/build/packaging
	@cp -r internal $(DIST)
	@cp -r pkg $(DIST)
	@cp -r third_party $(DIST)
	@cp -r cmd $(DIST)
	@cp -r docs $(DIST)
	@cp go.mod $(DIST)
	@cp go.sum $(DIST)
	@cp LICENSE README.md $(DIST)
	@cp build/packaging/suseconnect-keepalive* $(DIST)/build/packaging
	@cp build/packaging/suse-uptime-tracker* $(DIST)/build/packaging
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
	@echo -n "$(VERSION)" > internal/connect/version.txt

show-version: internal/connect/version.txt
	@cat internal/connect/version.txt

vet: internal/connect/version.txt
	$(GO) vet ./...

build: clean out internal/connect/version.txt
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/zypper-migration
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/zypper-search-packages
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suse-uptime-tracker
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/public-api-demo
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/validate-offline-certificate
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/offline-register-api
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect-mcp
	$(GO) build $(GOFLAGS) $(SOFLAGS) $(OUT) github.com/SUSE/connect-ng/third_party/libsuseconnect

# This "arm" means ARM64v8 little endian, the one being delivered currently on
# OBS.
build-arm: clean out internal/connect/version.txt
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect

build-s390: clean out internal/connect/version.txt
	GOOS=linux GOARCH=s390x $(GO) build $(GOFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect

build-ppc64le: clean out internal/connect/version.txt
	GOOS=linux GOARCH=ppc64le $(GO) build $(GOFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect

coverage-dirs:
	@mkdir -p $(COVERAGE_UNIT)

coverage-check-enabled:
	@if [ -z "$(filter true,$(strip $(COVERAGE)))" ]; then \
		echo "WARNING: Coverage generation and collection not enabled."; \
		echo "To enable it either add COVERAGE=true on the make command line arguments or set COVERAGE=true in the environment."; \
		exit 1; \
	fi

coverage-func: coverage-check-enabled coverage-dirs
	$(call go-tool-covdata,func)

coverage-percent: coverage-check-enabled coverage-dirs
	$(call go-tool-covdata,percent)

coverage: coverage-func

test: internal/connect/version.txt coverage-dirs
	$(GO) test ./internal/* ./cmd/suseconnect ./pkg/* $(call cover-test-flags)
	$(call go-tool-covdata,func)

ci-env:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(CONTAINER) bash

sle15sp6-env:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(SLE15SP6CONTAINER) bash

sle15sp6-migration-check:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(SLE15SP6CONTAINER) bash -c 'build/ci/run-sle-migration --query'

sle15sp6-migration-choice-1:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(SLE15SP6CONTAINER) bash -c 'build/ci/run-sle-migration --migration 1'

sle15sp6-legacy-migration-check:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(SLE15SP6CONTAINER) bash -c 'build/ci/run-legacy-sle-migration --query'

sle15sp6-legacy-migration-choice-1:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(SLE15SP6CONTAINER) bash -c 'build/ci/run-legacy-sle-migration --migration 1'

sle16-env:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(SLE16CONTAINER) bash

sle16-migration-check:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(SLE16CONTAINER) bash -c 'build/ci/run-sle-migration --query'

sle16-migration-choice-1:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(SLE16CONTAINER) bash -c 'build/ci/run-sle-migration --migration 1'

sle16-legacy-migration-check:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(SLE16CONTAINER) bash -c 'build/ci/run-legacy-sle-migration --query'

sle16-legacy-migration-choice-1:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(SLE16CONTAINER) bash -c 'build/ci/run-legacy-sle-migration --migration 1'

check-format:
	@test -z $(shell gofmt -l internal/* cmd/* pkg/* | tee /dev/stderr)

build-rpm:
	$(CRM) $(MOUNT) -w $(WORKDIR) $(CONTAINER) bash -c 'build/ci/build-rpm'

feature-tests:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(CONTAINER) bash -c 'build/ci/build-rpm && build/ci/configure && build/ci/run-feature-tests'

test-yast: build
	docker build -t go-connect-test-yast -f third_party/Dockerfile.yast . && docker run -t go-connect-test-yast

clean:
	go clean
	@rm -f internal/connect/version.txt
	@rm -rf connect-ng-$(NAME)-$(VERSION)/
	@rm -rf vendor.tar.xz
	@rm -f connect-ng-*.tar.xz
