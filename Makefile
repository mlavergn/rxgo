###############################################
#
# Makefile
#
###############################################

.DEFAULT_GOAL := build

.PHONY: test

VERSION := 0.15.0

GOPATH = "${PWD}"

ver:
	@sed -i '' 's/^const Version = "[0-9]\{1,3\}.[0-9]\{1,3\}.[0-9]\{1,3\}"/const Version = "${VERSION}"/' src/rx/rx.go

lint:
	GOPATH=${GOPATH} ~/go/bin/golint ./src/...

fmt:
	GOPATH=${GOPATH} go fmt rx ./src/...

vet:
	GOPATH=${GOPATH} go vet rx ./src/...

# PROFILE = -blockprofile
# PROFILE = -cpuprofile
PROFILE = -memprofile
profile:
	-rm -f rx.prof rx.test
	-GOPATH=${GOPATH} go test ${PROFILE}=rx.prof ./src/...
	GOPATH=${GOPATH} go tool pprof rx.prof

build:
	GOPATH=${GOPATH} go build -v ./src/...

clean:
	GOPATH=${GOPATH} go clean

demo: build
	GOPATH=${GOPATH} go run cmd/demo.go

race:
	GOPATH=${GOPATH} go build -race ./src/...

test: build
	GOPATH=${GOPATH} go test -v -count=1 ./src/...

TEST ?= TestMerge
COUNT ?= 1
testx: build
	GOPATH=${GOPATH} go test -v -count=${COUNT} ./src/... -run ${TEST}

bench: build
	GOPATH=${GOPATH} go test -bench=. -v ./src/...

github:
	open "https://github.com/mlavergn/rxgo"

release:
	zip rxgo.zip .
	hub release create -m "${VERSION} - rxgo" -a rxgo.zip -t master "${VERSION}"
	open "https://github.com/mlavergn/rxgo/releases"
