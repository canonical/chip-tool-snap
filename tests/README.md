# Run Tests

```bash
go test -v -failfast -count 1
```

where:

- `-v` is to enable verbose output
- `-failfast` makes the test stop after first failure
- `-count 1` is to avoid Go test caching for example when testing a rebuilt snap

### Generic Environment Variables

Environment variables can modify the test functionality. Refer to these in
[the documentation](https://pkg.go.dev/github.com/canonical/matter-snap-testing/env)
of the `matter-snap-testing` Go package.

## Run Thread tests

For running Thread tests, two Radio Co-Processors (RCPs) are needed for both local and remote machines.

For building and flashing RCP firmware, please refer
to [Build and flash RCP firmware on nRF52480 dongle](https://github.com/canonical/openthread-border-router-snap/wiki/Setup-OpenThread-Border-Router-with-nRF52840-Dongle#build-and-flash-rcp-firmware-on-nrf52480-dongle).

```bash
LOCAL_INFRA_IF="eno1" \
REMOTE_INFRA_IF="eth0" \
REMOTE_USER="ubuntu" \
REMOTE_PASSWORD="abcdef" \
REMOTE_HOST="192.168.178.95" \
go test -v -failfast -count 1 ./thread_tests
```

### Thread specific environment variables

 Variable name    | Required | Default value                   | Description                       
------------------|----------|---------------------------------|-----------------------------------
 LOCAL_INFRA_IF   | no       | wlan0                           | wlan0                             | Local backhaul network interface  
 LOCAL_RADIO_URL  | no       | spinel+hdlc+uart:///dev/ttyACM0 | Local RCP serial port             
 REMOTE_HOST      | yes      |                                 | Remote device IP or hostname      
 REMOTE_USER      | yes      |                                 | Remote device SSH username        
 REMOTE_PASSWORD  | yes      |                                 | Remote device SSH password        
 REMOTE_INFRA_IF  | no       | wlan0                           | Remote backhaul network interface 
 REMOTE_RADIO_URL | no       | spinel+hdlc+uart:///dev/ttyACM0 | Remote RCP serial port            
