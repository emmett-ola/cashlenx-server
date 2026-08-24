package open_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/util"
	"github.com/spf13/cobra"
)

var port int32

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the API server",
	Long:  `Start the CashLenX API server on the specified port (default: 8080)`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := util.ValidateConfiguredTimezone(); err != nil {
			return fmt.Errorf("TIMEZONE %w", err)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		startPort := port
		if !cmd.Flags().Changed("port") {
			startPort = int32(util.GetConfigInt("server.port", 8080))
		}
		startOpenServer(startPort)
	},
}

func init() {
	startCmd.Flags().Int32VarP(
		&port, "port", "p", 8080, "API server port (default: 8080)")
}
