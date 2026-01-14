package admin_cmd

// TODO: This command should be moved to a user-specific statistic module
// with proper data isolation so each user can import their own data.
// Currently in admin for temporary use until statistic module is implemented.

import (
	"github.com/macar-x/cashlenx-server/service/manage_service"
	"github.com/spf13/cobra"
)

var importFilePath string

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data from Excel",
	Long:  `Import cash flow data from Excel file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return manage_service.ImportService(importFilePath)
	},
}

func init() {
	importCmd.Flags().StringVarP(&importFilePath, "input", "i", "", "input path, e.g. ~/export.xlsx")
}
