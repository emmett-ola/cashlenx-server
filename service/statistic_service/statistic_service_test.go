package statistic_service

import (
	"testing"
	"time"
)

func TestGetDateRange(t *testing.T) {
	baseDate := time.Date(2026, time.May, 12, 15, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		period    string
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "daily",
			period:    "daily",
			wantStart: time.Date(2026, time.May, 12, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, time.May, 13, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "monthly",
			period:    "monthly",
			wantStart: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "yearly",
			period:    "yearly",
			wantStart: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := getDateRange(tt.period, baseDate)
			if !gotStart.Equal(tt.wantStart) {
				t.Fatalf("start = %s, want %s", gotStart, tt.wantStart)
			}
			if !gotEnd.Equal(tt.wantEnd) {
				t.Fatalf("end = %s, want %s", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestAnalyzeTrendDirection(t *testing.T) {
	tests := []struct {
		name             string
		dataPoints       []TrendDataPoint
		expenseCount     int
		totalExpense     float64
		wantIncomeTrend  string
		wantExpenseTrend string
		wantAverage      float64
	}{
		{
			name: "increasing income and expense",
			dataPoints: []TrendDataPoint{
				{Income: 100, Expense: 50},
				{Income: 100, Expense: 50},
				{Income: 130, Expense: 70},
				{Income: 140, Expense: 80},
			},
			expenseCount:     4,
			totalExpense:     250,
			wantIncomeTrend:  "increasing",
			wantExpenseTrend: "increasing",
			wantAverage:      62.5,
		},
		{
			name: "decreasing income and expense",
			dataPoints: []TrendDataPoint{
				{Income: 150, Expense: 80},
				{Income: 140, Expense: 70},
				{Income: 100, Expense: 50},
				{Income: 100, Expense: 50},
			},
			expenseCount:     4,
			totalExpense:     250,
			wantIncomeTrend:  "decreasing",
			wantExpenseTrend: "decreasing",
			wantAverage:      62.5,
		},
		{
			name: "stable with single point",
			dataPoints: []TrendDataPoint{
				{Income: 100, Expense: 50},
			},
			expenseCount:     1,
			totalExpense:     50,
			wantIncomeTrend:  "stable",
			wantExpenseTrend: "stable",
			wantAverage:      50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeTrendDirection(tt.dataPoints, tt.expenseCount, tt.totalExpense)
			if got.IncomeTrend != tt.wantIncomeTrend {
				t.Fatalf("IncomeTrend = %q, want %q", got.IncomeTrend, tt.wantIncomeTrend)
			}
			if got.ExpenseTrend != tt.wantExpenseTrend {
				t.Fatalf("ExpenseTrend = %q, want %q", got.ExpenseTrend, tt.wantExpenseTrend)
			}
			if got.AverageMonthlyExpense != tt.wantAverage {
				t.Fatalf("AverageMonthlyExpense = %v, want %v", got.AverageMonthlyExpense, tt.wantAverage)
			}
		})
	}
}
