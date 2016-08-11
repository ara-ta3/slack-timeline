
help:
	@cat Makefile

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build
