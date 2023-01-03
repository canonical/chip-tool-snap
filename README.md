# Chip Tool Snap

Chip Tool is a Matter controller being developed as part of the [Connected Home IP project](https://github.com/project-chip/connectedhomeip.git).

The snap packaging makes it easy to run the Chip Tool on Linux distributions.

This snap has been tested on amd64 and arm64 architecture for WiFi/Ethernet/DNS-SD commissioning and control. BLE and Thread has not been tested.

## Usage

### Setup

```bash
snap install chip-tool --edge
```

Connect the [`avahi-observe`](https://snapcraft.io/docs/avahi-observe-interface) interface to allow DNS-SD based discovery:
```bash
snap connect chip-tool:avahi-observe
```

### Commissioning
Discover and pair:
```bash
sudo chip-tool pairing onnetwork 110 20202021
```

Or, pair directly by giving the IP address:
```bash
sudo chip-tool pairing ethernet 110 20202021 3840 192.168.1.110 5543
```

where:

-   `110` is the assigned node id
-   `20202021` is the pin code for the bridge app
-   `3840` is the discriminator id
-   `192.168.1.111` is the IP address of the host for the bridge
-   `5540` the the port for the bridge


### Control
Toggle:
```bash
sudo chip-tool onoff toggle 110 1
```

where:

-   `onoff` is the matter cluster name
-   `on`/`off`/`toggle` is the command name. The `toggle` command is RECOMMENDED
    because it is stateless. The bridge does not synchronize the actual state of
    devices.
-   `110` is the node id of the bridge app assigned during the commissioning
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
