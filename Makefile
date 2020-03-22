###############################################
#
# Makefile
#
###############################################

.DEFAULT_GOAL := build

.PHONY: test

VERSION := 0.40.0

ver:
	@sed -i '' 's/^const Version = "[0-9]\{1,3\}.[0-9]\{1,3\}.[0-9]\{1,3\}"/const Version = "${VERSION}"/' rx.go

lint:
	$(shell go env GOPATH)/bin/golint .

format:
	go fmt .

vet:
	go vet .

sec:
	$(shell go env GOPATH)/bin/gosec .

# PROFILE = -blockprofile
# PROFILE = -cpuprofile
PROFILE = -memprofile
profile:
	-rm -f rx.prof rx.test
	-go test ${PROFILE}=rx.prof .
	go tool pprof rx.prof

ppcpu:
	go tool pprof http://localhost/debug/pprof/profile

ppmem:
	go tool pprof http://localhost/debug/pprof/heap

build:
	go build -v .

clean:
	go clean ...

demo: build
	go run cmd/demo.go

race:
	go build -race .

test: build
	go test -v -count=${COUNT} .

TEST ?= TestMerge
COUNT ?= 1
testx: build
	go test -v -count=${COUNT} . -run ${TEST}

bench: build
	go test -bench=. -v .

github:
	open "https://github.com/mlavergn/rxgo"

release:
	zip -r rxgo.zip LICENSE README.md Makefile cmd *.go go.mod
	hub release create -m "${VERSION} - RxGo" -a rxgo.zip -t master "v${VERSION}"
	open "https://github.com/mlavergn/rxgo/releases"

st:
	open -a SourceTree .
