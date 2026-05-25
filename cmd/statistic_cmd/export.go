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
	exportFormat   string
	exportUserId   string // For CLI, will be required parameter or from config
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export your data",
	Long: `Export your own cash flow data for the specified date range.
Only exports data that belongs to your account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureStatisticUser(&exportUserId); err != nil {
			return err
		}

		if exportFormat == "" {
			exportFormat = "xlsx"
		}
		if exportFilePath == "" {
			exportFilePath = "./export." + exportFormat
		}

		var err error
		switch exportFormat {
		case "xlsx":
			err = statistic_service.ExportForUser(exportFromDate, exportToDate, exportFilePath, exportUserId)
		case "csv":
			err = statistic_service.ExportToCSVForUser(exportFromDate, exportToDate, exportFilePath, exportUserId)
		case "pdf":
			err = statistic_service.ExportToPDFForUser(exportFromDate, exportToDate, exportFilePath, exportUserId)
		default:
			return fmt.Errorf("format must be xlsx, csv, or pdf")
		}
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
	exportCmd.Flags().StringVar(&exportFormat, "format", "xlsx", "export format: xlsx, csv, or pdf")
	exportCmd.Flags().StringVarP(&exportFilePath, "output", "o", "", "output path (default: ./export.<format>)")
	exportCmd.Flags().StringVarP(&exportUserId, "user", "u", "", "user ID; must match the logged-in user")
}
