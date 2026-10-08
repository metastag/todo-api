## Implementation Plan — Go Todo CRUD API

### 1. Define the scope

Build a REST API for managing TODO items.

Each Todo should have:

```text
id          integer
title       string
description string
completed   boolean
created_at  timestamp
updated_at  timestamp
```

Required operations:

| Method | Endpoint | Purpose |
|---|---|---|
| `POST` | `/todos` | Create a todo |
| `GET` | `/todos` | List todos |
| `GET` | `/todos/:id` | Get one todo |
| `PUT` | `/todos/:id` | Update a todo |
| `DELETE` | `/todos/:id` | Delete a todo |

Use:

- **Go**
- **Echo** for HTTP/server/routing
- **SQLite** for persistence
- A SQLite file such as `todos.db`

Keep the project intentionally simple. **No authentication, Docker, Redis, ORM, migrations framework, repository generator, or microservices.**

---

### 2. Project structure

Use a small layered structure:

```text
todo-api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handler/
│   │   └── todo_handler.go
│   ├── service/
│   │   └── todo_service.go
│   ├── repository/
│   │   └── todo_repository.go
│   ├── model/
│   │   └── todo.go
│   └── database/
│       └── sqlite.go
├── .gitignore
├── go.mod
└── README.md
```

Responsibilities:

```text
HTTP Request
     ↓
Echo Handler
     ↓
Service
     ↓
Repository
     ↓
SQLite
```

The handler should **not directly query SQLite**.

---

### 3. Initialize the Go project

Create the Go module.

Add dependencies for:

- Echo
- SQLite driver

Use Go's standard `database/sql` package rather than introducing an ORM.

The database should be represented through `*sql.DB`.

---

### 4. Define the Todo model

Create the `Todo` model containing:

```text
ID
Title
Description
Completed
CreatedAt
UpdatedAt
```

Also define request DTOs separately where useful.

For example:

```text
CreateTodoRequest
UpdateTodoRequest
```

Don't expose database-specific implementation details unnecessarily through HTTP.

---

### 5. Implement SQLite initialization

Create a database initialization function responsible for:

1. Opening the SQLite file.
2. Configuring the database connection.
3. Creating the `todos` table if it doesn't exist.
4. Returning the database connection.

The schema should roughly be:

```sql
CREATE TABLE IF NOT EXISTS todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    completed BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
```

Use SQL parameters/placeholders for user-provided values.

Do **not** construct SQL by concatenating request data.

---

### 6. Implement the repository layer

The repository owns all database operations.

Define methods conceptually like:

```text
Create(todo)
GetByID(id)
GetAll()
Update(todo)
Delete(id)
```

Responsibilities:

- Execute SQL.
- Map database rows → Go models.
- Handle `sql.ErrNoRows`.
- Return database errors upward.

The repository should **not know anything about Echo or HTTP**.

For example, it shouldn't return HTTP status codes.

---

### 7. Implement the service layer

The service contains application/business logic.

Define methods corresponding to the CRUD operations:

```text
CreateTodo()
GetTodo()
GetTodos()
UpdateTodo()
DeleteTodo()
```

Responsibilities include:

- Validate business rules.
- Call the repository.
- Set timestamps.
- Decide what should happen when a todo doesn't exist.
- Keep HTTP concerns out of the business logic.

For this project, keep the business logic minimal rather than artificially adding complexity.

---

### 8. Implement Echo handlers

Create handlers for each endpoint.

For each request:

```text
HTTP request
    ↓
Parse path parameters
    ↓
Parse JSON body
    ↓
Validate input
    ↓
Call service
    ↓
Convert result to JSON
    ↓
Return appropriate HTTP status
```

Expected responses:

### Create

```http
POST /todos
```

Return:

```http
201 Created
```

with the created Todo.

### Get all

```http
GET /todos
```

Return:

```http
200 OK
```

with an array of Todos.

### Get one

```http
GET /todos/:id
```

Return:

```http
200 OK
```

or:

```http
404 Not Found
```

### Update

```http
PUT /todos/:id
```

Return:

```http
200 OK
```

or:

```http
404 Not Found
```

### Delete

```http
DELETE /todos/:id
```

Return:

```http
204 No Content
```

if successful.

---

### 9. Define error handling

Use consistent JSON errors.

For example:

```json
{
  "error": "todo not found"
}
```

Handle at least:

- malformed JSON → `400`
- invalid input → `400`
- invalid ID → `400`
- todo doesn't exist → `404`
- unexpected database/server failure → `500`

Don't expose raw SQLite errors to API clients.

---

### 10. Wire everything together

`main.go` should roughly perform:

```text
Initialize database
       ↓
Create repository
       ↓
Create service
       ↓
Create Echo instance
       ↓
Register routes
       ↓
Start HTTP server
```

Keep `main.go` as the composition/root wiring layer rather than putting application logic there.

---

### 11. Add basic middleware

Use Echo's basic middleware:

- Logger
- Recover

Optionally add request ID if useful, but don't add middleware just for the sake of having it.

---

### 12. Test the API

Create tests for the important behavior.

At minimum:

**Repository tests**

- create todo
- retrieve todo
- retrieve nonexistent todo
- update todo
- delete todo

**Service tests**

- validation
- CRUD behavior

**Handler/API tests**

- `POST /todos`
- `GET /todos`
- `GET /todos/:id`
- `PUT /todos/:id`
- `DELETE /todos/:id`
- malformed requests
- nonexistent IDs

For tests, use a temporary/in-memory SQLite database rather than the production `todos.db`.

---

### 13. Manual API verification

After tests pass, run the server and verify the complete flow with curl/Postman:

```text
Create
  ↓
List
  ↓
Get individual
  ↓
Update
  ↓
Get again
  ↓
Delete
  ↓
Get → 404
```

This verifies that the entire stack works together.

---

### 14. README

Document:

1. Project description.
2. Requirements.
3. How to install dependencies.
4. How to run.
5. API endpoints.
6. Example requests/responses.
7. How to run tests.
8. Project architecture.

Include a small architecture diagram:

```text
                ┌──────────────┐
                │    Client    │
                └──────┬───────┘
                       │ HTTP
                       ▼
                ┌──────────────┐
                │    Echo      │
                │   Handlers   │
                └──────┬───────┘
                       ▼
                ┌──────────────┐
                │   Service    │
                └──────┬───────┘
                       ▼
                ┌──────────────┐
                │ Repository   │
                └──────┬───────┘
                       ▼
                ┌──────────────┐
                │   SQLite     │
                │  todos.db    │
                └──────────────┘
```

## Agent execution order

execute **one stage at a time**, not generate the entire project in one shot:

```text
1. Initialize project
2. Create model + DTOs
3. Implement SQLite initialization
4. Implement repository
5. Implement service
6. Implement handlers
7. Wire Echo server
8. Add middleware
9. Add tests
10. Run tests + fix failures
11. Manually verify API
12. Write README
13. Final code-quality review
```

At each stage, you should **inspect the existing code before modifying it**, make the smallest necessary change, run relevant tests, and explain what it changed.