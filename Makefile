CGO_ENABLED 	:= 1
CGO_CFLAGS		:=-I$(CURDIR)/vendor/libvix/include -Werror
CGO_LDFLAGS		:=-L$(CURDIR)/vendor/libvix -lvixAllProducts -ldl -lpthread

DYLD_LIBRARY_PATH	:=$(CURDIR)/vendor/libvix
LD_LIBRARY_PATH		:=$(CURDIR)/vendor/libvix

NAME 		:= go-osx-builder
VERSION 	:= v1.0.0
PLATFORM 	:= $(shell go env | grep GOHOSTOS | cut -d '"' -f 2)
ARCH 		:= $(shell go env | grep GOARCH | cut -d '"' -f 2)

export CGO_CFLAGS CGO_LDFLAGS DYLD_LIBRARY_PATH LD_LIBRARY_PATH CGO_ENABLED

build:
	go build -ldflags "-X main.Version $(VERSION)" -o build/$(NAME)

install:
	go install -ldflags "-X main.Version $(VERSION)" -v

dist: build
	$(eval FILES := $(shell ls build))
	@rm -rf dist && mkdir dist
	for f in $(FILES); do \
		(cd $(shell pwd)/build && tar -cvzf ../dist/$${f}_$(VERSION)_$(PLATFORM)_$(ARCH).tar.gz *); \
		(cd $(shell pwd)/dist && shasum -a 512 $${f}_$(VERSION)_$(PLATFORM)_$(ARCH).tar.gz > $${f}_$(VERSION)_$(PLATFORM)_$(ARCH).tar.gz.sha512); \
		echo $$f; \
	done

deps:
	go get github.com/c4milo/github-release
	go get github.com/mitchellh/gox
	go get github.com/satori/go.uuid
	go get gopkg.in/unrolled/render.v1

release: dist
	@latest_tag=$$(git describe --tags `git rev-list --tags --max-count=1`); \
	comparison="$$latest_tag..HEAD"; \
	if [ -z "$$latest_tag" ]; then comparison=""; fi; \
	changelog=$$(git log $$comparison --oneline --no-merges --reverse); \
	github-release c4milo/$(NAME) $(VERSION) "$$(git rev-parse --abbrev-ref HEAD)" "**Changelog**<br/>$$changelog" 'dist/*'; \
	git pull

test:
	        go test ./...

clean:
	        go clean ./...

.PHONY: dist release build test install clean
