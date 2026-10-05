---
layout: '@layouts/BlogPost.astro'
title: 'Installing Debian Packages Offline'
draft: false
pubDate: 2026-10-05
description: 'A guide on how to install a Debian package including all of its dependencies on a machine without internet access.'
author: 'Clowa'
tags: ["debian", "ubuntu", "apt", "tutorial"]
---

Sooner or later you'll run into a box without internet access, like an air-gapped server or a machine in an isolated lab network. `apt install` won't help you there, so you have to bring the package to the box yourself, **including all of its dependencies**. I'll use `openvswitch-switch` as an example, but the steps work for any package.

> [!TIP]
> Do you have to keep multiple offline machines installed or up to date? Have a look at [`apt-offline`](https://github.com/rickysarraf/apt-offline). It records what each offline machine is missing, so you only download exactly that.

## Prerequisites

- An **online** machine running the **same distribution, release and CPU architecture** as the offline box _(e.g. both Ubuntu 24.04 on `amd64`)_ - You could use a temporary Docker container.
- A USB drive or a local network connection to the offline box

> [!WARNING]
> Packages downloaded on Ubuntu 24.04 won't reliably install on Ubuntu 22.04 or Debian 12, and `arm64` packages won't install on an `amd64` machine. Compare both systems with `lsb_release -a` and `dpkg --print-architecture`.

## Step 1: Download the package and its dependencies

> [!TIP]
> No matching machine at hand? Use a Docker container with the same release and architecture instead:
>
> ```bash
> docker run --rm -it --platform linux/amd64 -v "$PWD:/debs" -w /debs ubuntu:24.04
> ```
>
> Inside the container, run the commands below without `sudo` and `cd $(mktemp -d)`. The packages end up in your current directory on the host.

On the **online** machine, refresh your package lists and change into a fresh temporary directory:

```bash
sudo apt-get update
cd $(mktemp -d)
```

Download the package together with its complete dependency tree:

```bash
export PACKAGE_NAME="openvswitch-switch"
apt-get download $(apt-cache depends --recurse --no-recommends --no-suggests --no-conflicts --no-breaks --no-replaces --no-enhances "$PACKAGE_NAME" | grep "^\w")
```

This way you receive the package itself and all necessary recursive dependencies of the package - Check what you've got:

```bash
ls -l *.deb
```

For `openvswitch-switch` that's 75 packages with roughly 25 MB. Many of them, like `dpkg` or `libc6`, are probably already installed on the offline box. It's hard to tell from the online machine what's missing on the other side.

## Step 2: Move the packages to the offline box

Copy all `.deb` files to the offline box via USB drive or `scp`. If you use `scp`, replace `/media/usb` in the next step with the directory you copied the files to.

```bash
ssh user@offline-box mkdir -p /tmp/debs
scp *.deb user@offline-box:/tmp/debs/
```

If you use a USB drive, plug it into the offline box and identify it with `lsblk` - usually by its size. In the following, I'll assume the partition holding your files is `/dev/sdb1`. Create a mount point and mount the drive:

```bash
sudo mkdir /media/usb
sudo mount /dev/sdb1 /media/usb
```

## Step 3: Install the packages

```bash
sudo apt install /media/usb/*.deb
```

`apt` installs local `.deb` files without network access and figures out the correct installation order by itself.

Verify the installation:

```bash
dpkg -s openvswitch-switch | grep Status
```

The output should be `Status: install ok installed`.

Finally, if you have used `scp` cleanup the temporary directory.

```bash
ssh user@offline-box rm -r /tmp/debs
```

... or unmount the USB drive and remove the mount point:

```bash
sudo umount /media/usb
sudo rmdir /media/usb
```
