# Chip Tool Snap
[![chip-tool](https://snapcraft.io/chip-tool/badge.svg)](https://snapcraft.io/chip-tool)

Chip Tool is a Matter controller being developed as part of the [Connected Home IP project](https://github.com/project-chip/connectedhomeip.git).

The snap packaging makes it easy to run the Chip Tool on Linux distributions.

This snap has been tested on amd64/arm64 architectures for WiFi/Ethernet/DNS-SD/BLE/Thread commissioning and control.

Usage instructions can be found in the [documentation](https://canonical-matter.readthedocs-hosted.com/en/latest/how-to/chip-tool-commission-and-control/).

## Development

### Build the snap

Build locally for the same architecture as the host:
```bash
snapcraft -v
```

Build remotely for all supported architectures:
```bash
snapcraft remote-build
```

### Install the built snap

Install the local snap:
```bash
sudo snap install --dangerous *.snap
```

## Notes

### Process Control permission

You may connect the [`process-control`](https://snapcraft.io/docs/process-control-interface) interface to allow system-wide process management.
This is needed to grant Chip Tool access to make [sched_setattr](https://man7.org/linux/man-pages/man2/sched_setattr.2.html) system calls.
This may improve the reliability of the commissioning and control operations (see [#8](https://github.com/canonical/chip-tool-snap/issues/8)).

```bash
sudo snap connect chip-tool:process-control
```

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

## Test

Refer to [tests](./tests).
