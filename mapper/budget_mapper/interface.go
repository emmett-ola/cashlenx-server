package budget_mapper

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BudgetMapper interface {
	Insert(model.BudgetEntity) (model.BudgetEntity, error)
	ListByUserAndPeriod(primitive.ObjectID, string) ([]model.BudgetEntity, error)
	GetByIDAndUser(primitive.ObjectID, primitive.ObjectID) (model.BudgetEntity, error)
	GetByScope(primitive.ObjectID, primitive.ObjectID, string) (model.BudgetEntity, error)
	Update(model.BudgetEntity) (model.BudgetEntity, error)
	Delete(primitive.ObjectID, primitive.ObjectID, primitive.ObjectID) (bool, error)
}

var INSTANCE BudgetMapper

func init() {
	switch util.GetConfigByKey("db.type") {
	case "mongodb":
		INSTANCE = MongoMapper{}
	case "mysql":
		INSTANCE = MySQLMapper{}
	default:
		panic("database type not supported")
	}
}
