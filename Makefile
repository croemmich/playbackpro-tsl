.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o "bin/playbackpro-tsl-darwin-amd64"

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o "bin/playbackpro-tsl-linux-amd64"

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -o "bin/playbackpro-tsl-windows-amd64"

.PHONY: build
build: build-darwin build-linux build-windows

.PHONY: clean
clean:
	-rm -rf bin