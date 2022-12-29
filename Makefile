GO ?= go

BINNAME := backee
OUTPATH := build/$(BINNAME)

.PHONY: build
build:
	$(GO) build -o $(OUTPATH) main.go

.PHONY: check
check:
	$(GO) test ./...

.PHONY: clean
clean:
	rm $(OUTPATH)
