package tests

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
)

func TestUpgrade(t *testing.T) {
	start := time.Now()

	// Remove snaps and logs at end of test, even if it failed
	t.Cleanup(func() {
		utils.SnapRemove(nil, allClustersSnap)
		utils.SnapDumpLogs(nil, start, allClustersSnap)
		utils.SnapRemove(nil, chipToolSnap)
		utils.SnapDumpLogs(nil, start, chipToolSnap)
	})

	// Start clean
	utils.SnapRemove(t, allClustersSnap)
	utils.SnapRemove(t, chipToolSnap)

	// Install stable chip tool from store
	utils.SnapInstallFromStore(t, chipToolSnap, "latest/stable")

	// Setup chip-tool
	utils.SnapConnect(t, chipToolSnap+":avahi-observe", "")
	utils.SnapConnect(t, chipToolSnap+":bluez", "")
	utils.SnapConnect(t, chipToolSnap+":process-control", "")

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

	// Pair device
	t.Run("Commission", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "sudo chip-tool pairing onnetwork 110 20202021 2>&1")
		assert.NoError(t,
			os.WriteFile(t.Name()+"-chip-tool-pairing.log", []byte(stdout), 0644),
		)
	})

	// Control device
	t.Run("Control with stable snap", func(t *testing.T) {
		snapVersion := utils.SnapVersion(t, chipToolSnap)
		snapRevision := utils.SnapRevision(t, chipToolSnap)
		log.Printf("%s installed version %s build %s\n", chipToolSnap, snapVersion, snapRevision)

		start := time.Now()
		stdout, _, _ := utils.Exec(t, "sudo chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t,
			os.WriteFile(t.Name()+"-chip-tool-onoff.log", []byte(stdout), 0644),
		)

		utils.WaitForLogMessage(t,
			allClustersSnap, "CHIP:ZCL: Toggle ep1 on/off", start)
	})

	// Upgrade chip-tool to local snap or edge
	t.Run("Refresh snap", func(t *testing.T) {
		if utils.LocalServiceSnap() {
			utils.SnapInstallFromFile(t, utils.LocalServiceSnapPath)
		} else {
			utils.SnapRefresh(t, chipToolSnap, "latest/edge")
		}
	})

	// Control device again
	t.Run("Control upgraded snap", func(t *testing.T) {
		snapVersion := utils.SnapVersion(t, chipToolSnap)
		snapRevision := utils.SnapRevision(t, chipToolSnap)
		log.Printf("%s installed version %s build %s\n", chipToolSnap, snapVersion, snapRevision)

		start := time.Now()
		stdout, _, _ := utils.Exec(t, "sudo chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t,
			os.WriteFile(t.Name()+"-chip-tool-onoff.log", []byte(stdout), 0644),
		)

		utils.WaitForLogMessage(t,
			allClustersSnap, "CHIP:ZCL: Toggle ep1 on/off", start)
	})

}
