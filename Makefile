all: bin/main 

bin:
	mkdir -p bin

bin/main: dist/linux_amd64/main dist/darwin_amd64/main bin
	sh -c 'ln -sf ../dist/`go env GOOS`_`go env GOARCH`/main bin/main'

dist/linux_amd64/main:
	mkdir -p dist/linux_amd64
	GOOS=linux GOARCH=amd64 GOBIN=dist/linux_amd64 go install src/github.com/jonseymour/osx-loopback/main.go

dist/darwin_amd64/main:
	mkdir -p dist/darwin_amd64
	GOOS=darwin GOARCH=amd64 GOBIN=dist/darwin_amd64 go install src/github.com/jonseymour/osx-loopback/main.go

clean:
	-rm -rf dist bin

server: bin/main
	bin/main -role server

client: bin/main
	bin/main

debug:
	echo ${GOOS}