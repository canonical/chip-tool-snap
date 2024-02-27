package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	remoteUserEnv     = "REMOTE_USER"
	remotePasswordEnv = "REMOTE_PASSWORD"
	remoteIPEnv       = "REMOTE_IP"

	remoteInfraInterfaceEnv = "REMOTE_INFRA_IF"

	defaultSSHConnectionPort = "22"
	defaultSSHConnectionUser = "ubuntu"
)

var (
	remoteUser          = ""
	remotePassword      = ""
	remoteIP            = ""
	infraInterfaceValue = ""
)

func setupRemoteThreadNetwork(t *testing.T) string {
	// Set remote configuration from environment variables
	setRemoteConfigFromEnv()

	// Connect to the remote device
	client := connectToRemoteDevice(t)
	defer client.Close()

	// Install OTBR and form Thread network
	if err := installOTBR(client, t); err != nil {
		t.Fatalf("failed to setup remote Thread network: %s", err)
	}

	// Get Thread network active dataset
	activeDataset, err := getActiveDataset(client, t)
	if err != nil {
		t.Fatalf("failed to retrieve active dataset: %s", err)
	}

	return activeDataset
}

func cleanupRemoteDevice(t *testing.T) string {
	client := connectToRemoteDevice(t)

	if _, err := runRemoteCommand(client, "sudo snap remove --purge openthread-border-router"); err != nil {
		t.Fatalf("remote commands failed due to err: %s", err)
	}

	return ""
}

// runRemoteCommand executes a command remotely via SSH
func runRemoteCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("error executing command '%s': %v\nOutput: %s", command, err, output)
	}

	return string(output), nil
}

// connectToRemoteDevice establishes SSH connection to the remote device
func connectToRemoteDevice(t *testing.T) *ssh.Client {
	config := &ssh.ClientConfig{
		User: remoteUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(remotePassword),
		},
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to the remote machine
	client, err := ssh.Dial("tcp", remoteIP+":"+defaultSSHConnectionPort, config) // Replace with your remote machine IP
	if err != nil {
		t.Fatalf("Failed to dial: %v\n", err)
	}
	t.Logf("SSH connection to %s established successfully", remoteIP)

	return client
}

// setRemoteConfigFromEnv sets configuration from environment variables
func setRemoteConfigFromEnv() {
	// Get and set remote user
	if v := os.Getenv(remoteUserEnv); v != "" {
		remoteUser = v
	} else {
		remoteUser = defaultSSHConnectionUser
	}

	// Get and set remote password
	if v := os.Getenv(remotePasswordEnv); v != "" {
		remotePassword = v
	}

	// Get and set remote IP
	if v := os.Getenv(remoteIPEnv); v != "" {
		remoteIP = v
	}

	// Get and set infrastructure interface
	if v := os.Getenv(remoteInfraInterfaceEnv); v != "" {
		infraInterfaceValue = v
	}
}

// installOTBR installs OTBR and configures Thread network
func installOTBR(client *ssh.Client, t *testing.T) error {
	commands := []string{
		"sudo snap remove --purge openthread-border-router",
		"sudo snap install openthread-border-router --edge",
		"sudo snap set openthread-border-router infra-if=" + infraInterfaceValue,
		"sudo snap connect openthread-border-router:avahi-control",
		"sudo snap connect openthread-border-router:firewall-control",
		"sudo snap connect openthread-border-router:raw-usb",
		"sudo snap connect openthread-border-router:network-control",
		"sudo snap connect openthread-border-router:bluetooth-control",
		"sudo snap connect openthread-border-router:bluez",
		"sudo snap start openthread-border-router",
		"sleep 10",
		"sudo openthread-border-router.ot-ctl dataset init new",
		"sudo openthread-border-router.ot-ctl dataset commit active",
		"sudo openthread-border-router.ot-ctl ifconfig up",
		"sudo openthread-border-router.ot-ctl thread start",
	}

	for _, cmd := range commands {
		if _, err := runRemoteCommand(client, cmd); err != nil {
			return err
		}
	}
	return nil
}

// getActiveDataset retrieves the active dataset of Thread network from the remote device
func getActiveDataset(client *ssh.Client, t *testing.T) (string, error) {
	output, err := runRemoteCommand(client, "sudo openthread-border-router.ot-ctl dataset active -x | awk '{print $NF}' | grep --invert-match \"Done\"")
	if err != nil {
		return "", err
	}
	return output, nil
}
