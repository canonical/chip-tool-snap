package tests

import (
	"os"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
)

const otbrSnap = "openthread-border-router"

func TestAllClustersAppThread(t *testing.T) {
	t.Cleanup(func() {
		utils.SnapRemove(t, otbrSnap)
		cleanupRemoteDevice(t)
	})

	// Start clean
	utils.SnapRemove(t, otbrSnap)
	cleanupRemoteDevice(t)

	start := time.Now()

	// Local device setup
	localDeviceSetup(t)
	trimmedActiveDataset := getActiveDataset(t)

	// Remote device setup
	RemoteDeviceSetup(t)

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

		// TODO: check log on remote device
		// utils.WaitForLogMessage(t,
		// 	allClusterSnap, "CHIP:ZCL: Toggle ep1 on/off", start)
	})

}

func cleanupRemoteDevice(t *testing.T) string {
	t.Helper()

	commands := []string{
		"sudo snap remove --purge matter-all-clusters-app",
		"sudo snap remove --purge openthread-border-router",
	}

	executeRemoteCommands(t, commands)

	return ""
}
