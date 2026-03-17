MODULE := github.com/asiffer/netspot
SRCS := $(shell find ./app ./collector ./analyzer ./register -type f -name '*.go')

.DEFAULT_GOAL := netspot
.PHONY: clean

go.sum: go.mod 

go.mod:
	$(GO) mod init $(MODULE)
	$(GO) mod tidy

collector/xdp/xdp_bpf.go: collector/xdp/hook.c
	go run github.com/cilium/ebpf/cmd/bpf2go -go-package xdp -target bpf -output-dir collector/xdp XDP $< -- -D__x86_64__ -Icollector/xdp

netspot: go.sum collector/xdp/xdp_bpf.go $(SRCS)
	CGO_ENABLED=1 GOEXPERIMENT=jsonv2 go build -o $@ ./app/*.go 

clean:
	rm -f netspot
	rm -f collector/xdp/xdp_bpf.go collector/xdp/xdp_bpf.o
