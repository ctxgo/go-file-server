.PHONY: build image help

# Set global environment
export GOPROXY=https://goproxy.cn,direct

# Build the Docker image
VERSION ?= latest
IMG ?= ctxgo/go-file-server:$(VERSION)



# Set global settings
CONTAINER_RUNTIME ?= $(if $(shell command -v docker),docker,podman)

# Go toolchain setup
GO ?= go
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GIT_COMMIT := $(shell git rev-parse HEAD 2> /dev/null || echo "unknown")
EXTRA_LDFLAGS ?=
LDFLAGS := -ldflags '-X main.gitCommit=${GIT_COMMIT} $(EXTRA_LDFLAGS)'


ifeq ($(DEBUG), 1)
    override GOGCFLAGS += -N -l
endif

ifeq ($(GOOS), linux)
    ifneq ($(GOARCH),$(filter $(GOARCH),mips mipsle mips64 mips64le ppc64 riscv64))
        GO_DYN_FLAGS="-buildmode=pie"
    endif
endif

# Build the binary
build:
	$(GO) build  ${GO_DYN_FLAGS} ${LDFLAGS} -gcflags "$(GOGCFLAGS)"  -o build/go-file-server


image:
	$(CONTAINER_RUNTIME) build -t $(IMG) -f Dockerfile .


# Help command
help:
	@echo "Makefile commands:"
	@echo
	@echo "  build         - Build the binary using the Go toolchain."
	@echo "                 - Example: make build"
	@echo "                 - This will compile the Go application and output the binary to 'build/go-file-server'."
	@echo "                 - Go Proxy: $(GOPROXY)"
	@echo "                 - Go OS/ARCH: $(GOOS)/$(GOARCH)"
	@echo "                 - Git Commit: $(GIT_COMMIT)"
	@echo
	@echo "  image         - Build a Docker image using the specified container runtime."
	@echo "                 - Example: make image [VERSION=v1.0.0]"
	@echo "                 - This will build a Docker image. The 'VERSION' parameter is optional."
	@echo "                 - If not specified, the image will be tagged as '$(VERSION)'."
	@echo "                 - Default Image Name: $(VERSION)"
	@echo "                 - Container Runtime: $(CONTAINER_RUNTIME)"