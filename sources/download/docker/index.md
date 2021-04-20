---
title: Docker
---

`netspot` is now available through a docker image, hosted on Github. You can have a look to the [local registry](https://github.com/users/asiffer/packages/container/package/netspot) to pull the image.

Once you have pulled the image, you can run `netspot` interactively through:
```sh
docker run -it --name netspot --cap-add NET_ADMIN --network host netspot:latest
```

The capabilities `NET_ADMIN` allows to run `netspot` through a non-root user inside the container. In addition, here we use the `host` network because in practice you may want to deploy the IDS on host interfaces.

You can tune the container a little through two environment variables:

| Environment variable  | Default value           |
| --------------------- | ----------------------- |
| `NETSPOT_ENDPOINT`    | `tcp://127.0.0.1:11000` |
| `NETSPOT_CONFIG_FILE` | `/etc/netspot.toml`     |

Hence, you can load your config file (e.g. `/tmp/config.toml`) by mounting it on the container.
```sh
docker run -it --name netspot --cap-add NET_ADMIN --network host -v /tmp/config.toml:/etc/netspot.toml netspot:latest
```