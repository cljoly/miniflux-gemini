GOFILES      := $(shell find . -type f -name '*.go')

miniflux-gemini: $(GOFILES) gemtext/templates/*
	GOPRIVATE=git.sr.ht go build .

all: miniflux-gemini
.PHONY: all

srv: miniflux-gemini
	./miniflux-gemini

test:
	GOPRIVATE=git.sr.ht go test
