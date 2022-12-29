GO ?= go

BINNAME := backee
OUTPATH := build/$(BINNAME)

.PHONY: build
build:
	$(GO) build -o $(OUTPATH) main.go

.PHONY: clean
clean:
	rm $(OUTPATH)
