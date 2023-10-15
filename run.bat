mkdir bin
rsrc -ico starteq.ico -manifest starteq.exe.manifest
copy /y starteq.exe.manifest bin\starteq.exe.manifest
go build -o starteq.exe main.go
move starteq.exe bin/starteq.exe
cd bin && starteq.exe