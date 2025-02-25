# Notes about the implementation 

    1. Assuming the UI will always start ordering from 1.
    2. Accepting only UUIDs for each new ID.
    3. When the ID of a new item is missing, creating automatically a new UUID for it.
    4. All the logic is based on the drag & drop functionality.
    5. Kept the new reorder API (todolist/{id}/reorder) as simple as possible for the UI, based on the instructions. Only a body with the new order is needed (see testing below).


# Testing the functionality of the new ordering system

Assuming we want to provide the 3 integer as the new order to an existing todolist's ID the UUID: 304cc3f8-7b31-43d9-a28f-1d90b529642e. 

    1. Curl command

    curl -X PUT http://localhost:8080/todolist/304cc3f8-7b31-43d9-a28f-1d90b529642e/reorder \
         -H "Content-Type: application/json" \
             -d '{"Order": 3}'

    2. Postman approach 

    Selecting the PUT method, adding the URL: "localhost:8080/todolist/304cc3f8-7b31-43d9-a28f-1d90b529642e/reorder" and providing the following body:

    {
      "Order": 3
    }


# Cmd for quick generation of a list with items

    curl -X POST http://localhost:8080/todolist -H "Content-Type: application/json" -d '{"Item": "panos", "id": "304cc3f8-7b31-43d9-a28f-1d90b529642e", "Order": 1}' ; curl -X POST http://localhost:8080/todolist -H "Content-Type: application/json" -d '{"Item": "geo", "Id": "2bceaaa4-198d-4180-9ad8-2ceaa452b8f3", "Order": 2}' ; curl -X POST http://localhost:8080/todolist -H "Content-Type: application/json" -d '{"Item": "stavr", "id": "a94ca515-622a-4fac-9df0-96c54c039ca8", "Order": 3}' ; curl -X POST http://localhost:8080/todolist -H "Content-Type: application/json" -d '{"Item": "kostas", "Order": 4}' ; curl -X POST http://localhost:8080/todolist -H "Content-Type: application/json" -d '{"Item": "nekta", "Order": 5}'

    Also available through the `make run` rule.


# Checking the diff of my changes 

I have initialized a git project into this code base, in order to monitor my changes locally. FYI because I have also executed some `go fmt ./...` to fix the formatting of any changes of mine, I used the `git diff --ignore-space-change` or the `git diff -b` command to have a clear view of my changes.


# Note based on the instructions

I have some concerns and more enhancements about this project, but I am not putting them here and I will ask / mention them in the next interview (if any), since I have observed your note "After the task there will be a discussion on the design and implementation of the system".
