package thread_tests

import (
	"os"
	"strings"
	"testing"
	"time"

	tests "chip-tool-snap-tests"

	"github.com/canonical/matter-snap-testing/utils"
)

func setup(t *testing.T) {
	tests.InstallChipTool(t)

	// Clean
	utils.SnapRemove(t, otbrSnap)

	// Install OTBR
	otbrInstallTime := time.Now()
	utils.SnapInstallFromStore(t, otbrSnap, "latest/beta")
	t.Cleanup(func() {
		logs := utils.SnapLogs(t, otbrInstallTime, otbrSnap)
		utils.WriteLogFile(t, otbrSnap, logs)
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

	// Set radio url
	if v := os.Getenv(localRadioUrlEnv); v != "" {
		radioUrlValue := v
		utils.SnapSet(nil, otbrSnap, radioUrlKey, radioUrlValue)
	} else {
		utils.SnapSet(nil, otbrSnap, radioUrlKey, defaultRadioUrl)
	}

	// Start OTBR
	otbrStartTime := time.Now()
	utils.SnapStart(t, otbrSnap)
	utils.WaitForLogMessage(t, otbrSnap, "Start Thread Border Agent: OK", otbrStartTime)

	// Form Thread network
	utils.Exec(t, "sudo "+OTCTL+" dataset init new")
	utils.Exec(t, "sudo "+OTCTL+" dataset commit active")
	utils.Exec(t, "sudo "+OTCTL+" ifconfig up")
	utils.Exec(t, "sudo "+OTCTL+" thread start")
	utils.WaitForLogMessage(t, otbrSnap, "Thread Network", otbrStartTime)
}

func getActiveDataset(t *testing.T) string {
	activeDataset, _, _ := utils.Exec(t, "sudo "+OTCTL+" dataset active -x | awk '{print $NF}' | grep --invert-match \"Done\"")
	trimmedActiveDataset := strings.TrimSpace(activeDataset)

	return trimmedActiveDataset
}
