package store

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.altair.com/todolist/pkg/structs"
)

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestDelete(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	store := NewSqlStore(db)
	ctx := context.Background()
	id := uuid.New().String()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM TODOLIST WHERE ID=\?`).WithArgs(id).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := store.Update(func(tx Txn) error {
		return tx.Delete(ctx, id)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	store := NewSqlStore(db)
	ctx := context.Background()

	todoItem := &structs.TodoItem{
		Id:    uuid.New().String(),
		Item:  "Updated Item",
		Order: 1,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT "ORDER" FROM TODOLIST WHERE "ORDER" = \? AND ID != \? LIMIT 1`).WithArgs(todoItem.Order, todoItem.Id).WillReturnRows(sqlmock.NewRows([]string{"ORDER"}))
	mock.ExpectExec(`UPDATE TODOLIST SET ITEM = \?, "ORDER" = \? WHERE ID = \?`).WithArgs(todoItem.Item, todoItem.Order, todoItem.Id).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := store.Update(func(tx Txn) error {
		return tx.Update(ctx, todoItem)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGet(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	store := NewSqlStore(db)
	ctx := context.Background()
	id := uuid.New().String()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT ID, ITEM, "ORDER" FROM TODOLIST WHERE ID=\?`).WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"ID", "ITEM", "ORDER"}).AddRow(id, "Test Item", 1))
	mock.ExpectCommit()

	var todoItem structs.TodoItem
	err := store.Update(func(tx Txn) error {
		return tx.Get(ctx, id, &todoItem)
	})

	assert.NoError(t, err)
	assert.Equal(t, "Test Item", todoItem.Item)
	assert.Equal(t, 1, todoItem.Order)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestList(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	store := NewSqlStore(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT ID, ITEM, "ORDER" FROM TODOLIST`).WillReturnRows(sqlmock.NewRows([]string{"ID", "ITEM", "ORDER"}).AddRow(uuid.New().String(), "panos", 1).AddRow(uuid.New().String(), "geo", 2))
	mock.ExpectCommit()

	var todoItemList structs.TodoItemList
	err := store.Update(func(tx Txn) error {
		return tx.List(ctx, &todoItemList)
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, todoItemList.Count)
	assert.Equal(t, "panos", todoItemList.Items[0].Item)
	assert.Equal(t, "geo", todoItemList.Items[1].Item)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestAdd(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	store := NewSqlStore(db)
	ctx := context.Background()

	t.Run("Add first item", func(t *testing.T) {
		todoItem := &structs.TodoItem{
			Id:    uuid.New().String(),
			Item:  "panos",
			Order: 1,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM TODOLIST`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec(`INSERT INTO TODOLIST`).WithArgs(todoItem.Id, todoItem.Item, todoItem.Order).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := store.Update(func(tx Txn) error {
			return tx.Add(ctx, todoItem)
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Add subsequent item with correct order", func(t *testing.T) {
		todoItem := &structs.TodoItem{
			Id:    uuid.New().String(),
			Item:  "geo",
			Order: 2,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM TODOLIST`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(`SELECT MAX\("ORDER"\) FROM TODOLIST`).WillReturnRows(sqlmock.NewRows([]string{"ORDER"}).AddRow(1))
		mock.ExpectExec(`INSERT INTO TODOLIST`).WithArgs(todoItem.Id, todoItem.Item, todoItem.Order).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := store.Update(func(tx Txn) error {
			return tx.Add(ctx, todoItem)
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Add subsequent item with incorrect order", func(t *testing.T) {
		todoItem := &structs.TodoItem{
			Id:    uuid.New().String(),
			Item:  "stavroula",
			Order: 3,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM TODOLIST`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(`SELECT MAX\("ORDER"\) FROM TODOLIST`).WillReturnRows(sqlmock.NewRows([]string{"ORDER"}).AddRow(1))
		mock.ExpectRollback()

		err := store.Update(func(tx Txn) error {
			return tx.Add(ctx, todoItem)
		})

		assert.Error(t, err)
		assert.Equal(t, "order should be 2 and you provided 3", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Add item with empty ID", func(t *testing.T) {
		todoItem := &structs.TodoItem{
			Id:    "",
			Item:  "Test Item 4",
			Order: 1,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM TODOLIST`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec(`INSERT INTO TODOLIST`).WithArgs(sqlmock.AnyArg(), todoItem.Item, todoItem.Order).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := store.Update(func(tx Txn) error {
			return tx.Add(ctx, todoItem)
		})

		assert.NoError(t, err)
		assert.NotEmpty(t, todoItem.Id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCheckId(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	store := NewSqlStore(db)
	ctx := context.Background()
	id := uuid.New().String()

	t.Run("ID exists", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT ID FROM TODOLIST WHERE ID = \? LIMIT 1`).
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(id))

		mock.ExpectCommit()

		err := store.Update(func(tx Txn) error {
			// Call the CheckId method, which will execute the query
			return tx.CheckId(ctx, id)
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ID does not exist", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT ID FROM TODOLIST WHERE ID = \? LIMIT 1`).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectRollback()

		err := store.Update(func(tx Txn) error {
			return tx.CheckId(ctx, id)
		})

		assert.Error(t, err)
		assert.Equal(t, err.Error(), "ID "+id+" does not exist")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT ID FROM TODOLIST WHERE ID = \? LIMIT 1`).
			WithArgs(id).
			WillReturnError(assert.AnError)

		mock.ExpectRollback()

		err := store.Update(func(tx Txn) error {
			return tx.CheckId(ctx, id)
		})

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
