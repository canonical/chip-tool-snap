# Chip Tool Snap
[![chip-tool](https://snapcraft.io/chip-tool/badge.svg)](https://snapcraft.io/chip-tool)

Chip Tool is a Matter controller being developed as part of the [Connected Home IP project](https://github.com/project-chip/connectedhomeip.git).

The snap packaging makes it easy to run the Chip Tool on Linux distributions.

This snap has been tested on amd64/arm64 architectures for WiFi/Ethernet/DNS-SD/BLE/Thread commissioning and control.

## Usage

### Setup

```bash
sudo snap install chip-tool
```

When installing from the Snap Store, the following interfaces auto-connect:
- [`avahi-observe`](https://snapcraft.io/docs/avahi-observe-interface) to allow DNS-SD based discovery.
- [`bluez`](https://snapcraft.io/docs/bluez-interface) for device discovery over Bluetooth Low Energy (BLE).

> **Note**  
> DNS-SD and Bluetooth depend on Avahi and Bluez.
> To install:
> - Ubuntu: `sudo apt install bluez avahi-daemon`
> - Ubuntu Core: `sudo snap install avahi bluez`


You may connect the [`process-control`](https://snapcraft.io/docs/process-control-interface) interface to allow system-wide process management.
This is needed to grant Chip Tool access to make [sched_setattr](https://man7.org/linux/man-pages/man2/sched_setattr.2.html) system calls. This may improve the reliability of the commissioning and control operations (see [#8](https://github.com/canonical/chip-tool-snap/issues/8)).
```bash
sudo snap connect chip-tool:process-control
```

### Commissioning into IP network
Discover using DNS-SD and pair:
```bash
chip-tool pairing onnetwork 110 20202021
```

where:

-   `110` is the node id being assigned to the app
-   `20202021` is the pin code set on the app

### Commissioning into Thread network over BLE
This depends on an OpenThread Border Router (OTBR) with an active network.
For guidance on setting that up using the [OTBR snap](https://snapcraft.io/openthread-border-router), refer to [this tutorial](https://github.com/canonical/openthread-border-router-snap/wiki/Commission-and-control-a-Matter-Thread-device-via-the-OTBR-Snap).

Use the [OpenThread CLI](https://openthread.io/reference/cli) to obtain Thread network credential:
```bash
$ sudo ot-ctl dataset active -x
0e08...f7f8
Done
```

Discover and pair:
```bash
chip-tool pairing ble-thread 110 hex:0e08...f7f8 20202021 3840
```

where:

-   `110` is the node id being assigned to the app
-   `0e08...f7f8` is the Thread network credential operational dataset, truncated for readability.
-   `20202021` is the pin code set on the app
-   `3840` is the discriminator id


### Control
Toggle:
```bash
chip-tool onoff toggle 110 1
```

where:

-   `onoff` is the matter cluster name
-   `on`/`off`/`toggle` is the command name.
-   `110` is the node id of the app assigned during the commissioning
-   `1` is the endpoint of the configured device


### Note on sudo

The latest version of the chip-tool snap does not require the use of sudo (root access). If you have updated the snap from a previous version it will still work with sudo. If you run it as a normal user, the previous state of provisioned devices will not be available.

To change from running with sudo to running without sudo, you need to copy the database files from the root user to your user, and update the file ownerships. This can be done with these two commands:

```
sudo cp /var/snap/chip-tool/common/mnt/chip_tool_* ~/snap/chip-tool/common/
sudo chown $USER:$USER ~/snap/chip-tool/common/*
```

If you run chip-tool again without sudo and get an error similar to `CHIP Error 0x000000AF: Write to file failed`, either restart your computer to clear all temporary files, or run the following commands to delete them:

```
# Open a shell inside the chip-tool snap sandbox
sudo snap run --shell chip-tool.chip-tool
# Inside this shell, delete the temporary files
rm /tmp/chip_*
```


## Build

Build locally for the architecture same as the host:
```bash
snapcraft -v
```

Build remotely for all supported architectures:
```
snapcraft remote-build
```

### Install the built snap

Install the local snap:
```bash
sudo snap install --dangerous *.snap
```

Connect the following interfaces:
```bash
sudo snap connect chip-tool:avahi-observe
sudo snap connect chip-tool:bluez
```

> **Note**  
> On **Ubuntu Core**, the `avahi-observe` and `bluez` interfaces are not provided by the system.
> These interfaces are provided by other snaps, such as the [Avahi](https://snapcraft.io/avahi) and [BlueZ](https://snapcraft.io/bluez) snaps.
> To install the snaps and connect the interfaces, run:
> ```bash
> sudo snap install avahi bluez
> sudo snap connect chip-tool:avahi-observe avahi:avahi-observe
> sudo snap connect chip-tool:bluez bluez:service
> ```


Continue the [setup](#setup).

## Test

Refer to [tests](./tests).
