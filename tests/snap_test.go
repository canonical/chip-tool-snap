package tests

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/require"
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

func TestAllClustersApp(t *testing.T) {
	const (
		allClustersAppBin = "bin/chip-all-clusters-minimal-app-commit-1536ca2"
		allClustersAppLog = "chip-all-clusters-minimal-app.log"
		chipToolLog       = "chip-tool.log"
	)

	// Setup: remove exisiting log files
	// if err := os.Remove("./" + chipAllClusterMinimalAppLog); err != nil && !os.IsNotExist(err) {
	// 	t.Fatalf("Error deleting log file: %s\n", err)
	// }
	// if err := os.Remove("./chip-tool.log"); err != nil && !os.IsNotExist(err) {
	// 	t.Fatalf("Error deleting log file: %s\n", err)
	// }

	// remove existing logs
	// utils.Exec(t, "rm -fr *.log")

	// // Setup: run and log chip-all-clusters-minimal-app in the background
	// logFile, err := os.Create(chipAllClusterMinimalAppLog)
	// if err != nil {
	// 	t.Fatalf("Error creating log file: %s\n", err)
	// }

	// cmd := exec.Command("./" + chipAllClusterMinimalAppFile)
	// cmd.Stdout = logFile
	// cmd.Stderr = logFile

	// err = cmd.Start()
	// if err != nil {
	// 	t.Fatalf("Error starting application: %s\n", err)
	// }

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		stdout, _, _ := utils.ExecContext(t, ctx, allClustersAppBin+" 2>&1")
		require.NoError(t,
			os.WriteFile(allClustersAppLog, []byte(stdout), 0644),
		)
	}()
	t.Log("Waiting for app startup")
	waitForLogMessage(t, allClustersAppLog, "CHIP minimal mDNS started advertising", start)

	t.Cleanup(func() {
		utils.Exec(t, "rm /tmp/chip_*")
	})

	// t.Log("Sleep")
	// time.Sleep(2 * time.Minute)

	t.Run("Commission", func(t *testing.T) {
		// utils.Exec(t, "(sudo echo 'hi4' && sleep 5) > "+chipToolLog)
		utils.Exec(t, "sudo chip-tool pairing onnetwork 110 20202021")
		// utils.ExecVerbose(t, "sudo chip-tool pairing onnetwork 110 20202021")
	})

	t.Run("Control", func(t *testing.T) {
		utils.Exec(t, "sudo chip-tool onoff toggle 110 1")
		waitForLogMessage(t, allClustersAppLog, "CHIP:ZCL: Toggle ep1 on/off", start)
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

func waitForLogMessage(t *testing.T, appLogPath, expectedLog string, since time.Time) {
	const maxRetry = 10

	for i := 1; i <= maxRetry; i++ {
		time.Sleep(1 * time.Second)
		t.Logf("Retry %d/%d: Waiting for expected content in logs: '%s'", i, maxRetry, expectedLog)

		logs, err := os.ReadFile(appLogPath)
		if err != nil {
			t.Fatalf("Error reading log file: %s\n", err)
			continue
		}

		if strings.Contains(string(logs), expectedLog) {
			t.Logf("Found expected content in logs: '%s'", expectedLog)
			return
		}
	}

	t.Fatalf("Time out: reached max %d retries.", maxRetry)
}
