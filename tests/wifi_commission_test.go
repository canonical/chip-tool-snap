package tests

import (
	"os"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
)

const allClusterSnap = "matter-all-clusters-app"

func TestAllClustersAppWiFi(t *testing.T) {
	start := time.Now()

	// Start clean
	utils.SnapRemove(t, allClusterSnap)

	t.Cleanup(func() {
		utils.SnapRemove(t, allClusterSnap)
		utils.SnapDumpLogs(nil, start, allClusterSnap)
	})

	// Install all clusters app
	utils.SnapInstallFromStore(t, allClusterSnap, utils.ServiceChannel)

	// Setup all clusters app
	utils.SnapSet(t, allClusterSnap, "args", "--wifi")
	utils.SnapConnect(t, allClusterSnap+":avahi-control", "")
	utils.SnapConnect(t, allClusterSnap+":bluez", "")

	// Start all clusters app
	utils.SnapStart(t, allClusterSnap)
	utils.WaitForLogMessage(t,
		allClusterSnap, "CHIP minimal mDNS started advertising", start)

	t.Run("Commission", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "sudo chip-tool pairing onnetwork 110 20202021 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-pairing.log", []byte(stdout), 0644),
		)
	})

	t.Run("Control", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "sudo chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-onoff.log", []byte(stdout), 0644),
		)

		utils.WaitForLogMessage(t,
			allClusterSnap, "CHIP:ZCL: Toggle ep1 on/off", start)
	})

}
