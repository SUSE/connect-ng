all: test build build-so

out:
	mkdir -p out

build: out
	go build -v -o out/ github.com/SUSE/connect-ng/suseconnect

test:
	go test -v ./connect

build-so: out
	go build -v -buildmode=c-shared -o out/libsuseconnect.so github.com/SUSE/connect-ng/libsuseconnect

build-so-example: build-so
	gcc libsuseconnect-examples/use-lib.c -o out/use-lib -Lout -lsuseconnect

build-arm: out
	GOOS=linux GOARCH=arm64 GOARM=7 go build -v -o out/ github.com/SUSE/connect-ng/suseconnect

clean:
	go clean

