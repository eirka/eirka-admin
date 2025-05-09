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

## Development Notes

1. The project uses Go modules for dependency management
2. Configuration is loaded from `/etc/pram/pram.conf` JSON file
3. Database connections are pooled and configured via the config file
4. Redis connections are also pooled
5. Actions are audited through the audit module from eirka-libs