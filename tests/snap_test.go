package tests

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
)

func TestMain(m *testing.M) {
	teardown, err := setup()
	if err != nil {
		log.Fatalf("Failed to setup tests: %s", err)
	}

	code := m.Run()
	teardown()

	os.Exit(code)
}

func TestAllClustersApp(t *testing.T) {
	const (
		allClustersAppBin = "bin/chip-all-clusters-minimal-app-commit-1536ca2"
		allClustersAppLog = "chip-all-clusters-minimal-app.log"
	)

	// clean
	utils.Exec(t, "rm -fr /tmp/chip_*")

	logFile, err := os.Create(allClustersAppLog)
	if err != nil {
		t.Fatalf("Error creating log file: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, allClustersAppBin)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	start := time.Now()
	if err := cmd.Start(); err != nil {
		t.Fatalf("Error starting application: %s", err)
	}

	waitForLogMessage(t,
		allClustersAppLog, "CHIP minimal mDNS started advertising", start)

	t.Cleanup(func() {
		utils.Exec(t, "rm /tmp/chip_*")
	})

	t.Run("Commission", func(t *testing.T) {
		utils.ExecVerbose(t, "sudo chip-tool pairing onnetwork 110 20202021")
	})

	t.Run("Control", func(t *testing.T) {
		utils.ExecVerbose(t, "sudo chip-tool onoff toggle 110 1")

		waitForLogMessage(t,
			allClustersAppLog, "CHIP:ZCL: Toggle ep1 on/off", start)
	})

}

func setup() (teardown func(), err error) {
	const chipToolSnap = "chip-tool"

	log.Println("[CLEAN]")
	utils.SnapRemove(nil, chipToolSnap)

	log.Println("[SETUP]")

	teardown = func() {
		log.Println("[TEARDOWN]")

		log.Println("Removing installed snap:", !utils.SkipTeardownRemoval)
		if !utils.SkipTeardownRemoval {
			utils.SnapRemove(nil, chipToolSnap)
		}
	}

	if utils.LocalServiceSnap() {
		err = utils.SnapInstallFromFile(nil, utils.LocalServiceSnapPath)
	} else {
		err = utils.SnapInstallFromStore(nil, chipToolSnap, utils.ServiceChannel)
	}
	if err != nil {
		teardown()
		return
	}

	// connect interfaces
	utils.SnapConnect(nil, chipToolSnap+":avahi-observe", "")
	utils.SnapConnect(nil, chipToolSnap+":bluez", "")
	utils.SnapConnect(nil, chipToolSnap+":process-control", "")

	return
}

func waitForLogMessage(t *testing.T, logPath, expectedMsg string, since time.Time) {
	const maxRetry = 10

	for i := 1; i <= maxRetry; i++ {
		time.Sleep(1 * time.Second)
		t.Logf("Retry %d/%d: Waiting log message: '%s'", i, maxRetry, expectedMsg)

		logs, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Error reading log file: %s\n", err)
			continue
		}

		if strings.Contains(string(logs), expectedMsg) {
			t.Logf("Found expected log message: '%s'", expectedMsg)
			return
		}
	}

	t.Fatalf("Time out: reached max %d retries.", maxRetry)
}
