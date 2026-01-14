package cash_flow_mapper

import (
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var INSTANCE CashFlowMapper

type CashFlowMapper interface {
	// Legacy methods without user filtering (for admin/system operations)
	GetCashFlowByObjectId(plainId string) model.CashFlowEntity
	GetCashFlowsByObjectIdArray(plainIdList []string) []model.CashFlowEntity
	GetCashFlowsByBelongsDate(belongsDate time.Time) []model.CashFlowEntity
	GetCashFlowsByDateRange(from, to time.Time) []model.CashFlowEntity
	GetCashFlowsByCategoryId(categoryPlainId string) []model.CashFlowEntity
	GetCashFlowsByExactDesc(description string) []model.CashFlowEntity
	GetCashFlowsByFuzzyDesc(description string) []model.CashFlowEntity
	CountCashFLowsByCategoryId(categoryPlainId string) int64
	InsertCashFlowByEntity(newEntity model.CashFlowEntity) string
	BulkInsertCashFlows(entities []model.CashFlowEntity) ([]string, error)
	UpdateCashFlowByEntity(plainId string, updatedEntity model.CashFlowEntity) model.CashFlowEntity
	GetAllCashFlows(limit, offset int) []model.CashFlowEntity
	CountAllCashFlows() int64
	DeleteCashFlowByObjectId(plainId string) model.CashFlowEntity
	DeleteCashFlowByBelongsDate(belongsDate time.Time) []model.CashFlowEntity
	TruncateCashFlows() error

	// User-specific methods for data isolation
	GetCashFlowByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CashFlowEntity
	GetCashFlowsByBelongsDateAndUser(belongsDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity
	GetCashFlowsByDateRangeAndUser(from, to time.Time, userId primitive.ObjectID) []model.CashFlowEntity
	GetCashFlowsByCategoryIdAndUser(categoryPlainId string, userId primitive.ObjectID) []model.CashFlowEntity
	GetAllCashFlowsByUser(userId primitive.ObjectID, limit, offset int) []model.CashFlowEntity
	CountAllCashFlowsByUser(userId primitive.ObjectID) int64
	DeleteCashFlowByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CashFlowEntity
	DeleteCashFlowsByBelongsDateAndUser(belongsDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity
}

func init() {
	switch util.GetConfigByKey("db.type") {
	case "mongodb":
		INSTANCE = CashFlowMongoDbMapper{}
	case "mysql":
		INSTANCE = CashFlowMySqlMapper{}
	default:
		panic("database type not supported")
	}
}
