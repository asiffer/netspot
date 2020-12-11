# ---------------- #
# NetSpot Makefile #
# ---------------- #

# Shell for $shell commands
SHELL        := /bin/bash

# Fancyness
SEP  := $(shell printf "%80s" | tr " " "-")
OK   := "[\033[92mOK\033[0m]"

# environment
GOEXEC := $(shell which go)
GOPATH := ${GOPATH}
ARCH   ?= $(shell $(GOEXEC) env | grep GOARCH= | sed -e 's/GOARCH=//' -e 's/"//g' )
OS     ?= $(shell $(GOEXEC) env | grep GOOS=   | sed -e 's/GOOS=//'   -e 's/"//g' )
CC     ?= $(shell command -v cc)
AR     ?= $(shell command -v ar)
LD     ?= $(shell command -v ld)

# package details
PACKAGE_NAME := netspot
VERSION      := $(shell grep Version cmd/app.go | sed -e 's/[",]//g' -e 's/Version://g' -e 's/[\t\ ]//g')
PACKAGE_DESC := "A simple IDS with statistical learning"
MAINTAINER   := asiffer


# sources
SRC_DIR   := $(shell pwd)
EXTRA_DIR := $(SRC_DIR)/extra

# golang compiler
GO                   := CC=$(CC) LD=$(LD) GOARCH=$(ARCH) GOOS=$(OS) $(shell which go)
GO_LDFLAGS           := -s -w
GO_BUILD_EXTRA_FLAGS := -a

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

# Print environment variable
$(info $(SEP))
$(info GO="$(GOEXEC)")
$(info GOPATH="$(GOPATH)")
$(info ARCH="$(ARCH)")
$(info OS="$(OS)")
$(info CC="$(CC)")
$(info AR="$(AR)")
$(info LD="$(LD)")
$(info SRC_DIR="$(SRC_DIR)")
$(info VERSION="$(VERSION)")
$(info $(SEP))

# PHONY actions
.PHONY: build snap

# main actions
default: build
build: build_netspot
install: install_bin

deps:
	@echo -n "Retrieving dependencies...           "
	@$(GO) get -u ./...
	@echo -e $(OK)

build_netspot:
	@echo -e "\033[93m[Building netspot]\033[0m"
	@export GOPATH=$(GOPATH)
	@echo -n "Building go package...               "
	@$(GO) build $(GO_BUILD_EXTRA_FLAGS) -ldflags='$(GO_LDFLAGS)' -o $(BIN_DIR)/netspot-$(VERSION)-$(ARCH)-$(OS) $(SRC_DIR)/*.go
	@echo -e $(OK)

build_netspot_static:
	@echo -e "\033[93m[Building netspot (musl)]\033[0m"
	@export GOPATH=$(GOPATH)
	$(eval GO_LDFLAGS += -extldflags "-static")
	@echo -n "Building go package...               "
	@$(GO) build $(GO_BUILD_EXTRA_FLAGS) -ldflags='$(GO_LDFLAGS)' -o $(BIN_DIR)/netspot-$(VERSION)-$(ARCH)-$(OS)-static $(SRC_DIR)/*.go
	@echo -e $(OK)

install_bin:
	@echo -e "\033[93m[Installing binaries]\033[0m"
	@echo -en "Creating directory...                "
	@mkdir -p $(INSTALL_BIN_DIR)
	@echo $(OK)
	@echo -en "Installing netspot...                "
	@install $(BIN_DIR)/netspot $(INSTALL_BIN_DIR)/
	@echo -e $(OK)

snap:
	snapcraft
	mkdir -p $(SNAP_DIR)
	mv *.snap $(SNAP_DIR)

clean:
	@echo -en "Removing netspot binary   "
	@rm -f $(BIN_DIR)/netspot
	@echo -e $(OK)

