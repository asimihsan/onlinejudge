DEPENDENCIES := \
	github.com/smugmug/godynamo \
	github.com/gorilla/mux \
	github.com/stretchr/graceful \
	github.com/gorilla/sessions \
	github.com/nu7hatch/gouuid

all: deps build

all-linux: deps build-linux

build:
	go build -o user_data

build-linux:
	GOOS=linux GOARCH=amd64 go build -o user_data.linux

deps:
	go get $(DEPENDENCIES)