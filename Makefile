VERSION?=$(shell git describe --tags --always --dirty)
LD_FLAGS?="-X main.version=$(VERSION)"

all: build

build:
	go build -ldflags $(LD_FLAGS)

install: build
	go install -ldflags $(LD_FLAGS)

dist: dist/annotate_darwin_amd64 dist/annotate_linux_amd64

dist/annotate_darwin_amd64:
	GOOS=darwin GOARCH=amd64 go build -o $@ -ldflags $(LD_FLAGS)

dist/annotate_linux_amd64:
	GOOS=linux GOARCH=amd64 go build -o $@ -ldflags $(LD_FLAGS)

rel: dist
	hub release create -a dist $(VERSION)
