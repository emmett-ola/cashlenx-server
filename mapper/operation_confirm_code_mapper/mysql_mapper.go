package operation_confirm_code_mapper

import (
	"database/sql"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OperationConfirmCodeMySQLMapper implements OperationConfirmCodeMapper for MySQL.
type OperationConfirmCodeMySQLMapper struct{}

func (mapper *OperationConfirmCodeMySQLMapper) CreateCode(code model.OperationConfirmCode) error {
	query := `INSERT INTO ` + database.OperationConfirmCodeTableName + `
		(id, user_id, token, operation_type, payload, expires_at, used_at,
		 create_user_id, create_time, update_user_id, update_time, delete_user_id, delete_time, is_delete)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	_, err := connection.Exec(
		query,
		code.Id,
		code.UserId,
		code.Code,
		code.OperationType,
		nullableString(code.Payload),
		code.ExpiresTime,
		code.UsedTime,
		objectIDHexOrNil(code.CreateUserId),
		code.CreateTime,
		objectIDHexOrNil(code.UpdateUserId),
		code.UpdateTime,
		objectIDPointerHexOrNil(code.DeleteUserId),
		code.DeleteTime,
		code.IsDelete,
	)
	return err
}

func (mapper *OperationConfirmCodeMySQLMapper) GetCodeByToken(token string) model.OperationConfirmCode {
	query := `SELECT id, user_id, token, operation_type, payload, expires_at, used_at,
		create_user_id, create_time, update_user_id, update_time, delete_user_id, delete_time, is_delete
		FROM ` + database.OperationConfirmCodeTableName + `
		WHERE token = ? AND is_delete = FALSE`

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	var entity model.OperationConfirmCode
	var payload sql.NullString
	var usedAt sql.NullTime
	var createUserId, updateUserId string
	var deleteUserId sql.NullString
	var deleteTime sql.NullTime

	err := connection.QueryRow(query, token).Scan(
		&entity.Id,
		&entity.UserId,
		&entity.Code,
		&entity.OperationType,
		&payload,
		&entity.ExpiresTime,
		&usedAt,
		&createUserId,
		&entity.CreateTime,
		&updateUserId,
		&entity.UpdateTime,
		&deleteUserId,
		&deleteTime,
		&entity.IsDelete,
	)
	if err != nil {
		if err != sql.ErrNoRows {
			util.Logger.Errorw("Failed to get operation confirmation code", "error", err)
		}
		return model.OperationConfirmCode{}
	}

	if payload.Valid {
		entity.Payload = payload.String
	}
	if usedAt.Valid {
		entity.UsedTime = &usedAt.Time
	}
	if createUserId != "" {
		entity.CreateUserId = util.Convert2ObjectId(createUserId)
	}
	if updateUserId != "" {
		entity.UpdateUserId = util.Convert2ObjectId(updateUserId)
	}
	if deleteUserId.Valid && deleteUserId.String != "" {
		deleteObjectId := util.Convert2ObjectId(deleteUserId.String)
		entity.DeleteUserId = &deleteObjectId
	}
	if deleteTime.Valid {
		entity.DeleteTime = &deleteTime.Time
	}

	return entity
}

func (mapper *OperationConfirmCodeMySQLMapper) InvalidateActiveCodes(userId string, operationType string) error {
	query := `UPDATE ` + database.OperationConfirmCodeTableName + `
		SET is_delete = TRUE, delete_time = ?, update_time = ?
		WHERE user_id = ? AND operation_type = ? AND is_delete = FALSE AND used_at IS NULL`

	now := time.Now()
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	_, err := connection.Exec(query, now, now, userId, operationType)
	return err
}

func (mapper *OperationConfirmCodeMySQLMapper) MarkCodeAsUsed(id string) error {
	query := `UPDATE ` + database.OperationConfirmCodeTableName + `
		SET used_at = ?, update_time = ?
		WHERE id = ? AND is_delete = FALSE`

	now := time.Now()
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	_, err := connection.Exec(query, now, now, id)
	return err
}

func (mapper *OperationConfirmCodeMySQLMapper) DeleteCode(id string) error {
	query := `UPDATE ` + database.OperationConfirmCodeTableName + `
		SET is_delete = TRUE, delete_time = ?, update_time = ?
		WHERE id = ?`

	now := time.Now()
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	_, err := connection.Exec(query, now, now, id)
	return err
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func objectIDHexOrNil(id primitive.ObjectID) interface{} {
	if id.IsZero() {
		return nil
	}
	return id.Hex()
}

func objectIDPointerHexOrNil(id *primitive.ObjectID) interface{} {
	if id == nil || id.IsZero() {
		return nil
	}
	return id.Hex()
}
