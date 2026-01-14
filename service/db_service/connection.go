package db_service

import (
	"strings"

	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
)

// ConnectionInfo represents database connection information
type ConnectionInfo struct {
	Type     string
	Host     string
	Database string
	Status   string
}

// TestConnection tests the database connection
func TestConnection() (*ConnectionInfo, error) {
	info := &ConnectionInfo{
		Type:     util.GetConfigByKey("db.type"),
		Database: util.GetConfigByKey("db.name"),
		Status:   "disconnected",
	}

	// Get host from connection string
	switch info.Type {
	case "mongodb":
		uri := util.GetConfigByKey("mongodb.uri")
		// Extract host from MongoDB URI (simplified)
		if strings.Contains(uri, "@") {
			parts := strings.Split(uri, "@")
			if len(parts) > 1 {
				hostPart := strings.Split(parts[1], "/")[0]
				info.Host = hostPart
			}
		}

		// Test connection by opening and closing
		database.OpenMongoDbConnection(database.CashFlowTableName)
		database.CloseMongoDbConnection()
		info.Status = "connected"

	case "mysql":
		uri := util.GetConfigByKey("mysql.uri")
		// Extract host from MySQL URI (simplified)
		if strings.Contains(uri, "@tcp(") {
			parts := strings.Split(uri, "@tcp(")
			if len(parts) > 1 {
				hostPart := strings.Split(parts[1], ")")[0]
				info.Host = hostPart
			}
		}

		// Test connection by getting connection and closing
		_ = database.GetMySqlConnection()
		database.CloseMySqlConnection()
		info.Status = "connected"
	}

	return info, nil
}
