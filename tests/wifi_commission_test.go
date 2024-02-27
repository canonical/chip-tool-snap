package tests

import (
	"os"
	"testing"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
)

func TestAllClustersAppWiFi(t *testing.T) {
	startAllClustersApp(t, "--wifi")

	// wait for startup
	waitForLogMessage(t,
		allClustersAppLog, "CHIP minimal mDNS started advertising")

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
			allClustersAppLog, "CHIP:ZCL: Toggle ep1 on/off")
	})

}
