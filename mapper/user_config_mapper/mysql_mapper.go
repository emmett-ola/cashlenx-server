package user_config_mapper

import (
	"bytes"
	"database/sql"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserConfigurationMySqlMapper struct{}

func (m UserConfigurationMySqlMapper) GetByUserId(plainUserId string) model.UserConfigurationEntity {
	query := `SELECT id, belongs_user_id, display_language, currency_code, active_theme_color, create_user_id, create_time, update_user_id, update_time FROM ` + database.UserConfigurationTableName + ` WHERE belongs_user_id = ? AND is_delete = FALSE`

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	row := connection.QueryRow(query, plainUserId)

	var config model.UserConfigurationEntity
	var id, belongsUserId, createUserId, updateUserId string
	var createTime, updateTime time.Time
	err := row.Scan(&id, &belongsUserId, &config.DisplayLanguage, &config.CurrencyCode, &config.ActiveThemeColor, &createUserId, &createTime, &updateUserId, &updateTime)
	if err != nil {
		if err == sql.ErrNoRows {
			util.Logger.Debugw("User configuration not found", "userId", plainUserId)
			return model.UserConfigurationEntity{}
		}
		util.Logger.Errorw("Failed to get user configuration", "error", err, "userId", plainUserId)
		return model.UserConfigurationEntity{}
	}

	config.Id = util.Convert2ObjectId(id)
	config.BelongsUserId = util.Convert2ObjectId(belongsUserId)
	config.CreateUserId = util.Convert2ObjectId(createUserId)
	config.CreateTime = createTime
	config.UpdateUserId = util.Convert2ObjectId(updateUserId)
	config.UpdateTime = updateTime

	return config
}

func (m UserConfigurationMySqlMapper) InsertByEntity(newEntity model.UserConfigurationEntity) string {
	now := time.Now()
	if newEntity.Id.IsZero() {
		newEntity.Id = primitive.NewObjectID()
	}
	if newEntity.CreateTime.IsZero() {
		newEntity.CreateTime = now
	}
	if newEntity.UpdateTime.IsZero() {
		newEntity.UpdateTime = now
	}

	query := `INSERT INTO ` + database.UserConfigurationTableName + ` (id, belongs_user_id, display_language, currency_code, active_theme_color, create_user_id, create_time, update_user_id, update_time, is_delete) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	result, err := connection.Exec(
		query,
		newEntity.Id.Hex(),
		newEntity.BelongsUserId.Hex(),
		newEntity.DisplayLanguage,
		newEntity.CurrencyCode,
		newEntity.ActiveThemeColor,
		newEntity.CreateUserId.Hex(),
		newEntity.CreateTime,
		newEntity.UpdateUserId.Hex(),
		newEntity.UpdateTime,
		false,
	)
	if err != nil {
		util.Logger.Errorw("Failed to insert user configuration", "error", err, "userId", newEntity.BelongsUserId.Hex())
		return ""
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		util.Logger.Errorw("Failed to insert user configuration", "error", err, "userId", newEntity.BelongsUserId.Hex())
		return ""
	}

	return newEntity.Id.Hex()
}

func (m UserConfigurationMySqlMapper) UpdateByUserId(plainUserId string, updatedEntity model.UserConfigurationEntity) model.UserConfigurationEntity {
	updatedEntity.UpdateTime = time.Now()
	query := `UPDATE ` + database.UserConfigurationTableName + ` SET display_language = ?, currency_code = ?, active_theme_color = ?, update_user_id = ?, update_time = ? WHERE belongs_user_id = ? AND is_delete = FALSE`

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	result, err := connection.Exec(
		query,
		updatedEntity.DisplayLanguage,
		updatedEntity.CurrencyCode,
		updatedEntity.ActiveThemeColor,
		updatedEntity.UpdateUserId.Hex(),
		updatedEntity.UpdateTime,
		plainUserId,
	)
	if err != nil {
		util.Logger.Errorw("Failed to update user configuration", "error", err, "userId", plainUserId)
		return model.UserConfigurationEntity{}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		util.Logger.Errorw("Failed to update user configuration", "error", err, "userId", plainUserId)
		return model.UserConfigurationEntity{}
	}

	return m.GetByUserId(plainUserId)
}

func (m UserConfigurationMySqlMapper) GetAllIncludeDeleted(limit, offset int) []model.UserConfigurationEntity {
	query := `SELECT id, belongs_user_id, display_language, currency_code, active_theme_color, create_user_id, create_time, update_user_id, update_time FROM ` + database.UserConfigurationTableName + ` ORDER BY update_time DESC`
	if limit > 0 {
		query += ` LIMIT ? OFFSET ?`
	}

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	var rows *sql.Rows
	var err error
	if limit > 0 {
		rows, err = connection.Query(query, limit, offset)
	} else {
		rows, err = connection.Query(query)
	}
	if err != nil {
		util.Logger.Errorw("Failed to get user configurations", "error", err)
		return []model.UserConfigurationEntity{}
	}
	defer rows.Close()

	var configs []model.UserConfigurationEntity
	for rows.Next() {
		var config model.UserConfigurationEntity
		var id, belongsUserId, createUserId, updateUserId string
		var createTime, updateTime time.Time
		if err := rows.Scan(&id, &belongsUserId, &config.DisplayLanguage, &config.CurrencyCode, &config.ActiveThemeColor, &createUserId, &createTime, &updateUserId, &updateTime); err != nil {
			util.Logger.Errorw("Failed to scan user configuration", "error", err)
			continue
		}
		config.Id = util.Convert2ObjectId(id)
		config.BelongsUserId = util.Convert2ObjectId(belongsUserId)
		config.CreateUserId = util.Convert2ObjectId(createUserId)
		config.CreateTime = createTime
		config.UpdateUserId = util.Convert2ObjectId(updateUserId)
		config.UpdateTime = updateTime
		configs = append(configs, config)
	}

	return configs
}

func (m UserConfigurationMySqlMapper) CountAll() int64 {
	query := `SELECT COUNT(1) FROM ` + database.UserConfigurationTableName

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	var count int64
	if err := connection.QueryRow(query).Scan(&count); err != nil {
		util.Logger.Errorw("Failed to count user configurations", "error", err)
		return 0
	}
	return count
}

func (m UserConfigurationMySqlMapper) Truncate() error {
	var sqlString bytes.Buffer
	sqlString.WriteString("TRUNCATE TABLE ")
	sqlString.WriteString(database.UserConfigurationTableName)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	if _, err := connection.Exec(sqlString.String()); err != nil {
		util.Logger.Errorw("Failed to truncate user configurations", "error", err)
		return err
	}

	util.Logger.Infow("User configurations truncated successfully")
	return nil
}
