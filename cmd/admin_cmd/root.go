package admin_cmd

import (
	"github.com/spf13/cobra"
)

var adminToken string

var AdminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin-only commands (requires admin privileges)",
	Long: `Admin-only commands that require admin authentication.
These commands are restricted and mirror the /api/admin/* endpoints.

Available sub-commands:
  backup  - Create database backup
  restore - Restore database from backup`,
}

func init() {
	// Add global admin-token flag for dangerous operations
	AdminCmd.PersistentFlags().StringVarP(
		&adminToken, "admin-token", "t", "", "Admin token for dangerous operations")

	// Register all admin commands directly
	AdminCmd.AddCommand(backupCmd)
	AdminCmd.AddCommand(restoreBackupCmd)
}
