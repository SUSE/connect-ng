all: test build build-so

out:
	mkdir -p out

internal/connect/version.txt:
	# this is the equivalent from _service of: @PARENT_TAG@~git@TAG_OFFSET@.%h
	parent=$$(git describe --tags --abbrev=0 --match='v*' | sed 's:^v::' ); \
	       offset=$$(git rev-list --count "v$${parent}..HEAD"); \
	       git log --no-show-signature -n1 --date='format:%Y%m%d' --pretty="format:$${parent}~git$${offset}.%h" > internal/connect/version.txt
	cat -v internal/connect/version.txt
	@echo

build: out internal/connect/version.txt
	GOEXPERIMENT=boringcrypto go build -v -o out/ github.com/SUSE/connect-ng/suseconnect

test: internal/connect/version.txt
	go test -v ./internal/connect ./suseconnect

test-yast: build-so
	docker build -t go-connect-test-yast -f Dockerfile.yast . && docker run -t go-connect-test-yast

test-scc: connect-ruby
	docker build -t connect.ng-sle15sp3 -f integration/Dockerfile.ng-sle15sp3 .
	docker run --privileged --rm -t connect.ng-sle15sp3 ./integration/run.sh

connect-ruby:
	git clone https://github.com/SUSE/connect connect-ruby

gofmt:
	@if [ ! -z "$$(gofmt -l ./)" ]; then echo "Formatting errors..."; gofmt -d ./; exit 1; fi

build-so: out internal/connect/version.txt
	GOEXPERIMENT=boringcrypto go build -v -buildmode=c-shared -o out/libsuseconnect.so github.com/SUSE/connect-ng/libsuseconnect

build-arm: out internal/connect/version.txt
	GOOS=linux GOARCH=arm64 GOARM=7 GOEXPERIMENT=boringcrypto go build -v -o out/ github.com/SUSE/connect-ng/suseconnect

build-s390: out internal/connect/version.txt
	GOOS=linux GOARCH=s390x GOEXPERIMENT=boringcrypto go build -v -o out/ github.com/SUSE/connect-ng/suseconnect

clean:
	go clean
	rm internal/connect/version.txt
