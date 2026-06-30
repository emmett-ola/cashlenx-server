package database

import (
	"database/sql"
	"log"
	"reflect"
	"strings"
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
	dsn := buildMySqlDSN(defaultDatabaseUri, defaultDatabaseName)
	connection, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	connection.SetConnMaxLifetime(time.Minute * 3)
	connection.SetMaxOpenConns(10)
	connection.SetMaxIdleConns(10)

	isConnected = true
	util.Logger.Debugln("database connection created")
}

func buildMySqlDSN(databaseURI, databaseName string) string {
	parts := strings.SplitN(databaseURI, "?", 2)
	dsn := parts[0] + "/" + databaseName
	options := []string{}
	if len(parts) == 2 && parts[1] != "" {
		options = strings.Split(parts[1], "&")
	}

	foundParseTime := false
	for index, option := range options {
		if strings.HasPrefix(strings.ToLower(option), "parsetime=") {
			options[index] = "parseTime=true"
			foundParseTime = true
		}
	}
	if !foundParseTime {
		options = append(options, "parseTime=true")
	}
	return dsn + "?" + strings.Join(options, "&")
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

// Common SQL conditions
const (
	SqlExcludeDeleted = " AND IS_DELETE = FALSE "
)
