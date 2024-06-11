package tests

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/require"
)

const allClusterSnap = "matter-all-clusters-app"
const chipToolSnap = "chip-tool"

func InstallChipTool(t *testing.T) {

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
}

func PrintSnapVersion(t *testing.T, snapName string) error {
	delimiter := regexp.MustCompile("\\s+")

	snapInfo, _, err := utils.Exec(t, fmt.Sprintf("snap list %s --color=never --unicode=never", snapName))
	if err != nil {
		return err
	}

	lines := strings.Split(snapInfo, "\n")
	for _, line := range lines {
		columns := delimiter.Split(line, -1)
		if columns[0] == snapName {
			if t != nil {
				t.Logf("%s installed version %s\n", snapName, columns[1])
			} else {
				log.Printf("%s installed version %s\n", snapName, columns[1])
			}
			return nil
		}
	}

	return errors.New(fmt.Sprintf("snap '%s' not found", snapName))
}
