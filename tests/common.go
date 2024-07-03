package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const allClustersSnap = "matter-all-clusters-app"
const chipToolSnap = "chip-tool"

func InstallChipTool(t *testing.T) {

	// clean
	utils.SnapRemove(t, chipToolSnap)

	if utils.LocalServiceSnap() {
		require.NoError(t,
			utils.SnapInstallFromFile(nil, utils.LocalServiceSnapPath),
		)
	} else {
		require.NoError(t,
			utils.SnapInstallFromStore(nil, chipToolSnap, utils.ServiceChannel),
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
