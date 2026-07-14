.PHONY: build test lint clean install

build:
	go build -o racore-cli .

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f racore-cli

install: build
	cp racore-cli $(GOPATH)/bin/
