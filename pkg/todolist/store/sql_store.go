package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"go.altair.com/todolist/pkg/structs"
)

func NewSqlStore(db *sqlx.DB) Store {
	return &sqlStore{
		db: db,
	}
}

type sqlStore struct {
	db *sqlx.DB
}

func (s *sqlStore) Update(action func(tx Txn) error) error {
	dbtx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = dbtx.Rollback()
			panic(r)
		}
	}()

	tx := &sqlStoreTxn{
		txn: dbtx,
	}

	err = action(tx)
	if err != nil {
		_ = dbtx.Rollback()
		log.Debug().Msg(fmt.Sprintf("Transaction rollback due to error: %v", err))
		return err
	}

	return dbtx.Commit()
}

type sqlStoreTxn struct {
	txn *sqlx.Tx
}

func readRecord(rows *sql.Rows, record *structs.TodoItem) error {
	return rows.Scan(
		&record.Id,
		&record.Item,
		&record.Order,
	)
}

func (tx *sqlStoreTxn) DbTx() interface{} {
	return tx.txn
}

func (tx *sqlStoreTxn) CheckId(ctx context.Context, id string) error {
	var existingId string
	err := tx.txn.GetContext(ctx, &existingId, `SELECT ID FROM TODOLIST WHERE ID = ? LIMIT 1`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Msg(fmt.Sprintf("ID %s does not exist", id))
			return fmt.Errorf("ID %s does not exist", id)
		}
		log.Debug().Msg(fmt.Sprintf("Failed to check existing ID: %v", err))
		return err
	}
	return nil
}

func (tx *sqlStoreTxn) checkIfOrderExists(ctx context.Context, order int, id string) (bool, error) {
	var existingOrder int
	query := `SELECT "ORDER" FROM TODOLIST WHERE "ORDER" = ? AND ID != ? LIMIT 1`
	err := tx.txn.GetContext(ctx, &existingOrder, query, order, id)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	log.Debug().Msg(fmt.Sprintf("Failed to check existing order: %v", err))
	return false, err
}

func (tx *sqlStoreTxn) Add(ctx context.Context, record *structs.TodoItem) error {

	// Generate an UUID if the ID is empty
	if record.Id == "" {
		record.Id = uuid.New().String()
	}

	// check if any other item exists in sql
	var count int
	err := tx.txn.GetContext(ctx, &count, `SELECT COUNT(*) FROM TODOLIST`)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to get the count of items: %v", err))
	}
	if count != 0 {

		// the provided order of the new item has to be greater by 1, than the current max order
		var maxOrder int
		err = tx.txn.GetContext(ctx, &maxOrder, `SELECT MAX("ORDER") FROM TODOLIST`)
		if err != nil {
			log.Debug().Msg(fmt.Sprintf("Failed to get the max order: %v", err))
			return err
		}
		if record.Order != maxOrder+1 {
			return fmt.Errorf("order should be %d", maxOrder+1)
		}

	}

	_, err = tx.txn.ExecContext(ctx,
		tx.txn.Rebind(`INSERT INTO TODOLIST(ID, ITEM, "ORDER") VALUES(?, ?, ?)`),
		record.Id,
		record.Item,
		record.Order,
	)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to add item: %v", err))
	}
	return err
}

func (tx *sqlStoreTxn) Delete(ctx context.Context, id string) error {

	result, err := tx.txn.ExecContext(ctx, tx.txn.Rebind("DELETE FROM TODOLIST WHERE ID=?"), id)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to delete item with ID %s: %v", id, err))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		log.Debug().Msg(fmt.Sprintf("Unknown ID %s", id))
		return fmt.Errorf("unknown id")
	}
	return nil
}

func (tx *sqlStoreTxn) Update(ctx context.Context, record *structs.TodoItem) error {

	// check if the order provided to the new body is already taken by another item
	orderNeedsChange, err := tx.checkIfOrderExists(ctx, record.Order, record.Id)
	if err != nil {
		return err
	}
	if orderNeedsChange {
		// if the order has changed, we need shift again all the items with order
		// greater than the new order, like we did in the reorder method
		err := tx.Reorder(ctx, record.Id, record.Order)
		if err != nil {
			return err
		}
	}

	result, err := tx.txn.ExecContext(ctx,
		tx.txn.Rebind(`UPDATE TODOLIST SET
            ITEM = ?,
            "ORDER" = ?
            WHERE ID = ?`),
		record.Item,
		record.Order,
		record.Id,
	)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to update item: %v", err))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		log.Debug().Msg(fmt.Sprintf("Unknown ID %s", record.Id))
		return fmt.Errorf("unknown id")
	}

	return nil
}

func (tx *sqlStoreTxn) reorderItems(ctx context.Context, reorderQuery string, newOrder, currentOrder int) error {

	rows, err := tx.txn.QueryContext(ctx, tx.txn.Rebind(reorderQuery), newOrder, currentOrder)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to get the current order list of items: %v", err))
		return err
	}
	defer rows.Close()

	reordered := false
	for rows.Next() {
		reordered = true
		var record structs.TodoItem
		if err := readRecord(rows, &record); err != nil {
			log.Debug().Msg(fmt.Sprintf("Failed to read record: %v", err))
			return err
		}
		var newOrderValue int
		if newOrder > currentOrder {
			newOrderValue = record.Order - 1
		} else {
			newOrderValue = record.Order + 1
		}
		_, err = tx.txn.ExecContext(ctx,
			tx.txn.Rebind(`UPDATE TODOLIST SET "ORDER" = ? WHERE ID = ?`),
			newOrderValue,
			record.Id,
		)
		if err != nil {
			log.Debug().Msg(fmt.Sprintf("Failed to update items with new order: %v", err))
			return err
		}
	}

	if !reordered {
		return fmt.Errorf("no items were reordered")
	}

	return nil
}

func (tx *sqlStoreTxn) Reorder(ctx context.Context, id string, newOrder int) error {

	err := tx.CheckId(ctx, id)
	if err != nil {
		return err
	}

	// Get the current order of the item
	var currentOrder int
	err = tx.txn.GetContext(ctx, &currentOrder, `SELECT "ORDER" FROM TODOLIST WHERE ID = ?`, id)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to get the current order of the item: %v", err))
		return err
	}

	// newOrder > currentOrder: Get all the items which have Order smaller or equal than the newOrder and
	// greater or equal than the current Order of the item with the given ID and update their Order to Order-1.
	// newOrder < currentOrder: Get all the items which have Order greater or equal than the newOrder and
	// smaller or equal than the current Order of the item with the given ID and update their Order to Order+1.

	var reorderQuery string
	if newOrder > currentOrder {
		reorderQuery = `SELECT ID, ITEM, "ORDER" FROM TODOLIST WHERE "ORDER" <= ? AND "ORDER" >= ? ORDER BY "ORDER"`
	} else {
		reorderQuery = `SELECT ID, ITEM, "ORDER" FROM TODOLIST WHERE "ORDER" >= ? AND "ORDER" <= ? ORDER BY "ORDER"`
	}

	rows, err := tx.txn.QueryContext(ctx, tx.txn.Rebind(reorderQuery), newOrder, currentOrder)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to get the current Order list of items: %v", err))
		return err
	}
	defer rows.Close()

	err = tx.reorderItems(ctx, reorderQuery, newOrder, currentOrder)
	if err != nil {
		return err
	}

	// Update the order value for the specified item
	_, err = tx.txn.ExecContext(ctx,
		tx.txn.Rebind(`UPDATE TODOLIST SET "ORDER" = ? WHERE ID = ?`),
		newOrder,
		id,
	)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to reorder item: %v", err))
		return err
	}

	return nil
}

func (tx *sqlStoreTxn) Get(ctx context.Context, id string, item *structs.TodoItem) error {
	queryStmt := `SELECT ID, ITEM, "ORDER" FROM TODOLIST WHERE ID=?`

	rows, err := tx.txn.QueryContext(ctx, tx.txn.Rebind(queryStmt), id)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to get item with ID %s: %v", id, err))
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		log.Debug().Msg(fmt.Sprintf("Unknown ID %s", id))
		return fmt.Errorf("unknown id")
	}

	if err := readRecord(rows, item); err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to read record for ID %s: %v", id, err))
		return err
	}

	return nil
}

func (tx *sqlStoreTxn) List(ctx context.Context, items *structs.TodoItemList) error {
	queryStmt := `SELECT ID, ITEM, "ORDER" FROM TODOLIST`

	rows, err := tx.txn.QueryContext(ctx, tx.txn.Rebind(queryStmt))
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("Failed to list items: %v", err))
		return err
	}
	defer rows.Close()

	items.Items = make([]structs.TodoItem, 0)
	var record structs.TodoItem
	for rows.Next() {
		if err := readRecord(rows, &record); err != nil {
			log.Debug().Msg(fmt.Sprintf("Failed to read record: %v", err))
			return err
		}
		items.Items = append(items.Items, record)
		items.Count++
	}
	return nil
}
