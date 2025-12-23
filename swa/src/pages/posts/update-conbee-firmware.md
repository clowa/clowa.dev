---
layout: '@layouts/BlogPost.astro'
title: 'Update the firmware of your ConBee'
draft: false
pubDate: 2025-12-23
description: 'A guide on how to update the firmware of your ConBee USB stick for Zigbee connectivity.'
author: 'Clowa'
tags: ["home assistant", "conBee", "zigbee", "matter"]
---

Updating the firmware of your ConBee USB stick is essential to ensure compatibility with the latest Zigbee devices and to benefit from new features and bug fixes. Newer firmware versions can also unlock exciting new features like [Matter support](https://phoscon.de/en/openthread/doc), which opens up your smart home to an even broader ecosystem of devices.

**This guide will not walk you though the whole process of the firmware-update itself**, there are offical docs for that in the [deCONZ documentation](https://github.com/dresden-elektronik/deconz-rest-plugin/wiki/Update-deCONZ-manually) or if you want to install the OpenThread Router you can find instructions on the [Phoscon website](https://phoscon.de/en/openthread/doc). It's more like a short summary of the steps you need to take and some tips and tricks I found useful while updating my ConBee firmware.

>[!NOTE]
> I'm using the ConBee via the [deCONZ integration in Home Assistant](https://www.home-assistant.io/integrations/deconz/), so the steps might differ slightly if you are using it with other software like Zigbee2MQTT, ZHA or others. But most of the steps should be similar expect maybe for the backup part.

## The right firmware

Dresden Elektronik publishes different firmware versions for the different Hardware like ConBee II, ConBee III and RaspBee. Additionally, there are different firmware files depending on if you are using Zigbee or Thread / Matter. Make sure to download the correct firmware for your use case.

Firmware files for Zigbee can be found on the [Dresden Elektronik website](https://deconz.dresden-elektronik.de/deconz-firmware)
Firmware files for Thread / Matter can be found on the [Phoscon website](https://phoscon.de/downloads/openthread/firmware/).

## Verifing the download

After downloading the firmware file, it's a good idea to verify its integrity. Dresden Elektronik provides MD5 checksums together with their firmware files you can easily use to check if the file was downloaded correctly and is not corrupted.

On Windows you can use Powershell Desktop or Powershell Core

```powershell
$file = "./ot-rcp-cb2_0x01010700.GCF"
$checksumFile = "$file.md5"
(Get-FileHash -Algorithm MD5 -Path $file).Hash -eq ((Get-Content -Path $checksumFile) -split " ")[0]
```

On Linux use md5sum

```bash
FILE="ot-rcp-cb2_0x01010700.GCF"
CHECKSUM_FILE="$FILE.md5"
EXPECTED_CHECKSUM=$(cut -d ' ' -f1 "$CHECKSUM_FILE")
CALCULATED_CHECKSUM=$(md5sum "$FILE" | cut -d ' ' -f1)
[ "$EXPECTED_CHECKSUM" == "xxx" ] && echo "True" || echo "False"
```

In both cases, the output should be `True` if the file is valid or `False` if something is wrong with the download.

## Backup your current configuration

Before you start the firmware update process, it's a good idea to back up your current configuration. This way, if anything goes wrong during the update, you can restore your previous setup without having to re-pair all your devices. The [phoscon webUI](https://github.com/dresden-elektronik/deconz-rest-plugin/wiki/Backup-&-Restore) provides an option to export your current configuration to a file, which you can later import after the firmware update.

## Flash the new firmware

Dresden Elektronik provides a tool called [`GCFFlasher`](https://github.com/dresden-elektronik/gcfflasher) to flash the new firmware onto your ConBee device. You can download the Windows version at the [Dresden Elektronik website](https://deconz.dresden-elektronik.de/win/) and the Linux version at their [GitHub releases page](https://github.com/dresden-elektronik/gcfflasher/releases).

I was thinking about creating a docker image for this tool - let me know if you would find that useful by creating a [feature request](https://github.com/clowa/clowa.dev/issues)!

Once you have `GCFFlasher` installed, you can use it to flash the new firmware. Make sure to follow the instructions provided in the [deCONZ documentation](https://github.com/dresden-elektronik/deconz-rest-plugin/wiki/Update-deCONZ-manually)
