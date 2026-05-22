package user_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/spf13/cobra"
)

var (
	configLanguage string
	configCurrency string
	configTheme    string
)

var configurationCmd = &cobra.Command{
	Use:   "configuration",
	Short: "Get or update user configuration",
}

var configurationGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := user_service.GetConfigurationService(userId)
		if err != nil {
			return err
		}
		printConfiguration(config)
		return nil
	},
}

var configurationUpsertCmd = &cobra.Command{
	Use:   "upsert",
	Short: "Create or update configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		request := model.UserConfigurationRequest{}
		if cmd.Flags().Changed("display-language") {
			request.DisplayLanguage = &configLanguage
		}
		if cmd.Flags().Changed("currency-code") {
			request.CurrencyCode = &configCurrency
		}
		if cmd.Flags().Changed("active-theme-color") {
			request.ActiveThemeColor = &configTheme
		}

		config, err := user_service.UpsertConfigurationService(userId, request)
		if err != nil {
			return err
		}
		fmt.Println("Configuration saved successfully")
		printConfiguration(config)
		return nil
	},
}

func printConfiguration(config model.UserConfigurationEntity) {
	fmt.Printf("ID:                  %s\n", config.Id.Hex())
	fmt.Printf("User ID:             %s\n", config.BelongsUserId.Hex())
	fmt.Printf("Display Language:    %s\n", config.DisplayLanguage)
	fmt.Printf("Currency Code:       %s\n", config.CurrencyCode)
	fmt.Printf("Active Theme Color:  %s\n", config.ActiveThemeColor)
}

func init() {
	configurationCmd.AddCommand(configurationGetCmd)
	configurationCmd.AddCommand(configurationUpsertCmd)

	configurationUpsertCmd.Flags().StringVar(&configLanguage, "display-language", "", "display language")
	configurationUpsertCmd.Flags().StringVar(&configCurrency, "currency-code", "", "3-letter ISO 4217 currency code")
	configurationUpsertCmd.Flags().StringVar(&configTheme, "active-theme-color", "", "active theme color")
}
