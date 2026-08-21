package budget_mapper

import (
	"database/sql"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MySQLMapper struct{}

func (MySQLMapper) Insert(entity model.BudgetEntity) (model.BudgetEntity, error) {
	if entity.Id.IsZero() {
		entity.Id = primitive.NewObjectID()
	}
	_, err := database.GetMySqlConnection().Exec(`INSERT INTO budgets (id, belongs_user_id, category_id, period, limit_amount, create_user_id, create_time, update_user_id, update_time, is_delete) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, FALSE)`, entity.Id.Hex(), entity.BelongsUserId.Hex(), entity.CategoryId.Hex(), entity.Period, entity.LimitAmount, entity.CreateUserId.Hex(), entity.CreateTime, entity.UpdateUserId.Hex(), entity.UpdateTime)
	return entity, err
}

func (MySQLMapper) ListByUserAndPeriod(userID primitive.ObjectID, period string) ([]model.BudgetEntity, error) {
	rows, err := database.GetMySqlConnection().Query(`SELECT id, belongs_user_id, category_id, period, limit_amount, create_user_id, create_time, update_user_id, update_time FROM budgets WHERE belongs_user_id = ? AND period = ? AND is_delete = FALSE ORDER BY create_time`, userID.Hex(), period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.BudgetEntity{}
	for rows.Next() {
		entity, err := scanBudget(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, entity)
	}
	return items, rows.Err()
}

type budgetScanner interface{ Scan(...interface{}) error }

func scanBudget(row budgetScanner) (model.BudgetEntity, error) {
	var entity model.BudgetEntity
	var id, userID, categoryID, createUserID, updateUserID string
	err := row.Scan(&id, &userID, &categoryID, &entity.Period, &entity.LimitAmount, &createUserID, &entity.CreateTime, &updateUserID, &entity.UpdateTime)
	if err != nil {
		return model.BudgetEntity{}, err
	}
	entity.Id, _ = primitive.ObjectIDFromHex(id)
	entity.BelongsUserId, _ = primitive.ObjectIDFromHex(userID)
	entity.CategoryId, _ = primitive.ObjectIDFromHex(categoryID)
	entity.CreateUserId, _ = primitive.ObjectIDFromHex(createUserID)
	entity.UpdateUserId, _ = primitive.ObjectIDFromHex(updateUserID)
	return entity, nil
}

func (MySQLMapper) GetByIDAndUser(id, userID primitive.ObjectID) (model.BudgetEntity, error) {
	row := database.GetMySqlConnection().QueryRow(`SELECT id, belongs_user_id, category_id, period, limit_amount, create_user_id, create_time, update_user_id, update_time FROM budgets WHERE id = ? AND belongs_user_id = ? AND is_delete = FALSE`, id.Hex(), userID.Hex())
	entity, err := scanBudget(row)
	if err == sql.ErrNoRows {
		return model.BudgetEntity{}, nil
	}
	return entity, err
}

func (MySQLMapper) GetByScope(userID, categoryID primitive.ObjectID, period string) (model.BudgetEntity, error) {
	row := database.GetMySqlConnection().QueryRow(`SELECT id, belongs_user_id, category_id, period, limit_amount, create_user_id, create_time, update_user_id, update_time FROM budgets WHERE belongs_user_id = ? AND category_id = ? AND period = ? AND is_delete = FALSE`, userID.Hex(), categoryID.Hex(), period)
	entity, err := scanBudget(row)
	if err == sql.ErrNoRows {
		return model.BudgetEntity{}, nil
	}
	return entity, err
}

func (MySQLMapper) Update(entity model.BudgetEntity) (model.BudgetEntity, error) {
	result, err := database.GetMySqlConnection().Exec(`UPDATE budgets SET category_id = ?, period = ?, limit_amount = ?, update_user_id = ?, update_time = ? WHERE id = ? AND belongs_user_id = ? AND is_delete = FALSE`, entity.CategoryId.Hex(), entity.Period, entity.LimitAmount, entity.UpdateUserId.Hex(), entity.UpdateTime, entity.Id.Hex(), entity.BelongsUserId.Hex())
	if err != nil {
		return model.BudgetEntity{}, err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return model.BudgetEntity{}, err
	}
	return entity, nil
}

func (MySQLMapper) Delete(id, userID, actorID primitive.ObjectID) (bool, error) {
	result, err := database.GetMySqlConnection().Exec(`UPDATE budgets SET is_delete = TRUE, delete_user_id = ?, delete_time = ? WHERE id = ? AND belongs_user_id = ? AND is_delete = FALSE`, actorID.Hex(), time.Now().UTC(), id.Hex(), userID.Hex())
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return err == nil && count == 1, err
}
