package admin_cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/macar-x/cashlenx-server/service/manage_service"
	"github.com/spf13/cobra"
)

var (
	restorePath  string
	forceRestore bool
)

var restoreBackupCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore database from backup",
	Long: `Restore database from a backup file.
WARNING: This will replace all existing data unless --merge is used.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if restorePath == "" {
			return errors.New("backup file path is required")
		}

		if !forceRestore {
			fmt.Println("WARNING: This will replace all existing data!")
			fmt.Print("Are you sure you want to continue? (yes/no): ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "yes" && response != "y" {
				fmt.Println("Restore cancelled")
				return nil
			}
		}

		stats, err := manage_service.RestoreBackup(restorePath)
		if err != nil {
			return err
		}

		fmt.Printf("Database restored successfully from: %s\n", restorePath)
		fmt.Println("\nStatistics:")
		fmt.Printf("  Users: %d success, %d failed\n", stats.Users.Success, stats.Users.Failed)
		fmt.Printf("  Categories: %d success, %d failed\n", stats.Categories.Success, stats.Categories.Failed)
		fmt.Printf("  Cash Flows: %d success, %d failed\n", stats.CashFlows.Success, stats.CashFlows.Failed)
		return nil
	},
}

func init() {
	restoreBackupCmd.Flags().StringVarP(
		&restorePath, "input", "i", "", "backup file path (required)")
	restoreBackupCmd.Flags().BoolVarP(
		&forceRestore, "force", "f", false, "skip confirmation prompt")

	restoreBackupCmd.MarkFlagRequired("input")
}
