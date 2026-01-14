package admin_cmd

// TODO: This command should be moved to a user-specific statistic module
// with proper data isolation so each user can export their own data.
// Currently in admin for temporary use until statistic module is implemented.

import (
	"github.com/macar-x/cashlenx-server/service/manage_service"
	"github.com/spf13/cobra"
)

var (
	fromDate string
	toDate   string
	filePath string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data to Excel",
	Long:  `Export cash flow data to Excel file for the specified date range.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return manage_service.ExportService(fromDate, toDate, filePath)
	},
}

func init() {
	exportCmd.Flags().StringVarP(&fromDate, "from", "f", "", "from date (include), e.g. 19700101")
	exportCmd.Flags().StringVarP(&toDate, "to", "t", "", "to date (include), e.g. 19700101")
	exportCmd.Flags().StringVarP(&filePath, "output", "o", "", "output path, default ./export.xlsx")
}
