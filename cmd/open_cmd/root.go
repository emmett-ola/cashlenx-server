package open_cmd

import (
	"github.com/spf13/cobra"
)

var OpenCmd = &cobra.Command{
	Use:   "open",
	Short: "Public commands (no authentication required)",
	Long: `Public commands that don't require authentication.
These commands are available to all users and mirror the /api/open/* endpoints.

Available sub-commands:
  health  - Check system health
  version - Show version information
  start   - Start the API server`,
}

func init() {
	OpenCmd.AddCommand(healthCmd)
	OpenCmd.AddCommand(versionCmd)
	OpenCmd.AddCommand(startCmd)
}
