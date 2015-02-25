DEPENDENCIES := \
	github.com/stretchr/graceful

all: deps build

all-linux: deps build-linux

build:
	go build -o runner

build-linux:
	GOOS=linux GOARCH=amd64 go build -o runner.linux

deps:
	go get $(DEPENDENCIES)