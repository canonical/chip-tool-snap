# Run Tests

```bash
go test -v -failfast -count 1
```

where:
- `-v` is to enable verbose output
- `-failfast` makes the test stop after first failure
- `-count 1` is to avoid Go test caching for example when testing a rebuilt snap

# Run local Thread commission test
For running manual local Thread commissioning test, two Radio Co-Processors (RCPs) are needed for both local and remote machines. 
For building and flashing RCP firmware, please refer to [Build and flash RCP firmware on nRF52480 dongle](https://github.com/canonical/openthread-border-router-snap/wiki/Setup-OpenThread-Border-Router-with-nRF52840-Dongle#build-and-flash-rcp-firmware-on-nrf52480-dongle)

```bash
LOCAL_INFRA_IF="eno1" REMOTE_INFRA_IF="eth0" REMOTE_USER="ubuntu" REMOTE_PASSWORD="abcdef" REMOTE_IP="192.168.178.95" go test -v -failfast -count 1
```