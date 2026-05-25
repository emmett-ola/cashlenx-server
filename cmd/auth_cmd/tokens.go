package auth_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/cmd/cli_auth"
	"github.com/macar-x/cashlenx-server/service/refresh_token_service"
	"github.com/spf13/cobra"
)

var tokensUserId string

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "List refresh tokens for a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		claims, err := cli_auth.RequireUser()
		if err != nil {
			return err
		}
		if tokensUserId != "" && tokensUserId != claims.UserID {
			return fmt.Errorf("user flag must match logged-in user")
		}
		tokensUserId = claims.UserID

		tokens := refresh_token_service.GetUserRefreshTokens(tokensUserId)
		if len(tokens) == 0 {
			fmt.Println("No refresh tokens found")
			return nil
		}

		fmt.Println("ID                                   | Device              | IP              | Expires At           | Revoked")
		fmt.Println("-------------------------------------|---------------------|-----------------|----------------------|--------")
		for _, token := range tokens {
			fmt.Printf("%-36s | %-19s | %-15s | %-20s | %t\n",
				token.Id,
				truncate(token.DeviceName, 19),
				token.IPAddress,
				token.ExpiresAt.Format("2006-01-02 15:04:05"),
				token.RevokedAt != nil,
			)
		}
		return nil
	},
}

func init() {
	tokensCmd.Flags().StringVarP(&tokensUserId, "user", "u", "", "user ID; must match the logged-in user")
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}
