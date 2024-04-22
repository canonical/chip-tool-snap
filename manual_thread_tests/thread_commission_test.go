package tests

import (
	"os"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
)

const (
	otbrSnap       = "openthread-border-router"
	allClusterSnap = "matter-all-clusters-app"
)

func TestAllClustersAppThread(t *testing.T) {
	t.Cleanup(func() {
		utils.SnapRemove(t, otbrSnap)
		cleanupRemoteDevice(t)
		remoteSSHClient.Close()
	})

	// Start clean
	utils.SnapRemove(t, otbrSnap)

	// Local device setup
	localDeviceOTBRSetup(t)
	trimmedActiveDataset := getActiveDataset(t)

	// Remote device setup
	RemoteDeviceSetup(t)

	t.Run("Commission", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "sudo chip-tool pairing code-thread 110 hex:"+trimmedActiveDataset+" 34970112332 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-pairing.log", []byte(stdout), 0644),
		)
	})

	t.Run("Control", func(t *testing.T) {
		start := time.Now()
		stdout, _, _ := utils.Exec(t, "sudo chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-onoff.log", []byte(stdout), 0644),
		)

		waitForLogMessageOnRemoteDevice(t, allClusterSnap, "CHIP:ZCL: Toggle ep1 on/off", start)
	})

}

func cleanupRemoteDevice(t *testing.T) {
	t.Helper()

	commands := []string{
		"sudo snap remove --purge " + allClusterSnap,
		"sudo snap remove --purge " + otbrSnap,
	}

	executeRemoteCommands(t, commands)
}
