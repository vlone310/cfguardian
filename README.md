# GoConfig Guardian

**Distributed Configuration Management Service** built in Go, focusing on strong consistency (CP), data integrity, and developer workflow efficiency using Raft-based consensus.

## 🎯 Features

- **Strong Consistency (CP)**: Raft-based consensus for configuration data
- **Optimistic Locking**: Version-based conflict prevention
- **Schema Enforcement**: JSON Schema validation for type safety
- **Role-Based Access Control**: Admin, Editor, and Viewer roles
- **Multi-tenancy**: Project-based configuration isolation
- **Configuration History**: Full audit trail with rollback capability
- **High Performance**: Go concurrency primitives for low-latency operations

## 🏗️ Architecture

This project follows **Hexagonal Architecture** (Ports and Adapters):

```
internal/
├── domain/          # Pure business logic
├── usecases/        # Application business rules
├── ports/           # Interface definitions (inbound/outbound)
├── adapters/        # Implementations (inbound/outbound)
└── infrastructure/  # Cross-cutting concerns
```

## 🚀 Quick Start

### Prerequisites

- Go 1.25.4+
- PostgreSQL 16+
- Docker & Docker Compose (for local development)
- Make

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/vlone310/cfguardian.git
   cd cfguardian
   ```

2. **Install development tools**
   ```bash
   make install-tools
   ```

3. **Setup environment**
   ```bash
   make setup
   # Edit .env with your configuration
   ```

4. **Start infrastructure services**
   ```bash
   make docker-up
   ```

5. **Run migrations**
   ```bash
   make migrate-up
   ```

6. **Run the application**
   ```bash
   make run
   ```

The application will start on `http://localhost:8080`

## 📖 Development

### Available Commands

```bash
make help          # Show all available commands
make build         # Build the application
make run           # Run the application
make dev           # Run with live reload
make test          # Run all tests
make lint          # Run linters
make format        # Format code
```

### Project Structure

```
cfguardian/
├── cmd/server/              # Application entry point
├── internal/
│   ├── domain/              # Business entities & logic
│   ├── usecases/            # Application use cases
│   ├── ports/               # Interface definitions
│   ├── adapters/            # Implementations
│   └── infrastructure/      # Configuration, logging, etc.
├── db/
│   ├── migrations/          # Database migrations
│   └── queries/             # SQL queries for sqlc
├── api/                     # OpenAPI specifications
├── docker/                  # Docker configurations
└── docs/                    # Documentation
```

## 🔧 Technology Stack

- **Language**: Go 1.25.4
- **Router**: chi/v5
- **Database**: PostgreSQL + sqlc
- **Consensus**: Raft (etcd/hashicorp/raft)
- **Logging**: log/slog
- **Observability**: OpenTelemetry
- **API**: OpenAPI 3.0 + oapi-codegen
- **Validation**: JSON Schema

## 📊 API Documentation

API documentation is available at:
- OpenAPI Spec: `/api/openapi.yaml`
- Swagger UI: `http://localhost:8080/docs` (when running)

### Key Endpoints

- `POST /v1/auth/login` - User authentication
- `GET /v1/projects` - List projects
- `POST /v1/projects/{id}/configs` - Create configuration
- `PUT /v1/projects/{id}/configs/{key}` - Update configuration
- `GET /v1/read/{apiKey}/{key}` - Public read API

## 🧪 Testing

```bash
# Run all tests
make test

# Run unit tests
make test-unit

# Run integration tests
make test-integration

# Run with coverage
make test-coverage
```

## 🚢 Deployment

### Docker

```bash
# Build Docker image
make docker-build

# Run with Docker Compose
make docker-up
```

### Kubernetes

```bash
# Apply Kubernetes manifests
make k8s-apply
```

See `docs/deployment/` for detailed deployment guides.

## 📝 Configuration

Configuration is managed through environment variables. See `.env.example` for all available options.

Key configuration areas:
- Server settings
- Database connection
- Raft cluster
- JWT authentication
- OpenTelemetry
- Rate limiting

## 🔒 Security

- JWT-based authentication
- Role-based access control (RBAC)
- API key authentication for read endpoints
- Bcrypt password hashing
- Request rate limiting
- Input validation

## 📈 Monitoring

- **Metrics**: Prometheus metrics at `/metrics`
- **Tracing**: OpenTelemetry traces exported to Jaeger
- **Logging**: Structured JSON logging with slog
- **Health Checks**: `/health` and `/ready` endpoints

## 🤝 Contributing

Contributions are welcome! Please see `docs/CONTRIBUTING.md` for guidelines.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with Go's powerful concurrency primitives
- Inspired by modern distributed systems design
- Follows Clean Architecture principles

## 📞 Support

- Documentation: `/docs`
- Issues: GitHub Issues
- Discussions: GitHub Discussions

---

**Status**: 🚧 Under Development

For detailed development plan, see [PLAN.md](PLAN.md)

