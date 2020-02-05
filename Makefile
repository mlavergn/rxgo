###############################################
#
# Makefile
#
###############################################

.DEFAULT_GOAL := build

.PHONY: test

VERSION := 0.37.1

ver:
	@sed -i '' 's/^const Version = "[0-9]\{1,3\}.[0-9]\{1,3\}.[0-9]\{1,3\}"/const Version = "${VERSION}"/' src/rx/rx.go

lint:
	$(shell go env GOPATH)/bin/golint ./src/...

fmt:
	go fmt ./src/...

vet:
	go vet ./src/...

# PROFILE = -blockprofile
# PROFILE = -cpuprofile
PROFILE = -memprofile
profile:
	-rm -f rx.prof rx.test
	-go test ${PROFILE}=rx.prof ./src/...
	go tool pprof rx.prof

ppcpu:
	go tool pprof http://localhost/debug/pprof/profile

ppmem:
	go tool pprof http://localhost/debug/pprof/heap

build:
	go build -v ./src/...

clean:
	go clean ...

demo: build
	go run cmd/demo.go

race:
	go build -race ./src/...

test: build
	go test -v -count=${COUNT} ./src/...

TEST ?= TestMerge
COUNT ?= 1
testx: build
	go test -v -count=${COUNT} ./src/... -run ${TEST}

bench: build
	go test -bench=. -v ./src/...

github:
	open "https://github.com/mlavergn/rxgo"

release:
	zip -r rxgo.zip LICENSE README.md Makefile cmd src
	hub release create -m "${VERSION} - RxGo" -a rxgo.zip -t master "v${VERSION}"
	open "https://github.com/mlavergn/rxgo/releases"
