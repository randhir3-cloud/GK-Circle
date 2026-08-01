package database

import "database/sql"

// SetTestingDB sets the global database connection pool for unit tests.
func SetTestingDB(testingDB *sql.DB) {
	db = testingDB
}
