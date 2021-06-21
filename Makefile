SHELL := /bin/bash
GO := GO111MODULE=on GO15VENDOREXPERIMENT=1 go
#GO := go
#GO_NOMOD := GO111MODULE=off go

# set dev version unless VERSION is explicitly set via environment
VERSION ?= $(shell echo "$$(git describe --abbrev=0 --tags 2>/dev/null)-dev+$(REV)" | sed 's/^v//')

GO_VERSION       := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
PACKAGE_DIRS     := $(shell $(GO) list ./... | grep -v /vendor/ | grep -v e2e)
PEGOMOCK_PACKAGE := github.com/petergtz/pegomock
GO_DEPENDENCIES  := $(shell find . -type f -name '*.go')

BUILDFLAGS := -trimpath
CGO_ENABLED = 0
BUILDTAGS :=

GOPATH1=$(firstword $(subst :, ,$(GOPATH)))

export PATH := $(PATH):$(GOPATH1)/bin

build: $(GO_DEPENDENCIES)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(BUILDTAGS) $(BUILDFLAGS) -o action cmd/action.go

test:
	DISABLE_SSO=true CGO_ENABLED=$(CGO_ENABLED) $(GO) test -coverprofile=coverage.out $(PACKAGE_DIRS)
