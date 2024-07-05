package tests

import (
	"log"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
)

func TestUpgrade(t *testing.T) {
	start := time.Now()

	t.Cleanup(func() {
		utils.SnapRemove(nil, allClustersSnap)
		utils.SnapDumpLogs(nil, start, allClustersSnap)

		utils.SnapRemove(nil, chipToolSnap)
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

	t.Run("Commission", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "chip-tool pairing onnetwork 110 20202021 2>&1")
		writeLogFile(t, "chip-tool-pairing", []byte(stdout))
	})

	t.Run("Control before upgrade", func(t *testing.T) {
		snapVersion := utils.SnapVersion(t, chipToolSnap)
		snapRevision := utils.SnapRevision(t, chipToolSnap)
		log.Printf("%s installed version %s build %s\n", chipToolSnap, snapVersion, snapRevision)

		start := time.Now()
		stdout, _, _ := utils.Exec(t, "chip-tool onoff on 110 1 2>&1")
		writeLogFile(t, "chip-tool-onoff", []byte(stdout))

		waitForOnOffHandlingByAllClustersApp(t, start)
	})

	t.Run("Upgrade snap", func(t *testing.T) {
		if utils.LocalServiceSnap() {
			utils.SnapInstallFromFile(t, utils.LocalServiceSnapPath)
		} else {
			utils.SnapRefresh(t, chipToolSnap, "latest/edge")
		}
	})

	t.Run("Control upgraded snap", func(t *testing.T) {
		snapVersion := utils.SnapVersion(t, chipToolSnap)
		snapRevision := utils.SnapRevision(t, chipToolSnap)
		log.Printf("%s installed version %s build %s\n", chipToolSnap, snapVersion, snapRevision)

		start := time.Now()
		stdout, _, _ := utils.Exec(t, "chip-tool onoff off 110 1 2>&1")
		writeLogFile(t, "chip-tool-onoff", []byte(stdout))

		waitForOnOffHandlingByAllClustersApp(t, start)
	})

}
