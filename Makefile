.PHONY: todolist
todolist:
	go build -o build/todolist ./cmd/todolist/

.PHONY: test
test:
	go test -v ./...

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

.PHONY: staticcheck
staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	$(GOPATH)/bin/staticcheck ./...

.PHONY: run
run: todolist
	@echo "Starting the application..."
	@./build/todolist &
	@sleep 2
	@echo "Executing curl command..."
	@curl -X POST http://localhost:8080/todolist -H "Content-Type: application/json" \
		-d '{"Item": "panos", "id": "304cc3f8-7b31-43d9-a28f-1d90b529642e", "Order": 1}' ; \
	curl -X POST http://localhost:8080/todolist -H "Content-Type: application/json" \
		-d '{"Item": "geo", "Id": "2bceaaa4-198d-4180-9ad8-2ceaa452b8f3", "Order": 2}' ; \
	curl -X POST http://localhost:8080/todolist -H "Content-Type: application/json" \
		-d '{"Item": "stavr", "id": "a94ca515-622a-4fac-9df0-96c54c039ca8", "Order": 3}' ; \
	curl -X POST http://localhost:8080/todolist -H "Content-Type: application/json" \
		-d '{"Item": "kostas", "Order": 4}' ; \
	curl -X POST http://localhost:8080/todolist -H "Content-Type: application/json" \
		-d '{"Item": "nekta", "Order": 5}'