# Todo-App V2

A high-performance, concurrent todo list application written in Go with both CLI and HTTP API server modes.  
Built with an actor pattern for safe concurrent operations and comprehensive test coverage.

## 📋 Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
  - [CLI Mode](#cli-mode)
  - [Server Mode](#server-mode)
- [API Documentation](#api-documentation)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Development](#development)
- [Technical Details](#technical-details)
- [Contributing](#contributing)
- [License](#license)
- [Author](#author)

## ✨ Features

- **Dual Mode Operation**: Run as CLI tool or HTTP API server
- **Actor Pattern**: Safe concurrent operations using Go channels
- **Persistent Storage**: JSON-based file storage with automatic reload
- **RESTful API**: Full HTTP API with JSON endpoints
- **Comprehensive Logging**: Structured logging with trace IDs
- **Thread-Safe**: Built-in concurrency support for multiple operations
- **Full CRUD Operations**: Create, Read, Update, Delete todo items
- **Status Management**: Track items as not_started, in_progress, or completed
- **High Test Coverage**: Extensive unit tests including concurrency tests

## 🏗️ Architecture

The application uses a clean architecture with the following layers:

```
┌─────────────────────────────────────────────┐
│           CLI / HTTP Handler Layer          │
├─────────────────────────────────────────────┤
│            Actor Layer (Channels)           │
├─────────────────────────────────────────────┤
│             Storage Layer (JSON)            │
├─────────────────────────────────────────────┤
│            File System / Logging            │
└─────────────────────────────────────────────┘
```

### Key Components

- **Actor Pattern**: Serializes all storage operations through channels for thread safety
- **Storage Layer**: Handles JSON persistence with automatic file management
- **Handler Layer**: Provides HTTP endpoints and routes
- **Logging Layer**: Structured logging with context and trace IDs

## 🔧 Prerequisites

- Go 1.25.1 or higher
- Windows, Linux, or macOS

## 📦 Installation

1. Clone the repository:
```bash
git clone https://github.com/LeoRidgwellCGI/Todo-App-V2.git
cd Todo-AppV2
```

2. Build the application:
```bash
go build -o todo-app.exe .
```

## 🚀 Usage

### CLI Mode

The application runs in CLI mode by default. All data is stored in `%USERPROFILE%\AppData\Local\tododata\todos.json`.

#### List all items:
```bash
go run . -list
```

#### List a specific item by ID:
```bash
go run . -list -itemid 1
```

#### Create a new todo item:
```bash
go run . -create "Buy groceries" -status not_started
```

#### Update an existing item:
```bash
go run . -update 1 -description "Buy groceries and cook dinner" -status in_progress
```

#### Delete an item:
```bash
go run . -delete 1
```

#### Valid status values:
- `not_started` - Task hasn't been started
- `in_progress` - Task is in progress
- `completed` - Task has been completed

### Server Mode

Start the HTTP API server:
```bash
go run . -server
```

The server starts on `http://localhost:8080`

## 📡 API Documentation

### Endpoints

#### GET /get
List all todo items

**Response:**
```json
[
  {
    "id": 1,
    "description": "Buy groceries",
    "status": "not_started",
    "created": "2025-11-14T10:00:00Z"
  }
]
```

#### GET /get/{itemid}
Get a specific todo item by ID

**Response:**
```json
{
  "id": 1,
  "description": "Buy groceries",
  "status": "not_started",
  "created": "2025-11-14T10:00:00Z"
}
```

#### POST /create
Create a new todo item

**Request Body:**
```json
{
  "description": "Buy groceries",
  "status": "not_started"
}
```

**Response:**
```json
{
  "id": 1,
  "description": "Buy groceries",
  "status": "not_started",
  "created": "2025-11-14T10:00:00Z"
}
```

#### PUT /update
Update an existing todo item

**Request Body:**
```json
{
  "id": 1,
  "description": "Buy groceries and cook dinner",
  "status": "in_progress"
}
```

**Response:**
```json
{
  "id": 1,
  "description": "Buy groceries and cook dinner",
  "status": "in_progress",
  "created": "2025-11-14T10:00:00Z"
}
```

#### DELETE /delete/{itemid}
Delete a todo item

**Response:**
```json
{
  "deleted": 1
}
```

#### GET /list
HTML view of all todo items (dynamic web page)

#### GET /about
Static about page

## 📁 Project Structure

```
Todo-App-V2/
├── main.go                 # Application entry point and CLI handling
├── main_test.go           # Main package tests
├── go.mod                 # Go module definition
├── README.md             # This file
│
├── actor/                # Actor pattern implementation
│   ├── actor.go         # Channel-based concurrency handling
│   └── actor_test.go    # Actor tests with concurrency tests
│
├── handler/             # HTTP handlers
│   ├── handler.go      # API endpoints and routing
│   └── handler_test.go # Handler tests with concurrency tests
│
├── storage/            # Data persistence layer
│   ├── storage.go     # JSON file storage operations
│   └── storage_test.go # Storage tests (if exists)
│
└── logging/           # Logging utilities
    ├── logging.go    # Logger setup and utilities
    └── logging_test.go # Logging tests
```

## 🧪 Testing

The project includes comprehensive test coverage with unit tests, integration tests, and concurrency tests.

### Run all tests:
```bash
go test ./... -v
```

### Run tests for a specific package:
```bash
go test ./actor -v
go test ./handler -v
go test ./logging -v
```

### Run with coverage:
```bash
go test ./... -cover
```

### Run only concurrency tests:
```bash
go test ./handler -v -run TestConcurrency
go test ./actor -v -run TestActor_Concurrent
```

### Test Categories

- **Unit Tests**: Test individual functions and methods
- **Concurrency Tests**: Test thread-safety with multiple goroutines (20-50 concurrent operations)
- **Integration Tests**: Test interaction between components
- **Edge Case Tests**: Test error handling and boundary conditions

## 💻 Development

### Adding a new feature:

1. Create your feature branch:
```bash
git checkout -b feature/my-new-feature
```

2. Implement your changes with tests

3. Run tests:
```bash
go test ./... -v
```

4. Commit your changes:
```bash
git commit -am "Add new feature"
```

5. Push to the branch:
```bash
git push origin feature/my-new-feature
```

### Code Style

- Follow Go conventions and best practices
- Use `gofmt` for code formatting
- Add tests for new features
- Update documentation as needed

## 🔬 Technical Details

### Actor Pattern Implementation

The application uses an actor pattern to ensure thread-safe operations:

- All CRUD operations are serialized through a single goroutine
- Commands are sent via channels with response channels for results
- Automatic storage reload before each read operation
- Automatic persistence after each write operation

### Storage Strategy

- **Format**: JSON
- **Location**: User's AppData folder (`%USERPROFILE%\AppData\Local\tododata\`)
- **Persistence**: Automatic save after each modification
- **Reload**: Automatic reload before each read operation (ensures data consistency)

### Concurrency Model

- **Actor Pattern**: Single goroutine processes all storage operations
- **Channels**: Request/response pattern for safe concurrent access
- **Thread-Safe**: Handles multiple concurrent requests safely
- **Tested**: Comprehensive concurrency tests with 20-50 parallel operations

### Logging

- **Structured Logging**: Uses Go's `log/slog` package
- **Trace IDs**: Each request gets a unique trace ID for tracking
- **Context Propagation**: Trace IDs flow through all operations
- **Log Location**: `%USERPROFILE%\AppData\Local\tododata\todos.log`

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to:
- Update tests as appropriate
- Follow the existing code style
- Add documentation for new features
- Ensure all tests pass

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👤 Author

**Leo Ridgwell**  
Software Engineer @ CGI  
GitHub: [@LeoRidgwellCGI](https://github.com/LeoRidgwellCGI)

---

## 📊 Project Statistics

- **Lines of Code**: ~1500+
- **Test Coverage**: High (includes concurrency tests)
- **Concurrent Operations Tested**: 20-50 simultaneous operations
- **Go Version**: 1.25.1
- **Architecture**: Clean architecture with actor pattern

## 🎯 Future Enhancements

- [ ] Add authentication and authorization
- [ ] Implement database backend (PostgreSQL/MySQL)
- [ ] Add due dates and priorities
- [ ] Support for tags and categories
- [ ] Web UI with React/Vue
- [ ] Docker containerization
- [ ] API rate limiting
- [ ] GraphQL API support
- [ ] Real-time updates with WebSockets
- [ ] Multi-user support

---

**Built with ❤️ using Go**