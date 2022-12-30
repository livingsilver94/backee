GO ?= go

BINNAME := backee
OUTPATH := $(CURDIR)/build/$(BINNAME)

$(OUTPATH):
	$(GO) build -o $(OUTPATH) main.go

.PHONY: check
check:
	$(GO) test ./...

DESTDIR ?= /
prefix  ?= /usr/local
bindir  ?= $(prefix)/bin

install: $(OUTPATH)
	install -Dm00755 $(OUTPATH) -t $(DESTDIR)/$(bindir)

.PHONY: clean
clean:
	rm $(OUTPATH)
