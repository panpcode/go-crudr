package structs

type TodoItem struct {
	Id    string `json:"id" validate:"uuid4_or_empty"`
	Item  string `json:"item" validate:"required"`
	Order int    `json:"order" validate:"required,min=1"`
}

type TodoItemList struct {
	Items []TodoItem `json:"items"`
	Count int        `json:"count"`
}

type ReorderRequest struct {
	Order int `json:"order" validate:"required,min=1"`
}
