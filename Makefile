NAME ?= launcheq
VERSION ?= 0.0.9
FILELIST_URL ?= https://raw.githubusercontent.com/xackery/launcheq/rof
PATCHER_URL ?= https://github.com/xackery/launcheq/releases/latest/download/

# CICD triggers this
.PHONY: set-variable
set-version:
	@echo "VERSION=${VERSION}" >> $$GITHUB_ENV

#go install golang.org/x/tools/cmd/goimports@latest
#go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
#go install golang.org/x/lint/golint@latest
#go install honnef.co/go/tools/cmd/staticcheck@v0.2.2

sanitize:
	@echo "sanitize: checking for errors"
	rm -rf vendor/
	go vet -tags ci ./...
	test -z $(goimports -e -d . | tee /dev/stderr)
	gocyclo -over 30 .
	golint -set_exit_status $(go list -tags ci ./...)
	staticcheck -go 1.14 ./...
	go test -tags ci -covermode=atomic -coverprofile=coverage.out ./...
    coverage=`go tool cover -func coverage.out | grep total | tr -s '\t' | cut -f 3 | grep -o '[^%]*'`

run: sanitize
	@echo "run: building"
	mkdir -p bin
	cd bin && go run ../main.go

#go install github.com/tc-hib/go-winres@latest
bundle:
	go-winres simply --icon launcheq.png

.PHONY: build-all
build-all: sanitize build-prepare build-linux build-darwin build-windows	
.PHONY: build-prepare
build-prepare:
	@echo "Preparing talkeq ${VERSION}"
	@rm -rf bin/*
	@-mkdir -p bin/
.PHONY: build-darwin
build-darwin:
	@echo "Building darwin ${VERSION}"
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -s -w" -o bin/${NAME}-darwin-x64 main.go
.PHONY: build-linux
build-linux:
	@echo "Building Linux ${VERSION}"
	@GOOS=linux GOARCH=amd64 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -w" -o bin/${NAME}-linux-x64 main.go		
.PHONY: build-windows
build-windows:
	@echo "Building Windows ${VERSION}"
	go install github.com/akavel/rsrc@latest
	rsrc -ico launcheq.ico -manifest launcheq.exe.manifest
	GOOS=windows GOARCH=amd64 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -X main.PatcherUrl=${PATCHER_URL} -s -w" -o bin/${NAME}.exe