# ---------------- #
# NetSpot Makefile #
# ---------------- #

# package details
PACKAGE_NAME := netspot
VERSION := 1.2
PACKAGE_DESC := "A basic IDS with statistical learning"

# sources
SRC_DIR := $(GOPATH)/src/netspot
EXTRA_DIR = $(SRC_DIR)/extra

# build directories
BIN_DIR := $(SRC_DIR)/bin
PKG_DIR := $(SRC_DIR)/dist
PKG_BUILD_DIR = $(PKG_DIR)/$(OS)/$(ARCH)

# installation
DEST_DIR :=
INSTALL_BIN_DIR := $(DEST_DIR)/usr/local/bin
INSTALL_CONF_DIR := $(DEST_DIR)/etc/netspot
INSTALL_SERVICE_DIR := $(DEST_DIR)/lib/systemd/system
CTL_INSTALL_CONF_DIR := $(DEST_DIR)/etc/netspotctl

# environment
ARCH := amd64
OS := linux

# fancyness
OK := "[\033[32mOK\033[0m]"

# main actions
default: build

build: build_netspot build_netspotctl

install: install_bin install_config install_bin install_service post_install

package: pre_debian debian


# atomic ations 
build_netspot:
	@echo "\033[34m[Building netspot]\033[0m"
	@echo -n "Building go package...               "
	@GOOS=$(OS) GOARCH=$(ARCH) go build -o $(BIN_DIR)/netspot $(SRC_DIR)/netspot.go
	@echo $(OK)


build_netspotctl:
	@echo "\033[34m[Building netspotctl]\033[0m"
	@echo -n "Building go package...               "
	@GOOS=$(OS) GOARCH=$(ARCH) go build -o $(BIN_DIR)/netspotctl $(SRC_DIR)/netspotctl/netspotctl.go 
	@echo $(OK)


pre_install:
	@

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

post_install:
	@echo -n "Creating user 'netspot'...           "
	@adduser --no-create-home --disabled-password --disabled-login netspot
	@echo $(OK)
	@echo -n "Setting CAP_NET_RAW, CAP_NET_ADMIN capabilities... "
	@setcap 'CAP_NET_RAW+eip CAP_NET_ADMIN+eip' $(INSTALL_BIN_DIR)/netspot
	@echo $(OK)


pre_debian:
	@mkdir -p $(PKG_DIR)
	@mkdir -p $(PKG_DIR)/$(OS)
	@mkdir -p $(PKG_DIR)/$(OS)/$(ARCH)

debian:
	@echo "\033[34m[Packaging]\033[0m"
	@echo "Running fpm...                       "
	@fpm -f -s dir \
			-t deb \
			-n $(PACKAGE_NAME) \
			-v $(VERSION) \
			-a $(ARCH) \
			-m "alban.siffer@irisa.fr" \
			-p $(PKG_BUILD_DIR)/ \
			--category "network" \
			--description $(PACKAGE_DESC) \
			-d "libspot" \
			--deb-suggests "influxdb" \
			--deb-suggests "grafana" \
			--deb-systemd $(EXTRA_DIR)/netspot.service \
			--after-install $(EXTRA_DIR)/post-install.sh \
			$(BIN_DIR)/netspot=$(INSTALL_BIN_DIR)/ \
			$(BIN_DIR)/netspotctl=$(INSTALL_BIN_DIR)/ \
			$(EXTRA_DIR)/netspot.toml=$(INSTALL_CONF_DIR)/ \
			$(EXTRA_DIR)/netspotctl.toml=$(CTL_INSTALL_CONF_DIR)/
	@echo $(OK)"\n"


purge:
	rm -rf $(PKG_DIR)