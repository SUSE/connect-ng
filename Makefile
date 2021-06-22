all: test build build-so

build:
	go build cmd/suseconnect.go

test:
	go test -v ./connect

build-so:
	go build -buildmode=c-shared -o libsuseconnect.so ext/main.go

build-arm:
	GOOS=linux GOARCH=arm64 GOARM=7 go build cmd/suseconnect.go

clean:
	go clean

