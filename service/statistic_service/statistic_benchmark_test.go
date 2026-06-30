package statistic_service

import (
	"fmt"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	statisticSummaryBenchmarkResult   *Summary
	statisticDashboardBenchmarkResult *DashboardOverview
)

type statisticBenchmarkCashFlowMapper struct {
	*statisticCashFlowMapperFake
}

func (mapper *statisticBenchmarkCashFlowMapper) GetCashFlowsByDateRangeAndUser(_, _ time.Time, _ primitive.ObjectID) []model.CashFlowEntity {
	return mapper.flows
}

func BenchmarkStatisticServiceGetSummaryForUser(b *testing.B) {
	benchmarkStatisticService(b, func(b *testing.B, service *StatisticService, userID primitive.ObjectID) {
		result, err := service.GetSummaryForUser("monthly", "202605", userID.Hex())
		if err != nil {
			b.Fatal(err)
		}
		statisticSummaryBenchmarkResult = result
	})
}

func BenchmarkStatisticServiceGetDashboardOverviewForUser(b *testing.B) {
	benchmarkStatisticService(b, func(b *testing.B, service *StatisticService, userID primitive.ObjectID) {
		result, err := service.GetDashboardOverviewForUser("monthly", "202605", userID.Hex())
		if err != nil {
			b.Fatal(err)
		}
		statisticDashboardBenchmarkResult = result
	})
}

func benchmarkStatisticService(b *testing.B, run func(*testing.B, *StatisticService, primitive.ObjectID)) {
	for _, transactionCount := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("transactions_%d", transactionCount), func(b *testing.B) {
			service, userID := newStatisticBenchmarkService(transactionCount)
			run(b, service, userID)

			b.ReportAllocs()
			b.ReportMetric(float64(transactionCount), "transactions/op")
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				run(b, service, userID)
			}
		})
	}
}

func newStatisticBenchmarkService(transactionCount int) (*StatisticService, primitive.ObjectID) {
	userID := statisticBenchmarkObjectID(1)
	categoryIDs := make([]primitive.ObjectID, 10)
	categories := make(map[primitive.ObjectID]model.CategoryEntity, len(categoryIDs))

	for i := range categoryIDs {
		categoryID := statisticBenchmarkObjectID(byte(i + 2))
		categoryType := model.FlowTypeExpense
		if i >= 8 {
			categoryType = model.FlowTypeIncome
		}
		categoryIDs[i] = categoryID
		categories[categoryID] = model.CategoryEntity{
			Id:            categoryID,
			BelongsUserId: userID,
			Name:          fmt.Sprintf("Category %02d", i),
			Type:          categoryType,
		}
	}

	fixture := make([]model.CashFlowEntity, transactionCount)
	for i := range fixture {
		fixture[i] = model.CashFlowEntity{
			Id:            statisticBenchmarkObjectID(byte(i%240 + 12)),
			BelongsUserId: userID,
			CategoryId:    categoryIDs[i%len(categoryIDs)],
			BelongsDate:   time.Date(2026, time.May, i%31+1, 0, 0, 0, 0, time.UTC),
			Amount:        float64(i%500+1) / 10,
		}
	}

	cashMapper := &statisticBenchmarkCashFlowMapper{
		statisticCashFlowMapperFake: &statisticCashFlowMapperFake{flows: fixture},
	}
	categoryMapper := &statisticCategoryMapperFake{categories: categories}
	return NewStatisticService(cashMapper, categoryMapper), userID
}

func statisticBenchmarkObjectID(seed byte) primitive.ObjectID {
	var id primitive.ObjectID
	id[len(id)-1] = seed
	return id
}
