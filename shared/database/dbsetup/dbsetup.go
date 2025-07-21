package dbsetup

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/bluelock-go/shared"
	database "github.com/bluelock-go/shared/database/generated"
)

var db database.DBTX
var querier database.Querier

func InitializeDb() (*sql.DB, error) {
	customLogger := shared.AcquireCustomLogger()
	if db != nil {
		return nil, fmt.Errorf("database already initialized")
	}

	var err error
	db, err = sql.Open("sqlite3", filepath.Join(shared.RootDir, "database.db"))
	if err != nil {
		customLogger.Logger.Error("Failed to initialize SQLC DB", "error", err)
		os.Exit(1)
	}
	querier = database.New(db)
	return db.(*sql.DB), nil
}

func AcquireQuerier() database.Querier {
	if db == nil {
		panic("database not initialized, call InitializeDb first")
	}
	if querier == nil {
		panic("queries not initialized, call InitializeDb first")
	}
	return querier
}
