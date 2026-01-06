---
layout: '@layouts/BlogPost.astro'
title: 'Accessing Home Assistant via SSH'
description: 'A guide on how to access the node of your Home Assistant system using SSH.'
author: Clowa
draft: false
pubDate: 2025-12-17
tags: ["home assistant", "SSH", "tutorial"]
---

If you're looking to access the Host of your Home Assistant via SSH for development or troubleshooting purposes, this guide will walk you through the necessary steps.

The home assistant team itself published a guide on how to do this in their [developer documentation](https://developers.home-assistant.io/docs/operating-system/debugging/) and only recommend this for development and debugging purposes **not for regular use**.

> [!WARNING]
> Please note that accessing the host system is not intended for regular use. This method will grant you access to almost everything on your Home Assistant host system including the docker containers of Home Assistant and it's add-ons. Be very careful with what you do as you can easily break your Home Assistant installation to an unrecoverable state.

## Prerequisites

- Home Assistant installation
- USB drive _(for storing your ssh key)_
- A computer with an ssh client installed

## Step 1: Generate SSH Keys

First, you'll need to generate an SSH key pair on your local machine if you don't already have one. You can do this using the following command:

```bash
ssh-keygen -t rsa -b 4096
```

This command will create a public and private key pair. By default, the keys will be named `id_rsa` _(private key)_and `id_rsa.pub` _(public key)_ and stored in the `~/.ssh/` directory, but you will be prompted to specify a different location if desired.

## Step 2: Prepare the USB Drive

Next, format a USB drive as `FAT32`, `ext4` or `NTFS` and name this partition `CONFIG`. At the root of the partition, create a text file named `authorized_keys` (without any file extension) on the USB drive and copy the contents of your public key (the contents of `id_rsa.pub`) into this file. If you have multiple keys, you can add them on separate lines.

> [!NOTE]
> Ensure the `authorized_keys` file is saved with ASCII encoding and Unix-style line endings (LF) to avoid any issues. Windows uses CRLF by default, which will not work.

## Step 3: Insert the USB into Home Assistant

Now, safely eject the USB drive from your local machine and insert it into the Home Assistant host system. The Home Assistant OS will automatically detect the `authorized_keys` file on the USB drive during reboot and configure SSH access accordingly.

In case your Home Assistant host is already running, you must reboot it to load the keys.

## Step 4: Connect via SSH

Once the Home Assistant host has rebooted, you can connect to it via SSH using the following command:

```bash
ssh root@homeassistant.local -p 22222 -i /path/to/your/private/key/id_rsa
```

Replace `/path/to/your/private/key/id_rsa` with the actual path to your private key file. If you are using a different hostname or IP address for your Home Assistant host, replace `homeassistant.local` with that address or hostname.

You should now be connected to the Home Assistant host system via SSH. 🎉

## Step 5: Make things comfortable

Optionally, you can enhance your SSH experience by creating or updating the SSH configuration file at `~/.ssh/config` on your local machine. Add the following configuration:

```plaintext
Host homeassistant
    HostName homeassistant.local
    Port 22222
    User root
    IdentityFile /path/to/your/private/key/id_rsa
```

>[!Tip]
> If you use a password manager with SSH key support (such as 1Password or Bitwarden), or if your SSH key is loaded in an SSH agent via some other method, you can omit the `IdentityFile` line. For detailed steps, consult your password manager’s documentation.

With this configuration, you can simply connect to your Home Assistant host by running:

```bash
ssh homeassistant
```

## Conclusion

You have successfully set up SSH access to your Home Assistant host system. Remember to use this access responsibly and avoid making changes that could compromise the stability or security of your Home Assistant installation.

Happy home automating! 🚀
