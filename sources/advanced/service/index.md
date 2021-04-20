---
title: Service
weight: 10
summary: Running netspot as a service
---


Even if it is not the main way to use `netspot`, it can run as a service, exposing a minimal REST API.

```sh
netspot serve
```

By default it listens at `tcp://localhost:11000`, and you can visit `http://localhost:11000` to look at the simple dashboard that displays
the current config of `netspot`.

![dashboard](/images/dashboard.png)

Naturally, depending on the interface(s) you monitor, you would like to change the API endpoint not to pollute what `netspot` is analyzing.
You can be changed it with the `-e` flag. For instance, you can consider a unix socket.

```sh
netspot serve -e unix:///tmp/netspot.sock
```

The server exposes few methods that allows to do roughly everything. 

| Method | Path           | Description                               |
| ------ | -------------- | ----------------------------------------- |
| `GET`  | `/api/config`  | Get the current config (JSON output)      |
| `POST` | `/api/config`  | Change the config (JSON expected)         |
| `POST` | `/api/run`     | Manage the status of netspot (start/stop) |
| `GET`  | `/api/stats`   | Get the list of available statistics      |
| `GET`  | `/api/devices` | Get the list of available interfaces      |

In addition, a `Go` client is available in the `api/client` subpackage.

```sh
go get -u github.com/asiffer/netspot/api/client
```