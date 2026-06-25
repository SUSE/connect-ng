NAME          = suseconnect-ng
VERSION       = $(shell bash -c "cat build/packaging/suseconnect-ng.spec | sed -n 's/^Version:\s*\(.*\)/\1/p'")
DIST          = $(NAME)-$(VERSION)
PWD           = $(shell pwd)
GO            = go
OUT           = -o out/
# coverage testing enabled by default
COVERAGE     ?= true
COVERAGE_BIN ?= false
COVERAGE_LIB ?= false
GOFLAGS       = -v -mod=vendor
BINFLAGS      = -buildmode=pie
SOFLAGS       = -buildmode=c-shared

GOCONTAINER   = registry.suse.com/bci/golang:1.24-openssl
RUSTCONTAINER = registry.suse.com/bci/rust:1.95
CRM           = docker run --rm -it --privileged
ENVFILE       = .env
WORKDIR       = /usr/src/connect-ng
MOUNT         = -v $(PWD):$(WORKDIR)
AGAMA_SOURCES = $(PWD)/testdata/agama_srcs
AGAMA_REPO    = https://github.com/agama-project/agama
AGAMA_MOUNT   = -v $(AGAMA_SOURCES):/usr/src/agama

COVERAGE_DIR     = $(PWD)/coverage
COVERAGE_MERGED  = $(COVERAGE_DIR)/merged
COVERAGE_UNIT    = $(COVERAGE_DIR)/unit
COVERAGE_FEATURE = $(COVERAGE_DIR)/feature
# these directories aren't being used yet
COVERAGE_YAST    = $(COVERAGE_DIR)/yast
COVERAGE_AGAMA   = $(COVERAGE_DIR)/agama

define go-tool-covdata-merge
@if $(if $(filter true,$(strip $(COVERAGE))),true,false); then \
	mkdir -p $(COVERAGE_MERGED); \
	$(GO) tool covdata merge -i=$(COVERAGE_UNIT),$(COVERAGE_FEATURE),$(COVERAGE_YAST),$(COVERAGE_AGAMA) -o $(COVERAGE_MERGED); \
fi
endef

define go-tool-covdata-report
@if [ -z "$(strip $(1))" ]; then \
	echo "ERROR: no go tool covdata action specified."; \
	exit 1; \
fi
@if [ -z "$(strip $(2))" ]; then \
	echo "ERROR: no test type target directory specified."; \
	exit 1; \
fi
@if $(if $(filter true,$(strip $(COVERAGE))),true,false); then \
	mkdir -p $(2); \
	$(GO) tool covdata $(1) -i=$(strip $(2)); \
fi
endef

define cover-test-flags
$(if $(filter true,$(strip $(COVERAGE))),-cover -args -test.gocoverdir=$(COVERAGE_UNIT))
endef

define cover-bin-flags
$(if $(filter true,$(strip $(COVERAGE))),$(if $(filter true,$(strip $(COVERAGE_BIN))),-cover))
endef

define cover-lib-flags
$(if $(filter true,$(strip $(COVERAGE))),$(if $(filter true,$(strip $(COVERAGE_LIB))),-cover))
endef

.PHONY: all build build-arm build-ppc64le build-rpm build-s390 check-format
.PHONY: ci-env clean dist feature-tests out show-version test test-yast vendor vet
.PHONY: coverage coverage-check-enabled coverage-dirs coverage-func coverage-percent
.PHONY: agama-sources agama-tests bci-build go-env rust-env run-tests coverage-clean
.PHONY: fix-ownership real-clean unit-test-coverage feature-tests-coverage

all: clean build test

run-tests: coverage-clean test feature-tests test-yast agama-tests
	$(call go-tool-covdata-merge)
	$(call go-tool-covdata-report,percent,$(COVERAGE_MERGED))
	$(call go-tool-covdata-report,func,$(COVERAGE_MERGED))

dist: clean internal/connect/version.txt vendor
	@mkdir -p $(DIST)/build/packaging
	@cp -r internal $(DIST)/
	@cp -r pkg $(DIST)/
	@cp -r third_party $(DIST)/
	@cp -r cmd $(DIST)/
	@cp -r docs $(DIST)/
	@cp go.mod $(DIST)/
	@cp go.sum $(DIST)/
	@cp LICENSE README.md $(DIST)/
	@cp build/packaging/suseconnect-keepalive* $(DIST)/build/packaging
	@cp build/packaging/suse-uptime-tracker* $(DIST)/build/packaging
	@cp -r build/packaging/suseconnect-ng* $(DIST)/

	@tar cfvj vendor.tar.xz vendor
	@tar cfvj $(DIST).tar.xz $(DIST)/

	@rm -r $(DIST)/

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
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(call cover-bin-flags) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(call cover-bin-flags) $(OUT) github.com/SUSE/connect-ng/cmd/zypper-migration
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(call cover-bin-flags) $(OUT) github.com/SUSE/connect-ng/cmd/zypper-search-packages
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(call cover-bin-flags) $(OUT) github.com/SUSE/connect-ng/cmd/suse-uptime-tracker
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(call cover-bin-flags) $(OUT) github.com/SUSE/connect-ng/cmd/public-api-demo
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(call cover-bin-flags) $(OUT) github.com/SUSE/connect-ng/cmd/validate-offline-certificate
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(call cover-bin-flags) $(OUT) github.com/SUSE/connect-ng/cmd/offline-register-api
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(call cover-bin-flags) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect-mcp
	$(GO) build $(GOFLAGS) $(SOFLAGS) $(call cover-lib-flags) $(OUT)/libsuseconnect.so github.com/SUSE/connect-ng/third_party/libsuseconnect

bci-build:
	$(CRM) $(MOUNT) -w $(WORKDIR) $(GOCONTAINER) bash -c 'git config --global --add safe.directory $(WORKDIR); make vendor build'

# This "arm" means ARM64v8 little endian, the one being delivered currently on
# OBS.
build-arm: clean out internal/connect/version.txt
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect

build-s390: clean out internal/connect/version.txt
	GOOS=linux GOARCH=s390x $(GO) build $(GOFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect

build-ppc64le: clean out internal/connect/version.txt
	GOOS=linux GOARCH=ppc64le $(GO) build $(GOFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect

coverage-dirs:
	@mkdir -p $(COVERAGE_UNIT) ${COVERAGE_FEATURE} ${COVERAGE_YAST} ${COVERAGE_AGAMA} ${COVERAGE_MERGED}

coverage-check-enabled:
	@if [ -z "$(filter true,$(strip $(COVERAGE)))" ]; then \
		echo "WARNING: Coverage generation and collection not enabled."; \
		echo "To enable it either add COVERAGE=true on the make command line arguments or set COVERAGE=true in the environment."; \
		exit 1; \
	fi

coverage-func: coverage-check-enabled coverage-dirs
	$(call go-tool-covdata-merge)
	$(call go-tool-covdata-report,func,$(COVERAGE_MERGED))

coverage-percent: coverage-check-enabled coverage-dirs
	$(call go-tool-covdata-merge)
	$(call go-tool-covdata-report,percent,$(COVERAGE_MERGED))

coverage: coverage-func

test: internal/connect/version.txt coverage-dirs
	$(GO) test ./internal/* ./cmd/suseconnect ./pkg/* $(call cover-test-flags)
	$(call go-tool-covdata-report,func,$(COVERAGE_UNIT))

unit-test-coverage: coverage-dirs
	$(call go-tool-covdata-report,percent,$(COVERAGE_UNIT))
	$(call go-tool-covdata-report,func,$(COVERAGE_UNIT))

ci-env go-env:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(GOCONTAINER) bash

agama-sources:
	@if [ ! -d $(AGAMA_SOURCES) ]; then \
		echo "Cloning $(AGAMA_REPO) under $(AGAMA_SOURCES) ..."; \
		git clone $(AGAMA_REPO) $(AGAMA_SOURCES); \
	fi

agama-tests: agama-sources bci-build
	$(CRM) $(MOUNT) $(AGAMA_MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(RUSTCONTAINER) bash -c 'build/ci/run-agama-rust-tests'


rust-env: agama-sources
	$(CRM) $(MOUNT) $(AGAMA_MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(RUSTCONTAINER) bash

check-format:
	@test -z $(shell gofmt -l internal/* cmd/* pkg/* | tee /dev/stderr)

build-rpm:
	$(CRM) $(MOUNT) -w $(WORKDIR) $(GOCONTAINER) bash -c 'build/ci/build-rpm'

feature-tests: coverage-dirs
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(GOCONTAINER) bash -c 'make vendor build COVERAGE_BIN=true && build/ci/run-feature-tests'
	$(call go-tool-covdata-report,percent,$(COVERAGE_FEATURE))
	$(call go-tool-covdata-report,func,$(COVERAGE_FEATURE))

feature-tests-coverage: coverage-dirs
	$(call go-tool-covdata-report,percent,$(COVERAGE_FEATURE))
	$(call go-tool-covdata-report,func,$(COVERAGE_FEATURE))

test-yast: bci-build
	docker build -t go-connect-test-yast -f third_party/Dockerfile.yast . && docker run -t go-connect-test-yast

coverage-clean:
	@rm -rf $(COVERAGE_DIR)

real-clean: clean coverage-clean
	@rm -rf vendor/
	@rm -rf out/

clean:
	go clean
	@rm -f internal/connect/version.txt
	@rm -rf $(DIST)/
	@rm -f vendor.tar.xz
	@rm -f $(DIST).tar.xz

fix-ownership:
	@sudo chown -R $$(id -u):$$(id -g) artifacts coverage out vendor internal/connect/version.txt testdata/agama_srcs
	@for o in artifacts coverage internal/connect/version.txt out testdata/agama_srcs vendor; do \
		[ -e $${o} ] || continue; \
		sudo chown -R $$(id -u):$$(id -g) $${o}; \
	done