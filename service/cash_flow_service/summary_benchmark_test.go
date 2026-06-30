package cash_flow_service

import (
	"fmt"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var cashFlowSummaryBenchmarkResult *Summary

type cashFlowSummaryBenchmarkMapper struct {
	*cashFlowMapperFake
	fixture []model.CashFlowEntity
}

func (mapper *cashFlowSummaryBenchmarkMapper) GetCashFlowsByDateRangeAndUser(_, _ time.Time, _ primitive.ObjectID) []model.CashFlowEntity {
	return mapper.fixture
}

func BenchmarkCashFlowServiceGetSummaryForUser(b *testing.B) {
	for _, transactionCount := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("transactions_%d", transactionCount), func(b *testing.B) {
			service, userID := newCashFlowSummaryBenchmarkService(transactionCount)

			if _, err := service.GetSummaryForUser("monthly", "2026-05", userID.Hex()); err != nil {
				b.Fatalf("warm-up summary failed: %v", err)
			}

			b.ReportAllocs()
			b.ReportMetric(float64(transactionCount), "transactions/op")
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result, err := service.GetSummaryForUser("monthly", "2026-05", userID.Hex())
				if err != nil {
					b.Fatal(err)
				}
				cashFlowSummaryBenchmarkResult = result
			}
		})
	}
}

func newCashFlowSummaryBenchmarkService(transactionCount int) (*CashFlowService, primitive.ObjectID) {
	userID := cashFlowBenchmarkObjectID(1)
	categoryMapper := newCashFlowCategoryMapperFake()
	categoryIDs := make([]primitive.ObjectID, 10)

	for i := range categoryIDs {
		categoryID := cashFlowBenchmarkObjectID(byte(i + 2))
		categoryType := model.FlowTypeExpense
		if i >= 8 {
			categoryType = model.FlowTypeIncome
		}
		categoryIDs[i] = categoryID
		categoryMapper.categories[categoryID] = model.CategoryEntity{
			Id:            categoryID,
			BelongsUserId: userID,
			Name:          fmt.Sprintf("Category %02d", i),
			Type:          categoryType,
		}
	}

	fixture := make([]model.CashFlowEntity, transactionCount)
	for i := range fixture {
		fixture[i] = model.CashFlowEntity{
			Id:            cashFlowBenchmarkObjectID(byte(i%240 + 12)),
			BelongsUserId: userID,
			CategoryId:    categoryIDs[i%len(categoryIDs)],
			BelongsDate:   time.Date(2026, time.May, i%31+1, 0, 0, 0, 0, time.UTC),
			Amount:        float64(i%500+1) / 10,
		}
	}

	cashMapper := &cashFlowSummaryBenchmarkMapper{
		cashFlowMapperFake: newCashFlowMapperFake(),
		fixture:            fixture,
	}
	return NewCashFlowService(cashMapper, categoryMapper), userID
}

func cashFlowBenchmarkObjectID(seed byte) primitive.ObjectID {
	var id primitive.ObjectID
	id[len(id)-1] = seed
	return id
}
