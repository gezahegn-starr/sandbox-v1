package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"text/tabwriter"
	"os"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List sandboxes",
	Long:  `List all Apple container sandboxes (containers named copilot-*).`,
	Args:  cobra.NoArgs,
	RunE:  runLs,
}

var lsAll bool

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().BoolVarP(&lsAll, "all", "a", false, "Show all containers, not just copilot-* sandboxes")
}

func runLs(_ *cobra.Command, _ []string) error {
	out, err := exec.Command(containerBin(), "ls", "--all", "--format", "json").CombinedOutput()
	if err != nil {
		return fmt.Errorf("listing containers: %w\n%s", err, out)
	}

	var containers []Container
	if err := json.Unmarshal(out, &containers); err != nil {
		return fmt.Errorf("parsing container list: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "SANDBOX\tSTATUS\tIMAGE\tWORKSPACE")

	for _, c := range containers {
		if !lsAll && !c.isSandbox() {
			continue
		}
		image := c.Config.Image.Reference
		if image == "" {
			image = "-"
		}
		// Trim digest suffix for readability
		if idx := strings.Index(image, "@"); idx != -1 {
			image = image[:idx]
		}
		workspace := c.workspacePath()
		if workspace == "" {
			workspace = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Config.ID, c.Status, image, workspace)
	}

	return w.Flush()
}
