package cash_flow_cmd

import (
	"path/filepath"
	"testing"
)

func TestCashRootRequiresSubcommand(t *testing.T) {
	if err := CashCmd.RunE(CashCmd, nil); err == nil {
		t.Fatal("expected subcommand error")
	}
}

func TestCashCommandsRequireLoggedInUser(t *testing.T) {
	t.Setenv("CASHLENX_CLI_SESSION_FILE", filepath.Join(t.TempDir(), "session.json"))

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "expense", run: func() error { return expenseCmd.RunE(expenseCmd, nil) }},
		{name: "income", run: func() error { return incomeCmd.RunE(incomeCmd, nil) }},
		{name: "list", run: func() error { return listCmd.RunE(listCmd, nil) }},
		{name: "query", run: func() error { return queryCmd.RunE(queryCmd, nil) }},
		{name: "range", run: func() error { return rangeCmd.RunE(rangeCmd, nil) }},
		{name: "summary", run: func() error { return summaryCmd.RunE(summaryCmd, nil) }},
		{name: "update", run: func() error { return updateCmd.RunE(updateCmd, nil) }},
		{name: "delete", run: func() error { return deleteCmd.RunE(deleteCmd, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cashUserId = ""
			if err := tt.run(); err == nil {
				t.Fatal("expected missing CLI session error")
			}
		})
	}
}
