NAME          = suseconnect-ng
VERSION       = $(shell bash -c "cat build/packaging/suseconnect-ng.spec | sed -n 's/^Version:\s*\(.*\)/\1/p'")
DIST          = $(NAME)-$(VERSION)/

all: test build build-so

dist: clean internal/connect/version.txt
	@mkdir -p $(DIST)
	@cp -r internal $(DIST)
	@cp -r third_party $(DIST)
	@cp -r cmd $(DIST)
	@cp -r docs $(DIST)
	@cp go.mod $(DIST)
	@cp LICENSE LICENSE.LGPL README.md $(DIST)
	@cp SUSEConnect.example $(DIST)
	@cp -r build/packaging/* $(DIST)
	@tar cfvj $(NAME)-$(VERSION).tar.xz $(NAME)-$(VERSION)/
	@rm -r $(NAME)-$(VERSION)

out:
	mkdir -p out

internal/connect/version.txt:
	@echo "$(VERSION)" > internal/connect/version.txt

build: out internal/connect/version.txt
	go build -v -o out/ github.com/SUSE/connect-ng/cmd/suseconnect
	go build -v -o out/ github.com/SUSE/connect-ng/cmd/zypper-migration
	go build -v -o out/ github.com/SUSE/connect-ng/cmd/zypper-search-packages

test: internal/connect/version.txt
	go test -v ./internal/* ./cmd/suseconnect

test-yast: build-so
	docker build -t go-connect-test-yast -f third_party/Dockerfile.yast . && docker run -t go-connect-test-yast

test-scc: connect-ruby
	docker build -t connect.ng-sle15sp3 -f integration/Dockerfile.ng-sle15sp3 .
	docker run --privileged --rm -t connect.ng-sle15sp3 ./integration/run.sh

connect-ruby:
	git clone https://github.com/SUSE/connect connect-ruby

gofmt:
	@if [ ! -z "$$(gofmt -l ./)" ]; then echo "Formatting errors..."; gofmt -d ./; exit 1; fi

build-so: out internal/connect/version.txt
	go build -v -buildmode=c-shared -o out/libsuseconnect.so github.com/SUSE/connect-ng/third_party/libsuseconnect

build-arm: out internal/connect/version.txt
	GOOS=linux GOARCH=arm64 GOARM=7 go build -v -o out/ github.com/SUSE/connect-ng/cmd/suseconnect

build-s390: out internal/connect/version.txt
	GOOS=linux GOARCH=s390x go build -v -o out/ github.com/SUSE/connect-ng/cmd/suseconnect

clean:
	go clean
	@rm -f internal/connect/version.txt
	@rm -rf connect-ng-$(NAME)-$(VERSION)/
	@rm -f connect-ng-*.tar.xz
