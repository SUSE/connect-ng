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
	go build -v -o out/ github.com/SUSE/connect-ng/suseconnect

test: internal/connect/version.txt
	go test -v ./internal/connect

gofmt:
	@if [ ! -z "$$(gofmt -l ./)" ]; then echo "Formatting errors..."; gofmt -d ./; exit 1; fi

build-so: out internal/connect/version.txt
	go build -v -buildmode=c-shared -o out/libsuseconnect.so github.com/SUSE/connect-ng/libsuseconnect

build-so-example: build-so
	gcc libsuseconnect-examples/use-lib.c -o out/use-lib -Lout -lsuseconnect

build-arm: out internal/connect/version.txt
	GOOS=linux GOARCH=arm64 GOARM=7 go build -v -o out/ github.com/SUSE/connect-ng/suseconnect

build-s390: out internal/connect/version.txt
	GOOS=linux GOARCH=s390x go build -v -o out/ github.com/SUSE/connect-ng/suseconnect

clean:
	go clean
	rm internal/connect/version.txt
