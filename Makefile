SHELL := /bin/bash

#TAG=`git describe --exact-match --tags $(git log -n1 --pretty='%h') 2>/dev/null || git rev-parse --abbrev-ref HEAD`
TAG=`git describe --abbrev=0 --tag`
LAST_TAG := $(shell git describe --abbrev=0 --tags)

PASSWORD ?= $(shell stty -echo; read -p "new tag: " tag; stty echo; echo $$tag)
NEW_TAG ?= $(shell read -p "new tag: " tag; stty echo; echo $$tag)
GITHUB_API_KEY ?= $(shell read -p "please enter the github api key: " key; stty echo; echo $$key)

all: build test
clean:
	rm -fv bin/neosp*
build: clean
	go build -o bin/neosproxy cmd/neosproxy/main.go
build-arch: clean
	GOOS=darwin GOARCH=amd64 go build -o bin/neosproxy-darwin-amd64 cmd/neosproxy/main.go
build-linux: clean
	GOOS=linux GOARCH=amd64 go build -o bin/neosproxy-linux-amd64 cmd/neosproxy/main.go
build-docker: clean
	docker build -t foomo/neosproxy:dev .
prepare-docker:
	curl -o files/cacert.pem https://curl.haxx.se/ca/cacert.pem
test:
	go test ./...

run: clean
	go run cmd/neosproxy/main.go -config-file ./config-example.yaml


latest-tag:
	@echo "last tagged version: $(LAST_TAG)"

release-notes: latest-tag
	@git log $(LAST_TAG)..HEAD --no-merges --format="%h: %s" > changelog.md
	@echo create new tag: $(NEW_TAG)
	@echo GitHub API key is: $(GITHUB_API_KEY)


#	echo The password is $(PASSWORD)
#--release-notes=FILE