mkdir bin
rsrc -ico starteq.ico -manifest starteq.exe.manifest
copy /y starteq.exe.manifest bin\starteq.exe.manifest
go build -buildmode=pie -ldflags="-s -w -H=windowsgui" -o starteq.exe main.go
move starteq.exe bin/starteq.exe
cd bin && starteq.exe