GO ?= go

BINNAME := backee
OUTPATH := $(CURDIR)/build/$(BINNAME)

.PHONY: build
build:
	$(GO) build -o $(OUTPATH) main.go

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
