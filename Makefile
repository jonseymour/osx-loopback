MODULE=vbox-portforward
OPEN=open

all: bin/${MODULE} 

bin:
	mkdir -p bin

bin/${MODULE}: dist/linux_amd64/${MODULE} dist/darwin_amd64/${MODULE} dist/windows_amd64/${MODULE}.exe dist/windows_386/${MODULE}.exe bin
	sh -c 'ln -sf ../dist/`go env GOOS`_`go env GOARCH`/${MODULE} bin/${MODULE}'

dist/linux_amd64/${MODULE}:
	mkdir -p dist/linux_amd64
	GOOS=linux GOARCH=amd64 GOBIN=dist/linux_amd64 go install ${MODULE}.go

dist/darwin_amd64/${MODULE}:
	mkdir -p dist/darwin_amd64
	GOOS=darwin GOARCH=amd64 GOBIN=dist/darwin_amd64 go install ${MODULE}.go

dist/windows_amd64/${MODULE}.exe:
	mkdir -p dist/windows_amd64
	GOOS=windows GOARCH=amd64 GOBIN=dist/windows_amd64 go install ${MODULE}.go

dist/windows_386/${MODULE}.exe:
	mkdir -p dist/windows_386
	GOOS=windows GOARCH=386 GOBIN=dist/windows_386 go install ${MODULE}.go

dist:
	mkdir -p dist

clean:
	-rm -rf dist bin

doc:	dist
	markdown README.md > dist/README.html

server: bin/${MODULE}
	bin/${MODULE} -role server

client: bin/${MODULE}
	bin/${MODULE}

debug:
	echo ${GOOS}

readme:	doc
	${OPEN} dist/README.html