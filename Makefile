export GO15VENDOREXPERIMENT=1
PACKAGES=$(shell GO15VENDOREXPERIMENT=1 go list ./... | grep -v vendor)
NOVENDOR=$(shell find . -path ./vendor -prune -o -name '*.go' -print)

all: lint build

build:
	gox -arch=amd64 -os="linux darwin windows"

lint: format
	cd -P . && go vet $(PACKAGES)
	for package in $(PACKAGES); do \
		golint -min_confidence .25 $$package ; \
	done

format:
	gofmt -w -s $(NOVENDOR)
