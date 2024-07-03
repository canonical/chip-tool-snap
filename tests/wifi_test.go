package tests

import (
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
)

func TestAllClustersAppWiFi(t *testing.T) {
	InstallChipTool(t)

	start := time.Now()

	// Start clean
	utils.SnapRemove(t, allClustersSnap)

	t.Cleanup(func() {
		utils.SnapRemove(t, allClustersSnap)
		utils.SnapDumpLogs(t, start, allClustersSnap)
	})

	// Install all clusters app
	utils.SnapInstallFromStore(t, allClustersSnap, "latest/edge")

	// Setup all clusters app
	utils.SnapSet(t, allClustersSnap, "args", "--wifi")
	utils.SnapConnect(t, allClustersSnap+":avahi-control", "")
	utils.SnapConnect(t, allClustersSnap+":bluez", "")

	// Start all clusters app
	utils.SnapStart(t, allClustersSnap)
	utils.WaitForLogMessage(t,
		allClustersSnap, "CHIP minimal mDNS started advertising", start)

	t.Run("Commission", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "sudo chip-tool pairing onnetwork 110 20202021 2>&1")
		writeLogFile(t, "chip-tool-pairing", []byte(stdout))
	})

	t.Run("Control", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "sudo chip-tool onoff toggle 110 1 2>&1")
		writeLogFile(t, "chip-tool-toggle", []byte(stdout))

		// 0x6 is the cluster ID for on-off
		// Using cluster ID here because of a buffering issue in the log stream:
		// https://github.com/canonical/chip-tool-snap/pull/69#issuecomment-2207189962
		utils.WaitForLogMessage(t, allClustersSnap, "ClusterId = 0x6", start)
	})

}
