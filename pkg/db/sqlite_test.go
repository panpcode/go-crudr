package db

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func createInMemoryDB() (*sql.DB, error) {
	// Open an in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Create the necessary tables for testing
	_, err = db.Exec(`
        CREATE TABLE todolist (
            id TEXT PRIMARY KEY,
            item TEXT,
            "order" INTEGER
        )
    `)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func TestCreateDb(t *testing.T) {
	// Create an in-memory SQLite database
	database, err := createInMemoryDB()
	assert.NoError(t, err, "Expected no error, but got one")

	// Ensure the database connection is closed after the test
	defer func() {
		if database != nil {
			database.Close()
		}
	}()

	// Example test: Insert a record and verify it
	_, err = database.Exec(`INSERT INTO todolist (id, item, "order") VALUES (?, ?, ?)`, "1", "Test Item", 1)
	assert.NoError(t, err, "Expected no error, but got one")

	var item string
	var order int
	err = database.QueryRow(`SELECT item, "order" FROM todolist WHERE id = ?`, "1").Scan(&item, &order)
	assert.NoError(t, err, "Expected no error, but got one")
	assert.Equal(t, "Test Item", item, "Expected item to be 'Test Item'")
	assert.Equal(t, 1, order, "Expected order to be 1")
}
