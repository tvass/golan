BIN     := golan
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: all build build-arm64 test clean run

all: build

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BIN) .

build-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BIN)-arm64 .


run:
	./$(BIN)

clean:
	rm -f $(BIN) $(BIN)-arm64
