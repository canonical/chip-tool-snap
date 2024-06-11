package tests

import (
	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestUpgrade(t *testing.T) {
	if !utils.LocalServiceSnap() {
		t.Skipf("LOCAL_SERVICE_SNAP needs to be defined")
	}

	start := time.Now()

	// Start clean
	utils.SnapRemove(t, allClusterSnap)
	utils.SnapRemove(t, chipToolSnap)

	t.Cleanup(func() {
		utils.SnapRemove(t, allClusterSnap)
		utils.SnapDumpLogs(nil, start, allClusterSnap)
	})

	// Install stable chip tool from store
	require.NoError(t,
		utils.SnapInstallFromStore(nil, chipToolSnap, "latest/stable"),
	)
	t.Cleanup(func() {
		utils.SnapRemove(t, chipToolSnap)
	})

	// Setup chip-tool
	utils.SnapConnect(t, chipToolSnap+":avahi-observe", "")
	utils.SnapConnect(t, chipToolSnap+":bluez", "")
	utils.SnapConnect(t, chipToolSnap+":process-control", "")

	// Install all clusters app
	require.NoError(t,
		utils.SnapInstallFromStore(t, allClusterSnap, utils.ServiceChannel),
	)
	t.Cleanup(func() {
		utils.SnapRemove(t, allClusterSnap)
	})

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
	t.Run("Control Stable", func(t *testing.T) {
		err := PrintSnapVersion(t, chipToolSnap)
		if err != nil {
			t.Fatalf("snap not installed")
		}

		stdout, _, _ := utils.Exec(t, "sudo chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-onoff.log", []byte(stdout), 0644),
		)

		utils.WaitForLogMessage(t,
			allClusterSnap, "CHIP:ZCL: Toggle ep1 on/off", start)
	})

	// Upgrade chip-tool to local snap
	require.NoError(t,
		utils.SnapInstallFromFile(nil, utils.LocalServiceSnapPath),
	)

	// Control device again
	t.Run("Control Local", func(t *testing.T) {
		err := PrintSnapVersion(t, chipToolSnap)
		if err != nil {
			t.Fatalf("snap not installed")
		}

		stdout, _, _ := utils.Exec(t, "sudo chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-onoff.log", []byte(stdout), 0644),
		)

		utils.WaitForLogMessage(t,
			allClusterSnap, "CHIP:ZCL: Toggle ep1 on/off", start)
	})

}
