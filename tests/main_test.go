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

const (
	allClustersAppBin = "bin/chip-all-clusters-minimal-app-commit-1536ca2"
	allClustersAppLog = "chip-all-clusters-minimal-app.log"
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

func startAllClustersApp(t *testing.T, args ...string) {
	// remove existing temp files
	utils.Exec(t, "rm -fr /tmp/chip_*")

	logFile, err := os.Create(allClustersAppLog)
	if err != nil {
		t.Fatalf("Error creating log file: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cmd := exec.CommandContext(ctx, allClustersAppBin, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	t.Logf("[exec] %s\n", cmd)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Error starting application: %s", err)
	}

	t.Cleanup(func() {
		utils.Exec(t, "rm -f /tmp/chip_*")
	})
}

func waitForLogMessage(t *testing.T, logPath, expectedMsg string) {
	const maxRetry = 10

	for i := 1; i <= maxRetry; i++ {
		time.Sleep(1 * time.Second)
		t.Logf("Retry %d/%d: Find log message: '%s'", i, maxRetry, expectedMsg)

		logs, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Error reading log file: %s\n", err)
		}

		if strings.Contains(string(logs), expectedMsg) {
			t.Logf("Found log message: '%s'", expectedMsg)
			return
		}
	}

	t.Fatalf("Time out: reached max %d retries.", maxRetry)
}
