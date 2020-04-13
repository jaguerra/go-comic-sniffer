all: build

run-dev: build
	export $(cat .env | xargs) && ./go-comic-sniffer

build:
	go build

clean:
	rm go-comic-sniffer
