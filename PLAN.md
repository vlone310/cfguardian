# GoConfig Guardian - Development Plan

## Project Overview
**GoConfig Guardian** is a Distributed Configuration Management Service built in Go, focusing on strong consistency (CP), data integrity, and developer workflow efficiency using Raft-based consensus.

## Tech Stack Summary
- **Language**: Go v1.25.4
- **Architecture**: Hexagonal (Ports and Adapters)
- **Router**: chi/v5
- **Database**: PostgreSQL with sqlc v1.30.0
- **Logging**: log/slog
- **Observability**: OpenTelemetry
- **API**: OpenAPI 3.0 with oapi-codegen
- **Validation**: JSON Schema (gojsonschema)
- **Consensus**: Raft (etcd/hashicorp/raft)

---

## Phase 1: Project Setup and Infrastructure

### 1.1 Project Initialization ✅
- [x] Initialize Go module (`go mod init github.com/vlone310/cfguardian`)
- [x] Set up project directory structure (hexagonal architecture)
  ```
  cfguardian/
  ├── cmd/
  │   └── server/
  │       └── main.go                    # Application entry point
  │
  ├── internal/
  │   ├── domain/                        # Enterprise Business Rules
  │   │   ├── entities/                  # Core business entities
  │   │   │   ├── user.go
  │   │   │   ├── project.go
  │   │   │   ├── config.go
  │   │   │   ├── config_schema.go
  │   │   │   ├── role.go
  │   │   │   └── config_revision.go
  │   │   ├── valueobjects/              # Value objects
  │   │   │   ├── email.go
  │   │   │   ├── role_level.go
  │   │   │   ├── version.go
  │   │   │   └── api_key.go
  │   │   ├── services/                  # Domain services (pure logic)
  │   │   │   ├── version_manager.go
  │   │   │   ├── schema_validator.go
  │   │   │   ├── password_hasher.go
  │   │   │   └── api_key_generator.go
  │   │   └── events/                    # Domain events
  │   │       ├── config_created.go
  │   │       └── config_updated.go
  │   │
  │   ├── usecases/                      # Application Business Rules
  │   │   ├── auth/
  │   │   │   ├── login_user.go
  │   │   │   ├── register_user.go
  │   │   │   └── validate_api_key.go
  │   │   ├── user/
  │   │   │   ├── create_user.go
  │   │   │   ├── list_users.go
  │   │   │   ├── get_user.go
  │   │   │   └── delete_user.go
  │   │   ├── project/
  │   │   │   ├── create_project.go
  │   │   │   ├── list_projects.go
  │   │   │   ├── get_project.go
  │   │   │   └── delete_project.go
  │   │   ├── role/
  │   │   │   ├── assign_role.go
  │   │   │   ├── revoke_role.go
  │   │   │   └── check_permission.go
  │   │   ├── schema/
  │   │   │   ├── create_schema.go
  │   │   │   ├── list_schemas.go
  │   │   │   ├── update_schema.go
  │   │   │   └── delete_schema.go
  │   │   └── config/
  │   │       ├── create_config.go
  │   │       ├── update_config.go
  │   │       ├── get_config.go
  │   │       ├── delete_config.go
  │   │       ├── rollback_config.go
  │   │       └── read_config_by_api_key.go
  │   │
  │   ├── ports/                         # Interface definitions
  │   │   ├── inbound/                   # Ports for inbound adapters
  │   │   │   ├── http_handler.go        # Interface for HTTP handlers
  │   │   │   └── grpc_handler.go        # Interface for gRPC (future)
  │   │   └── outbound/                  # Ports for outbound adapters
  │   │       ├── user_repository.go
  │   │       ├── project_repository.go
  │   │       ├── role_repository.go
  │   │       ├── config_repository.go
  │   │       ├── config_schema_repository.go
  │   │       ├── config_revision_repository.go
  │   │       ├── raft_store.go
  │   │       ├── cache.go
  │   │       └── event_publisher.go
  │   │
  │   ├── adapters/
  │   │   ├── inbound/                   # Driving adapters (what calls us)
  │   │   │   ├── http/
  │   │   │   │   ├── router.go
  │   │   │   │   ├── middleware/
  │   │   │   │   │   ├── auth.go
  │   │   │   │   │   ├── authorization.go
  │   │   │   │   │   ├── logging.go
  │   │   │   │   │   ├── rate_limit.go
  │   │   │   │   │   └── recovery.go
  │   │   │   │   └── handlers/
  │   │   │   │       ├── auth_handler.go
  │   │   │   │       ├── user_handler.go
  │   │   │   │       ├── project_handler.go
  │   │   │   │       ├── role_handler.go
  │   │   │   │       ├── schema_handler.go
  │   │   │   │       ├── config_handler.go
  │   │   │   │       └── read_handler.go
  │   │   │   └── grpc/                  # Future gRPC handlers
  │   │   │
  │   │   └── outbound/                  # Driven adapters (what we call)
  │   │       ├── postgres/              # PostgreSQL implementation
  │   │       │   ├── user_repository.go
  │   │       │   ├── project_repository.go
  │   │       │   ├── role_repository.go
  │   │       │   ├── config_repository.go
  │   │       │   ├── config_schema_repository.go
  │   │       │   ├── config_revision_repository.go
  │   │       │   ├── transaction.go
  │   │       │   └── connection.go
  │   │       ├── raft/                  # Raft consensus implementation
  │   │       │   ├── config_store.go
  │   │       │   ├── fsm.go
  │   │       │   ├── snapshot.go
  │   │       │   └── cluster.go
  │   │       ├── redis/                 # Cache implementation
  │   │       │   └── cache.go
  │   │       └── eventbus/              # Event publishing
  │   │           └── publisher.go
  │   │
  │   └── infrastructure/                # Cross-cutting concerns
  │       ├── config/                    # Configuration loading
  │       │   └── config.go
  │       ├── logger/                    # Logging setup
  │       │   └── logger.go
  │       ├── telemetry/                 # OpenTelemetry setup
  │       │   ├── tracer.go
  │       │   └── metrics.go
  │       └── errors/                    # Error handling utilities
  │           └── errors.go
  │
  ├── pkg/                               # Public libraries (if any)
  │   └── validator/
  │       └── json_schema.go
  │
  ├── db/                                # Database related files
  │   ├── migrations/                    # SQL migrations
  │   │   ├── 001_create_users.up.sql
  │   │   ├── 001_create_users.down.sql
  │   │   └── ...
  │   ├── queries/                       # SQLC queries
  │   │   ├── users.sql
  │   │   ├── projects.sql
  │   │   ├── roles.sql
  │   │   ├── configs.sql
  │   │   ├── config_schemas.sql
  │   │   └── config_revisions.sql
  │   └── sqlc.yaml                      # SQLC configuration
  │
  ├── api/                               # API specifications
  │   ├── openapi.yaml                   # OpenAPI 3.0 spec
  │   └── generated/                     # Generated code from oapi-codegen
  │
  ├── config/                            # Configuration files
  │   ├── development.yaml
  │   ├── staging.yaml
  │   └── production.yaml
  │
  ├── docker/                            # Docker related files
  │   ├── Dockerfile
  │   ├── Dockerfile.dev
  │   └── docker-compose.yml
  │
  ├── k8s/                               # Kubernetes manifests
  │   ├── deployment.yaml
  │   ├── service.yaml
  │   ├── configmap.yaml
  │   └── ingress.yaml
  │
  ├── scripts/                           # Build and utility scripts
  │   ├── build.sh
  │   ├── migrate.sh
  │   └── generate.sh
  │
  ├── test/                              # Integration and E2E tests
  │   ├── integration/
  │   └── e2e/
  │
  ├── docs/                              # Documentation
  │   ├── architecture/
  │   │   └── adr/                       # Architecture Decision Records
  │   ├── api/
  │   └── deployment/
  │
  ├── .github/                           # GitHub workflows
  │   └── workflows/
  │       ├── ci.yml
  │       └── deploy.yml
  │
  ├── go.mod
  ├── go.sum
  ├── Makefile
  ├── README.md
  ├── PLAN.md
  └── .gitignore
  ```
- [x] Create `.gitignore` for Go projects
- [x] Set up `Makefile` with common commands
- [x] Create `docker-compose.yml` for local development
- [x] Set up `.env.example` for configuration templates

### 1.2 Development Environment ✅
- [x] Install and configure PostgreSQL (local/Docker)
- [x] Install sqlc (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0`)
- [x] Install oapi-codegen
- [x] Install golang-migrate for database migrations
- [x] Set up pre-commit hooks (golangci-lint, gofmt, tests)

### 1.3 Infrastructure Layer Setup ✅
- [x] Create infrastructure components in `internal/infrastructure/`:
  - [x] `config/config.go` - Configuration management (using viper or envconfig)
  - [x] `logger/logger.go` - Structured logging setup (slog)
  - [x] `telemetry/tracer.go` - OpenTelemetry tracer setup
  - [x] `telemetry/metrics.go` - OpenTelemetry metrics setup
  - [x] `errors/errors.go` - Error handling utilities and custom error types
- [x] Define configuration structs for:
  - Database connection
  - Server settings (host, port)
  - Raft cluster settings
  - JWT/Auth settings
  - Logging levels
  - OpenTelemetry settings

---

## Phase 2: Database Layer

### 2.1 Database Schema Design ✅
- [x] Create migration files for all tables:
  - [x] `001_create_users_table.up.sql`
  - [x] `002_create_projects_table.up.sql`
  - [x] `003_create_roles_table.up.sql`
  - [x] `004_create_config_schemas_table.up.sql`
  - [x] `005_create_configs_table.up.sql`
  - [x] `006_create_config_revisions_table.up.sql`
  - [x] Add indexes for performance
  - [x] Add foreign key constraints
  - [x] Create down migrations for each

### 2.2 SQLC Configuration ✅
- [x] Create `sqlc.yaml` configuration
- [x] Define SQL queries for all operations:
  - [x] User CRUD operations (`db/queries/users.sql`)
  - [x] Project CRUD operations (`db/queries/projects.sql`)
  - [x] Role management (`db/queries/roles.sql`)
  - [x] ConfigSchema operations (`db/queries/config_schemas.sql`)
  - [x] Config CRUD with optimistic locking (`db/queries/configs.sql`)
  - [x] ConfigRevision operations (`db/queries/config_revisions.sql`)
- [x] Run `sqlc generate` to generate Go code
- [x] Create database connection pool management

### 2.3 Repository Pattern Implementation
- [x] Define repository interfaces in `internal/ports/outbound/`:
  - [x] `user_repository.go` - UserRepository interface
  - [x] `project_repository.go` - ProjectRepository interface
  - [x] `role_repository.go` - RoleRepository interface
  - [x] `config_schema_repository.go` - ConfigSchemaRepository interface
  - [x] `config_repository.go` - ConfigRepository interface
  - [x] `config_revision_repository.go` - ConfigRevisionRepository interface
- [ ] Implement PostgreSQL adapters in `internal/adapters/outbound/postgres/`
- [ ] Add transaction support for multi-table operations
- [ ] Implement connection pooling and retry logic

---

## Phase 3: Domain Layer (Core Business Logic) ✅

### 3.1 Domain Entities ✅
- [x] Create domain entities in `internal/domain/entities/`:
  - [x] `user.go` - User entity with validation
  - [x] `project.go` - Project entity with API key generation
  - [x] `role.go` - Role entity with role level enum
  - [x] `config_schema.go` - ConfigSchema entity with JSON Schema validation
  - [x] `config.go` - Config entity with version management
  - [x] `config_revision.go` - ConfigRevision entity
- [x] Define value objects in `internal/domain/valueobjects/`:
  - [x] `email.go` - Email value object (with validation)
  - [x] `role_level.go` - RoleLevel enum (admin, editor, viewer)
  - [x] `version.go` - Version value object (optimistic locking)
  - [x] `api_key.go` - APIKey value object (generation and validation)

### 3.2 Domain Services ✅
- [x] Create domain services in `internal/domain/services/`:
  - [x] `version_manager.go` - ConfigVersionManager (optimistic locking logic)
  - [x] `schema_validator.go` - SchemaValidator (JSON Schema validation)
  - [x] `password_hasher.go` - PasswordHasher (bcrypt)
  - [x] `api_key_generator.go` - APIKeyGenerator

### 3.3 Domain Events ✅
- [x] Define domain events in `internal/domain/events/`:
  - [x] `event.go` - DomainEvent interface
  - [x] `config_created.go` - ConfigCreated event
  - [x] `config_updated.go` - ConfigUpdated event
  - [x] `config_deleted.go` - ConfigDeleted event
  - [x] `config_rolledback.go` - ConfigRolledBack event

---

## Phase 4: Application Layer (Use Cases) ✅

### 4.1 Authentication & Authorization Use Cases ✅
- [x] Implement use cases in `internal/usecases/auth/`:
  - [x] `login_user.go` - LoginUser use case
  - [x] `register_user.go` - RegisterUser use case
  - [x] `validate_api_key.go` - ValidateAPIKey use case (for read API)
- [ ] Create middleware in `internal/adapters/inbound/http/middleware/` (Phase 6):
  - [ ] `auth.go` - Authentication (JWT validation)
  - [ ] `authorization.go` - Authorization (role-based access control)
  - [ ] `rate_limit.go` - Rate limiting
  - [ ] `logging.go` - Request logging

### 4.2 User Management Use Cases ✅
- [x] Implement use cases in `internal/usecases/user/`:
  - [x] `create_user.go` - CreateUser
  - [x] `list_users.go` - ListUsers
  - [x] `get_user.go` - GetUserByID
  - [x] `delete_user.go` - DeleteUser

### 4.3 Project Management Use Cases ✅
- [x] Implement use cases in `internal/usecases/project/`:
  - [x] `create_project.go` - CreateProject (with auto admin role)
  - [x] `list_projects.go` - ListProjects (with owner filter)
  - [x] `get_project.go` - GetProjectByID
  - [x] `delete_project.go` - DeleteProject

### 4.4 Role Management Use Cases ✅
- [x] Implement use cases in `internal/usecases/role/`:
  - [x] `assign_role.go` - AssignRole (upsert)
  - [x] `revoke_role.go` - RevokeRole
  - [x] `check_permission.go` - CheckUserPermission (with helpers)

### 4.5 Schema Management Use Cases ✅
- [x] Implement use cases in `internal/usecases/schema/`:
  - [x] `create_schema.go` - CreateConfigSchema (with validation)
  - [x] `list_schemas.go` - ListConfigSchemas (with usage count)
  - [x] `update_schema.go` - UpdateConfigSchema
  - [x] `delete_schema.go` - DeleteConfigSchema (with safety check)

### 4.6 Config Management Use Cases ✅
- [x] Implement use cases in `internal/usecases/config/`:
  - [x] `create_config.go` - CreateConfig (with schema validation)
  - [x] `get_config.go` - GetConfig
  - [x] `update_config.go` - UpdateConfig (**optimistic locking** + validation)
  - [x] `delete_config.go` - DeleteConfig
  - [x] `rollback_config.go` - RollbackConfig (to previous version)
  - [x] `read_config_by_api_key.go` - ReadConfigByAPIKey (public client API)

---

## Phase 5: Raft Consensus Integration ✅

### 5.1 Raft Setup ✅
- [x] Choose Raft implementation (**hashicorp/raft** selected)
- [x] Create Raft cluster configuration (StoreConfig)
- [x] Implement Raft node initialization (with BoltDB backend)
- [x] Set up leader election handling (automatic)
- [x] Implement log replication for Config operations

### 5.2 Config Store with Raft ✅
- [x] Create Raft implementation in `internal/adapters/outbound/raft/`:
  - [x] `store.go` - Raft store implementation with cluster management
  - [x] `fsm.go` - Finite State Machine for Config operations
  - [x] `config_repository.go` - Raft-backed ConfigRepository
  - [x] `README.md` - Comprehensive Raft documentation
- [x] Implement Raft FSM operations:
  - [x] Apply log entries (CREATE_CONFIG, UPDATE_CONFIG, DELETE_CONFIG)
  - [x] Create snapshots (FSMSnapshot implementation)
  - [x] Restore from snapshots (FSM.Restore)
- [x] Integrate Raft store with ConfigRepository interface
- [x] Add conflict resolution for concurrent updates (optimistic locking in FSM)
- [x] Implement read consistency guarantees (local FSM reads)

### 5.3 Cluster Management ✅
- [x] Implement node join/leave mechanisms (Join/Leave methods)
- [x] Add health checks for Raft nodes (IsLeader, GetLeader, Stats)
- [x] Implement graceful node addition/removal (via Raft API)
- [x] Leader detection and failover (WaitForLeader)

---

## Phase 6: HTTP API Layer ✅

### 6.1 OpenAPI Specification ✅
- [x] Define complete OpenAPI 3.0 spec in `api/openapi.yaml`:
  - [x] All endpoint definitions (20+ endpoints)
  - [x] Request/response schemas
  - [x] Authentication schemes (JWT, API Key)
  - [x] Error responses
  - [x] Examples
- [x] Manual implementation (following OpenAPI spec)
- [ ] Generate client SDK (optional - future enhancement)

### 6.2 HTTP Handlers (Inbound Adapters) ✅
- [x] Set up chi router in `internal/adapters/inbound/http/`:
  - [x] `router.go` - Main router setup and route registration
- [x] Implement handler groups in `internal/adapters/inbound/http/handlers/`:
  - [x] `auth_handler.go` - `/v1/auth` - Authentication handlers
  - [x] `user_handler.go` - `/v1/users` - User management
  - [x] `project_handler.go` - `/v1/projects` - Project management
  - [x] `role_handler.go` - `/v1/projects/{id}/roles` - Role management
  - [x] `schema_handler.go` - `/v1/schemas` - Schema management
  - [x] `config_handler.go` - `/v1/projects/{id}/configs` - Config management
  - [x] `read_handler.go` - `/v1/read/{apiKey}/{key}` - Public read API

### 6.3 Middleware Implementation ✅
- [x] Implement middleware in `internal/adapters/inbound/http/middleware/`:
  - [x] `request_id.go` - Request ID middleware
  - [x] `logging.go` - Structured logging middleware (slog)
  - [x] `auth.go` - Authentication middleware (JWT)
  - [x] `authorization.go` - Authorization middleware (role-based)
  - [x] `cors.go` - CORS middleware
  - [x] `rate_limit.go` - Rate limiting middleware
  - [x] `recovery.go` - Panic recovery middleware
  - [x] Chi built-in compression and timeout

### 6.4 Error Handling ✅
- [x] Define standard error response format
- [x] Create response helpers in `common/response.go`
- [x] Implement error responses with codes
- [x] Add version conflict detection (409 Conflict)

---

## Phase 7: Observability

### 7.1 Structured Logging
- [ ] Set up slog with appropriate handlers
- [ ] Define log levels (Debug, Info, Warn, Error)
- [ ] Add contextual logging throughout application
- [ ] Implement log sampling for high-volume endpoints
- [ ] Add correlation IDs for request tracing

### 7.2 OpenTelemetry Integration
- [ ] Initialize OpenTelemetry SDK
- [ ] Set up trace provider
- [ ] Add tracing to:
  - [ ] HTTP handlers
  - [ ] Database operations
  - [ ] Raft operations
  - [ ] External API calls
- [ ] Configure trace exporters (Jaeger/Zipkin)
- [ ] Add custom spans for critical operations

### 7.3 Metrics
- [ ] Define custom metrics:
  - [ ] Request count by endpoint
  - [ ] Request duration
  - [ ] Error rates
  - [ ] Config update success/failure
  - [ ] Optimistic locking conflicts
  - [ ] Raft cluster health
  - [ ] Database connection pool stats
- [ ] Expose `/metrics` endpoint (Prometheus format)
- [ ] Set up metric exporters

### 7.4 Health Checks
- [ ] Implement `/health` endpoint
- [ ] Implement `/ready` endpoint
- [ ] Add health checks for:
  - [ ] Database connectivity
  - [ ] Raft cluster status
  - [ ] Dependent services

---

## Phase 8: Security

### 8.1 Authentication & Authorization
- [ ] Implement JWT token generation and validation
- [ ] Add refresh token mechanism
- [ ] Implement password hashing (bcrypt)
- [ ] Add API key validation for read endpoint
- [ ] Implement role-based access control (RBAC)
- [ ] Add permission checks in use cases

### 8.2 Input Validation
- [ ] Validate all request payloads
- [ ] Implement JSON Schema validation
- [ ] Add SQL injection protection (via sqlc)
- [ ] Sanitize user inputs
- [ ] Add request size limits

### 8.3 Security Headers
- [ ] Add security headers middleware
- [ ] Implement HTTPS enforcement
- [ ] Add CORS configuration
- [ ] Implement rate limiting per user/IP

### 8.4 Secrets Management
- [ ] Use environment variables for secrets
- [ ] Integrate with secrets manager (optional)
- [ ] Implement secret rotation mechanism
- [ ] Never log sensitive data

---

## Phase 9: Testing

### 9.1 Unit Tests
- [ ] Test domain entities and value objects
- [ ] Test domain services
- [ ] Test use cases with mocked repositories
- [ ] Test HTTP handlers with mocked services
- [ ] Aim for >80% code coverage

### 9.2 Integration Tests
- [ ] Test database operations with test containers
- [ ] Test full HTTP API flows
- [ ] Test Raft consensus operations
- [ ] Test optimistic locking scenarios
- [ ] Test concurrent config updates

### 9.3 End-to-End Tests
- [ ] Test complete user workflows:
  - [ ] User registration and login
  - [ ] Project creation and management
  - [ ] Config CRUD operations
  - [ ] Config rollback scenarios
  - [ ] Read API access

### 9.4 Performance Tests
- [ ] Load testing with k6 or vegeta
- [ ] Test config read throughput
- [ ] Test concurrent write scenarios
- [ ] Test Raft cluster performance
- [ ] Identify bottlenecks

---

## Phase 10: Documentation

### 10.1 Code Documentation
- [ ] Add godoc comments to all exported functions
- [ ] Document complex algorithms
- [ ] Add examples in documentation
- [ ] Generate API documentation from OpenAPI spec

### 10.2 User Documentation
- [ ] Write README.md with:
  - [ ] Project overview
  - [ ] Quick start guide
  - [ ] Installation instructions
  - [ ] Configuration guide
- [ ] Create API usage guide
- [ ] Document authentication flows
- [ ] Add deployment guide
- [ ] Create troubleshooting guide

### 10.3 Developer Documentation
- [ ] Architecture decision records (ADRs)
- [ ] Database schema documentation
- [ ] Raft cluster setup guide
- [ ] Contributing guidelines
- [ ] Development environment setup

---

## Phase 11: Deployment & Operations

### 11.1 Containerization
- [ ] Create production Dockerfile
- [ ] Create multi-stage build
- [ ] Optimize image size
- [ ] Add health check in Dockerfile
- [ ] Create docker-compose for full stack

### 11.2 Kubernetes Deployment
- [ ] Create Kubernetes manifests:
  - [ ] Deployment
  - [ ] Service (ClusterIP for internal)
  - [ ] Service (LoadBalancer for external)
  - [ ] ConfigMap
  - [ ] Secret
  - [ ] PersistentVolumeClaim (for Raft data)
  - [ ] Ingress
- [ ] Create Helm chart (optional)
- [ ] Set up horizontal pod autoscaling
- [ ] Configure resource limits and requests

### 11.3 CI/CD Pipeline
- [ ] Set up GitHub Actions / GitLab CI:
  - [ ] Run tests on PR
  - [ ] Run linters
  - [ ] Build Docker image
  - [ ] Push to registry
  - [ ] Deploy to staging
  - [ ] Deploy to production (manual approval)
- [ ] Implement semantic versioning
- [ ] Add changelog generation

### 11.4 Monitoring & Alerting
- [ ] Set up Prometheus for metrics
- [ ] Set up Grafana dashboards
- [ ] Configure alerts for:
  - [ ] High error rates
  - [ ] Slow response times
  - [ ] Raft cluster issues
  - [ ] Database connection issues
- [ ] Set up log aggregation (ELK or Loki)
- [ ] Configure on-call rotation

---

## Phase 12: Optimization & Polish

### 12.1 Performance Optimization
- [ ] Profile application with pprof
- [ ] Optimize database queries
- [ ] Add caching layer (Redis) for read-heavy endpoints
- [ ] Implement database connection pooling tuning
- [ ] Optimize JSON serialization/deserialization
- [ ] Add compression for large responses

### 12.2 Error Handling & Resilience
- [ ] Implement circuit breakers for external calls
- [ ] Add retry logic with exponential backoff
- [ ] Implement graceful degradation
- [ ] Add request timeout handling
- [ ] Implement bulkhead pattern

### 12.3 Additional Features
- [ ] Implement webhook notifications for config changes
- [ ] Add audit logging for all mutations
- [ ] Implement config diff view
- [ ] Add batch operations support
- [ ] Implement config export/import
- [ ] Add config validation before apply

---

## Phase 13: Launch Preparation

### 13.1 Pre-Launch Checklist
- [ ] Security audit
- [ ] Performance benchmarking
- [ ] Load testing
- [ ] Disaster recovery plan
- [ ] Backup and restore procedures
- [ ] Documentation review
- [ ] License and legal compliance

### 13.2 Launch
- [ ] Deploy to production
- [ ] Monitor metrics and logs
- [ ] Gather user feedback
- [ ] Create runbook for operations
- [ ] Set up support channels

### 13.3 Post-Launch
- [ ] Address bugs and issues
- [ ] Gather performance metrics
- [ ] Plan for future iterations
- [ ] Implement feedback from users

---

## Success Criteria

### Functional Requirements
- ✅ All API endpoints operational
- ✅ Optimistic locking prevents data conflicts
- ✅ JSON Schema validation enforced
- ✅ Role-based access control working
- ✅ Config versioning and rollback functional
- ✅ Raft consensus ensures strong consistency

### Non-Functional Requirements
- ✅ API response time < 100ms (p95)
- ✅ System availability > 99.9%
- ✅ Config read throughput > 10k req/s
- ✅ Database queries < 50ms (p95)
- ✅ Test coverage > 80%
- ✅ Zero downtime deployments

---

## Timeline Estimate

| Phase | Estimated Duration |
|-------|-------------------|
| Phase 1-2: Setup & Database | 1 week |
| Phase 3-4: Domain & Application | 2 weeks |
| Phase 5: Raft Integration | 2 weeks |
| Phase 6: HTTP API | 1 week |
| Phase 7: Observability | 1 week |
| Phase 8: Security | 1 week |
| Phase 9: Testing | 2 weeks |
| Phase 10-11: Docs & Deployment | 1 week |
| Phase 12-13: Optimization & Launch | 1 week |
| **Total** | **12 weeks** |

---

## Hexagonal Architecture Flow

### Request Flow Example
```
HTTP Request (POST /v1/projects/{id}/configs)
  ↓
1. inbound/http/handlers/config_handler.go
   - Validates HTTP request
   - Extracts parameters
   - Calls use case
  ↓
2. usecases/config/create_config.go
   - Contains business logic
   - Validates business rules
   - Calls domain services
   - Calls outbound ports
  ↓
3. ports/outbound/config_repository.go (interface)
   - Defines what operations are needed
  ↓
4. adapters/outbound/postgres/config_repository.go
   - Implements the interface
   - Handles database operations
  ↓
Database (PostgreSQL)
```

### Dependency Rules
1. **Domain Layer** (`domain/`)
   - NO dependencies on other layers
   - Pure business logic
   - No infrastructure concerns
   - Framework-agnostic

2. **Use Cases Layer** (`usecases/`)
   - Depends ONLY on domain layer and ports (interfaces)
   - Orchestrates business logic
   - Calls domain services
   - Uses repository interfaces

3. **Ports Layer** (`ports/`)
   - Defines interfaces only
   - Split into inbound (driven by) and outbound (drives)
   - No implementations

4. **Adapters Layer** (`adapters/`)
   - Implements port interfaces
   - **Inbound**: HTTP, gRPC (what drives us)
   - **Outbound**: Database, Raft, Cache (what we drive)
   - Converts between external formats and domain models

5. **Infrastructure Layer** (`infrastructure/`)
   - Cross-cutting concerns
   - Configuration, logging, telemetry
   - Used by all layers

### Key Principles
- **Inbound adapters** call use cases (driving side)
- **Use cases** call outbound ports (driven side)
- **Outbound adapters** implement outbound ports
- **Domain never depends on adapters**
- **Dependencies point inward** (toward domain)

---

## Notes

- Follow hexagonal architecture principles strictly
- Use dependency injection for testability
- Keep domain layer pure (no infrastructure dependencies)
- Prioritize code readability and maintainability
- Write tests alongside implementation
- Document architectural decisions
- Regular code reviews
- Continuous integration from day one
- Each use case should be a single file with clear responsibility
- Interfaces (ports) define contracts, implementations (adapters) fulfill them
- Infrastructure setup happens in `cmd/server/main.go` via dependency injection

---

**Last Updated**: 2025-12-02

