###############################################
#
# Makefile
#
###############################################

.DEFAULT_GOAL := build

.PHONY: test

GOPATH = "${PWD}"

lint:
	GOPATH=${GOPATH} ~/go/bin/golint .

fmt:
	GOPATH=${GOPATH} go fmt rxgo

vet:
	GOPATH=${GOPATH} go vet rxgo

build:
	GOPATH=${GOPATH} go build ./...

demo: build
	GOPATH=${GOPATH} go run cmd/demo.go

race: build
	GOPATH=${GOPATH} go run -race cmd/demo.go

test: build
	GOPATH=${GOPATH} go test -v -count=1 ./src/...

bench: build
	GOPATH=${GOPATH} go test -bench=. -v ./src/...

github:
	open "https://github.com/mlavergn/rxgo"