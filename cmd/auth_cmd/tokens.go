package auth_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/service/refresh_token_service"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/spf13/cobra"
)

var tokensUserId string

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "List refresh tokens for a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		if tokensUserId == "" {
			var err error
			tokensUserId, err = user_service.GetDefaultAdminUserId()
			if err != nil {
				return err
			}
		}

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
	tokensCmd.Flags().StringVarP(&tokensUserId, "user", "u", "", "user ID (required)")
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
