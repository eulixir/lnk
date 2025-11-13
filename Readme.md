# LNK - URL Shortener Service

A high-performance URL shortener service built with Go, using Cassandra for persistent storage and Redis for counter management. This service provides a RESTful API to create short URLs and retrieve the original URLs.

## Features

- 🔗 **URL Shortening**: Convert long URLs into short, memorable codes
- 🔄 **URL Retrieval**: Get the original URL from a short code
- 📊 **Counter Management**: Uses Redis for distributed counter management
- 💾 **Persistent Storage**: Cassandra for reliable, scalable data storage
- 🚀 **High Performance**: Optimized database queries with partition key design
- 📝 **API Documentation**: Swagger/OpenAPI documentation in development mode
- 🏥 **Health Checks**: Built-in health check endpoint
- 🧪 **Test Coverage**: Comprehensive test suite with isolated test databases

## Architecture

The project follows a clean architecture pattern with clear separation of concerns:

- **Domain Layer**: Business logic and entities
- **Gateway Layer**: External integrations (HTTP, Cassandra)
- **Extension Layer**: Infrastructure utilities (config, logger, Redis)

## Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose (for running Cassandra and Redis)
- Make (optional, for using Makefile commands)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd lnk
```

2. Install dependencies:
```bash
go mod download
```

3. Start required services using Docker Compose:
```bash
docker-compose up -d
```

This will start:
- Redis on port `6379`
- Cassandra on port `9042`

4. Create a `.env` file in the project root with the following configuration:

```env
# Application Configuration
ENV=development
PORT=8080
GIN_MODE=debug
BASE62_SALT=your-secret-salt-here

# Cassandra Configuration
CASSANDRA_HOST=localhost
CASSANDRA_PORT=9042
CASSANDRA_USERNAME=cassandra
CASSANDRA_PASSWORD=cassandra
CASSANDRA_KEYSPACE=lnk
CASSANDRA_AUTO_MIGRATE=true

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
COUNTER_KEY=url_counter
COUNTER_START_VAL=1000000

# Logger Configuration
LOG_LEVEL=info
```

**Note**: Make sure to set a secure `BASE62_SALT` value in production.

## Running the Application

### Development Mode

Run the application directly:
```bash
go run main.go
```

The server will start on `http://localhost:8080` (or the port specified in your `.env` file).

### Using Make

The project includes a Makefile with useful commands:

```bash
# Run tests
make test

# Generate test coverage report
make coverage

# Generate Swagger documentation
make swagger

# Generate migrations (requires migrate tool)
make generate-migration NAME=your_migration_name

# Generate all (swagger + mocks)
make generate
```

## API Endpoints

### Health Check

**GET** `/health`

Check if the API is running.

**Response:**
```json
{
  "message": "OK"
}
```

### Create Short URL

**POST** `/shorten`

Create a short URL from a long URL.

**Request Body:**
```json
{
  "url": "https://www.example.com/very/long/url/path"
}
```

**Response:**
```json
{
  "short_url": "abc123",
  "original_url": "https://www.example.com/very/long/url/path"
}
```

### Get Original URL

**GET** `/{short_url}`

Retrieve the original URL from a short code.

**Response:**
```json
{
  "url": "https://www.example.com/very/long/url/path"
}
```

**Status Codes:**
- `200`: Success
- `308`: Permanent Redirect (original URL found)
- `404`: URL not found
- `500`: Internal server error

### API Documentation

In development mode, Swagger documentation is available at:
```
http://localhost:8080/swagger/index.html
```

## Testing

Run the full test suite:
```bash
go test ./...
```

Or use the Makefile:
```bash
make test
```

Generate test coverage:
```bash
make coverage
```

The tests use isolated test databases that are automatically created and cleaned up for each test.

## Project Structure

```
lnk/
├── domain/                  # Domain layer
│   └── entities/
│       ├── helpers/        # URL encoding/decoding utilities
│       └── usecases/       # Business logic
├── gateways/               # Gateway layer
│   ├── gocql/             # Cassandra integration
│   │   ├── migrations/    # Database migrations
│   │   └── repositories/  # Data access layer
│   └── http/              # HTTP handlers and router
├── extensions/             # Infrastructure extensions
│   ├── config/            # Configuration management
│   ├── logger/            # Logging utilities
│   ├── redis/             # Redis client
│   └── gocqltesting/      # Testing utilities
├── docs/                   # Swagger documentation
├── main.go                 # Application entry point
├── docker-compose.yml      # Docker services configuration
├── Makefile               # Build automation
└── README.md              # This file
```

## Database Schema

### URLs Table

The `urls` table uses `short_code` as the partition key for optimal query performance:

```sql
CREATE TABLE urls (
    short_code TEXT,
    long_url TEXT,
    created_at TIMESTAMP,
    PRIMARY KEY (short_code)
);
```

This design ensures fast lookups when retrieving URLs by their short code.

## Technologies Used

- **Go 1.24**: Programming language
- **Gin**: HTTP web framework
- **Cassandra (gocql)**: Database for URL storage
- **Redis**: Counter management for URL generation
- **Zap**: Structured logging
- **Swagger/OpenAPI**: API documentation
- **Docker Compose**: Local development environment
- **Testify**: Testing framework

## Configuration

The application uses environment variables for configuration. All configuration options can be set in a `.env` file or as environment variables.

### Required Environment Variables

- `BASE62_SALT`: Secret salt for URL encoding
- `CASSANDRA_HOST`: Cassandra host address
- `CASSANDRA_PORT`: Cassandra port
- `CASSANDRA_USERNAME`: Cassandra username
- `CASSANDRA_PASSWORD`: Cassandra password
- `CASSANDRA_KEYSPACE`: Cassandra keyspace name
- `REDIS_HOST`: Redis host address
- `REDIS_PORT`: Redis port
- `REDIS_PASSWORD`: Redis password
- `REDIS_DB`: Redis database number
- `COUNTER_KEY`: Redis key for URL counter
- `COUNTER_START_VAL`: Starting value for URL counter

### Optional Environment Variables

- `ENV`: Environment name (default: `development`)
- `PORT`: Server port (default: `8080`)
- `GIN_MODE`: Gin mode (default: `debug`)
- `LOG_LEVEL`: Logging level (default: `info`)
- `CASSANDRA_AUTO_MIGRATE`: Auto-run migrations (default: `false`)

## Development

### Adding a New Migration

```bash
make generate-migration NAME=your_migration_name
```

This will create up and down migration files in `gateways/gocql/migrations/`.

### Generating Swagger Documentation

After updating API endpoints with Swagger annotations:

```bash
make swagger
```

### Code Generation

Generate mocks and documentation:

```bash
make generate
```



