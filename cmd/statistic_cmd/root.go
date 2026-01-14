package statistic_cmd

import (
	"github.com/spf13/cobra"
)

var StatisticCmd = &cobra.Command{
	Use:   "statistic",
	Short: "User-specific statistics and analytics (requires authentication)",
	Long: `User-specific statistics and analytics with data isolation.
All commands only operate on your own data.

Available sub-commands:
  export     - Export your data to Excel
  import     - Import data to your account
  summary    - Financial summary (daily/monthly/yearly)
  breakdown  - Category breakdown analysis
  trends     - Spending trends over time
  top        - Top N expenses`,
}

func init() {
	// Register all statistic subcommands
	StatisticCmd.AddCommand(exportCmd)
	StatisticCmd.AddCommand(importCmd)
	StatisticCmd.AddCommand(summaryCmd)
	StatisticCmd.AddCommand(breakdownCmd)
	StatisticCmd.AddCommand(trendsCmd)
	StatisticCmd.AddCommand(topCmd)
}
