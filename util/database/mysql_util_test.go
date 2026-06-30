package database

import "testing"

func TestBuildMySqlDSNEnablesTimeParsing(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want string
	}{
		{name: "no options", uri: "user:pass@tcp(localhost:3306)", want: "user:pass@tcp(localhost:3306)/cashlenx?parseTime=true"},
		{name: "existing option", uri: "user:pass@tcp(localhost:3306)?charset=utf8mb4", want: "user:pass@tcp(localhost:3306)/cashlenx?charset=utf8mb4&parseTime=true"},
		{name: "parse time corrected", uri: "user:pass@tcp(localhost:3306)?parseTime=false", want: "user:pass@tcp(localhost:3306)/cashlenx?parseTime=true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildMySqlDSN(tt.uri, "cashlenx"); got != tt.want {
				t.Fatalf("buildMySqlDSN() = %q, want %q", got, tt.want)
			}
		})
	}
}
