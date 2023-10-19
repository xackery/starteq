mkdir bin
mkdir .gocache
rsrc -ico starteq.ico -manifest starteq.exe.manifest
copy /y starteq.exe.manifest bin\starteq.exe.manifest
docker run --rm -v %cd%:/src -v %cd%/.gocache:/go -w /src golang:1.21.1-alpine sh -c "GOOS=windows time go build -buildmode=pie -ldflags=\"-s -w\" -o bin/starteq.exe main.go"
cd bin && starteq.exe