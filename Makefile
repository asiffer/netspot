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
ARCH := 
OS := 
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
SRC_DIR := $(GOPATH)/src/netspot
EXTRA_DIR = $(SRC_DIR)/extra

$(info SRC_DIR="$(SRC_DIR)")

# golang compiler
#GO := /usr/local/go/bin/go
GO := env GOARCH=$(ARCH) GOOS=$(OS) $(shell which go)

# build directories
BIN_DIR := $(SRC_DIR)/bin
PKG_DIR := $(SRC_DIR)/dist
PKG_BUILD_DIR = $(PKG_DIR)/$(OS)/$(ARCH)

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

# main actions
default: build

deps: netspot_deps netspotctl_deps

build: build_netspot build_netspotctl

install: install_bin install_config install_service # post_install

package: build pre_debian debian


# atomic actions

netspot_deps:
	@echo "\033[34m[Retrieving netspot build dependencies]\033[0m"
	@for dep in $(GO_DEP_NETSPOT) ; do echo "Getting "$$dep"... "; go get -u github.com/$$dep; done

netspotctl_deps:
	@echo "\033[34m[Retrieving netspotctl build dependencies]\033[0m"
	@for dep in $(GO_DEP_NETSPOT) ; do echo "Getting "$$dep"... "; go get -u github.com/$$dep; done

build_netspot:
	@echo "\033[34m[Building netspot]\033[0m"
	@export GOPATH=$(GOPATH)
	@echo -n "Building go package...               "
	@$(GO) build -o $(BIN_DIR)/netspot $(SRC_DIR)/*.go
	@echo $(OK)

build_netspotctl:
	@echo "\033[34m[Building netspotctl]\033[0m"
	@echo -n "Building go package...               "
	@$(GO) build -o $(BIN_DIR)/netspotctl $(SRC_DIR)/netspotctl/*.go 
	@echo $(OK)


install_config:
	@echo "\033[34m[Installing configurations]\033[0m"
	@echo -n "Creating config directories...       "
	@mkdir -p $(INSTALL_CONF_DIR)
	@mkdir -p $(CTL_INSTALL_CONF_DIR)
	@echo $(OK)
	@echo -n "Installing netspot config file...    "
	@install $(EXTRA_DIR)/netspot.toml $(INSTALL_BIN_DIR)/
	@echo $(OK)
	@echo -n "Installing netspotctl config file... "
	@install $(EXTRA_DIR)/netspotctl.toml $(INSTALL_BIN_DIR)/
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

# post_install:
# 	@echo -n "Creating user 'netspot'...           "
# 	@adduser --no-create-home --disabled-password --disabled-login netspot
# 	@echo $(OK)
# 	@echo -n "Setting CAP_NET_RAW, CAP_NET_ADMIN capabilities... "
# 	@setcap 'CAP_NET_RAW+eip CAP_NET_ADMIN+eip' $(INSTALL_BIN_DIR)/netspot
# 	@echo $(OK)


# pre_debian:
# 	@mkdir -p $(PKG_DIR)
# 	@mkdir -p $(PKG_DIR)/$(OS)
# 	@mkdir -p $(PKG_DIR)/$(OS)/$(ARCH)

# debian:
# 	@echo "\033[34m[Packaging]\033[0m"
# 	@echo "Running fpm...                       "
# 	@fpm -f -s dir \
# 			-t deb \
# 			-n $(PACKAGE_NAME) \
# 			-v $(VERSION) \
# 			-a $(ARCH) \
# 			-m "alban.siffer@irisa.fr" \
# 			-p $(PKG_BUILD_DIR)/ \
# 			--category "network" \
# 			--description $(PACKAGE_DESC) \
# 			--deb-suggests "influxdb" \
# 			--deb-suggests "grafana" \
# 			--deb-systemd $(EXTRA_DIR)/netspot.service \
# 			--after-install $(EXTRA_DIR)/post-install.sh \
# 			$(BIN_DIR)/netspot=$(INSTALL_BIN_DIR)/ \
# 			$(BIN_DIR)/netspotctl=$(INSTALL_BIN_DIR)/ \
# 			$(EXTRA_DIR)/netspot.toml=$(INSTALL_CONF_DIR)/ \
# 			$(EXTRA_DIR)/netspotctl.toml=$(CTL_INSTALL_CONF_DIR)/
# 	@echo $(OK)"\n"

docker:
# Build a docker image according to the OS and the architecture
# They can be modified through ARCH and OS
	docker build -t $(MAINTAINER)/netspot-$(ARCH) --build-arg GOARCH=$(ARCH) --build-arg GOOS=$(OS) ./
	docker tag $(MAINTAINER)/netspot-$(ARCH) $(MAINTAINER)/netspot-$(ARCH):$(VERSION)
	docker save -o images/docker-netspot-$(ARCH)_$(VERSION).tar.gz $(MAINTAINER)/netspot-$(ARCH):$(VERSION)

purge:
	rm -rf $(PKG_DIR)