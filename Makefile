MKFILE_DIR := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))
BUILDOPTS?=-v
BUILD_DIR?=build
BINARY?=$(BUILD_DIR)/in2un
SYSTEM:=

all: test build

.PHONY: test
test:
	go test -v ./...

.PHONY: fmt
fmt:
	go fmt -mod=mod *.go
	git diff --exit-code

.PHONY: vet
vet:
	go vet *.go

.PHONY: staticheck
staticcheck:
	staticcheck ./...

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

bin: clean $(BINARY)

build: test fmt vet staticcheck bin

$(BINARY):
	$(SYSTEM) go build $(BUILDOPTS) -o $(MKFILE_DIR)/$(BINARY)

