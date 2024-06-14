package tests

import (
	"os"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
)

func TestAllClustersAppWiFi(t *testing.T) {
	InstallChipTool(t)

	start := time.Now()

	// Start clean
	utils.SnapRemove(t, allClustersSnap)

	t.Cleanup(func() {
		utils.SnapRemove(t, allClustersSnap)
		utils.SnapDumpLogs(nil, start, allClustersSnap)
	})

	// Install all clusters app
	utils.SnapInstallFromStore(t, allClustersSnap, utils.ServiceChannel)

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
			allClustersSnap, "CHIP:ZCL: Toggle ep1 on/off", start)
	})

}
