build:
	GOBIN=./bin go install src/github.com/jonseymour/osx-loopback/main.go

bin/main: build

server: bin/main
	bin/main -role server

client: bin/main
	bin/main