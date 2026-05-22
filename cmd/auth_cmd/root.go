package auth_cmd

import "github.com/spf13/cobra"

var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticated token commands",
	Long:  `Authenticated token commands that mirror non-open /api/auth/* endpoints.`,
}

func init() {
	AuthCmd.AddCommand(tokensCmd)
}
