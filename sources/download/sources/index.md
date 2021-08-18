---
title: From sources
---

You can get `netspot` by building it from sources. By default, it will dynamically link netspot
to `libpcap`. At least, you will need `libpcap-dev` on your system (and a `C` compiler).

### Go magic

Look at this magic!

```bash
go install github.com/asiffer/netspot
```

It installs `netspot` to `$GOPATH/bin/netspot`, so you must ensure it is in your path to use it directly from cli.

<!-- prettier-ignore -->
!!! warning
    This method basically run the `Go` compiler on the sources. However it differs from the `make` build as the latter use additional "release" flags. In addition, the version output by the binary won't embed the git hash (only the major version).

### Classic make

<!-- prettier-ignore -->
!!! warning
    As `netspot` uses new `Go` features, you should use a recent version of the `Go` compiler (`>=1.16`).
    You can run the following script to update your version (`curl`, `sudo` and `tar` commands are required):

    ```bash
    # Get latest version ID
    LAST_VERSION=$(curl -s https://golang.org/dl/ | grep -e 'go[0-9]*\.[0-9]*\.[0-9]*' -om 1 | sed 's/go//g')

    # prepare the dl link
    TAR="go${LAST_VERSION}.linux-amd64.tar.gz"
    LINK="https://dl.google.com/go/${TAR}"

    # save the archive to /tmp
    curl -so "/tmp/${TAR}" "${LINK}"

    # install it (you should remove the previous /usr/local/go folder)
    sudo tar -C /usr/local -xzf "/tmp/${TAR}"

    # add to path
    export PATH=$PATH:/usr/local/go/bin

    # check version
    go version
    ```

On Debian 10/Ubuntu 20.04

```bash
apt update
apt install git make gcc libpcap-dev
git clone https://github.com/asiffer/netspot.git
cd netspot
make
```

It builds the binary on the `bin/` folder. You can check the result:

```bash
./bin/netspot-{{ netspot.version }}-amd64-linux --version
```

Finally you can install it (as root)

```bash
make install
```
