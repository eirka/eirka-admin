# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

eirka-admin is a Go project that provides admin functionality for an eirka board. It's a backend administrative API that uses Gin as its web framework and connects to MySQL and Redis.

## Building and Running

```bash
# Build the project
go build -o eirka-admin

# Run the project
./eirka-admin

# Run tests
go test -v ./...
```

## Code Architecture

The project follows a typical MVC-like structure:

1. **Controllers** - Handle HTTP requests, validate input, call models, and return responses
   - Located in `/controllers/`
   - Each controller handles a specific administrative action (delete post, ban IP, etc.)

2. **Models** - Handle business logic and database operations
   - Located in `/models/`
   - Correspond to controllers with matching names
   - Validate data, perform database operations, and return results

3. **Utils** - Helper functions for common tasks
   - Located in `/utils/`
   - Include CloudFlare integration, CRON jobs, pagination, etc.

4. **Config** - Application configuration
   - Located in `/config/`
   - Loads from `/etc/pram/pram.conf` or uses defaults
   - Defines database, Redis, CORS, and other settings

## Key Dependencies

- **gin-gonic/gin** - Web framework
- **eirka/eirka-libs** - Shared libraries for eirka projects (handles auth, validation, etc.)
- **facebookgo/grace** - Graceful HTTP server restart
- **robfig/cron** - Cron job scheduling
- **go-sql-driver/mysql** - MySQL driver
- **gomodule/redigo** - Redis client
- **gopkg.in/DATA-DOG/go-sqlmock.v1** - SQL mocking for tests
- **github.com/stretchr/testify** - Test assertions and mocking

## Database Structure

The application interacts with MySQL tables including:
- threads - Thread information
- posts - Post information
- analytics - Usage statistics
- tags - Image tags

## Authentication

Authentication is handled via the eirka-libs/user package, which uses JWT tokens. Admin routes are protected by middleware that checks for valid authentication and proper permissions.

## Redis Caching

The application uses Redis for caching. Many controllers clear Redis caches when data is modified (e.g., when a post is deleted).

## Testing

The project uses Go's standard testing package along with some helper libraries:

1. **Controller Tests**
   - Tests are in corresponding `*_test.go` files in the controllers directory
   - Use `httptest.ResponseRecorder` to capture HTTP responses
   - Use mock middleware to simulate authenticated users and validated parameters
   - Mock Redis with `redis.NewRedisMock()` and SQL with `db.NewTestDb()`
   - Test both success and error paths

2. **Model Tests**
   - Tests are in corresponding `*_test.go` files in the models directory
   - Uses `go-sqlmock` to mock database interactions
   - Test validation logic with both valid and invalid input data
   - Test database interactions, including error cases
   - Test query parameter handling

3. **Helper Test Functions**
   - `performRequest`: Create HTTP requests for controller testing
   - `errorMessage`: Format error message JSON for assertions
   - `successMessage`: Format success message JSON for assertions
   - `mockAdminMiddleware`: Simulate authenticated admin middleware

4. **Test Patterns**
   - Always test both success and error paths
   - Mock external dependencies (database, Redis, etc.)
   - Verify that SQL expectations are met after test execution
   - Reset mocks between tests
   - Test different response status codes
   - Test response body content
   - Follow the same patterns used in eirka-post for consistent testing

## Development Notes

1. The project uses Go modules for dependency management
2. Configuration is loaded from `/etc/pram/pram.conf` JSON file
3. Database connections are pooled and configured via the config file
4. Redis connections are also pooled
5. Actions are audited through the audit module from eirka-libs