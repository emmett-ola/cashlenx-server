package cash_flow_cmd

import (
	"errors"
	"fmt"

	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "query for cash_flow data",
	RunE: func(cmd *cobra.Command, args []string) error {
		plainId, _ := cmd.Flags().GetString("id")
		belongsDate, _ := cmd.Flags().GetString("date")
		descriptionExact, _ := cmd.Flags().GetString("exact")
		descriptionFuzzy, _ := cmd.Flags().GetString("fuzzy")

		// Valid params through command.
		if cash_flow_service.IsQueryFieldsConflicted(plainId, belongsDate, descriptionExact, descriptionFuzzy) {
			return errors.New("should have one and only one query type")
		}

		// if id is not empty, use it for query.
		if plainId != "" {
			cashFlowEntity, err := cash_flow_service.QueryById(plainId)
			if err != nil {
				return err
			}
			fmt.Println("cash_flow ", 0, ": ", cashFlowEntity.ToString())
			return nil
		}

		// else if date is not empty, use it for query.
		if belongsDate != "" {
			cashFlowEntityList, err := cash_flow_service.QueryByDate(belongsDate)
			if err != nil {
				return err
			}
			if len(cashFlowEntityList) == 0 {
				fmt.Println("the day's flow is empty")
				return nil
			}
			for index, cashFlowEntity := range cashFlowEntityList {
				fmt.Println("cash_flow ", index, ": ", cashFlowEntity.ToString())
			}
			return nil
		}

		// else if exact_desc is not empty, use it for query.
		if descriptionExact != "" {
			cashFlowEntityList, err := cash_flow_service.QueryByExactDescription(descriptionExact)
			if err != nil {
				return err
			}
			if len(cashFlowEntityList) == 0 {
				fmt.Println("no matched cash_flows")
				return nil
			}

			for index, cashFlowEntity := range cashFlowEntityList {
				fmt.Println("cash_flow ", index, ": ", cashFlowEntity.ToString())
			}
		}

		// else if fuzzy_desc is not empty, use it for query.
		if descriptionFuzzy != "" {
			cashFlowEntityList, err := cash_flow_service.QueryByFuzzyDescription(descriptionFuzzy)
			if err != nil {
				return err
			}
			if len(cashFlowEntityList) == 0 {
				fmt.Println("no matched cash_flows")
				return nil
			}

			for index, cashFlowEntity := range cashFlowEntityList {
				fmt.Println("cash_flow ", index, ": ", cashFlowEntity.ToString())
			}
		}

		return errors.New("not supported query type")
	},
}

func init() {
	queryCmd.Flags().StringP(
		"id", "i", "", "query by id")
	queryCmd.Flags().StringP(
		"date", "b", "", "query by belongs-date")
	queryCmd.Flags().StringP(
		"exact", "e", "", "query by exact-description")
	queryCmd.Flags().StringP(
		"fuzzy", "f", "", "query by fuzzy-description")
	CashCmd.AddCommand(queryCmd)
}
