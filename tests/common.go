package tests

import (
	"github.com/canonical/matter-snap-testing/env"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/env"
	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const allClustersSnap = "matter-all-clusters-app"
const chipToolSnap = "chip-tool"

func InstallChipTool(t *testing.T) {

	// clean
	utils.SnapRemove(t, chipToolSnap)

	if env.SnapPath() != "" {
		require.NoError(t,
			utils.SnapInstallFromFile(nil, env.SnapPath()),
		)
	} else {
		require.NoError(t,
			utils.SnapInstallFromStore(nil, chipToolSnap, env.SnapChannel()),
		)
	}
	t.Cleanup(func() {
		utils.SnapRemove(t, chipToolSnap)
	})

	// connect interfaces
	utils.SnapConnect(t, chipToolSnap+":avahi-observe", "")
	utils.SnapConnect(t, chipToolSnap+":bluez", "")
	utils.SnapConnect(t, chipToolSnap+":process-control", "")
}

func writeLogFile(t *testing.T, label string, b []byte) {
	assert.NoError(t,
		os.WriteFile(strings.ReplaceAll(t.Name(), "/", "-")+"-"+label+".log", b, 0644),
	)
}

func waitForOnOffHandlingByAllClustersApp(t *testing.T, start time.Time) {
	// 0x6 is the Matter Cluster ID for on-off
	// Using cluster ID here because of a buffering issue in the log stream:
	// https://github.com/canonical/chip-tool-snap/pull/69#issuecomment-2207189962
	utils.WaitForLogMessage(t, allClustersSnap, "ClusterId = 0x6", start)
}
