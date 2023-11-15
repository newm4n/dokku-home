GO111MODULE=on

.PHONY: all test test-short build

build:
	go build ./...

test-short: build
	go test ./... -v -covermode=count -coverprofile=coverage.out -short

test: build
	go test ./... -covermode=count -coverprofile=coverage.out

push: build
	git push dokku main
