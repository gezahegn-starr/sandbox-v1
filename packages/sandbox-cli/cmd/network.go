package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// networkCmd is the parent management command for sandbox networking.
var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage sandbox networking",
	Long:  `Manage networking for Apple container sandboxes.`,
}

var networkLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List networks",
	Args:  cobra.NoArgs,
	RunE:  runNetworkLs,
}

var networkConnectCmd = &cobra.Command{
	Use:   "connect NETWORK SANDBOX",
	Short: "Connect a sandbox to a network",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runNetworkConnect,
}

var networkDisconnectCmd = &cobra.Command{
	Use:   "disconnect NETWORK SANDBOX",
	Short: "Disconnect a sandbox from a network",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runNetworkDisconnect,
}

func init() {
	rootCmd.AddCommand(networkCmd)
	networkCmd.AddCommand(networkLsCmd)
	networkCmd.AddCommand(networkConnectCmd)
	networkCmd.AddCommand(networkDisconnectCmd)
}

func runNetworkLs(_ *cobra.Command, _ []string) error {
	debugLog("exec: container network ls")
	c := exec.Command(containerBin(), "network", "ls")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("listing networks: %w", err)
	}
	return nil
}

func runNetworkConnect(_ *cobra.Command, args []string) error {
	network := args[0]
	sandboxID := ""
	if len(args) >= 2 {
		sandboxID = args[1]
	} else {
		var err error
		sandboxID, err = pickSandbox(fmt.Sprintf("Select a sandbox to connect to network %q", network))
		if err != nil {
			return err
		}
	}
	debugLog("exec: container network connect %s %s", network, sandboxID)
	c := exec.Command(containerBin(), "network", "connect", network, sandboxID)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("connecting %s to network %s: %w", sandboxID, network, err)
	}
	fmt.Fprintf(os.Stderr, "Connected %s to network %s\n", sandboxID, network)
	return nil
}

func runNetworkDisconnect(_ *cobra.Command, args []string) error {
	network := args[0]
	sandboxID := ""
	if len(args) >= 2 {
		sandboxID = args[1]
	} else {
		var err error
		sandboxID, err = pickSandbox(fmt.Sprintf("Select a sandbox to disconnect from network %q", network))
		if err != nil {
			return err
		}
	}
	debugLog("exec: container network disconnect %s %s", network, sandboxID)
	c := exec.Command(containerBin(), "network", "disconnect", network, sandboxID)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("disconnecting %s from network %s: %w", sandboxID, network, err)
	}
	fmt.Fprintf(os.Stderr, "Disconnected %s from network %s\n", sandboxID, network)
	return nil
}
