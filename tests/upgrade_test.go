package tests

import (
	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

func TestUpgrade(t *testing.T) {
	start := time.Now()

	// Remove snaps and logs at end of test, even if it failed
	t.Cleanup(func() {
		utils.SnapRemove(nil, allClusterSnap)
		utils.SnapDumpLogs(nil, start, allClusterSnap)
		utils.SnapRemove(nil, chipToolSnap)
		utils.SnapDumpLogs(nil, start, chipToolSnap)
	})

	// Start clean
	utils.SnapRemove(t, allClusterSnap)
	utils.SnapRemove(t, chipToolSnap)

	// Install stable chip tool from store
	utils.SnapInstallFromStore(t, chipToolSnap, "latest/stable")

	// Setup chip-tool
	utils.SnapConnect(t, chipToolSnap+":avahi-observe", "")
	utils.SnapConnect(t, chipToolSnap+":bluez", "")
	utils.SnapConnect(t, chipToolSnap+":process-control", "")

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

	// Pair device
	t.Run("Commission", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "sudo chip-tool pairing onnetwork 110 20202021 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-pairing.log", []byte(stdout), 0644),
		)
	})

	// Control device
	t.Run("Control with stable snap", func(t *testing.T) {
		snapVersion := utils.SnapVersion(t, chipToolSnap)
		snapRevision := utils.SnapRevision(t, chipToolSnap)
		log.Printf("%s installed version %s build %s\n", chipToolSnap, snapVersion, snapRevision)

		stdout, _, _ := utils.Exec(t, "sudo chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-onoff.log", []byte(stdout), 0644),
		)

		utils.WaitForLogMessage(t,
			allClusterSnap, "CHIP:ZCL: Toggle ep1 on/off", start)
	})

	// Upgrade chip-tool to local snap or edge
	if utils.LocalServiceSnap() {
		utils.SnapInstallFromFile(t, utils.LocalServiceSnapPath)
	} else {
		utils.SnapRefresh(t, chipToolSnap, "latest/edge")
	}

	// Control device again
	t.Run("Control upgraded snap", func(t *testing.T) {
		snapVersion := utils.SnapVersion(t, chipToolSnap)
		snapRevision := utils.SnapRevision(t, chipToolSnap)
		log.Printf("%s installed version %s build %s\n", chipToolSnap, snapVersion, snapRevision)

		stdout, _, _ := utils.Exec(t, "sudo chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-onoff.log", []byte(stdout), 0644),
		)

		utils.WaitForLogMessage(t,
			allClusterSnap, "CHIP:ZCL: Toggle ep1 on/off", start)
	})

}
