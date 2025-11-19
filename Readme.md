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
- 🔍 **Observability**: OpenTelemetry integration for distributed tracing
- ⚖️ **Load Balancing**: Nginx load balancer with 4 app replicas
- 🔄 **HAProxy**: Cassandra load balancing for high availability

## Architecture

The project follows a clean architecture pattern with clear separation of concerns:

- **Domain Layer**: Business logic and entities
- **Gateway Layer**: External integrations (HTTP, Cassandra)
- **Extension Layer**: Infrastructure utilities (config, logger, Redis, OpenTelemetry)

### System Architecture

```
┌─────────────┐
│  Frontend   │ → http://localhost:8888
└──────┬──────┘       
       │
┌──────▼──────┐
│    Nginx    │ (Port 8888) - Load Balancer
└──────┬──────┘      
       │
┌──────▼──────┐
│  App (x4)   │ (Load balanced replicas)
└──────┬──────┘        
   ┌───┴──┐     
┌──▼──┐ ┌─▼──┐
│Redis│ │ DB │
└─────┘ └──┬─┘   
           │      
      ┌────▼────┐
      │HAProxy  │ (Cassandra Load Balancer)
      └────┬────┘           
           │
      ┌────▼────┐
      │Cassandra│
      └─────────┘
```


## Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose (for running all services)
- Make (optional, for using Makefile commands)

## Installation

1. Clone the repository:
```bash
git clone git@github.com:eulixir/lnk.git
cd lnk
```

2. Install backend dependencies:
```bash
cd backend
go mod download
```

3. Copy `.env.example` to `.env` file in the `backend` directory with the following configuration:

```env
# Cassandra
CASSANDRA_HOST=cassandra-lb
CASSANDRA_PORT=9042
CASSANDRA_USERNAME=cassandra
CASSANDRA_PASSWORD=cassandra
CASSANDRA_KEYSPACE=lnk
CASSANDRA_AUTO_MIGRATE=true

# APP
ENV=development
PORT=8080
GIN_MODE=debug
BASE62_SALT=banana

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
COUNTER_KEY=short_url_counter
COUNTER_START_VAL=14000000

# Log
LOG_LEVEL=debug

# Otel
SERVICE_NAME=lnk-backend
OTEL_EXPORTER_OTLP_ENDPOINT=grafana:4317
```

**Note**: 
- Make sure to set a secure `BASE62_SALT` value in production
- Use Docker service names (e.g., `cassandra-lb`, `redis`, `grafana`) when running in Docker
- Use `localhost` when running services locally outside Docker

4. Start all services using Docker Compose:
```bash
cd backend
docker-compose up -d
```

This will start all services. See [Docker Services](#docker-services) section for details.

## Running the Application

### Docker Compose (Recommended)

All services are managed via Docker Compose:

```bash
cd backend
docker-compose up -d          # Start all services
docker-compose logs -f app    # View app logs
docker-compose down           # Stop all services
```

The API will be available at `http://localhost:8888` (via Nginx load balancer).

### Development Mode (Local)

If you want to run the application locally outside Docker:

1. Update `.env` to use `localhost` instead of Docker service names:
```env
CASSANDRA_HOST=localhost
REDIS_HOST=localhost
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
```

2. Start only the infrastructure services:
```bash
cd backend
docker-compose up -d cassandra redis grafana
```

3. Run the application:
```bash
cd backend
go run cmd/app/main.go
```

The server will start on `http://localhost:8080` (or the port specified in your `.env` file).


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
http://localhost:8888/swagger/index.html
```


## Testing

```bash
cd backend
make test        # Run tests
make coverage    # Generate coverage report
```

Tests use isolated test databases that are automatically created and cleaned up.

## Project Structure

```
lnk/
├── backend/
│   ├── cmd/
│   │   ├── app/
│   │   │   └── main.go           # Application entry point
│   │   └── migrator/
│   │       └── main.go           # Migration service
│   ├── domain/                   # Domain layer
│   │   └── entities/
│   │       ├── helpers/          # URL encoding/decoding utilities
│   │       └── usecases/         # Business logic
│   ├── gateways/                 # Gateway layer
│   │   ├── gocql/                # Cassandra integration
│   │   │   ├── migrations/       # Database migrations
│   │   │   └── repositories/     # Data access layer
│   │   └── http/                 # HTTP handlers and router
│   ├── extensions/               # Infrastructure extensions
│   │   ├── config/               # Configuration management
│   │   ├── logger/               # Logging utilities
│   │   ├── redis/                # Redis client
│   │   └── opentelemetry/        # OpenTelemetry setup
│   ├── nginx/
│   │   └── nginx.conf            # Nginx load balancer config
│   ├── haproxy/
│   │   └── haproxy.cfg           # HAProxy Cassandra LB config
│   ├── docker-compose.yml        # Docker services configuration
│   ├── Dockerfile                # Multi-stage build (migrator + app)
│   ├── Makefile                  # Build automation
│   └── .env                      # Environment configuration
├── frontend/                     # Next.js frontend application
└── Readme.md                     # This file
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


## Frontend

The frontend is a modern Next.js application that provides a user-friendly interface for the URL shortener service.

### Frontend Features

- 🎨 **Modern UI**: Built with Next.js 16 and React 19
- 🎯 **Type-Safe API Client**: Auto-generated TypeScript client from Swagger/OpenAPI
- 🎨 **Beautiful Components**: Uses shadcn/ui component library
- 📱 **Responsive Design**: Mobile-friendly interface
- ⚡ **Fast Performance**: Optimized with React Compiler

### Frontend Prerequisites

- Node.js 18+ or Bun
- Backend API running (see Backend section)

### Frontend Installation

1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
# Using npm
npm install

# Or using Bun
bun install
```

3. Generate API client from Swagger documentation:
```bash
npm run generate:api
# Or
bun run generate:api
```

**Note**: Make sure the backend is running and Swagger documentation is available at `http://localhost:8888/swagger/doc.json` before generating the API client.

### Running the Frontend

#### Development Mode

```bash
npm run dev
# Or
bun run dev
```

The frontend will start on `http://localhost:3000` (default Next.js port).

**Important**: Update the API base URL in `frontend/src/api/undici-instance.ts` to point to `http://localhost:8888` (Nginx load balancer).

#### Production Build

```bash
npm run build
npm run start
# Or
bun run build
bun run start
```

### Frontend Scripts

- `npm run dev` / `bun run dev`: Start development server
- `npm run build` / `bun run build`: Build for production
- `npm run start` / `bun run start`: Start production server
- `npm run lint` / `bun run lint`: Run linter (Biome)
- `npm run lint:fix` / `bun run lint:fix`: Fix linting issues
- `npm run format` / `bun run format`: Format code
- `npm run generate:api` / `bun run generate:api`: Generate API client from Swagger

### Frontend Project Structure

```
frontend/
├── src/
│   ├── app/              # Next.js App Router pages
│   │   ├── [shortUrl]/   # Dynamic route for URL redirection
│   │   └── page.tsx      # Home page
│   ├── api/              # API client and configuration
│   │   ├── lnk.ts        # Auto-generated API client
│   │   └── undici-instance.ts  # Custom fetch instance
│   ├── components/       # React components
│   │   ├── ui/           # shadcn/ui components
│   │   ├── url-dialog.tsx
│   │   ├── url-input.tsx
│   │   └── url-shortener.tsx
│   ├── hooks/            # Custom React hooks
│   ├── lib/              # Utility functions
│   └── types/            # TypeScript type definitions
├── public/               # Static assets
├── orval.config.ts       # API client generation config
├── next.config.ts        # Next.js configuration
└── package.json          # Dependencies and scripts
```

## Technologies Used

### Backend
- **Go 1.24**: Programming language
- **Gin**: HTTP web framework
- **Cassandra (gocql)**: Database for URL storage
- **Redis**: Counter management for URL generation
- **Zap**: Structured logging
- **OpenTelemetry**: Distributed tracing
- **Swagger/OpenAPI**: API documentation
- **Docker Compose**: Local development environment
- **Nginx**: Load balancer
- **HAProxy**: Cassandra load balancer
- **Testify**: Testing framework

### Frontend
- **Next.js 16**: React framework
- **React 19**: UI library
- **TypeScript**: Type safety
- **Tailwind CSS**: Styling
- **shadcn/ui**: Component library
- **Orval**: API client generator

## Development

### Makefile Commands

```bash
cd backend

# Testing
make test              # Run tests
make coverage          # Generate test coverage report

# Documentation
make swagger           # Generate Swagger documentation

# Migrations
make generate-migration NAME=your_migration_name  # Create new migration
# Migrations auto-run via migrator service when using docker-compose

# Code Generation
make generate          # Generate all (swagger + mocks)
```

### Database Migrations

Migrations run automatically via the `migrator` service before the app starts. To add a new migration:

```bash
cd backend
make generate-migration NAME=your_migration_name
```

This creates migration files in `gateways/gocql/migrations/` that will run automatically on next `docker-compose up`.

## Docker Services

### Service Overview

| Service | Port | Description |
|---------|------|-------------|
| **Nginx** | 8888 | Load balancer (frontend access point) |
| **App** | 8080 | Application service (4 replicas, internal) |
| **Migrator** | - | Runs database migrations, then exits |
| **HAProxy** | 9042 | Cassandra load balancer |
| **Cassandra** | - | Primary database (accessed via HAProxy) |
| **Redis** | 6379 | Counter management and caching |
| **Grafana** | 8081, 4317 | Observability UI (8081) and OTLP gRPC (4317) |

### Service Startup Order

Services start in the following order with health checks:

1. **Cassandra** → waits until healthy (`nodetool status`)
2. **HAProxy** → waits for Cassandra, checks with `pgrep haproxy`
3. **Redis** → checks with `redis-cli ping`
4. **Grafana** → observability stack
5. **Migrator** → waits for Cassandra, runs migrations, exits
6. **App** → waits for migrator completion, 4 replicas, HTTP health check
7. **Nginx** → waits for app service

Check service status:
```bash
docker-compose ps
```

### Observability

OpenTelemetry traces are automatically exported to Grafana:
- **OTLP Endpoint**: `grafana:4317` (gRPC) when running in Docker
- **Grafana UI**: `http://localhost:8081` (admin/admin)
- **Service Name**: `lnk-backend` (configurable via `SERVICE_NAME` env var)

## Troubleshooting

### Services not starting

Check service logs:
```bash
docker-compose logs -f [service-name]
```

Check service status:
```bash
docker-compose ps
```

### Migrations failing

Check migrator logs:
```bash
docker-compose logs migrator
```

Ensure Cassandra is healthy:
```bash
docker-compose exec cassandra nodetool status
```

### Can't connect to services

- Verify service names in `.env` match Docker service names
- Check that services are on the same Docker network (`lnk-network`)
- Verify ports are not already in use

### Frontend can't reach backend

- Ensure backend is running: `http://localhost:8888/health`
- Check CORS configuration in `gateways/http/middleware/middleware.go`
- Verify API base URL in frontend configuration
