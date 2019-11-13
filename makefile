SHELL = bash


COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags 2> /dev/null || echo v0)
PERMALINK ?= $(shell git name-rev --name-only --tags --no-undefined HEAD &> /dev/null && echo latest || echo canary)
REPOSITORY = qlik/qliksense-operator

BINDIR = bin/qliksense-operator



ifeq ($(CLIENT_PLATFORM),windows)
FILE_EXT=.exe
else ifeq ($(RUNTIME_PLATFORM),windows)
FILE_EXT=.exe
else
FILE_EXT=
endif

REGISTRY ?= $(USER)

.PHONY: build
build: build-client build-docker

push: build docker-push

build-client: generate
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/qliksense-operator$(FILE_EXT)

generate:
	go generate ./...

test: test-unit
	$(BINDIR)/qliksense-operator$(FILE_EXT) version

test-unit: build
	go test ./...


build-docker:
	docker build $(BINDIR) -t qlik/qliksense-operator:$(VERSION)

docker-push: build-docker docker-login
	docker push $(REPOSITORY)

docker-login:
	docker login

git tag: 
	git tag $1 
	make push
