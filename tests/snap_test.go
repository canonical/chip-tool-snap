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
	"github.com/stretchr/testify/assert"
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

func TestAllClustersApp(t *testing.T) {
	start := time.Now()
	startAllClustersApp(t)

	// wait for startup
	waitForLogMessage(t,
		allClustersAppLog, "CHIP minimal mDNS started advertising", start)

	t.Run("Commission", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "sudo chip-tool pairing onnetwork 110 20202021 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-pairing.log", []byte(stdout), 0644),
		)
	})

	t.Run("Control", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "sudo chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-onoff.log", []byte(stdout), 0644),
		)

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

func startAllClustersApp(t *testing.T) {
	// remove existing temp files
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

	if err := cmd.Start(); err != nil {
		t.Fatalf("Error starting application: %s", err)
	}

	t.Cleanup(func() {
		utils.Exec(t, "rm -f /tmp/chip_*")
	})
}

func waitForLogMessage(t *testing.T, logPath, expectedMsg string, since time.Time) {
	const maxRetry = 10

	for i := 1; i <= maxRetry; i++ {
		time.Sleep(1 * time.Second)
		t.Logf("Retry %d/%d: Find log message: '%s'", i, maxRetry, expectedMsg)

		logs, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Error reading log file: %s\n", err)
			continue
		}

		if strings.Contains(string(logs), expectedMsg) {
			t.Logf("Found log message: '%s'", expectedMsg)
			return
		}
	}

	t.Fatalf("Time out: reached max %d retries.", maxRetry)
}
