# ---------------- #
# NetSpot Makefile #
# ---------------- #

# package details
PACKAGE_NAME := netspot
VERSION      := 2.0
PACKAGE_DESC := "A simple IDS with statistical learning"
MAINTAINER   := asiffer

# Shell for $shell commands
SHELL        := /bin/bash

GOEXEC       := $(shell which go)
GOPATH       := ${GOPATH}
$(info GO="$(GOEXEC)")
$(info GOPATH="$(GOPATH)")

# environment
ARCH ?= $(shell $(GOEXEC) env | grep GOARCH= | sed -e 's/GOARCH=//' -e 's/"//g' )
OS   ?= $(shell $(GOEXEC) env | grep GOOS=   | sed -e 's/GOOS=//'   -e 's/"//g' )

# Print environment variable
$(info ARCH="$(ARCH)")
$(info OS="$(OS)")

# sources
SRC_DIR   := $(shell pwd)
EXTRA_DIR := $(SRC_DIR)/extra

$(info SRC_DIR="$(SRC_DIR)")

# API
API_DIR         := $(SRC_DIR)/api
PROTO_CC        := $(shell command -v protoc)
PROTO_INCLUDE   := -I$(API_DIR) -I/usr/include/google/protobuf/
# -I/home/asr/Documents/Work/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.9.0/third_party/googleapis/
PROTO_MODULES   := miner
PROTO_FILES     := $(foreach mod, $(PROTO_MODULES), $(API_DIR)/$(mod).proto) 
INTERFACE_FILES := $(foreach mod, $(PROTO_MODULES), $(API_DIR)/$(mod).pb.go) 

# golang compiler
GO                   := GOARCH=$(ARCH) GOOS=$(OS) $(shell which go)
GO_BUILD_EXTRA_FLAGS := 

# build directories
BIN_DIR    := $(SRC_DIR)/bin
DIST_DIR   := $(SRC_DIR)/dist
DOCKER_DIR := $(DIST_DIR)/docker
DEBIAN_DIR := $(DIST_DIR)/debian
SNAP_DIR   := $(DIST_DIR)/snap

# installation
DESTDIR              :=
INSTALL_BIN_DIR      := $(DESTDIR)/usr/bin
INSTALL_CONF_DIR     := $(DESTDIR)/etc/netspot
INSTALL_SERVICE_DIR  := $(DESTDIR)/lib/systemd/system



# fancyness
OK := "[\033[32mOK\033[0m]"

# PHONY actions
.PHONY: build debian docker snap api

# main actions
default: build
build: api build_netspot
install: install_bin install_config install_service # post_install

deps:
	@echo -n "Retrieving dependencies...           "
	@$(GO) get -u ./...
	@echo -e $(OK)

build_netspot:
	@echo -e "\033[34m[Building netspot]\033[0m"
	@export GOPATH=$(GOPATH)
	@echo -n "Building go package...               "
	@$(GO) build $(GO_BUILD_EXTRA_FLAGS) -o $(BIN_DIR)/netspot $(SRC_DIR)/*.go
	@echo -e $(OK)

install_config:
	@echo -e "\033[34m[Installing configurations]\033[0m"
	@echo -en "Creating config directories...       "
	@mkdir -p $(INSTALL_CONF_DIR)
	@mkdir -p $(CTL_INSTALL_CONF_DIR)
	@echo $(OK)
	@echo -en "Installing netspot config file...    "
	@install $(EXTRA_DIR)/netspot.toml $(INSTALL_CONF_DIR)/
	@echo -e $(OK)

install_bin:
	@echo -e "\033[34m[Installing binaries]\033[0m"
	@echo -en "Creating directory...                "
	@mkdir -p $(INSTALL_BIN_DIR)
	@echo $(OK)
	@echo -en "Installing netspot...                "
	@install $(BIN_DIR)/netspot $(INSTALL_BIN_DIR)/
	@echo -e $(OK)

install_service:
	@echo -en "Creating directory...                "
	@mkdir -p $(INSTALL_SERVICE_DIR)
	@echo -e $(OK)
	@echo -e "\033[34m[Installing service]\033[0m"
	@echo -en "Installing netspot service file...   "
	@install $(EXTRA_DIR)/netspot.service $(INSTALL_SERVICE_DIR)/
	@echo -e $(OK)

debian:
	mkdir -p $(DEBIAN_DIR)
	dpkg-buildpackage -us -uc -b
	cp -u ../*.deb $(DEBIAN_DIR)


docker: build
	# Build a docker image according to the OS and the architecture
	# They can be modified through ARCH and OS
	# It is likely to ask root privileges
	docker build --rm -t $(MAINTAINER)/netspot-$(ARCH) --build-arg GOARCH=$(ARCH) --build-arg GOOS=$(OS) ./
	docker tag $(MAINTAINER)/netspot-$(ARCH) $(MAINTAINER)/netspot-$(ARCH):$(VERSION)
	mkdir -p $(DOCKER_DIR)
	docker save -o $(DOCKER_DIR)/docker-netspot-$(ARCH)_$(VERSION).tar.gz $(MAINTAINER)/netspot-$(ARCH):$(VERSION)

snap:
	snapcraft
	mkdir -p $(SNAP_DIR)
	mv *.snap $(SNAP_DIR)


%.pb.go: %.proto
	@echo -en "Generating $@ \t"
	@$(PROTO_CC) $(PROTO_INCLUDE) $^ --proto_path=$(API_DIR) --go_out=$(API_DIR) --go_opt=paths=source_relative
	@gofmt -w $@
	@goimports -w $@
	@echo -e $(OK)

api: $(INTERFACE_FILES)


clean:
	@echo -en "Removing auto-generated gRPC files   "
	@rm -f $(API_DIR)/*.pb.go
	@echo -e $(OK)

purge:
	# rm -rf $(PKG_DIR)