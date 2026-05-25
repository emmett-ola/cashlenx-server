package statistic_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/service/statistic_service"
	"github.com/spf13/cobra"
)

var (
	importFilePath string
	importUserId   string // For CLI, will be required parameter or from config
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data to your account",
	Long: `Import cash flow data from Excel file to your account.
All imported records will be associated with your user account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureStatisticUser(&importUserId); err != nil {
			return err
		}

		err := statistic_service.ImportForUser(importFilePath, importUserId)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}

		fmt.Printf("✅ Data imported successfully from: %s\n", importFilePath)
		fmt.Println("Note: All records have been associated with your account")
		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&importFilePath, "input", "i", "", "input path, e.g. ~/export.xlsx (required)")
	importCmd.Flags().StringVarP(&importUserId, "user", "u", "", "user ID; must match the logged-in user")
	importCmd.MarkFlagRequired("input")
}
