package user_cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var accountForce bool

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Delete current user's account",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !accountForce {
			fmt.Print("This will delete the user account. Continue? (yes/no): ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "yes" && response != "y" {
				fmt.Println("Account deletion cancelled")
				return nil
			}
		}
		if err := deleteUserAccount(userId); err != nil {
			return err
		}
		fmt.Println("Account deleted successfully")
		return nil
	},
}

func init() {
	accountCmd.Flags().BoolVarP(&accountForce, "force", "f", false, "skip confirmation prompt")
}
