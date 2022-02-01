all: build
.PHONY: all build

build:
	@CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ./bin/kolony ./cmd/main.go
