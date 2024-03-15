package tests

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
)

const (
	defaultInfraInterfaceValue = "wlan0"
	infraInterfaceKey          = "infra-if"
	localInfraInterfaceEnv     = "LOCAL_INFRA_IF"
)

func localDeviceSetup(t *testing.T) {
	t.Helper()

	// Install OTBR
	utils.SnapInstallFromStore(t, otbrSnap, utils.ServiceChannel)

	// Setup OTBR
	utils.SnapConnect(nil, otbrSnap+":avahi-control", "")
	utils.SnapConnect(nil, otbrSnap+":firewall-control", "")
	utils.SnapConnect(nil, otbrSnap+":raw-usb", "")
	utils.SnapConnect(nil, otbrSnap+":network-control", "")
	utils.SnapConnect(nil, otbrSnap+":bluetooth-control", "")
	utils.SnapConnect(nil, otbrSnap+":bluez", "")

	if v := os.Getenv(localInfraInterfaceEnv); v != "" {
		infraInterfaceValue := v
		utils.SnapSet(nil, otbrSnap, infraInterfaceKey, infraInterfaceValue)
	} else {
		utils.SnapSet(nil, otbrSnap, infraInterfaceKey, defaultInfraInterfaceValue)
	}

	// Form Thread network
	start := time.Now()
	utils.SnapStart(t, otbrSnap)

	time.Sleep(5 * time.Second)
	utils.Exec(t, "sudo openthread-border-router.ot-ctl dataset init new")
	utils.Exec(t, "sudo openthread-border-router.ot-ctl dataset commit active")
	utils.Exec(t, "sudo openthread-border-router.ot-ctl ifconfig up")
	utils.Exec(t, "sudo openthread-border-router.ot-ctl thread start")

	utils.WaitForLogMessage(t, otbrSnap, "Thread Network", start)
}

func getActiveDataset(t *testing.T) string {
	t.Helper()

	activeDataset, _, _ := utils.Exec(t, "sudo openthread-border-router.ot-ctl dataset active -x | awk '{print $NF}' | grep --invert-match \"Done\"")
	trimmedActiveDataset := strings.TrimSpace(activeDataset)

	return trimmedActiveDataset
}
