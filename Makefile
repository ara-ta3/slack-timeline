GOOS=
GOARCH=
GLIDE=$(shell which glide)
config=config.json
goos_opt=GOOS=$(GOOS)
goarch_opt=GOARCH=$(GOARCH)
out=slacktimeline
out_opt="-o $(out)"

help:
	@cat Makefile

run: install $(config)
	go run main.go config.go -c $(config)

install: 
	$(GLIDE) install

build: install
	$(goos_opt) $(goarch_opt) go build $(out_opt)

build_for_linux:
	$(MAKE) build GOOS=linux GOARCH=amd64 out_opt=""

build_for_local:
	$(MAKE) build goos_opt= goarch_opt= out_opt=

test:
	go test -v ./timeline/...

$(config): config.sample.json
	cp -f $< $@
