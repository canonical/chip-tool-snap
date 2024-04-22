package thread_tests

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/require"
)

const (
	otbrSnap = "openthread-border-router"
	OTCTL    = otbrSnap + ".ot-ctl"
)

func setup(t *testing.T) {
	installChipTool(t)

	const (
		defaultInfraInterfaceValue = "wlan0"
		infraInterfaceKey          = "infra-if"
		localInfraInterfaceEnv     = "LOCAL_INFRA_IF"
	)

	// Clean
	utils.SnapRemove(t, otbrSnap)

	// Install OTBR
	utils.SnapInstallFromStore(t, otbrSnap, utils.ServiceChannel)
	t.Cleanup(func() {
		utils.SnapRemove(t, otbrSnap)
	})

	// Connect interfaces
	snapInterfaces := []string{"avahi-control", "firewall-control", "raw-usb", "network-control", "bluetooth-control", "bluez"}
	for _, interfaceSlot := range snapInterfaces {
		utils.SnapConnect(nil, otbrSnap+":"+interfaceSlot, "")
	}

	// Set infra interface
	if v := os.Getenv(localInfraInterfaceEnv); v != "" {
		infraInterfaceValue := v
		utils.SnapSet(nil, otbrSnap, infraInterfaceKey, infraInterfaceValue)
	} else {
		utils.SnapSet(nil, otbrSnap, infraInterfaceKey, defaultInfraInterfaceValue)
	}

	// Start OTBR
	start := time.Now()
	utils.SnapStart(t, otbrSnap)
	utils.WaitForLogMessage(t, otbrSnap, "Start Thread Border Agent: OK", start)

	// Form Thread network
	utils.Exec(t, "sudo "+OTCTL+" dataset init new")
	utils.Exec(t, "sudo "+OTCTL+" dataset commit active")
	utils.Exec(t, "sudo "+OTCTL+" ifconfig up")
	utils.Exec(t, "sudo "+OTCTL+" thread start")
	utils.WaitForLogMessage(t, otbrSnap, "Thread Network", start)
}

func getActiveDataset(t *testing.T) string {
	activeDataset, _, _ := utils.Exec(t, "sudo "+OTCTL+" dataset active -x | awk '{print $NF}' | grep --invert-match \"Done\"")
	trimmedActiveDataset := strings.TrimSpace(activeDataset)

	return trimmedActiveDataset
}

func installChipTool(t *testing.T) {
	const chipToolSnap = "chip-tool"

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

	return
}
