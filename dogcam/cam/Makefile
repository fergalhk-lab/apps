.PHONY: build build-arm64 test

build:
	go build -o ../../bin/dogcam-cam .

# Cross-compile for Raspberry Pi (64-bit)
build-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ../../bin/dogcam-cam-arm64 .

test:
	go test ./...
