all: build

run-dev: build
	export $(cat .env | xargs) && ./go-comic-sniffer

build: deps
	go build

deps:
	go get github.com/ericchiang/css
	go get golang.org/x/net/html

clean:
	rm go-comic-sniffer
