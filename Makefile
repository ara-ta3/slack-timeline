help:
	@cat Makefile

install:
	go get github.com/pkg/errors
	go get golang.org/x/net/websocket
	go get github.com/stretchr/testify/assert
	go get github.com/syndtr/goleveldb/leveldb

build: 
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build

build_for_linux:
	$(MAKE) build GOOS=linux GOARCH=amd64

test:
	go test -v ./timeline/...
