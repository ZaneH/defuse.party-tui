.PHONY: all build clean run test

all: build

build:
	go build -o tui-server ./cmd/server

clean:
	rm -f tui-server

run: build
	./tui-server

test:
	go test ./...

lint:
	golangci-lint run ./...
