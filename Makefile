NAME          = suseconnect-ng
VERSION       = $(shell bash -c "cat build/packaging/suseconnect-ng.spec | sed -n 's/^Version:\s*\(.*\)/\1/p'")
VERSION_TXT   = internal/connect/version.txt
DIST          = $(NAME)-$(VERSION)
PWD           = $(shell pwd)
GO            = go
OUT           = -o out/
COVERAGE     ?= false
GOFLAGS       = -v -mod=vendor $(if $(filter true,$(strip $(COVERAGE))),-cover)
BINFLAGS      = -buildmode=pie
SOFLAGS       = -buildmode=c-shared
COVERAGE_DIR  = $(PWD)/coverage
COV_UNIT      = $(COVERAGE_DIR)/unit
COV_FEATURE   = $(COVERAGE_DIR)/feature
COV_YAST      = $(COVERAGE_DIR)/yast
COV_MERGED    = $(COVERAGE_DIR)/merged

CONTAINER     = registry.suse.com/bci/golang:1.24-openssl
CRM           = docker run --rm -it --privileged -e COVERAGE=$(COVERAGE)
ENVFILE       = .env
WORKDIR       = /usr/src/connect-ng
MOUNT         = -v $(PWD):$(WORKDIR)

export COVERAGE

.PHONY: all dist real-clean clean go-clean ci-env check-format vendor generate-version
.PHONY: vet build build-arm build-ppc64le build-s390 test test-yast feature-tests deps
.PHONY: coverage coverage-dirs coverage-merged run-tests check-coverage-enabled

all: clean build test

dist: clean generate-version vendor
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

vet: generate-version vendor
	$(GO) vet ./...

deps:
	$(GO) get -u ./...
	$(GO) mod tidy

vendor:
	@$(GO) mod download
	@$(GO) mod verify
	@$(GO) mod vendor

out:
	mkdir -p out

generate-version: $(VERSION_TXT)

$(VERSION_TXT):
	@echo -n "$(VERSION)" > $(VERSION_TXT)

build: go-clean out vet
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/zypper-migration
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/zypper-search-packages
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suse-uptime-tracker
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/public-api-demo
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/validate-offline-certificate
	$(GO) build $(GOFLAGS) $(BINFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/offline-register-api
	$(GO) build $(GOFLAGS) $(SOFLAGS) $(OUT) github.com/SUSE/connect-ng/third_party/libsuseconnect

# This "arm" means ARM64v8 little endian, the one being delivered currently on
# OBS.
build-arm: go-clean out vet
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect

build-s390: go-clean out vet
	GOOS=linux GOARCH=s390x $(GO) build $(GOFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect

build-ppc64le: go-clean out vet
	GOOS=linux GOARCH=ppc64le $(GO) build $(GOFLAGS) $(OUT) github.com/SUSE/connect-ng/cmd/suseconnect

run-tests: test feature-tests test-yast

test: vet coverage-dirs
	$(GO) test -cover ./internal/* ./cmd/suseconnect ./pkg/* -args -test.gocoverdir=$(COV_UNIT)
	@if [ -n "$(filter true,$(strip $(COVERAGE)))" ]; then \
		$(GO) tool covdata textfmt -i=$(COV_UNIT) -o /dev/stdout | $(GO) tool cover -func=/dev/stdin; \
	fi

check-coverage-enabled:
	@if [ -z "$(filter true,$(strip $(COVERAGE)))" ]; then \
		echo "WARNING: Coverage generation and collection not enabled."; \
		echo "To enable it add COVERAGE=true on the command line or set it in the environment."; \
		exit 1; \
	fi

coverage-dirs:
	@mkdir -p $(COV_UNIT) $(COV_FEATURE) $(COV_YAST) $(COV_MERGED)

coverage-merged: coverage-dirs check-coverage-enabled
	$(GO) tool covdata merge -i=$(COV_UNIT),$(COV_FEATURE) -o $(COV_MERGED)

coverage-percent: coverage-merged
	$(GO) tool covdata percent -i=$(COV_MERGED)

coverage: coverage-merged
	$(GO) tool covdata textfmt -i=$(COV_MERGED) -o /dev/stdout | $(GO) tool cover -func=/dev/stdin

ci-env:
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(CONTAINER) bash

check-format: vet
	@test -z $(shell gofmt -l internal/* cmd/* pkg/* | tee /dev/stderr)

build-rpm: vet
	$(CRM) $(MOUNT) -w $(WORKDIR) $(CONTAINER) bash -c 'build/ci/build-rpm && rm -rf vendor'
	@$(GO) mod vendor

feature-tests: vet coverage-dirs
	$(CRM) $(MOUNT) --env-file $(ENVFILE) -w $(WORKDIR) $(CONTAINER) bash -c 'build/ci/build-rpm && build/ci/configure && build/ci/run-feature-tests && rm -rf vendor'
	@$(GO) mod vendor
	@if [ -n "$(filter true,$(strip $(COVERAGE)))" ]; then \
		$(GO) tool covdata textfmt -i=$(COV_FEATURE) -o /dev/stdout | $(GO) tool cover -func=/dev/stdin; \
	fi

test-yast: build coverage-dirs
	docker build -t go-connect-test-yast -f third_party/Dockerfile.yast .
	docker run -v $(COVERAGE_DIR):/coverage -t go-connect-test-yast
	@if [ -n "$(filter true,$(strip $(COVERAGE)))" ]; then \
		$(GO) tool covdata textfmt -i=$(COV_YAST) -o /dev/stdout | $(GO) tool cover -func=/dev/stdin; \
	fi

 go-clean:
	$(GO) clean

clean: go-clean
	@rm -f $(VERSION_TXT)
	@rm -rf $(DIST)/
	@rm -f vendor.tar.xz
	@rm -f $(DIST).tar.xz

real-clean: clean
	@rm -rf out
	@rm -rf vendor
	@rm -rf $(COVERAGE_DIR)
