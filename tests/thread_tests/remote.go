package thread_tests

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

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

func remote_deployOTBRAgent(t *testing.T) {

	t.Cleanup(func() {
		remote_exec(t, "sudo snap remove --purge openthread-border-router")
	})

	start := time.Now().UTC()

	commands := []string{
		"sudo snap remove --purge openthread-border-router",
		"sudo snap install openthread-border-router --edge",
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

	t.Cleanup(func() {
		remote_exec(t, "sudo snap remove --purge matter-all-clusters-app")
	})

	start := time.Now().UTC()

	commands := []string{
		"sudo apt install bluez",
		"sudo snap remove --purge matter-all-clusters-app",
		"sudo snap install matter-all-clusters-app --edge",
		"sudo snap set matter-all-clusters-app args='--thread'",
		"sudo snap connect matter-all-clusters-app:avahi-control",
		"sudo snap connect matter-all-clusters-app:bluez",
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

		command := fmt.Sprintf("sudo journalctl --utc --since \"%s\" --no-pager | grep \"%s\"|| true", start.UTC().Format("2006-01-02 15:04:05"), snap)
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
