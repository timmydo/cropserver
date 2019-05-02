HASH := $(shell git rev-parse --short HEAD)

.PHONY: all
all: build

.PHONY: build
build:
	go build -o httpd

.PHONY: version
version:
	echo $(HASH)

.PHONY: docker
docker:
	docker build -t timmydo/cropserver:git-$(HASH) .

.PHONY: dev
dev:
	docker run -p 8080:80 --rm -it -e GO111MODULE=on --workdir /go/src/github.com/timmydo/cropserver --volume $(CURDIR):/go/src/github.com/timmydo/cropserver quay.io/deis/go-dev:latest

.PHONY: run
run: build
	./httpd --port 80 --file image.png --url '/testimage'