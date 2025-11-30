package dbsetup

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/bluelock-go/shared"
	database "github.com/bluelock-go/shared/database/generated"
)

var db database.DBTX
var querier database.Querier

func InitializeDb() (*sql.DB, error) {
	if db != nil {
		return nil, fmt.Errorf("database already initialized")
	}

	var err error
	db, err = sql.Open("sqlite3", filepath.Join(shared.RootDir, "database.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
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
