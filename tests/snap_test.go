package tests

import (
	"log"
	"os"
	"testing"

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
