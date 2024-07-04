package thread_tests

import (
	"os"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
	"github.com/stretchr/testify/assert"
)

func TestAllClustersAppThread(t *testing.T) {
	setup(t)

	trimmedActiveDataset := getActiveDataset(t)

	remote_setup(t)

	t.Run("Commission", func(t *testing.T) {
		stdout, _, _ := utils.Exec(t, "sudo chip-tool pairing code-thread 110 hex:"+trimmedActiveDataset+" 34970112332 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-thread-pairing.log", []byte(stdout), 0644),
		)
	})

	t.Run("Control", func(t *testing.T) {
		start := time.Now()
		stdout, _, _ := utils.Exec(t, "sudo chip-tool onoff toggle 110 1 2>&1")
		assert.NoError(t,
			os.WriteFile("chip-tool-thread-onoff.log", []byte(stdout), 0644),
		)

		// 0x6 is the Matter Cluster ID for on-off
		// Using cluster ID here because of a buffering issue in the log stream:
		// https://github.com/canonical/chip-tool-snap/pull/69#issuecomment-2209530275
		remote_waitForLogMessage(t, "matter-all-clusters-app", "ClusterId = 0x6", start)
	})

}
