# base path used to install.
DESTDIR ?= /usr/local

# Used to populate variables in version package.
PKG=github.com/containerd/containerd

VERSION ?= $(shell git describe --match 'v[0-9]*' --dirty='.m' --always)
REVISION=$(shell git rev-parse HEAD)$(shell if ! git diff --no-ext-diff --quiet --exit-code; then echo .m; fi)

CONTAINERD_LDFLAGS=-ldflags '-X $(PKG)/version.Version=$(VERSION) -X $(PKG)/version.Revision=$(REVISION)'

# go build command
GO_BUILD_BINARY=go build -o $@ ./$<
GO_BUILD_ENV ?=
BPF_MAKE_ARGS ?=

GO_GENERATE_CMD=go generate ./...

COMMANDS=embedshim-containerd embedshim-runcext

# binaries
BINARIES=$(addprefix bin/,$(COMMANDS))

.PHONY: build binaries binaries-arm64-generic binaries-arm64-lse

binaries: $(BINARIES)

binaries-arm64-generic:
	$(MAKE) binaries GO_BUILD_ENV="GOOS=linux GOARCH=arm64 GOARM64=v8.0" BPF_MAKE_ARGS="ARCH=arm64"

binaries-arm64-lse:
	$(MAKE) binaries GO_BUILD_ENV="GOOS=linux GOARCH=arm64 GOARM64=v8.1,lse" BPF_MAKE_ARGS="ARCH=arm64"

# force to rebuild
REBUILD:

bin/embedshim-containerd: cmd/embedshim-containerd REBUILD
	@echo "$@"
	@make -C bpf $(BPF_MAKE_ARGS)
	$(GO_GENERATE_CMD)
	@$(GO_BUILD_ENV) go build -o $@ ${CONTAINERD_LDFLAGS} ./cmd/embedshim-containerd

bin/embedshim-runcext: cmd/embedshim-runcext REBUILD
	@echo "$@"
	@$(GO_BUILD_ENV) go build -o $@ ./cmd/embedshim-runcext

# install binaries
install:
	@echo "$@ $(DESTDIR)/$(BINARIES)"
	@mkdir -p $(DESTDIR)/bin
	@install $(BINARIES) $(DESTDIR)/bin

clean:
	@rm -rf ./bin
	@rm -rf ./bpf/.output
