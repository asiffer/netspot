# ---------------- #
# NetSpot Makefile #
# ---------------- #

# package details
PACKAGE_NAME := netspot
VERSION := 1.3
PACKAGE_DESC := "A simple IDS with statistical learning"
MAINTAINER := asiffer

GOEXEC := $(shell which go)
GOPATH := ${GOPATH}
$(info GO="$(GOEXEC)")
$(info GOPATH="$(GOPATH)")

# environment
ARCH =
OS =
ifndef ARCH
  ARCH := $(shell $(GOEXEC) env | grep GOARCH= | sed -e 's/GOARCH=//' -e 's/"//g' )
endif

ifndef OS
  OS := $(shell $(GOEXEC) env | grep GOOS= | sed -e 's/GOOS=//' -e 's/"//g' )
endif

# Print environment variable
$(info ARCH="$(ARCH)")
$(info OS="$(OS)")

# sources
# SRC_DIR := $(GOPATH)/src/netspot
SRC_DIR := $(shell pwd)
EXTRA_DIR = $(SRC_DIR)/extra

$(info SRC_DIR="$(SRC_DIR)")

# golang compiler
#GO := /usr/local/go/bin/go
GO := GOARCH=$(ARCH) GOOS=$(OS) $(shell which go)
GO_BUILD_EXTRA_FLAGS := 

# build directories
BIN_DIR := $(SRC_DIR)/bin
DIST_DIR := $(SRC_DIR)/dist
DOCKER_DIR := $(DIST_DIR)/docker
DEBIAN_DIR := $(DIST_DIR)/debian
SNAP_DIR := $(DIST_DIR)/snap
# PKG_BUILD_DIR = $(PKG_DIR)/$(OS)/$(ARCH)

# installation
DESTDIR :=
INSTALL_BIN_DIR := $(DESTDIR)/usr/bin
INSTALL_CONF_DIR := $(DESTDIR)/etc/netspot
INSTALL_SERVICE_DIR := $(DESTDIR)/lib/systemd/system
CTL_INSTALL_CONF_DIR := $(DESTDIR)/etc/netspotctl

# go dependencies
GO_DEP_NETSPOT := 	"fatih/color" \
					"fsnotify/fsnotify" \
					"rs/zerolog" \
					"spf13/viper" \
					"urfave/cli" \
					"google/gopacket" \
					"asiffer/gospot" \
					"influxdata/influxdb1-client/v2" \
					"gorilla/mux" \

GO_DEP_NETSPOTCTL := 	"rs/zerolog" \
						"spf13/viper" \
						"c-bata/go-prompt" \




# fancyness
OK := "[\033[32mOK\033[0m]"

# PHONY actions
.PHONY: build debian docker snap

# main actions
default: build

deps: netspot_deps netspotctl_deps

build: build_netspot build_netspotctl

install: install_bin install_config install_service # post_install



# atomic actions

netspot_deps:
	@echo "\033[34m[Retrieving netspot build dependencies]\033[0m"
	@for dep in $(GO_DEP_NETSPOT) ; do echo "Getting "$$dep"... "; go get -u github.com/$$dep; done

netspotctl_deps:
	@echo "\033[34m[Retrieving netspotctl build dependencies]\033[0m"
	@for dep in $(GO_DEP_NETSPOTCTL) ; do echo "Getting "$$dep"... "; go get -u github.com/$$dep; done

build_netspot:
	@echo "\033[34m[Building netspot]\033[0m"
	@export GOPATH=$(GOPATH)
	@echo -n "Building go package...               "
	@$(GO) build $(GO_BUILD_EXTRA_FLAGS) -o $(BIN_DIR)/netspot $(SRC_DIR)/netspot.go
	@echo $(OK)

build_netspotctl:
	@echo "\033[34m[Building netspotctl]\033[0m"
	@echo -n "Building go package...               "
	@$(GO) build $(GO_BUILD_EXTRA_FLAGS) -o $(BIN_DIR)/netspotctl $(SRC_DIR)/netspotctl/*.go 
	@echo $(OK)


install_config:
	@echo "\033[34m[Installing configurations]\033[0m"
	@echo -n "Creating config directories...       "
	@mkdir -p $(INSTALL_CONF_DIR)
	@mkdir -p $(CTL_INSTALL_CONF_DIR)
	@echo $(OK)
	@echo -n "Installing netspot config file...    "
	@install $(EXTRA_DIR)/netspot.toml $(INSTALL_CONF_DIR)/
	@echo $(OK)
	@echo -n "Installing netspotctl config file... "
	@install $(EXTRA_DIR)/netspotctl.toml $(INSTALL_CONF_DIR)/
	@echo $(OK)

install_bin:
	@echo "\033[34m[Installing binaries]\033[0m"
	@echo -n "Creating directory...                "
	@mkdir -p $(INSTALL_BIN_DIR)
	@echo $(OK)
	@echo -n "Installing netspot...                "
	@install $(BIN_DIR)/netspot $(INSTALL_BIN_DIR)/
	@echo $(OK)
	@echo -n "Installing netspotctl...             "
	@install $(BIN_DIR)/netspotctl $(INSTALL_BIN_DIR)/
	@echo $(OK)

install_service:
	@echo -n "Creating directory...                "
	@mkdir -p $(INSTALL_SERVICE_DIR)
	@echo $(OK)
	@echo "\033[34m[Installing service]\033[0m"
	@echo -n "Installing netspot service file...   "
	@install $(EXTRA_DIR)/netspot.service $(INSTALL_SERVICE_DIR)/
	@echo $(OK)

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

purge:
	# rm -rf $(PKG_DIR)