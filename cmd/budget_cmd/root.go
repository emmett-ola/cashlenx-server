package budget_cmd

import (
	"errors"
	"fmt"

	"github.com/macar-x/cashlenx-server/cmd/cli_auth"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/budget_service"
	"github.com/spf13/cobra"
)

var userID string

var BudgetCmd = &cobra.Command{
	Use:               "budget",
	Short:             "manage monthly category budgets",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return cli_auth.RequireUserID(&userID) },
	RunE:              func(cmd *cobra.Command, args []string) error { return errors.New("must provide a valid sub command") },
}

func init() {
	BudgetCmd.PersistentFlags().StringVarP(&userID, "user", "u", "", "user ID; must match the logged-in user")
	BudgetCmd.AddCommand(createCmd(), listCmd(), getCmd(), updateCmd(), deleteCmd())
}

func printBudget(view model.BudgetView) {
	fmt.Printf("%s  %s  %s  limit %.2f  spent %.2f  remaining %.2f\n", view.Id, view.Period, view.CategoryName, view.LimitAmount, view.SpentAmount, view.Remaining)
}

func budgetFlags(command *cobra.Command, categoryID, period *string, amount *float64) {
	command.Flags().StringVar(categoryID, "category", "", "expense category ID")
	command.Flags().StringVar(period, "period", "", "budget month in YYYY-MM")
	command.Flags().Float64Var(amount, "amount", 0, "monthly limit amount")
	_ = command.MarkFlagRequired("category")
	_ = command.MarkFlagRequired("period")
	_ = command.MarkFlagRequired("amount")
}

func createCmd() *cobra.Command {
	var categoryID, period string
	var amount float64
	command := &cobra.Command{Use: "create", Short: "create a monthly budget", RunE: func(cmd *cobra.Command, args []string) error {
		view, err := budget_service.CreateForUser(model.UpsertBudgetRequest{CategoryId: categoryID, Period: period, LimitAmount: amount}, userID)
		if err != nil {
			return err
		}
		printBudget(view)
		return nil
	}}
	budgetFlags(command, &categoryID, &period, &amount)
	return command
}

func listCmd() *cobra.Command {
	var period string
	command := &cobra.Command{Use: "list", Short: "list budgets for a month", RunE: func(cmd *cobra.Command, args []string) error {
		views, err := budget_service.ListForUser(period, userID)
		if err != nil {
			return err
		}
		for _, view := range views {
			printBudget(view)
		}
		return nil
	}}
	command.Flags().StringVar(&period, "period", "", "budget month in YYYY-MM; defaults to current month")
	return command
}

func getCmd() *cobra.Command {
	return &cobra.Command{Use: "get <id>", Args: cobra.ExactArgs(1), Short: "get a budget", RunE: func(cmd *cobra.Command, args []string) error {
		view, err := budget_service.GetForUser(args[0], userID)
		if err != nil {
			return err
		}
		printBudget(view)
		return nil
	}}
}

func updateCmd() *cobra.Command {
	var categoryID, period string
	var amount float64
	command := &cobra.Command{Use: "update <id>", Args: cobra.ExactArgs(1), Short: "replace a monthly budget", RunE: func(cmd *cobra.Command, args []string) error {
		view, err := budget_service.UpdateForUser(args[0], model.UpsertBudgetRequest{CategoryId: categoryID, Period: period, LimitAmount: amount}, userID)
		if err != nil {
			return err
		}
		printBudget(view)
		return nil
	}}
	budgetFlags(command, &categoryID, &period, &amount)
	return command
}

func deleteCmd() *cobra.Command {
	return &cobra.Command{Use: "delete <id>", Args: cobra.ExactArgs(1), Short: "delete a budget", RunE: func(cmd *cobra.Command, args []string) error { return budget_service.DeleteForUser(args[0], userID) }}
}
