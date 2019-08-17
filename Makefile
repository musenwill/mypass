GO_EXECUTABLE ?= go
VERSION = `git describe --always --tags --abbrev=0 | tr -d "[\r\n]"`
TIME = `date +%FT%T%z`

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
BINARY := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

LDFLAGS= -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${TIME}"

UNAME = $(shell uname)
ifeq (${UNAME}, Darwin)
	os=darwin
else
	os=linux
endif

build: check
	${GO_EXECUTABLE} build ${LDFLAGS} -o ${BINARY}

check:
	golint ./... | grep -v "exported" | exit 0
	go vet ./...
	gofmt -d -s `find . -name "*.go" -type f`
	go test ./...

clean:
	rm -f ${BINARY}
	rm -rf dist

build-all:
	gox -verbose \
	${LDFLAGS} \
	-os="linux darwin windows freebsd openbsd netbsd" \
	-arch="amd64 386 armv5 armv6 armv7 arm64" \
	-osarch="!darwin/arm64" \
	-output="dist/{{.OS}}-{{.Arch}}/${BINARY}" .

build-os:
	gox -verbose \
	${LDFLAGS} \
	-os="${os}" \
	-arch="amd64" \
	-output="dist/{{.OS}}-{{.Arch}}/${BINARY}" .

.PHONY: build build-all build-os clean check
