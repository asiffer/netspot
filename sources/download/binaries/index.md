---
title: Binaries
---

`netspot` is available on Linux platforms with different architectures: `amd64`, `arm` and `aarch64`. The `Go` ecosystem would allow to easily build `netspot` for Mac and Windows however
it has not been tested, and that is probably not very relevant.
You can download the latest binaries with the links below.

[netspot-{{ netspot.version }}-amd64-linux-static](https://github.com/asiffer/netspot/releases/download/v{{ netspot.version }}/netspot-{{ netspot.version }}-amd64-linux-static){ .md-button .md-button--primary .md-monospace }

[netspot-{{ netspot.version }}-arm-linux-static](https://github.com/asiffer/netspot/releases/download/v{{ netspot.version }}/netspot-{{ netspot.version }}-arm-linux-static){ .md-button .md-monospace }

[netspot-{{ netspot.version }}-arm64-linux-static](https://github.com/asiffer/netspot/releases/download/v{{ netspot.version }}/netspot-{{ netspot.version }}-arm64-linux-static){ .md-button .md-monospace }

If you don't like to click, you can draw your inspiration from the following command.

```bash
curl -Lo netspot 'https://github.com/asiffer/netspot/releases/download/v{{ netspot.version }}/netspot-{{ netspot.version }}-amd64-linux-static'
chmod +x netspot
```

Let us recall that you need some privilege to run `netspot` on network interfaces. So either
you can run it as root (not recommended) or you can add the sniffing capability.

```bash
sudo setcap cap_net_admin,cap_net_raw=eip ./netspot
```