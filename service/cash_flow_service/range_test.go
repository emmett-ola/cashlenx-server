package cash_flow_service

import (
	"testing"

	"github.com/macar-x/cashlenx-server/util"
)

func TestQueryByDateRange_Validation(t *testing.T) {
	dbConfigured := util.GetConfigByKey("db.mongodb.url") != "" || util.GetConfigByKey("db.mysql.url") != ""

	tests := []struct {
		name     string
		fromDate string
		toDate   string
		wantErr  bool
	}{
		{
			name:     "Valid date range",
			fromDate: "20241201",
			toDate:   "20241205",
			wantErr:  false,
		},
		{
			name:     "Valid date range with dashes",
			fromDate: "2024-12-01",
			toDate:   "2024-12-05",
			wantErr:  false,
		},
		{
			name:     "Same date",
			fromDate: "20241205",
			toDate:   "20241205",
			wantErr:  false,
		},
		{
			name:     "Invalid range (from after to)",
			fromDate: "20241205",
			toDate:   "20241201",
			wantErr:  true,
		},
		{
			name:     "Empty from date",
			fromDate: "",
			toDate:   "20241205",
			wantErr:  true,
		},
		{
			name:     "Empty to date",
			fromDate: "20241201",
			toDate:   "",
			wantErr:  true,
		},
		{
			name:     "Invalid from date format",
			fromDate: "invalid",
			toDate:   "20241205",
			wantErr:  true,
		},
		{
			name:     "Invalid to date format",
			fromDate: "20241201",
			toDate:   "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !dbConfigured && !tt.wantErr {
				t.Skip("database not configured")
			}
			_, err := QueryByDateRange(tt.fromDate, tt.toDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryByDateRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
