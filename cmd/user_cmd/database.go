package user_cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/macar-x/cashlenx-server/service/manage_service"
	"github.com/spf13/cobra"
)

var (
	userBackupPath  string
	userRestorePath string
)

var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "User-scoped backup and restore commands",
}

var databaseBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Export current user's data to JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		if userBackupPath == "" {
			userBackupPath = fmt.Sprintf("cashlenx_user_%s_backup_%s.json", userId, time.Now().Format("20060102_150405"))
		}
		stats, err := exportUserData(userId, userBackupPath)
		if err != nil {
			return err
		}
		fmt.Printf("User data exported successfully: %s\n", userBackupPath)
		printStats(stats)
		return nil
	},
}

var databaseRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Import current user's JSON backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		if userRestorePath == "" {
			return errors.New("backup file path is required")
		}
		stats, err := importUserData(userId, userRestorePath)
		if err != nil {
			return err
		}
		fmt.Printf("User data imported successfully from: %s\n", userRestorePath)
		printStats(stats)
		return nil
	},
}

func printStats(stats manage_service.OperationStats) {
	fmt.Println("\nStatistics:")
	fmt.Printf("  Users: %d success, %d failed\n", stats.Users.Success, stats.Users.Failed)
	fmt.Printf("  User Configurations: %d success, %d failed\n", stats.UserConfigs.Success, stats.UserConfigs.Failed)
	fmt.Printf("  Categories: %d success, %d failed\n", stats.Categories.Success, stats.Categories.Failed)
	fmt.Printf("  Cash Flows: %d success, %d failed\n", stats.CashFlows.Success, stats.CashFlows.Failed)
}

func init() {
	databaseCmd.AddCommand(databaseBackupCmd)
	databaseCmd.AddCommand(databaseRestoreCmd)

	databaseBackupCmd.Flags().StringVarP(&userBackupPath, "output", "o", "", "backup file path")
	databaseRestoreCmd.Flags().StringVarP(&userRestorePath, "input", "i", "", "backup file path (required)")
	databaseRestoreCmd.MarkFlagRequired("input")
}
