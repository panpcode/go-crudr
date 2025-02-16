package store

import (
	"context"
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

func TestAdd(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	store := NewSqlStore(db)
	ctx := context.Background()

	todoItem := &structs.TodoItem{
		Id:    uuid.New().String(),
		Item:  "Test Item",
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
	mock.ExpectQuery(`SELECT ID, ITEM, "ORDER" FROM TODOLIST`).WillReturnRows(sqlmock.NewRows([]string{"ID", "ITEM", "ORDER"}).AddRow(uuid.New().String(), "Test Item 1", 1).AddRow(uuid.New().String(), "Test Item 2", 2))
	mock.ExpectCommit()

	var todoItemList structs.TodoItemList
	err := store.Update(func(tx Txn) error {
		return tx.List(ctx, &todoItemList)
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, todoItemList.Count)
	assert.Equal(t, "Test Item 1", todoItemList.Items[0].Item)
	assert.Equal(t, "Test Item 2", todoItemList.Items[1].Item)
	assert.NoError(t, mock.ExpectationsWereMet())
}
