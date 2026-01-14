package statistic_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/service/statistic_service"
	"github.com/spf13/cobra"
)

var (
	exportFromDate string
	exportToDate   string
	exportFilePath string
	exportUserId   string // For CLI, will be required parameter or from config
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export your data to Excel",
	Long: `Export your own cash flow data to Excel file for the specified date range.
Only exports data that belongs to your account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Get userId from authentication/config
		// For now, require userId as parameter
		if exportUserId == "" {
			return fmt.Errorf("user ID is required (use --user flag)")
		}

		err := statistic_service.ExportForUser(exportFromDate, exportToDate, exportFilePath, exportUserId)
		if err != nil {
			return fmt.Errorf("export failed: %w", err)
		}

		fmt.Printf("✅ Data exported successfully to: %s\n", exportFilePath)
		fmt.Println("Note: Only your own data has been exported")
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFromDate, "from", "f", "", "from date (include), e.g. 20240101")
	exportCmd.Flags().StringVarP(&exportToDate, "to", "t", "", "to date (include), e.g. 20241231")
	exportCmd.Flags().StringVarP(&exportFilePath, "output", "o", "./export.xlsx", "output path (default: ./export.xlsx)")
	exportCmd.Flags().StringVarP(&exportUserId, "user", "u", "", "user ID (required)")
	exportCmd.MarkFlagRequired("user")
}
