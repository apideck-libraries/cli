BINARY_NAME=apideck
VERSION?=dev

.PHONY: build run test clean

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/$(BINARY_NAME) ./cmd/apideck

run: build
	./bin/$(BINARY_NAME)

test:
	go test ./... -v

clean:
	rm -rf bin/
