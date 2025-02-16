package structs

type TodoItem struct {
	Id   string
	Item string
}

type TodoItemList struct {
	Items []TodoItem
	Count int
}
