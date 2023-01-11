all:
	GOPRIVATE=git.sr.ht go build
.PHONY: all

test:
	GOPRIVATE=git.sr.ht go test
