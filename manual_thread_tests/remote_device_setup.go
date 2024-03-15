package tests

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/canonical/matter-snap-testing/utils"
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

	t.Cleanup(func() {
		defer remoteSSHClient.Close()
		defer remoteSession.Close()
	})

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

	start := time.Now()

	commands := []string{
		"sudo apt install bluez",
		"sudo snap remove --purge matter-all-clusters-app",
		"sudo snap install matter-all-clusters-app --edge",
		fmt.Sprintf("sudo snap set matter-all-clusters-app args='%s'", strings.Join(args, " ")),
		"sudo snap connect matter-all-clusters-app:avahi-control",
		"sudo snap connect matter-all-clusters-app:bluez",
		"sudo snap connect matter-all-clusters-app:otbr-dbus-wpan0 openthread-border-router:dbus-wpan0",
		"sudo snap start matter-all-clusters-app",
	}

	executeRemoteCommands(t, commands)

	utils.WaitForLogMessage(t, "matter-all-clusters-app", "CHIP minimal mDNS started advertising", start)
	// utils.WaitForLogMessage(t, "matter-all-clusters-app", "", start)

	t.Log("Running matter all clusters app")

	return nil
}

func executeRemoteCommand(t *testing.T, command string) string {
	t.Helper()

	if remoteSSHClient == nil {
		t.Fatalf("SSH client not initialized. Please connect to remote device first")
	}

	remoteSession, err := remoteSSHClient.NewSession()
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
		t.Logf("Executed the command remotely: %s\nOutput: %s\n", cmd, output)
	}
}
