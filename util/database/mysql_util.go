package database

import (
	"database/sql"
	"log"
	"reflect"
	"time"

	"github.com/macar-x/cashlenx-server/util"

	_ "github.com/go-sql-driver/mysql"
)

var connection *sql.DB

func GetMySqlConnection() *sql.DB {
	// check and init database setting
	once.Do(initMySqlConnection)
	if defaultDatabaseUri == "" {
		log.Fatal("environment value 'MYSQL_DB_URI' not set")
	}

	if isConnected {
		return connection
	}

	openMySqlConnection()
	return connection
}

func openMySqlConnection() {
	var err error
	connection, err = sql.Open("mysql", defaultDatabaseUri+"/"+defaultDatabaseName)
	if err != nil {
		panic(err)
	}
	connection.SetConnMaxLifetime(time.Minute * 3)
	connection.SetMaxOpenConns(10)
	connection.SetMaxIdleConns(10)

	isConnected = true
	util.Logger.Debugln("database connection created")
}

func CloseMySqlConnection() {
	// do nothing if not connected
	if !isConnected || reflect.DeepEqual(connection, sql.DB{}) {
		isConnected = false
		return
	}
	// close the connection
	if err := connection.Close(); err != nil {
		panic(err)
	}
	isConnected = false
	util.Logger.Debugln("database connection closed")
}
