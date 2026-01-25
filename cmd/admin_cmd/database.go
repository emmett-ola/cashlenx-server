package admin_cmd

import (
	"github.com/spf13/cobra"
)

// DatabaseCmd represents the database command
var DatabaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Database management commands",
	Long:  `Commands for managing the database, including backup and restore operations.`,
}

func init() {
	// Add database command to admin command
	AdminCmd.AddCommand(DatabaseCmd)

	// Add subcommands to database command
	DatabaseCmd.AddCommand(backupCmd)
	DatabaseCmd.AddCommand(restoreBackupCmd)
}
