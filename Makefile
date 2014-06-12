all: bin/main 

bin:
	mkdir -p bin

bin/main: dist/linux_amd64/main dist/darwin_amd64/main dist/windows_amd64/main.exe dist/windows_386/main.exe bin
	sh -c 'ln -sf ../dist/`go env GOOS`_`go env GOARCH`/main bin/main'

dist/linux_amd64/main:
	mkdir -p dist/linux_amd64
	GOOS=linux GOARCH=amd64 GOBIN=dist/linux_amd64 go install main.go

dist/darwin_amd64/main:
	mkdir -p dist/darwin_amd64
	GOOS=darwin GOARCH=amd64 GOBIN=dist/darwin_amd64 go install main.go

dist/windows_amd64/main.exe:
	mkdir -p dist/windows_amd64
	GOOS=windows GOARCH=amd64 GOBIN=dist/windows_amd64 go install main.go

dist/windows_386/main.exe:
	mkdir -p dist/windows_386
	GOOS=windows GOARCH=386 GOBIN=dist/windows_386 go install main.go

clean:
	-rm -rf dist bin

server: bin/main
	bin/main -role server

client: bin/main
	bin/main

debug:
	echo ${GOOS}
