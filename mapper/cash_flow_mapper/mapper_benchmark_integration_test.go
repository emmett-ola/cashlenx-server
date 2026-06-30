//go:build integration

package cash_flow_mapper

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const benchmarkInsertBatchSize = 500

var cashFlowMapperBenchmarkResult []model.CashFlowEntity

func BenchmarkCashFlowMapperDateRangeAndUser(b *testing.B) {
	benchmarkCashFlowMapperQueries(b, func(mapper CashFlowMapper, fixture cashFlowMapperBenchmarkFixture) ([]model.CashFlowEntity, error) {
		return mapper.GetCashFlowsByDateRangeAndUser(fixture.from, fixture.to, fixture.userID), nil
	}, func(fixture cashFlowMapperBenchmarkFixture) int {
		return fixture.size
	})
}

func BenchmarkCashFlowMapperFiltered(b *testing.B) {
	benchmarkCashFlowMapperQueries(b, func(mapper CashFlowMapper, fixture cashFlowMapperBenchmarkFixture) ([]model.CashFlowEntity, error) {
		return mapper.GetCashFlowsByFilter(model.CashFlowFilter{
			UserId:           fixture.userID,
			CategoryId:       fixture.categoryIDs[0].Hex(),
			ExactDescription: "benchmark-match",
			FromDate:         fixture.from,
			ToDate:           fixture.to,
			Limit:            100,
		})
	}, func(fixture cashFlowMapperBenchmarkFixture) int {
		expected := (fixture.size + len(fixture.categoryIDs) - 1) / len(fixture.categoryIDs)
		if expected > 100 {
			return 100
		}
		return expected
	})
}

type cashFlowMapperBenchmarkFixture struct {
	size        int
	userID      primitive.ObjectID
	categoryIDs []primitive.ObjectID
	from        time.Time
	to          time.Time
}

func benchmarkCashFlowMapperQueries(
	b *testing.B,
	query func(CashFlowMapper, cashFlowMapperBenchmarkFixture) ([]model.CashFlowEntity, error),
	expectedCount func(cashFlowMapperBenchmarkFixture) int,
) {
	mapper, databaseType := cashFlowBenchmarkMapper(b)
	defer closeCashFlowBenchmarkDatabase(databaseType)

	for _, fixtureSize := range []int{1_000, 10_000} {
		b.Run(fmt.Sprintf("%s/rows_%d", databaseType, fixtureSize), func(b *testing.B) {
			fixture := seedCashFlowBenchmarkFixture(b, mapper, fixtureSize)
			b.Cleanup(func() {
				if _, err := mapper.DeleteAllCashFlowsByUser(fixture.userID); err != nil {
					b.Errorf("clean benchmark fixture: %v", err)
				}
			})

			results, err := query(mapper, fixture)
			if err != nil {
				b.Fatalf("warm-up query: %v", err)
			}
			if len(results) != expectedCount(fixture) {
				b.Fatalf("warm-up returned %d rows, want %d", len(results), expectedCount(fixture))
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results, err = query(mapper, fixture)
				if err != nil {
					b.Fatal(err)
				}
				cashFlowMapperBenchmarkResult = results
			}
		})
	}
}

func cashFlowBenchmarkMapper(b *testing.B) (CashFlowMapper, string) {
	b.Helper()
	if os.Getenv("CASHLENX_BENCHMARKS") != "1" {
		b.Skip("set CASHLENX_BENCHMARKS=1 to allow disposable database fixtures")
	}

	databaseType := util.GetConfigByKey("db.type")
	switch databaseType {
	case "mongodb":
		if util.GetConfigByKey("db.mongodb.url") == "" {
			b.Skip("MONGO_DB_URI is not set")
		}
		if err := database.InitMongoDbConnection(); err != nil {
			b.Fatalf("connect to MongoDB: %v", err)
		}
		return CashFlowMongoDbMapper{}, databaseType
	case "mysql":
		if util.GetConfigByKey("db.mysql.url") == "" {
			b.Skip("MYSQL_DB_URI is not set")
		}
		if err := database.GetMySqlConnection().Ping(); err != nil {
			b.Fatalf("connect to MySQL: %v", err)
		}
		return CashFlowMySqlMapper{}, databaseType
	default:
		b.Fatalf("unsupported DB_TYPE %q", databaseType)
		return nil, ""
	}
}

func closeCashFlowBenchmarkDatabase(databaseType string) {
	if databaseType == "mongodb" {
		database.ShutdownMongoDbConnection()
		return
	}
	database.CloseMySqlConnection()
}

func seedCashFlowBenchmarkFixture(b *testing.B, mapper CashFlowMapper, fixtureSize int) cashFlowMapperBenchmarkFixture {
	b.Helper()
	fixture := cashFlowMapperBenchmarkFixture{
		size:        fixtureSize,
		userID:      primitive.NewObjectID(),
		categoryIDs: make([]primitive.ObjectID, 10),
		from:        time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC),
		to:          time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
	}
	for i := range fixture.categoryIDs {
		fixture.categoryIDs[i] = primitive.NewObjectID()
	}

	now := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	entities := make([]model.CashFlowEntity, fixtureSize)
	for i := range entities {
		description := "benchmark-other"
		if i%2 == 0 {
			description = "benchmark-match"
		}
		entities[i] = model.CashFlowEntity{
			Id:            primitive.NewObjectID(),
			BelongsUserId: fixture.userID,
			CategoryId:    fixture.categoryIDs[i%len(fixture.categoryIDs)],
			BelongsDate:   time.Date(2026, time.May, i%31+1, 0, 0, 0, 0, time.UTC),
			Amount:        float64(i%500 + 1),
			Description:   description,
			Remark:        "integration benchmark fixture",
			BaseEntity: model.BaseEntity{
				CreateUserId: fixture.userID,
				CreateTime:   now,
				UpdateUserId: fixture.userID,
				UpdateTime:   now,
			},
		}
	}

	for start := 0; start < len(entities); start += benchmarkInsertBatchSize {
		end := start + benchmarkInsertBatchSize
		if end > len(entities) {
			end = len(entities)
		}
		if _, err := mapper.BulkInsertCashFlows(entities[start:end]); err != nil {
			_, _ = mapper.DeleteAllCashFlowsByUser(fixture.userID)
			b.Fatalf("seed %d-row benchmark fixture: %v", fixtureSize, err)
		}
	}

	return fixture
}
