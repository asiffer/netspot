MODULE 	:= github.com/asiffer/netspot
VERSION := 3.0.0
SRCS 	:= $(shell find ./app ./collector ./analyzer ./register -type f -name '*.go')

GOARCH 	?= $(shell go env GOARCH)
BIN_DIR ?= $(CURDIR)/bin
GO      ?= $(shell command -v go)
# zig cross compiler for static build, we use zig to build libpcap and link it statically to netspot
CC 		 := $(CURDIR)/zig/zig cc
LD_FLAGS ?= -linkmode external -extldflags "-static" -s -w

# libcap build settings
LIBPCAP_VERSION 	:= 1.10.6
LIBPCAP_DIR 		:= $(CURDIR)/libpcap-$(LIBPCAP_VERSION)
LIBPCAP_CONFIGURE 	:= --without-libnl --enable-ipv6 --disable-shared --disable-usb --disable-netmap --disable-dbus --disable-bluetooth --disable-rdma
XDP_DIR 			:= collector/xdp
XDP_GENERATED 		:= $(XDP_DIR)/xdp_bpf.go $(XDP_DIR)/xdp_bpf.o

.DEFAULT_GOAL := $(BIN_DIR)/netspot-$(VERSION)-$(GOARCH)-linux

.PHONY: clean full-clean
.PRECIOUS: go.sum go.mod $(XDP_GENERATED) $(BIN_DIR)/netspot-$(VERSION)-%-linux

go.sum: go.mod
	$(GO) mod tidy

go.mod:
	$(GO) mod init $(MODULE)

$(XDP_GENERATED) &: $(XDP_DIR)/hook.c
	$(GO) tool github.com/cilium/ebpf/cmd/bpf2go -go-package xdp -target bpf -output-dir $(XDP_DIR) XDP $<

# cross compiler use to statically build both libpcap and netspot
zig/zig:
	curl -sL "https://ziglang.org/builds/zig-x86_64-linux-0.17.0-dev.296+a85a29ae4.tar.xz" | tar -xJC .
	mv ./zig-* zig

# libpcap sources
$(LIBPCAP_DIR):
	curl -sL "https://www.tcpdump.org/release/libpcap-$(LIBPCAP_VERSION).tar.xz" | tar -xJC .

# libpcap build (download libpcap only if missing)
$(LIBPCAP_DIR)/%/lib/libpcap.a: zig/zig | $(LIBPCAP_DIR)
	make -C $(LIBPCAP_DIR) distclean || true
	cd $(LIBPCAP_DIR) && ./configure CC='$(CC) --target=$*' --prefix=$(LIBPCAP_DIR)/$* --host=$* $(LIBPCAP_CONFIGURE)
	make -C $(LIBPCAP_DIR) -j$(nproc)
	make -C $(LIBPCAP_DIR) install

# amd64
$(BIN_DIR)/netspot-$(VERSION)-amd64-linux: main.go $(XDP_GENERATED) $(SRCS) go.sum $(LIBPCAP_DIR)/x86_64-linux-musl/lib/libpcap.a
	CGO_ENABLED=1 GOEXPERIMENT=jsonv2 GOOS=linux GOARCH=amd64 \
		CC="$(CC) --target=x86_64-linux-musl" \
		CGO_CFLAGS="-I$(LIBPCAP_DIR)/x86_64-linux-musl/include" \
		CGO_LDFLAGS="-L$(LIBPCAP_DIR)/x86_64-linux-musl/lib" \
		$(GO) build -o $@ -trimpath -ldflags='$(LD_FLAGS)' $<

# arm64
$(BIN_DIR)/netspot-$(VERSION)-arm64-linux: main.go $(XDP_GENERATED) $(SRCS) go.sum $(LIBPCAP_DIR)/aarch64-linux-musl/lib/libpcap.a
	CGO_ENABLED=1 GOEXPERIMENT=jsonv2 GOOS=linux GOARCH=arm64 \
		CC="$(CC) --target=aarch64-linux-musl" \
		CGO_CFLAGS="-I$(LIBPCAP_DIR)/aarch64-linux-musl/include" \
		CGO_LDFLAGS="-L$(LIBPCAP_DIR)/aarch64-linux-musl/lib" \
		$(GO) build -o $@ -trimpath -ldflags='$(LD_FLAGS)' $<

# armv7
$(BIN_DIR)/netspot-$(VERSION)-armv7-linux: main.go $(XDP_GENERATED) $(SRCS) go.sum $(LIBPCAP_DIR)/arm-linux-musleabihf/lib/libpcap.a
	CGO_ENABLED=1 GOEXPERIMENT=jsonv2 GOOS=linux GOARCH=arm GOARM=7 \
		CC="$(CC) --target=arm-linux-musleabihf" \
		CGO_CFLAGS="-I$(LIBPCAP_DIR)/arm-linux-musleabihf/include" \
		CGO_LDFLAGS="-L$(LIBPCAP_DIR)/arm-linux-musleabihf/lib" \
		$(GO) build -o $@ -trimpath -ldflags='$(LD_FLAGS)' $<

build-all: $(BIN_DIR)/netspot-$(VERSION)-amd64-linux $(BIN_DIR)/netspot-$(VERSION)-arm64-linux $(BIN_DIR)/netspot-$(VERSION)-armv7-linux

clean:
	rm -f $(BIN_DIR)/netspot*
	rm -f $(XDP_GENERATED)

full-clean: clean
	rm -rf zig $(LIBPCAP_DIR)

