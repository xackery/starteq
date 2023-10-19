NAME ?= starteq
VERSION ?= 0.0.9
FILELIST_URL ?= https://raw.githubusercontent.com/xackery/starteq/rof
PATCHER_URL ?= https://github.com/xackery/starteq/releases/latest/download/

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
	gocyclo -over 31 .
	golint -set_exit_status $(go list -tags ci ./...)
	staticcheck -go 1.14 ./...
	go test -tags ci -covermode=atomic -coverprofile=coverage.out ./...
    coverage=`go tool cover -func coverage.out | grep total | tr -s '\t' | cut -f 3 | grep -o '[^%]*'`

run: sanitize build-linux
	@echo "run: building"
	cd bin && ./${NAME}

run-windows: build-windows
	@echo "run-windows: starting"
	chmod +x bin/${NAME}.exe
	cd bin && ./${NAME}.exe

#go install github.com/tc-hib/go-winres@latest
bundle:
	go-winres simply --icon starteq.png

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
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -X main.PatcherURL=${PATCHER_URL} -s -w" -o bin/${NAME}-darwin main.go
.PHONY: build-linux
build-linux:
	@echo "Building Linux ${VERSION}"
	@GOOS=linux GOARCH=amd64 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -X main.PatcherURL=${PATCHER_URL} -w" -o bin/${NAME}-linux-x64 main.go		
.PHONY: build-windows
build-windows:
	@echo "Building Windows ${VERSION}"
	mkdir -p bin
	go install github.com/akavel/rsrc@latest
	#rsrc -ico starteq.ico -manifest starteq.exe.manifest
	cp starteq.exe.manifest bin/
	GOOS=windows GOARCH=amd64 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -X main.PatcherURL=${PATCHER_URL} -s -w -H=windowsgui" -o bin/${NAME}.exe