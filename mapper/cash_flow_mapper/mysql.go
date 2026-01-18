package cash_flow_mapper

import (
	"bytes"
	"database/sql"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CashFlowMySqlMapper struct{}

func (CashFlowMySqlMapper) GetCashFlowByObjectId(plainId string) model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE ID = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), plainId)
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
	}

	var cashFlowEntity model.CashFlowEntity
	for rows.Next() {
		cashFlowEntity = convertRow2CashFlowEntity(rows)
		break
	}
	return cashFlowEntity
}

func (CashFlowMySqlMapper) GetCashFlowsByObjectIdArray(plainIdList []string) []model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE ID in ")
	// fixme: pass the params by ? instead to avoid SQL inject.
	sqlString.WriteString("(" + util.CombiningWithComma(util.BatchSurroundingWithSingleQuotes(plainIdList)) + ") ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String())
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
	}

	targetEntityList := make([]model.CashFlowEntity, len(plainIdList))
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CashFlowEntity(rows))
	}
	return targetEntityList
}

func (CashFlowMySqlMapper) GetCashFlowsByBelongsDate(belongsDate time.Time) []model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE BELONGS_DATE = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), util.FormatDateToStringWithDash(belongsDate))
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
	}

	var targetEntityList []model.CashFlowEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CashFlowEntity(rows))
	}
	return targetEntityList
}

func (CashFlowMySqlMapper) GetCashFlowsByDateRange(from, to time.Time) []model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE BELONGS_DATE BETWEEN ? AND ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(),
		util.FormatDateToStringWithDash(from),
		util.FormatDateToStringWithDash(to))
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
	}

	var targetEntityList []model.CashFlowEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CashFlowEntity(rows))
	}
	return targetEntityList
}

func (CashFlowMySqlMapper) GetCashFlowsByCategoryId(categoryPlainId string) []model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE CATEGORY_ID = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), categoryPlainId)
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
	}

	var targetEntityList []model.CashFlowEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CashFlowEntity(rows))
	}
	return targetEntityList
}

func (CashFlowMySqlMapper) GetCashFlowsByExactDesc(description string) []model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE DESCRIPTION = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), description)
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
	}

	var targetEntityList []model.CashFlowEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CashFlowEntity(rows))
	}
	return targetEntityList
}

func (CashFlowMySqlMapper) GetCashFlowsByFuzzyDesc(description string) []model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE DESCRIPTION LIKE ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), "%"+description+"%")
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
	}

	var targetEntityList []model.CashFlowEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CashFlowEntity(rows))
	}
	return targetEntityList
}

func (CashFlowMySqlMapper) CountCashFLowsByCategoryId(categoryPlainId string) int64 {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT COUNT(1) FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE CATEGORY_ID = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), categoryPlainId)
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
	}

	var rowsAffected int64
	rows.Next()
	if err = rows.Scan(&rowsAffected); err != nil {
		util.Logger.Errorw("parse row affected failed", "error", err)
		return -1
	}
	return rowsAffected
}

func (CashFlowMySqlMapper) InsertCashFlowByEntity(newEntity model.CashFlowEntity) string {
	newPlainId := newEntity.Id.Hex()
	if newPlainId == "" || newEntity.Id == primitive.NilObjectID {
		newPlainId = util.GenerateObjectId()
	}

	// Set timestamps if not provided
	if newEntity.CreateTime.IsZero() || newEntity.UpdateTime.IsZero() {
		operatingTime := time.Now().UTC()
		if newEntity.CreateTime.IsZero() {
			newEntity.CreateTime = operatingTime
		}
		if newEntity.UpdateTime.IsZero() {
			newEntity.UpdateTime = operatingTime
		}
	}

	var sqlString bytes.Buffer
	sqlString.WriteString("INSERT ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" SET ID = ?, ")
	sqlString.WriteString(" USER_ID = ?, ") // Added USER_ID
	sqlString.WriteString(" CATEGORY_ID = ?, ")
	sqlString.WriteString(" BELONGS_DATE = ?, ")
	// sqlString.WriteString(" FLOW_TYPE = ?, ")
	sqlString.WriteString(" AMOUNT = ?, ")
	sqlString.WriteString(" DESCRIPTION = ?, ")
	sqlString.WriteString(" REMARK = ?, ")
	sqlString.WriteString(" CREATE_USER_ID = ?, ")
	sqlString.WriteString(" CREATE_TIME = ?, ")
	sqlString.WriteString(" UPDATE_USER_ID = ?, ")
	sqlString.WriteString(" UPDATE_TIME = ?, ")
	sqlString.WriteString(" IS_DELETE = FALSE ")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("insert failed", "error", err)
	}

	result, err := statement.Exec(newPlainId, newEntity.UserId.Hex(), newEntity.CategoryId.Hex(), newEntity.BelongsDate,
		newEntity.Amount, newEntity.Description, newEntity.Remark, newEntity.CreateUserId.Hex(), newEntity.CreateTime, newEntity.UpdateUserId.Hex(), newEntity.UpdateTime)
	if err != nil {
		util.Logger.Errorw("insert failed", "error", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		util.Logger.Errorw("insert failed", "error", err)
	}

	return newPlainId
}

func (CashFlowMySqlMapper) BulkInsertCashFlows(entities []model.CashFlowEntity) ([]string, error) {
	if len(entities) == 0 {
		return []string{}, nil
	}

	var sqlString bytes.Buffer
	sqlString.WriteString("INSERT INTO ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" (ID, USER_ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME, IS_DELETE) VALUES ")

	ids := make([]string, len(entities))
	values := make([]interface{}, 0, len(entities)*12)

	for i, entity := range entities {
		if i > 0 {
			sqlString.WriteString(", ")
		}
		sqlString.WriteString("(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		entityId := entity.Id.Hex()
		if entityId == "" || entity.Id == primitive.NilObjectID {
			entityId = util.GenerateObjectId()
		}
		ids[i] = entityId
		
		createTime := entity.CreateTime
		updateTime := entity.UpdateTime
		if createTime.IsZero() {
			createTime = time.Now().UTC()
		}
		if updateTime.IsZero() {
			updateTime = time.Now().UTC()
		}

		values = append(values, entityId, entity.UserId.Hex(), entity.CategoryId.Hex(), entity.BelongsDate,
			entity.Amount, entity.Description, entity.Remark, entity.CreateUserId.Hex(), createTime, entity.UpdateUserId.Hex(), updateTime, false)
	}

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("batch insert failed", "error", err)
		return []string{}, err
	}

	result, err := statement.Exec(values...)
	if err != nil {
		util.Logger.Errorw("batch insert failed", "error", err)
		return []string{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		util.Logger.Errorw("batch insert failed", "error", err)
		return []string{}, err
	}

	return ids, nil
}

func (CashFlowMySqlMapper) UpdateCashFlowByEntity(plainId string, updatedEntity model.CashFlowEntity) model.CashFlowEntity {
	// Update fields from updatedEntity while preserving ID and CreateTime
	targetEntity := INSTANCE.GetCashFlowByObjectId(plainId)
	if targetEntity.IsEmpty() {
		return model.CashFlowEntity{}
	}

	updatedEntity.Id = targetEntity.Id
	updatedEntity.UserId = targetEntity.UserId
	updatedEntity.CreateTime = targetEntity.CreateTime
	updatedEntity.CreateUserId = targetEntity.CreateUserId
	updatedEntity.UpdateTime = time.Now().UTC()

	var sqlString bytes.Buffer
	sqlString.WriteString("UPDATE ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" SET CATEGORY_ID = ?, ")
	sqlString.WriteString(" BELONGS_DATE = ?, ")
	// sqlString.WriteString(" FLOW_TYPE = ?, ")
	sqlString.WriteString(" AMOUNT = ?, ")
	sqlString.WriteString(" DESCRIPTION = ?, ")
	sqlString.WriteString(" REMARK = ?, ")
	sqlString.WriteString(" UPDATE_USER_ID = ?, ")
	sqlString.WriteString(" UPDATE_TIME = ? ")
	sqlString.WriteString(" WHERE ID = ? AND IS_DELETE = FALSE ")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("update failed", "error", err)
	}

	result, err := statement.Exec(updatedEntity.CategoryId.Hex(), updatedEntity.BelongsDate,
		updatedEntity.Amount, updatedEntity.Description, updatedEntity.Remark, updatedEntity.UpdateUserId.Hex(), updatedEntity.UpdateTime, plainId)
	if err != nil {
		util.Logger.Errorw("update failed", "error", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		util.Logger.Errorw("update failed", "error", err, "rows_affected", rowsAffected)
	}

	return INSTANCE.GetCashFlowByObjectId(plainId)
}

func (CashFlowMySqlMapper) DeleteCashFlowByObjectId(plainId string) model.CashFlowEntity {
	targetEntity := INSTANCE.GetCashFlowByObjectId(plainId)
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("cash_flow is not exist")
		return model.CashFlowEntity{}
	}

	var sqlString bytes.Buffer
	sqlString.WriteString("UPDATE ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" SET IS_DELETE = TRUE, DELETE_TIME = NOW() ")
	sqlString.WriteString(" WHERE ID = ? ")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
	}

	result, err := statement.Exec(plainId)
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		// fixme: maybe we should have a rollback here.
		util.Logger.Errorw("delete failed", "error", err, "rows_affected", rowsAffected)
	}

	targetEntity.IsDelete = true
	now := time.Now().UTC()
	targetEntity.DeleteTime = &now
	return targetEntity
}

func (CashFlowMySqlMapper) DeleteCashFlowByBelongsDate(belongsDate time.Time) []model.CashFlowEntity {
	cashFlowList := INSTANCE.GetCashFlowsByBelongsDate(belongsDate)
	if cashFlowList == nil {
		util.Logger.Infoln("no cash_flow(s) found")
		return cashFlowList
	}

	var sqlString bytes.Buffer
	sqlString.WriteString("UPDATE ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" SET IS_DELETE = TRUE, DELETE_TIME = NOW() ")
	sqlString.WriteString(" WHERE BELONGS_DATE = ? ")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
	}

	result, err := statement.Exec(util.FormatDateToStringWithDash(belongsDate))
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != int64(len(cashFlowList)) {
		// fixme: maybe we should have a rollback here.
		util.Logger.Errorw("delete failed", "error", err, "rows_affected", rowsAffected)
	}

	now := time.Now().UTC()
	for i := range cashFlowList {
		cashFlowList[i].IsDelete = true
		cashFlowList[i].DeleteTime = &now
	}
	return cashFlowList
}

func (CashFlowMySqlMapper) GetAllCashFlows(limit, offset int) []model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE TRUE ")
	sqlString.WriteString(database.SqlExcludeDeleted)
	sqlString.WriteString(" ORDER BY BELONGS_DATE DESC ")

	if limit > 0 {
		sqlString.WriteString(" LIMIT ? OFFSET ? ")
	}

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = connection.Query(sqlString.String(), limit, offset)
	} else {
		rows, err = connection.Query(sqlString.String())
	}

	if err != nil {
		util.Logger.Errorw("query all failed", "error", err)
		return []model.CashFlowEntity{}
	}

	var targetEntityList []model.CashFlowEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CashFlowEntity(rows))
	}
	return targetEntityList
}

func (CashFlowMySqlMapper) GetAllCashFlowsIncludeDeleted(limit, offset int) []model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE TRUE ")
	// No SqlExcludeDeleted
	sqlString.WriteString(" ORDER BY BELONGS_DATE DESC ")

	if limit > 0 {
		sqlString.WriteString(" LIMIT ? OFFSET ? ")
	}

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = connection.Query(sqlString.String(), limit, offset)
	} else {
		rows, err = connection.Query(sqlString.String())
	}

	if err != nil {
		util.Logger.Errorw("query all failed", "error", err)
		return []model.CashFlowEntity{}
	}

	var targetEntityList []model.CashFlowEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CashFlowEntity(rows))
	}
	return targetEntityList
}

func (CashFlowMySqlMapper) CountAllCashFlows() int64 {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT COUNT(1) FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE IS_DELETE = FALSE ")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String())
	if err != nil {
		util.Logger.Errorw("count all failed", "error", err)
		return 0
	}

	var count int64
	rows.Next()
	if err = rows.Scan(&count); err != nil {
		util.Logger.Errorw("parse count failed", "error", err)
		return 0
	}
	return count
}

func (CashFlowMySqlMapper) TruncateCashFlows() error {
	var sqlString bytes.Buffer
	sqlString.WriteString("TRUNCATE TABLE ")
	sqlString.WriteString(database.CashFlowTableName)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	_, err := connection.Exec(sqlString.String())
	if err != nil {
		util.Logger.Errorw("truncate cash flows failed", "error", err)
		return err
	}

	util.Logger.Infow("Cash flows truncated successfully")
	return nil
}

// User-specific methods for data isolation

func (CashFlowMySqlMapper) GetCashFlowByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, USER_ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE ID = ? AND USER_ID = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), plainId, userId.Hex())
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
		return model.CashFlowEntity{}
	}

	var cashFlowEntity model.CashFlowEntity
	for rows.Next() {
		cashFlowEntity = convertRow2CashFlowEntityWithUser(rows)
		break
	}
	return cashFlowEntity
}

func (CashFlowMySqlMapper) GetCashFlowsByBelongsDateAndUser(belongsDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, USER_ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE BELONGS_DATE = ? AND USER_ID = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), util.FormatDateToStringWithDash(belongsDate), userId.Hex())
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
		return []model.CashFlowEntity{}
	}

	var targetEntityList []model.CashFlowEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CashFlowEntityWithUser(rows))
	}
	return targetEntityList
}

func (CashFlowMySqlMapper) GetCashFlowsByDateRangeAndUser(from, to time.Time, userId primitive.ObjectID) []model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, USER_ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE BELONGS_DATE BETWEEN ? AND ? AND USER_ID = ? AND IS_DELETE = FALSE ")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(),
		util.FormatDateToStringWithDash(from),
		util.FormatDateToStringWithDash(to),
		userId.Hex())
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
		return []model.CashFlowEntity{}
	}

	var targetEntityList []model.CashFlowEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CashFlowEntityWithUser(rows))
	}
	return targetEntityList
}

func (CashFlowMySqlMapper) GetCashFlowsByCategoryIdAndUser(categoryPlainId string, userId primitive.ObjectID) []model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, USER_ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE CATEGORY_ID = ? AND USER_ID = ? AND IS_DELETE = FALSE ")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), categoryPlainId, userId.Hex())
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
		return []model.CashFlowEntity{}
	}

	var targetEntityList []model.CashFlowEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CashFlowEntityWithUser(rows))
	}
	return targetEntityList
}

func (CashFlowMySqlMapper) GetAllCashFlowsByUser(userId primitive.ObjectID, limit, offset int) []model.CashFlowEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, USER_ID, CATEGORY_ID, BELONGS_DATE, AMOUNT, DESCRIPTION FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE USER_ID = ? AND IS_DELETE = FALSE ")
	sqlString.WriteString(" ORDER BY BELONGS_DATE DESC ")

	if limit > 0 {
		sqlString.WriteString(" LIMIT ? OFFSET ? ")
	}

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = connection.Query(sqlString.String(), userId.Hex(), limit, offset)
	} else {
		rows, err = connection.Query(sqlString.String(), userId.Hex())
	}

	if err != nil {
		util.Logger.Errorw("query all by user failed", "error", err)
		return []model.CashFlowEntity{}
	}

	var targetEntityList []model.CashFlowEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CashFlowEntityWithUser(rows))
	}
	return targetEntityList
}

func (CashFlowMySqlMapper) CountAllCashFlowsByUser(userId primitive.ObjectID) int64 {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT COUNT(1) FROM ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" WHERE USER_ID = ? AND IS_DELETE = FALSE ")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), userId.Hex())
	if err != nil {
		util.Logger.Errorw("count by user failed", "error", err)
		return 0
	}

	var count int64
	rows.Next()
	if err = rows.Scan(&count); err != nil {
		util.Logger.Errorw("parse count failed", "error", err)
		return 0
	}
	return count
}

func (CashFlowMySqlMapper) DeleteCashFlowByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CashFlowEntity {
	targetEntity := INSTANCE.GetCashFlowByObjectIdAndUser(plainId, userId)
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("cash_flow is not exist or does not belong to user")
		return model.CashFlowEntity{}
	}

	var sqlString bytes.Buffer
	sqlString.WriteString("UPDATE ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" SET IS_DELETE = TRUE, DELETE_TIME = NOW(), DELETE_USER_ID = ? ")
	sqlString.WriteString(" WHERE ID = ? AND USER_ID = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
		return model.CashFlowEntity{}
	}

	result, err := statement.Exec(userId.Hex(), plainId, userId.Hex())
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
		return model.CashFlowEntity{}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		util.Logger.Errorw("delete failed", "error", err, "rows_affected", rowsAffected)
		return model.CashFlowEntity{}
	}

	targetEntity.IsDelete = true
	now := time.Now().UTC()
	targetEntity.DeleteTime = &now
	targetEntity.DeleteUserId = &userId
	return targetEntity
}

func (CashFlowMySqlMapper) DeleteCashFlowsByBelongsDateAndUser(belongsDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity {
	cashFlowList := INSTANCE.GetCashFlowsByBelongsDateAndUser(belongsDate, userId)
	if cashFlowList == nil || len(cashFlowList) == 0 {
		util.Logger.Infoln("no cash_flow(s) found")
		return []model.CashFlowEntity{}
	}

	var sqlString bytes.Buffer
	sqlString.WriteString("UPDATE ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" SET IS_DELETE = TRUE, DELETE_TIME = NOW(), DELETE_USER_ID = ? ")
	sqlString.WriteString(" WHERE BELONGS_DATE = ? AND USER_ID = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
		return cashFlowList
	}

	result, err := statement.Exec(userId.Hex(), util.FormatDateToStringWithDash(belongsDate), userId.Hex())
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
		return cashFlowList
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != int64(len(cashFlowList)) {
		util.Logger.Errorw("delete failed", "error", err, "rows_affected", rowsAffected)
	}

	now := time.Now().UTC()
	for i := range cashFlowList {
		cashFlowList[i].IsDelete = true
		cashFlowList[i].DeleteTime = &now
		cashFlowList[i].DeleteUserId = &userId
	}
	return cashFlowList
}

func (CashFlowMySqlMapper) DeleteCashFlowsByCategoryIdAndUser(categoryPlainId string, userId primitive.ObjectID) int64 {
	var sqlString bytes.Buffer
	sqlString.WriteString("UPDATE ")
	sqlString.WriteString(database.CashFlowTableName)
	sqlString.WriteString(" SET IS_DELETE = TRUE, DELETE_TIME = NOW(), DELETE_USER_ID = ? ")
	sqlString.WriteString(" WHERE CATEGORY_ID = ? AND USER_ID = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
		return 0
	}

	result, err := statement.Exec(userId.Hex(), categoryPlainId, userId.Hex())
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
		return 0
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err, "rows_affected", rowsAffected)
		return 0
	}
	return rowsAffected
}

// Helper functions

func convertRow2CashFlowEntity(rows *sql.Rows) model.CashFlowEntity {
	var id, categoryId, belongsDate, description, remark, createUserId, updateUserId string
	var amount float64
	var createTime, updateTime time.Time

	err := rows.Scan(&id, &categoryId, &belongsDate, &amount, &description, &remark, &createUserId, &createTime, &updateUserId, &updateTime)
	if err != nil {
		util.Logger.Errorw("covert into entity failed", "error", err)
	}

	return model.CashFlowEntity{
		Id:          util.Convert2ObjectId(id),
		CategoryId:  util.Convert2ObjectId(categoryId),
		BelongsDate: util.FormatDateFromStringWithDash(belongsDate),
		Amount:      amount,
		Description: description,
		Remark:      remark,
		BaseEntity: model.BaseEntity{
			CreateUserId: util.Convert2ObjectId(createUserId),
			CreateTime:   createTime,
			UpdateUserId: util.Convert2ObjectId(updateUserId),
			UpdateTime:   updateTime,
		},
	}
}

func convertRow2CashFlowEntityWithUser(rows *sql.Rows) model.CashFlowEntity {
	var id, userId, categoryId, belongsDate, description, remark, createUserId, updateUserId string
	var amount float64
	var createTime, updateTime time.Time

	err := rows.Scan(&id, &userId, &categoryId, &belongsDate, &amount, &description, &remark, &createUserId, &createTime, &updateUserId, &updateTime)
	if err != nil {
		util.Logger.Errorw("covert into entity failed", "error", err)
	}

	return model.CashFlowEntity{
		Id:          util.Convert2ObjectId(id),
		UserId:      util.Convert2ObjectId(userId),
		CategoryId:  util.Convert2ObjectId(categoryId),
		BelongsDate: util.FormatDateFromStringWithDash(belongsDate),
		Amount:      amount,
		Description: description,
		Remark:      remark,
		BaseEntity: model.BaseEntity{
			CreateUserId: util.Convert2ObjectId(createUserId),
			CreateTime:   createTime,
			UpdateUserId: util.Convert2ObjectId(updateUserId),
			UpdateTime:   updateTime,
		},
	}
}
