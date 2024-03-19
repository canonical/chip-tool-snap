package tests

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
	remoteIP             = ""
	remoteInfraInterface = ""

	remoteSSHClient *ssh.Client
	remoteSession   *ssh.Session
)

func RemoteDeviceSetup(t *testing.T) {
	t.Helper()

	setRemoteConfigFromEnv()

	establishSSHConnection(t)

	deployOTBRAgentOnRemoteDevice(t)

	startAllClustersAppOnRemoteDevice(t, "--thread")
}

func setRemoteConfigFromEnv() {
	const (
		remoteUserEnv           = "REMOTE_USER"
		remotePasswordEnv       = "REMOTE_PASSWORD"
		remoteIPEnv             = "REMOTE_IP"
		remoteInfraInterfaceEnv = "REMOTE_INFRA_IF"

		defaultSSHConnectionUser = "ubuntu"
	)

	if v := os.Getenv(remoteUserEnv); v != "" {
		remoteUser = v
	} else {
		remoteUser = defaultSSHConnectionUser
	}

	if v := os.Getenv(remotePasswordEnv); v != "" {
		remotePassword = v
	}

	if v := os.Getenv(remoteIPEnv); v != "" {
		remoteIP = v
	}

	if v := os.Getenv(remoteInfraInterfaceEnv); v != "" {
		remoteInfraInterface = v
	}
}

func establishSSHConnection(t *testing.T) {
	t.Helper()

	const defaultSSHConnectionPort = "22"

	if remoteSSHClient != nil {
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
	remoteSSHClient, err = ssh.Dial("tcp", remoteIP+":"+defaultSSHConnectionPort, config)
	if err != nil {
		t.Fatalf("Failed to dial: %v\n", err)
	}

	t.Logf("SSH connection to %s established successfully", remoteIP)
}

func deployOTBRAgentOnRemoteDevice(t *testing.T) error {
	t.Helper()

	commands := []string{
		"sudo snap remove --purge openthread-border-router",
		"sudo snap install openthread-border-router --edge",
		"sudo snap set openthread-border-router infra-if='" + remoteInfraInterface + "'",
		"sudo snap connect openthread-border-router:avahi-control",
		"sudo snap connect openthread-border-router:firewall-control",
		"sudo snap connect openthread-border-router:raw-usb",
		"sudo snap connect openthread-border-router:network-control",
		"sudo snap connect openthread-border-router:bluetooth-control",
		"sudo snap connect openthread-border-router:bluez",
		"sudo snap start openthread-border-router",
	}

	executeRemoteCommands(t, commands)

	return nil
}

func startAllClustersAppOnRemoteDevice(t *testing.T, args ...string) error {
	t.Helper()

	start := time.Now().UTC()

	commands := []string{
		"sudo apt install bluez",
		"sudo snap remove --purge " + allClusterSnap,
		"sudo snap install " + allClusterSnap + " --edge",
		fmt.Sprintf("sudo snap set "+allClusterSnap+" args='%s'", strings.Join(args, " ")),
		"sudo snap connect " + allClusterSnap + ":avahi-control",
		"sudo snap connect " + allClusterSnap + ":bluez",
		"sudo snap connect " + allClusterSnap + ":otbr-dbus-wpan0 openthread-border-router:dbus-wpan0",
		"sudo snap start " + allClusterSnap,
	}

	executeRemoteCommands(t, commands)

	waitForLogMessageOnRemoteDevice(t, allClusterSnap, "CHIP minimal mDNS started advertising", start)
	t.Log("Running matter all clusters app")

	return nil
}

func executeRemoteCommand(t *testing.T, command string) string {
	t.Helper()

	if remoteSSHClient == nil {
		t.Fatalf("SSH client not initialized. Please connect to remote device first")
	}

	var err error
	remoteSession, err = remoteSSHClient.NewSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	stdout, err := remoteSession.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	if err := remoteSession.Start(command); err != nil {
		t.Fatalf("Failed to start command '%s': %v", command, err)
	}

	output, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("Failed to read command output: %v", err)
	}

	if err := remoteSession.Wait(); err != nil {
		t.Fatalf("Command '%s' failed: %v", command, err)
	}

	return string(output)
}

func executeRemoteCommands(t *testing.T, commands []string) {
	t.Helper()

	for _, cmd := range commands {
		output := executeRemoteCommand(t, cmd)
		t.Logf("Executed the command remotely: %s", cmd)
		t.Logf("Output: %s", output)
	}
}

func waitForLogMessageOnRemoteDevice(t *testing.T, snap string, expectedLog string, start time.Time) {
	t.Helper()

	const maxRetry = 10
	for i := 1; i <= maxRetry; i++ {
		time.Sleep(1 * time.Second)
		t.Logf("Retry %d/%d: Waiting for expected content in logs: %s", i, maxRetry, expectedLog)

		command := fmt.Sprintf("sudo journalctl --since \"%s\" --no-pager | grep \"%s\"|| true", start.Format("2006-01-02 15:04:05"), snap)
		t.Logf(command)
		logs := executeRemoteCommand(t, command)
		if strings.Contains(logs, expectedLog) {
			t.Logf("Found expected content in logs: %s", expectedLog)
			return
		}
	}

	t.Fatalf("Time out: reached max %d retries.", maxRetry)
}
