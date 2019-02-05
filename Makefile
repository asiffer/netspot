
PACKAGE_NAME := netspot
VERSION := 1.2
PACKAGE_DESC := "A basic IDS with statistical learning"
SRC_DIR := $(GOPATH)/src/netspot
EXTRA_DIR = $(SRC_DIR)/extra
BIN_DIR := $(SRC_DIR)/bin
PKG_DIR := $(GOPATH)/src/netspot/dist


DEST_DIR :=
INSTALL_BIN_DIR := $(DEST_DIR)/usr/local/bin
INSTALL_CONF_DIR := $(DEST_DIR)/etc/netspot
CTL_INSTALL_CONF_DIR := $(DEST_DIR)/etc/netspotctl

ARCH := amd64
OS := linux

BUILD_DIR = $(PKG_DIR)/$(OS)/$(ARCH)

OK := "[\033[32mOK\033[0m]"
default: prebuild build package

prebuild:
	@mkdir -p $(PKG_DIR)
	@mkdir -p $(PKG_DIR)/$(OS)
	@mkdir -p $(PKG_DIR)/$(OS)/$(ARCH)

build_netspot:
	@echo "\033[34m[Building netspot]\033[0m"
	@echo -n "Building go package...               \n"
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(BIN_DIR)/netspot $(SRC_DIR)/netspot.go
	@echo $(OK)


build_netspotctl:
	@echo "\033[34m[Building netspotctl]\033[0m"
	@echo -n "Building go package...               "
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(BIN_DIR)/netspotctl $(SRC_DIR)/netspotctl/netspotctl.go 
	@echo $(OK)


build: build_netspot build_netspotctl



package:
	@echo "\033[34m[Packaging]\033[0m"
	@echo "Running fpm...                       "
	@fpm -f -s dir \
			-t deb \
			-n $(PACKAGE_NAME) \
			-v $(VERSION) \
			-a $(ARCH) \
			-m "alban.siffer@irisa.fr" \
			-p $(BUILD_DIR)/ \
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