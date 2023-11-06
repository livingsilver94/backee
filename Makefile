GO ?= go

BINNAME := backee
OUTPATH := $(CURDIR)/build/$(BINNAME)

VERSION := $(shell git describe --tags)
ifeq ($(VERSION),)
	VERSION = $(shell git rev-parse HEAD)
endif

.PHONY: build
build:
	$(GO) build -o $(OUTPATH) -ldflags "-X github.com/livingsilver94/backee/cli.Version=$(VERSION)" log.go main.go

.PHONY: check
check:
	$(GO) test ./...

DESTDIR ?= /
prefix  ?= /usr/local
bindir  ?= $(prefix)/bin

.PHONY: install
install: build
	install -Dm00755 $(OUTPATH) -t $(DESTDIR)/$(bindir)

.PHONY: clean
clean:
	rm $(OUTPATH)
