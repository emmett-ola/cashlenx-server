package category_mapper

import "testing"

func TestCategoryMapperForDatabase(t *testing.T) {
	tests := []struct {
		databaseType string
		want         bool
	}{
		{databaseType: "mongodb", want: true},
		{databaseType: "mysql", want: true},
		{databaseType: "unsupported", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.databaseType, func(t *testing.T) {
			mapper, ok := categoryMapperForDatabase(tt.databaseType)
			if ok != tt.want {
				t.Fatalf("supported = %v, want %v", ok, tt.want)
			}
			if ok && mapper == nil {
				t.Fatal("supported database returned a nil mapper")
			}
		})
	}
}
