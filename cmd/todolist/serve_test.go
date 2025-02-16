package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sqlitedb "go.altair.com/todolist/pkg/db"
	"go.altair.com/todolist/pkg/structs"
	"go.altair.com/todolist/pkg/todolist"
	"go.altair.com/todolist/pkg/todolist/store"
)

var testingT *testing.T

func TestTodoServe(t *testing.T) {
	testingT = t
	RegisterFailHandler(Fail)

	RunSpecs(t, "serve suite")
}

func testRequest(ts *httptest.Server, method, path string, requestBody interface{}, decodedRespBody interface{}) *http.Response {

	var body io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		Expect(err).NotTo(HaveOccurred())
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, ts.URL+path, body)
	Expect(err).NotTo(HaveOccurred())

	resp, err := http.DefaultClient.Do(req)
	Expect(err).NotTo(HaveOccurred())

	if decodedRespBody != nil {
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&decodedRespBody)
		Expect(err).NotTo(HaveOccurred())
	}
	defer resp.Body.Close()
	return resp
}

var _ = Describe("Todo Serve tests", func() {
	Context("When serving", Ordered, func() {
		var ts *httptest.Server
		BeforeAll(func() {
			tododb, err := sqlitedb.CreateDb()
			Expect(err).NotTo(HaveOccurred())
			todostore := store.NewSqlStore(tododb)
			todoService := todolist.NewItemsService(todostore)
			handler := &todolist.ItemsHandlers{
				ItemsService: todoService,
			}
			router := newRouter()
			handler.ConfigureRoutes(router)
			ts = httptest.NewServer(router)
		})

		AfterAll(func() {
			ts.Close()
		})

		Specify("List returns empty", func() {
			var items structs.TodoItemList
			resp := testRequest(ts, "GET", "/todolist", nil, &items)
			Expect(resp.StatusCode).To(Equal(200))
			Expect(items.Count).To(Equal(0))
		})

		Context("When todo item created", func() {
			var item structs.TodoItem
			BeforeEach(func() {
				item = structs.TodoItem{Id: "7efc0335-8da6-45f7-a9b6-d4a46ba3044b", Item: "Service motorbike", Order: 1}
				resp := testRequest(
					ts,
					"POST",
					"/todolist",
					&item,
					nil)
				Expect(resp.StatusCode).To(Equal(202))
			})

			AfterEach(func() {
				resp := testRequest(
					ts,
					"DELETE",
					"/todolist/7efc0335-8da6-45f7-a9b6-d4a46ba3044b",
					nil,
					nil)
				Expect(resp.StatusCode).To(Equal(204))
			})

			Specify("Item is returned from get", func() {
				var gItem structs.TodoItem
				resp := testRequest(
					ts,
					"GET",
					"/todolist/7efc0335-8da6-45f7-a9b6-d4a46ba3044b",
					nil,
					&gItem)
				Expect(resp.StatusCode).To(Equal(200))
				Expect(item).To(Equal(gItem))
			})

			Specify("Item is returned from List", func() {
				var items structs.TodoItemList
				resp := testRequest(ts, "GET", "/todolist", nil, &items)
				Expect(resp.StatusCode).To(Equal(200))
				Expect(items.Count).To(Equal(1))
				Expect(items.Items).To(ContainElement(item))
			})

			Context("When todo item modified", func() {
				var updatedItem structs.TodoItem
				BeforeEach(func() {
					updatedItem = structs.TodoItem{Id: "7efc0335-8da6-45f7-a9b6-d4a46ba3044b", Item: "Service motorbike and book MOT"}
					resp := testRequest(ts, "PUT", "/todolist/7efc0335-8da6-45f7-a9b6-d4a46ba3044b", updatedItem, nil)
					Expect(resp.StatusCode).To(Equal(202))
				})

				Specify("Item is returned from get", func() {
					var gItem structs.TodoItem
					resp := testRequest(
						ts,
						"GET",
						"/todolist/7efc0335-8da6-45f7-a9b6-d4a46ba3044b",
						nil,
						&gItem)

					Expect(resp.StatusCode).To(Equal(200))
					Expect(gItem).To(Equal(updatedItem))
					Expect(gItem).NotTo(Equal(item))
				})
			})

			Context("When second todo item created", func() {
				var secondItem structs.TodoItem
				BeforeEach(func() {
					secondItem = structs.TodoItem{Id: "dac2581f-9c76-47aa-877e-6c15ddcfb064", Item: "Book holiday", Order: 2}
					resp := testRequest(
						ts,
						"POST",
						"/todolist",
						&secondItem,
						nil)
					Expect(resp.StatusCode).To(Equal(202))
				})

				AfterEach(func() {
					resp := testRequest(
						ts,
						"DELETE",
						"/todolist/dac2581f-9c76-47aa-877e-6c15ddcfb064",
						nil,
						nil)
					Expect(resp.StatusCode).To(Equal(204))
				})

				Specify("Item is returned from get", func() {
					var gItem structs.TodoItem
					resp := testRequest(
						ts,
						"GET",
						"/todolist/dac2581f-9c76-47aa-877e-6c15ddcfb064",
						nil,
						&gItem)
					Expect(resp.StatusCode).To(Equal(200))
					Expect(secondItem).To(Equal(gItem))
				})

				Specify("Item is returned from List", func() {
					var items structs.TodoItemList
					resp := testRequest(ts, "GET", "/todolist", nil, &items)
					Expect(resp.StatusCode).To(Equal(200))
					Expect(items.Count).To(Equal(2))
					Expect(items.Items).To(ContainElements(item, secondItem))
				})
			})
		})
	})
})
