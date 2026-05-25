package admin_cmd

import (
	"github.com/macar-x/cashlenx-server/cmd/cli_auth"
	"github.com/spf13/cobra"
)

var AdminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin-only commands (requires admin privileges)",
	Long: `Admin-only commands that require admin authentication.
These commands are restricted and mirror the /api/admin/* endpoints.

Available sub-commands:
  user    - Manage users
  backup  - Create database backup
  restore - Restore database from backup`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		_, err := cli_auth.RequireAdmin()
		return err
	},
}

func init() {
	// Register all admin commands directly
	// AdminCmd.AddCommand(backupCmd) // Moved to database subcommand
	// AdminCmd.AddCommand(restoreBackupCmd) // Moved to database subcommand
}
