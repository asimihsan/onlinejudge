DEPENDENCIES := \
	github.com/smugmug/godynamo \
	github.com/gorilla/mux \
	github.com/BurntSushi/toml \
	github.com/stretchr/graceful

all: deps build

all-linux: deps build-linux

build:
	go build -o evaluator

build-linux:
	GOOS=linux GOARCH=amd64 go build -o evaluator.linux

deps:
	go get $(DEPENDENCIES)