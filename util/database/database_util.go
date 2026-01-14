package database

import (
	"log"
	"sync"

	"github.com/macar-x/cashlenx-server/util"
)

var (
	once                = sync.Once{}
	defaultDatabaseUri  string
	defaultDatabaseName string
	isConnected         = false
)

var (
	CashFlowTableName = "cash_flows"
	CategoryTableName = "categories"
	UserTableName     = "users"
)

func initMongoDbConnection() {
	defaultDatabaseUri = util.GetConfigByKey("db.mongodb.url")
	defaultDatabaseName = util.GetConfigByKey("db.name")
}

func initMySqlConnection() {
	defaultDatabaseUri = util.GetConfigByKey("db.mysql.url")
	defaultDatabaseName = util.GetConfigByKey("db.name")
}

func checkDbConnection() {
	if !isConnected {
		log.Fatal("empty database connection.")
	}
}
