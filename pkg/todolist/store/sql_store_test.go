package store

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sqlitedb "go.altair.com/todolist/pkg/db"
	"go.altair.com/todolist/pkg/structs"
)

var testingT *testing.T

func TestTodoSqlStorage(t *testing.T) {
	testingT = t
	RegisterFailHandler(Fail)

	RunSpecs(t, "store suite")
}

var _ = Describe("SQL Store tests", func() {
	var tododb *sqlx.DB
	var todostore Store
	var ctx context.Context

	Context("When database created", Ordered, func() {

		BeforeAll(func() {
			var err error
			tododb, err = sqlitedb.CreateDb()
			Expect(err).NotTo(HaveOccurred())
			todostore = NewSqlStore(tododb)
			ctx = context.Background()
		})

		AfterAll(func() {
			err := tododb.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		Specify("List returns empty", func() {
			var items structs.TodoItemList

			err := todostore.Update(func(tx Txn) error {
				return tx.List(ctx, &items)
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(items.Items).To(BeEmpty())
			Expect(items.Count).To(Equal(0))
		})

		Context("When todo item created", func() {
			var item structs.TodoItem
			BeforeEach(func() {
				item = structs.TodoItem{Id: "7efc0335-8da6-45f7-a9b6-d4a46ba3044b", Item: "Service motorbike"}
				err := todostore.Update(func(tx Txn) error {
					return tx.Add(ctx, &item)
				})
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				item = structs.TodoItem{Id: "7efc0335-8da6-45f7-a9b6-d4a46ba3044b", Item: "Service motorbike"}
				err := todostore.Update(func(tx Txn) error {
					return tx.Delete(ctx, item.Id)
				})
				Expect(err).NotTo(HaveOccurred())
			})

			Specify("Item is returned from get", func() {
				var gItem structs.TodoItem
				err := todostore.Update(func(tx Txn) error {
					return tx.Get(ctx, item.Id, &gItem)
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(gItem).To(Equal(item))
			})

			Specify("Item is returned from List", func() {
				var items structs.TodoItemList

				err := todostore.Update(func(tx Txn) error {
					return tx.List(ctx, &items)
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(items.Count).To(Equal(1))
				Expect(items.Items).To(ContainElement(item))
			})

			Context("When todo item modified", func() {
				var updatedItem structs.TodoItem
				BeforeEach(func() {
					updatedItem = structs.TodoItem{Id: "7efc0335-8da6-45f7-a9b6-d4a46ba3044b", Item: "Service motorbike and book MOT"}
					err := todostore.Update(func(tx Txn) error {
						return tx.Update(ctx, &updatedItem)
					})
					Expect(err).NotTo(HaveOccurred())
				})

				Specify("Item is returned from get", func() {
					var gItem structs.TodoItem
					err := todostore.Update(func(tx Txn) error {
						return tx.Get(ctx, item.Id, &gItem)
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(gItem).To(Equal(updatedItem))
					Expect(gItem).NotTo(Equal(item))
				})
			})

			Context("When second todo item created", func() {
				var secondItem structs.TodoItem
				BeforeEach(func() {
					secondItem = structs.TodoItem{Id: "dac2581f-9c76-47aa-877e-6c15ddcfb064", Item: "Book holiday"}
					err := todostore.Update(func(tx Txn) error {
						return tx.Add(ctx, &secondItem)
					})
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					err := todostore.Update(func(tx Txn) error {
						return tx.Delete(ctx, secondItem.Id)
					})
					Expect(err).NotTo(HaveOccurred())
				})

				Specify("Item is returned from get", func() {
					var gItem structs.TodoItem
					err := todostore.Update(func(tx Txn) error {
						return tx.Get(ctx, secondItem.Id, &gItem)
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(gItem).To(Equal(secondItem))
				})

				Specify("Item is returned from List", func() {
					var items structs.TodoItemList

					err := todostore.Update(func(tx Txn) error {
						return tx.List(ctx, &items)
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(items.Count).To(Equal(2))
					Expect(items.Items).To(ContainElements(item, secondItem))
				})
			})
		})
	})
})
