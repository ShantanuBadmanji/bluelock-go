package dbsetup

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bluelock-go/shared"
	database "github.com/bluelock-go/shared/database/generated"
)

var db database.DBTX
var queries *database.Queries

func InitializeDb() (*sql.DB, error) {
	customLogger := shared.AcquireCustomLogger()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	db, err := sql.Open("sqlite3", filepath.Join(shared.RootDir, "database.db"))
	if err != nil {
		customLogger.Logger.Error("Failed to initialize SQLC DB", "error", err)
		os.Exit(1)
	}
	queries = database.New(db)
	return db, nil
}

func AcquireQueries() *database.Queries {
	if queries == nil {
		panic("queries not initialized, call InitializeDb first")
	}
	return queries
}
