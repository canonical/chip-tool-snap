package tests

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
)

const OTCTL = "openthread-border-router.ot-ctl"

func localDeviceOTBRSetup(t *testing.T) {
	t.Helper()

	const (
		defaultInfraInterfaceValue = "wlan0"
		infraInterfaceKey          = "infra-if"
		localInfraInterfaceEnv     = "LOCAL_INFRA_IF"
	)

	// Install OTBR
	utils.SnapInstallFromStore(t, otbrSnap, utils.ServiceChannel)

	// Connect interfaces
	snapInterfaces := []string{":avahi-control", ":firewall-control", ":raw-usb", ":network-control", ":bluetooth-control", ":bluez"}
	for _, interf := range snapInterfaces {
		utils.SnapConnect(nil, otbrSnap+interf, "")
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
	t.Helper()

	activeDataset, _, _ := utils.Exec(t, "sudo "+OTCTL+" dataset active -x | awk '{print $NF}' | grep --invert-match \"Done\"")
	trimmedActiveDataset := strings.TrimSpace(activeDataset)

	return trimmedActiveDataset
}
