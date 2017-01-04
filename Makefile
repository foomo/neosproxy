SHELL := /bin/bash

#TAG=`git describe --exact-match --tags $(git log -n1 --pretty='%h') 2>/dev/null || git rev-parse --abbrev-ref HEAD`

all: build test
clean:
	rm -fv bin/neosp*
build: clean
	go build -o bin/neosproxy neosproxy.go
build-arch: clean build-linux
	GOOS=darwin GOARCH=amd64 go build -o bin/neosproxy-darwin-amd64 neosproxy.go
build-linux: clean
	GOOS=linux GOARCH=amd64 go build -o bin/neosproxy-linux-amd64 neosproxy.go
build-docker: clean build-arch prepare-docker
	docker build -t foomo/neosproxy:latest .
prepare-docker:
	curl -o files/cacert.pem https://curl.haxx.se/ca/cacert.pem
release: clean build-linux prepare-docker
	git add -f files/cacert.pem
	git add -f bin/neosproxy-linux-amd64
	git commit -m 'build release candidate - new binary added for docker autobuild'
	echo "-------------------------"
	echo "please make sure that version number has been bumped, then tag and push the git repo"
	echo "-------------------------"
test:
	go test ./...