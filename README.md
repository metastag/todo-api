# Todo API

REST API for managing todo list items built with Go, Echo, and SQLite.

It follows a clear seperation of concerns pattern - with repository, service and handler layers.

Additionally, it includes database migration, unit tests and stress tests using k6.

Note: This project has been made to serve as a reference for APIs written in golang.


## Installing Dependencies

To install project Go dependencies:
```bash
go mod download
```

To install the `migrate` CLI tool with SQLite support:
```bash
go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Database Migrations

Database schema migrations are stored as SQL files in the `migrations/` directory.

### Apply Migrations (Up)
```bash
migrate -database "sqlite://todos.db" -path migrations up
```

### Rollback Migrations (Down)
```bash
migrate -database "sqlite://todos.db" -path migrations down
```

## Running the Server

Start the application with:
```bash
go run ./cmd/server
```

By default, the server listens on port `8080` and connects to `todos.db`. You can optionally configure these via environment variables:
- `PORT`: Server port (default: `8080`)
- `DB_PATH`: SQLite database file path (default: `todos.db`)

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/todos` | Create a new todo |
| `GET` | `/todos` | Retrieve all todos |
| `GET` | `/todos/:id` | Retrieve a specific todo by ID |
| `PUT` | `/todos/:id` | Update an existing todo by ID |
| `DELETE` | `/todos/:id` | Delete a todo by ID |

## Running Tests

Execute the full suite of unit and integration tests:
```bash
go test -v ./...
```

### Load & Stress Testing with k6

In primary results, p(95) = 1.38s with 100 concurrent users doing write operations.

Ensure the server is running (`go run ./cmd/server`) and migrations have been applied.

1. **Load Testing** (verifies system performance under expected concurrency):
   ```bash
   k6 run tests/load.js
   ```

2. **Stress Testing** (verifies behavior and recovery under peak and extreme concurrency):
   ```bash
   k6 run tests/stress.js
   ```

You can customize the target URL via the `BASE_URL` environment variable:
```bash
k6 run -e BASE_URL=http://localhost:8080 tests/load.js
```

### Benchmark & Stress Test Results

Stress testing was performed using **k6** against the full CRUD lifecycle under sustained and peak loads up to **100 concurrent virtual users**:

| Metric | Result |
|---|---|
| **Peak Virtual Users (VUs)** | 100 concurrent users |
| **Total Requests Handled** | 13,672 |
| **Throughput** | ~170 requests/sec |
| **Error Rate** | **0.00%** (0 failures out of 13,672) |
| **Median Latency (`p50`)** | 132.06 ms |
| **Average Latency** | 286.67 ms |
| **95th Percentile Latency (`p95`)** | 1.12 s |

> **Key takeaway:** The single-connection SQLite configuration demonstrated **100% reliability with zero dropped or failed requests** under peak concurrency of 100 users, maintaining a median latency of ~132ms and a 95th-percentile response time of 1.12s during heavy write contention.


## Project Architecture

The codebase follows a clean layered architecture with unidirectional dependencies:

1. **Handler Layer (`internal/handler`)**: Receives HTTP requests through Echo, validates input parameters and request bodies, invokes the service layer, and formats JSON responses and HTTP status codes.
2. **Service Layer (`internal/service`)**: Encapsulates application and business logic, input sanitization (trimming whitespace), timestamp management, and orchestrates repository calls. Contains no HTTP framework dependencies.
3. **Repository Layer (`internal/repository`)**: Direct data access layer executing parameterized SQL queries against SQLite, mapping rows to Go domain models, and abstracting database specifics.
4. **Database Layer (`internal/database`)**: Manages SQLite connection lifecycle, driver configuration, and connection pooling.
5. **Model Layer (`internal/model`)**: Defines domain entities, request DTOs, and standard response payloads.
6. **Migrations (`migrations/`)**: Versioned SQL files for database schema management.
