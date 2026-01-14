package admin_cmd

import (
	"fmt"
	"time"

	"github.com/macar-x/cashlenx-server/service/manage_service"
	"github.com/spf13/cobra"
)

var backupPath string

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create database backup",
	Long: `Create a backup of all database data.
If no path is specified, creates backup in current directory with timestamp.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if backupPath == "" {
			backupPath = fmt.Sprintf("cashlenx_backup_%s.json", time.Now().Format("20060102_150405"))
		}

		stats, err := manage_service.CreateBackup(backupPath)
		if err != nil {
			return err
		}

		fmt.Printf("Backup created successfully: %s\n", backupPath)
		fmt.Println("\nStatistics:")
		fmt.Printf("  Users: %d success, %d failed\n", stats.Users.Success, stats.Users.Failed)
		fmt.Printf("  Categories: %d success, %d failed\n", stats.Categories.Success, stats.Categories.Failed)
		fmt.Printf("  Cash Flows: %d success, %d failed\n", stats.CashFlows.Success, stats.CashFlows.Failed)
		return nil
	},
}

func init() {
	backupCmd.Flags().StringVarP(
		&backupPath, "output", "o", "", "backup file path (optional, default: cashlenx_backup_TIMESTAMP.json)")
}
