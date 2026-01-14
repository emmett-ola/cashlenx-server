package open_cmd

import (
	"github.com/macar-x/cashlenx-server/controller"
	"github.com/spf13/cobra"
)

var port int32

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the API server",
	Long:  `Start the CashLenX API server on the specified port (default: 8080)`,
	Run: func(cmd *cobra.Command, args []string) {
		controller.StartServer(port)
	},
}

func init() {
	startCmd.Flags().Int32VarP(
		&port, "port", "p", 8080, "API server port (default: 8080)")
}
