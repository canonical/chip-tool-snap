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

Connect the [`avahi-observe`](https://snapcraft.io/docs/avahi-observe-interface) interface to allow DNS-SD based discovery:
```bash
sudo snap connect chip-tool:avahi-observe
```

Connect the [`bluez`](https://snapcraft.io/docs/bluez-interface) interface for device discovery over Bluetooth Low Energy (BLE):
```bash
sudo snap connect chip-tool:bluez
```

Connect the [`process-control`](https://snapcraft.io/docs/process-control-interface) interface for system-wide process management, such as sched_setattr system call:
```bash
sudo snap connect chip-tool:process-control
```

### Commissioning into IP network
Discover using DNS-SD and pair:
```bash
sudo chip-tool pairing onnetwork 110 20202021
```

where:

-   `110` is the node id being assigned to the app
-   `20202021` is the pin code set on the app

### Commissioning into Thread network over BLE
Obtain Thread network credential:
```bash
$ sudo ot-ctl dataset active -x
0e08...f7f8
Done
```
Discover and pair:
```bash
sudo chip-tool pairing ble-thread 110 hex:0e08...f7f8 20202021 3840
```

where:

-   `110` is the node id being assigned to the app
-   `0e08...f7f8` is the Thread network credential operational dataset, truncated for readability.
-   `20202021` is the pin code set on the app
-   `3840` is the discriminator id


### Control
Toggle:
```bash
sudo chip-tool onoff toggle 110 1
```

where:

-   `onoff` is the matter cluster name
-   `on`/`off`/`toggle` is the command name.
-   `110` is the node id of the app assigned during the commissioning
-   `1` is the endpoint of the configured device


## Build

Build locally for the architecture same as the host:
```bash
snapcraft -v
```

Build remotely for all supported architectures:
```
snapcraft remote-build
```
