# Andpad Transcode Demo

A Go-based microservice application demonstrating video transcoding capabilities with streaming, message queueing, and cloud storage integration.

## Overview

This project is a demonstration application that showcases a complete media transcoding pipeline using Go. It integrates multiple technologies including Kafka for message queueing, PostgreSQL for data persistence, AWS S3 for object storage, and provides a REST API for managing transcoding jobs.

## Tech Stack

### Core Framework
- **Go 1.26.4** - Main programming language
- **Fiber v3** - High-performance web framework for REST API
- **Cobra** - CLI command framework
- **Viper** - Configuration management

### Message Queue & Streaming
- **Confluent Kafka** - Primary message broker
- **Segmentio Kafka Go** - Alternative Kafka client
- **Snowflake** - ID generation for distributed systems

### Database & Persistence
- **PostgreSQL** - Primary relational database
- **GORM** - ORM for database operations
- **golang-migrate** - Database migration tool

### Cloud & Storage
- **AWS SDK v2** - AWS services integration
- **S3** - Object storage for media files

### Utilities & Tools
- **go-resty** - HTTP client library
- **Swagger/Swag** - API documentation
- **Validator** - Input validation
- **Uber Zap** - Structured logging
- **Ants** - Goroutine pool manager
- **Testify** - Testing utilities

## Project Structure

```
├── cmd/              # Command line entry points and main application
├── config/           # Configuration files and settings
├── data/             # Data files and fixtures
├── docker/           # Docker configurations and compose files
├── docs/             # Documentation
├── internal/         # Internal packages (application logic)
├── pkg/              # Public packages (reusable libraries)
├── samples/          # Sample data and examples
├── go.mod & go.sum   # Go module dependencies
├── makefile          # Build and deployment automation
└── README.md         # This file
```

## Getting Started

### Prerequisites
- Go 1.26.4 or higher
- Docker & Docker Compose
- PostgreSQL (optional, can run in Docker)
- AWS credentials (for S3 integration)
- Kafka (optional, can run in Docker)

### Local Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/tutybukuny/andpad-transcode-demo.git
   cd andpad-transcode-demo
   ```

2. **Build the application**
   ```bash
   make tdapp
   ```

3. **Start development environment**
   ```bash
   make local-dev-up
   ```
   This starts:
   - PostgreSQL database
   - Database migrations
   - LocalStack (AWS S3 emulation)

4. **Start the full environment**
   ```bash
   make local-up
   ```
   This includes all services needed for production-like testing

### Available Commands

```bash
# Build Docker image
make tdapp

# Start development environment (minimal services)
make local-dev-up

# Start full environment
make local-up

# Stop all services
make local-down

# Generate Swagger API documentation
make swagger
```

## API Documentation

The application includes Swagger/OpenAPI documentation. Generate it using:
```bash
make swagger
```

The API is built with Fiber v3 and provides endpoints for managing transcoding jobs and media processing.

## Architecture

### Core Components

- **API Server** - REST API built with Fiber for job management
- **Message Queue** - Kafka-based event streaming for asynchronous processing
- **Database Layer** - PostgreSQL with GORM ORM and migration support
- **Cloud Integration** - AWS S3 for media storage and retrieval
- **CLI Tools** - Command-line utilities for administrative tasks

### Data Flow

1. User submits transcoding request via REST API
2. Request is validated and stored in PostgreSQL
3. Job event is published to Kafka topic
4. Worker processes the job asynchronously
5. Media files are stored in S3
6. Status updates are streamed back via Kafka
7. Results are persisted in database

## Configuration

Configuration is managed through Viper and supports multiple sources:
- Environment variables
- Configuration files (YAML/TOML)
- Command-line flags

## Docker Deployment

### Build
```bash
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o ./bin/tdapp ./cmd
docker build . -f ./docker/tdapp_local.DockerFile -t tdapp:local
```

### Run with Docker Compose
```bash
cd docker
docker compose up -d
```

## Development

### Database Migrations
Migrations are handled by golang-migrate and run automatically on startup.

### Testing
Run tests using Go's standard testing framework:
```bash
go test ./...
```

### Logging
Structured logging is implemented using Uber Zap for efficient, high-performance logging.

## Dependencies

Key dependencies include:
- AWS SDK for cloud integration
- Kafka clients for messaging
- PostgreSQL driver
- GORM for ORM
- Fiber for web framework
- Swagger for API documentation

See `go.mod` for complete dependency list.

## License

[Add your license here]

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.

## Author

Created by [tutybukuny](https://github.com/tutybukuny)

## Support

For issues, questions, or contributions, please open an issue on GitHub.

---

**Last Updated**: August 2026
