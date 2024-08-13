GO ?= go # if using docker, should not need to be installed/linked
GOBINREL = build/bin
GOBIN = $(CURDIR)/$(GOBINREL)

GOBUILD = $(GO) build

default:
	@# Note: $* is replaced by the command name
	@echo "Building go-indexer"
	@cd ./src && $(GOBUILD) -o $(GOBIN)/go-indexer
	@echo "Run \"$(GOBIN)/go-indexer\" to launch go-indexer as a producer."