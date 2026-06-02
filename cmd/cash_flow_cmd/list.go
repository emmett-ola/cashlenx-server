package cash_flow_cmd

import (
	"fmt"
	"strings"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/spf13/cobra"
)

var (
	limit                int
	offset               int
	page                 int
	cashType             string
	listCategoryId       string
	listDescription      string
	listExactDescription string
	listFromDate         string
	listToDate           string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all cash_flow records",
	Long: `List all cash flow records with optional filtering and pagination.
Use --type to filter by income/expense, --limit for pagination.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureCashUser(); err != nil {
			return err
		}
		effectiveOffset := offset
		if page > 0 {
			effectiveOffset = (page - 1) * limit
		}

		cashFlowEntityList, totalCount, err := queryCashForUser(
			cashUserId,
			cashType,
			listCategoryId,
			listDescription,
			listExactDescription,
			listFromDate,
			listToDate,
			limit,
			effectiveOffset,
		)
		if err != nil {
			return err
		}

		if len(cashFlowEntityList) == 0 {
			fmt.Println("No cash flows found")
			return nil
		}

		var totalIncome, totalExpense float64
		for index, cashFlowEntity := range cashFlowEntityList {
			fmt.Println("cash_flow", index+effectiveOffset, ":", cashFlowEntity.ToString())
			if strings.EqualFold(cashFlowEntity.CategoryType, model.FlowTypeIncome) {
				totalIncome += cashFlowEntity.Amount
			} else if strings.EqualFold(cashFlowEntity.CategoryType, model.FlowTypeExpense) {
				totalExpense += cashFlowEntity.Amount
			}
		}

		fmt.Printf("\n--- Summary (showing %d of %d records) ---\n", len(cashFlowEntityList), totalCount)
		fmt.Printf("Total Income: %.2f\n", totalIncome)
		fmt.Printf("Total Expense: %.2f\n", totalExpense)
		fmt.Printf("Balance: %.2f\n", totalIncome-totalExpense)

		return nil
	},
}

func init() {
	listCmd.Flags().IntVarP(
		&limit, "limit", "l", 50, "maximum number of records to return")
	listCmd.Flags().IntVarP(
		&offset, "offset", "o", 0, "number of records to skip")
	listCmd.Flags().IntVar(
		&page, "page", 0, "page number; overrides offset when provided")
	listCmd.Flags().StringVarP(
		&cashType, "type", "t", "", "filter by type (income/expense)")
	listCmd.Flags().StringVar(
		&listCategoryId, "category-id", "", "filter by category ID")
	listCmd.Flags().StringVar(
		&listDescription, "description", "", "fuzzy description filter")
	listCmd.Flags().StringVar(
		&listExactDescription, "exact-description", "", "exact description filter")
	listCmd.Flags().StringVar(
		&listFromDate, "from-date", "", "start date filter (YYYY-MM-DD)")
	listCmd.Flags().StringVar(
		&listToDate, "to-date", "", "end date filter (YYYY-MM-DD)")

	CashCmd.AddCommand(listCmd)
}
