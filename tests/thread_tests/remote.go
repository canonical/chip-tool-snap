package thread_tests

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
	"golang.org/x/crypto/ssh"
)

var (
	remoteUser           = ""
	remotePassword       = ""
	remoteHost           = ""
	remoteInfraInterface = defaultInfraInterfaceValue
	remoteRadioUrl       = defaultRadioUrl

	SSHClient *ssh.Client
)

func remote_setup(t *testing.T) {
	remote_loadEnvVars()

	connectSSH(t)

	remote_deployOTBRAgent(t)

	remote_deployAllClustersApp(t)
}

func remote_loadEnvVars() {
	if v := os.Getenv(remoteUserEnv); v != "" {
		remoteUser = v
	}

	if v := os.Getenv(remotePasswordEnv); v != "" {
		remotePassword = v
	}

	if v := os.Getenv(remoteHostEnv); v != "" {
		remoteHost = v
	}

	if v := os.Getenv(remoteInfraInterfaceEnv); v != "" {
		remoteInfraInterface = v
	}

	if v := os.Getenv(remoteRadioUrlEnv); v != "" {
		remoteRadioUrl = v
	}
}

func connectSSH(t *testing.T) {
	if SSHClient != nil {
		return
	}

	config := &ssh.ClientConfig{
		User: remoteUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(remotePassword),
		},
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	var err error
	SSHClient, err = ssh.Dial("tcp", remoteHost+":22", config)
	if err != nil {
		t.Fatalf("Failed to dial: %s", err)
	}

	t.Cleanup(func() {
		SSHClient.Close()
	})

	t.Logf("SSH: connected to %s", remoteHost)
}

func getStartTimestamp(t *testing.T) time.Time {
	t.Helper()
	// Get the current unix timestamp on the remote device
	start := remote_exec(t, "date +%s")
	start = strings.TrimSpace(start)
	start = strings.TrimSuffix(start, "\n")
	startTimestamp, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		t.Fatalf("Failed to parse start timestamp: %v", err)
	}
	return time.Unix(startTimestamp, 0)
}

func remote_deployOTBRAgent(t *testing.T) {
	start := getStartTimestamp(t)

	t.Cleanup(func() {
		dumpRemoteLogs(t, "openthread-border-router", start)
		remote_exec(t, "sudo snap remove --purge openthread-border-router")
	})

	commands := []string{
		"sudo snap remove --purge openthread-border-router",
		"sudo snap install openthread-border-router --channel=latest/beta",
		fmt.Sprintf("sudo snap set openthread-border-router %s='%s'", infraInterfaceKey, remoteInfraInterface),
		fmt.Sprintf("sudo snap set openthread-border-router %s='%s'", radioUrlKey, remoteRadioUrl),
		// "sudo snap connect openthread-border-router:avahi-control",
		"sudo snap connect openthread-border-router:firewall-control",
		"sudo snap connect openthread-border-router:raw-usb",
		"sudo snap connect openthread-border-router:network-control",
		// "sudo snap connect openthread-border-router:bluetooth-control",
		// "sudo snap connect openthread-border-router:bluez",
		"sudo snap start openthread-border-router",
	}
	for _, cmd := range commands {
		remote_exec(t, cmd)
	}

	remote_waitForLogMessage(t, otbrSnap, "Start Thread Border Agent: OK", start)
	t.Log("OTBR on remote device is ready")
}

func remote_deployAllClustersApp(t *testing.T) {
	start := getStartTimestamp(t)

	t.Cleanup(func() {
		dumpRemoteLogs(t, "matter-all-clusters-app", start)
		remote_exec(t, "sudo snap remove --purge matter-all-clusters-app")
	})

	commands := []string{
		// "sudo apt install -y bluez",
		"sudo snap remove --purge matter-all-clusters-app",
		"sudo snap install matter-all-clusters-app --channel=latest/beta",
		"sudo snap set matter-all-clusters-app args='--thread'",
		"sudo snap connect matter-all-clusters-app:avahi-control",
		// "sudo snap connect matter-all-clusters-app:bluez",
		"sudo snap connect matter-all-clusters-app:otbr-dbus-wpan0 openthread-border-router:dbus-wpan0",
		"sudo snap start matter-all-clusters-app",
	}
	for _, cmd := range commands {
		remote_exec(t, cmd)
	}

	remote_waitForLogMessage(t, "matter-all-clusters-app", "CHIP minimal mDNS started advertising", start)
	t.Log("Matter All Clusters App is ready")
}

func remote_exec(t *testing.T, command string) string {
	t.Helper()

	t.Logf("[exec-ssh] %s", command)

	// Remote commands that require sudo might ask for the password. Always pass it in. See https://stackoverflow.com/a/11955358
	if strings.HasPrefix(command, "sudo ") {
		command = strings.TrimPrefix(command, "sudo ")
		escapedPassword := strings.ReplaceAll(remotePassword, `"`, `\"`)
		command = fmt.Sprintf(`echo "%s" | sudo -S %s`, escapedPassword, command)
	}

	if SSHClient == nil {
		t.Fatalf("SSH client not initialized. Please connect to remote device first")
	}

	session, err := SSHClient.NewSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	if err := session.Start(command); err != nil {
		t.Fatalf("Failed to start session with command '%s': %v", command, err)
	}

	output, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("Failed to read command output: %v", err)
	}

	if err := session.Wait(); err != nil {
		t.Fatalf("Command '%s' failed: %v", command, err)
	}

	return string(output)
}

func remote_waitForLogMessage(t *testing.T, snap string, expectedLog string, start time.Time) {
	t.Helper()

	const maxRetry = 10
	for i := 1; i <= maxRetry; i++ {
		time.Sleep(1 * time.Second)
		t.Logf("Retry %d/%d: Waiting for expected content in logs: '%s'", i, maxRetry, expectedLog)

		// Use Unix timestamp which is timezone-independent
		// journalctl accepts timestamps in the format @UNIX_TIMESTAMP
		command := fmt.Sprintf("sudo journalctl --since @%d --no-pager | grep \"%s\" || true", start, snap)
		logs := remote_exec(t, command)
		if strings.Contains(logs, expectedLog) {
			t.Logf("Found expected content in logs: '%s'", expectedLog)
			return
		}
	}

	t.Logf("Time out: reached max %d retries.", maxRetry)
	t.Log(remote_exec(t, "journalctl --no-pager --lines=10 --unit=snap.openthread-border-router.otbr-agent --priority=notice"))
	t.FailNow()
}

func dumpRemoteLogs(t *testing.T, label string, start time.Time) error {
	command := fmt.Sprintf("sudo journalctl --since @%d --no-pager | grep \"%s\" || true", start.Unix(), label)
	logs := remote_exec(t, command)
	return utils.WriteLogFile(t, "remote-"+label, logs)
}
