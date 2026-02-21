.PHONY: unit-test ticat
.DEFAULT_GOAL := default

REPO    := github.com/innerr/ticat

GOOS    := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
GOARCH  := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
GOENV   := GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO      := $(GOENV) go
GOBUILD := $(GO) build $(BUILD_FLAG)
GOTEST  := GO111MODULE=on CGO_ENABLED=1 go test -p 3
SHELL   := /usr/bin/env bash

_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
_GITREF := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
_DIRTY  := $(shell git diff --stat 2>/dev/null)
_DIRTY_HASH := $(shell if [ -n "$(_DIRTY)" ]; then git diff | shasum -a 256 | cut -c1-8; fi)

COMMIT  := $(if $(COMMIT),$(COMMIT),$(_COMMIT))
GITREF  := $(if $(GITREF),$(GITREF),$(_GITREF))
DIRTY_HASH := $(if $(DIRTY_HASH),$(DIRTY_HASH),$(_DIRTY_HASH))

LDFLAGS := -w -s
LDFLAGS += -X "$(REPO)/pkg/version.GitHash=$(COMMIT)"
LDFLAGS += -X "$(REPO)/pkg/version.GitRef=$(GITREF)"
LDFLAGS += -X "$(REPO)/pkg/version.GitDirtyHash=$(DIRTY_HASH)"
LDFLAGS += $(EXTRA_LDFLAGS)

FILES   := $$(find . -name "*.go")

default: unit-test ticat

ticat:
	$(GOBUILD) -ldflags '$(LDFLAGS)' -o bin/ticat ./pkg/main

unit-test:
	mkdir -p bin/cover
	$(GOTEST) ./pkg/... -covermode=count -coverprofile bin/cover/cov.unit-test.out
