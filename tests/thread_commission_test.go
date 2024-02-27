package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
)

const (
	otbrSnap = "openthread-border-router"

	defaultInfraInterfaceValue = "wlan0"
	infraInterfaceKey          = "infra-if"
	localInfraInterfaceEnv     = "LOCAL_INFRA_IF"
)

func TestAllClustersAppThread(t *testing.T) {
	// Start clean
	utils.SnapRemove(t, otbrSnap)

	t.Cleanup(func() {
		cleanupRemoteDevice(t)
	})

	// Remote device: Access remote device, setup Thread network, and get active dataset
	activeDataset := setupRemoteThreadNetwork(t)
	trimmedActiveDataset := strings.TrimSpace(activeDataset)

	// Local device: Start application
	startAllClustersApp(t, "--thread")

	waitForLogMessage(t,
		allClustersAppLog, "CHIP minimal mDNS started advertising")

	// Local device: Setup and start OTBR
	utils.SnapInstallFromStore(t, otbrSnap, utils.ServiceChannel)
	connectOTBRInterfaces()

	// Local device: Get and set infrastructure interface
	if v := os.Getenv(localInfraInterfaceEnv); v != "" {
		infraInterfaceValue := v
		utils.SnapSet(nil, otbrSnap, infraInterfaceKey, infraInterfaceValue)
	} else {
		utils.SnapSet(nil, otbrSnap, infraInterfaceKey, defaultInfraInterfaceValue)
	}

	utils.SnapStart(t, otbrSnap)

	t.Run("Commission", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "sudo chip-tool pairing ble-thread 110 hex:"+trimmedActiveDataset+" 20202021 3840 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-pairing.log", []byte(stdout), 0644),
		)
	})

	t.Run("Control", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "sudo chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-onoff.log", []byte(stdout), 0644),
		)

		waitForLogMessage(t,
			allClustersAppLog, "CHIP:ZCL: Toggle ep1 on/off")
	})

}

func connectOTBRInterfaces() {
	utils.SnapConnect(nil, otbrSnap+":avahi-control", "")
	utils.SnapConnect(nil, otbrSnap+":firewall-control", "")
	utils.SnapConnect(nil, otbrSnap+":raw-usb", "")
	utils.SnapConnect(nil, otbrSnap+":network-control", "")
	utils.SnapConnect(nil, otbrSnap+":bluetooth-control", "")
	utils.SnapConnect(nil, otbrSnap+":bluez", "")
}
