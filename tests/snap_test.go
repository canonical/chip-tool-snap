package tests

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
)

var start = time.Now()

func TestMain(m *testing.M) {
	teardown, err := setup()
	if err != nil {
		log.Fatalf("Failed to setup tests: %s", err)
	}

	code := m.Run()
	teardown()

	os.Exit(code)
}

func TestMatterDeviceOperations(t *testing.T) {
	const (
		chipAllClusterMinimalAppFile = "bin/chip-all-clusters-minimal-app-commit-1536ca2"
		chipAllClusterMinimalAppLog  = "chip-all-clusters-minimal-app.log"
	)

	// Setup: remove exisiting log files
	if err := os.Remove("./" + chipAllClusterMinimalAppLog); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Error deleting log file: %s\n", err)
	}
	if err := os.Remove("./chip-tool.log"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Error deleting log file: %s\n", err)
	}

	// Setup: run and log chip-all-clusters-minimal-app in the background
	logFile, err := os.Create(chipAllClusterMinimalAppLog)
	if err != nil {
		t.Fatalf("Error creating log file: %s\n", err)
	}

	cmd := exec.Command("./" + chipAllClusterMinimalAppFile)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	if err != nil {
		t.Fatalf("Error starting application: %s\n", err)
	}

	t.Cleanup(func() {
		matches, err := filepath.Glob("/tmp/chip_*")
		if err != nil {
			t.Fatalf("Error finding tmp chip files: %s\n", err)
		}

		for _, match := range matches {
			if err := os.Remove(match); err != nil {
				t.Fatalf("Error removing tmp chip file %s: %s\n", match, err)
			}
		}

		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("Error killing process: %s\n", err)
		}

		if logFile != nil {
			logFile.Close()
		}
	})

	t.Run("Commission", func(t *testing.T) {
		utils.Exec(t, "sudo chip-tool pairing onnetwork 110 20202021")
	})

	t.Run("Control", func(t *testing.T) {
		utils.Exec(t, "sudo chip-tool onoff toggle 110 1")
		waitForAppMessage(t, "./"+chipAllClusterMinimalAppLog, "CHIP:ZCL: Toggle ep1 on/off", start)
	})
}

func setup() (teardown func(), err error) {
	const chipToolSnap = "chip-tool"

	log.Println("[CLEAN]")
	utils.SnapRemove(nil, chipToolSnap)

	log.Println("[SETUP]")

	teardown = func() {
		log.Println("[TEARDOWN]")
		utils.SnapDumpLogs(nil, start, chipToolSnap)

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

func waitForAppMessage(t *testing.T, appLogPath, expectedLog string, since time.Time) {
	const maxRetry = 10

	for i := 1; i <= maxRetry; i++ {
		time.Sleep(1 * time.Second)
		t.Logf("Retry %d/%d: Waiting for expected content in logs: %s", i, maxRetry, expectedLog)

		logs, err := readLogFile(appLogPath)
		if err != nil {
			t.Fatalf("Error reading log file: %s\n", err)
			continue
		}

		if strings.Contains(logs, expectedLog) {
			t.Logf("Found expected content in logs: %s", expectedLog)
			return
		}
	}

	t.Fatalf("Time out: reached max %d retries.", maxRetry)
}

func readLogFile(filePath string) (string, error) {
	text, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(text), nil
}
