GOOS=
GOARCH=
goos_opt=GOOS=$(GOOS)
goarch_opt=GOARCH=$(GOARCH)
out=slacktimeline
out_opt="-o $(out)"

help:
	@cat Makefile

install:
	go get github.com/pkg/errors
	go get golang.org/x/net/websocket
	go get github.com/stretchr/testify/assert
	go get github.com/syndtr/goleveldb/leveldb

build: 
	 $(goos_opt) $(goarch_opt) go build $(out_opt)

build_for_linux:
	$(MAKE) build GOOS=linux GOARCH=amd64

test:
	go test -v ./timeline/...
