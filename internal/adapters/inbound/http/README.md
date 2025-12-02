# HTTP API Layer

Complete REST API implementation for GoConfig Guardian with chi router, comprehensive middleware, and OpenAPI 3.0 specification.

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [API Endpoints](#api-endpoints)
- [Handlers](#handlers)
- [Middleware Stack](#middleware-stack)
- [Authentication](#authentication)
- [Authorization](#authorization)
- [Error Handling](#error-handling)
- [Request Flow](#request-flow)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Request                             │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                  Middleware Stack                           │
│  1. RequestID    - Generate unique request ID               │
│  2. Recovery     - Panic recovery                           │
│  3. Logging      - Structured logging (slog)                │
│  4. CORS         - Cross-origin resource sharing            │
│  5. Compress     - Response compression                     │
│  6. Timeout      - Request timeout (60s)                    │
│  7. RateLimit    - IP-based rate limiting                   │
│  8. Auth         - JWT validation (protected routes)        │
│  9. Authorization - RBAC permission check                   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    Router (chi)                             │
│  - Route matching                                           │
│  - URL parameter extraction                                 │
│  - Handler dispatch                                         │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Handlers                            │
│  - Request validation                                       │
│  - Parameter extraction                                     │
│  - Use case execution                                       │
│  - Response formatting                                      │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    Use Cases                                │
│  - Business logic orchestration                             │
│  - Domain service calls                                     │
│  - Repository operations                                    │
└─────────────────────────────────────────────────────────────┘
```

---

## API Endpoints

### Authentication (Public)

```
POST   /api/v1/auth/register     Register new user
POST   /api/v1/auth/login        Login and get JWT token
```

### Users (Protected - Admin)

```
GET    /api/v1/users             List all users
POST   /api/v1/users             Create user
GET    /api/v1/users/{userId}    Get user details
DELETE /api/v1/users/{userId}    Delete user
```

### Projects (Protected)

```
GET    /api/v1/projects                List projects
POST   /api/v1/projects                Create project
GET    /api/v1/projects/{projectId}    Get project
DELETE /api/v1/projects/{projectId}    Delete project
```

### Roles (Protected - Project-scoped)

```
POST   /api/v1/projects/{projectId}/roles            Assign role (Admin)
GET    /api/v1/projects/{projectId}/roles            List roles (Viewer+)
DELETE /api/v1/projects/{projectId}/roles/{userId}   Revoke role (Admin)
```

### Schemas (Protected - Global)

```
GET    /api/v1/schemas             List schemas
POST   /api/v1/schemas             Create schema (Admin)
PUT    /api/v1/schemas/{schemaId}  Update schema (Admin)
DELETE /api/v1/schemas/{schemaId}  Delete schema (Admin)
```

### Configs (Protected - Project-scoped)

```
GET    /api/v1/projects/{projectId}/configs                  List configs (Viewer+)
POST   /api/v1/projects/{projectId}/configs                  Create config (Editor+)
GET    /api/v1/projects/{projectId}/configs/{key}            Get config (Viewer+)
PUT    /api/v1/projects/{projectId}/configs/{key}            Update config (Editor+) *
DELETE /api/v1/projects/{projectId}/configs/{key}            Delete config (Admin)
POST   /api/v1/projects/{projectId}/configs/{key}/rollback   Rollback config (Admin) *
```

**\* Requires optimistic locking (expected_version)**

### Read API (Public - API Key)

```
GET    /api/v1/read/{apiKey}/{key}    Read config by API key
```

### Health & Status

```
GET    /health                     Health check
GET    /                           Service info
```

---

## Handlers

### Handler Structure

All handlers follow a consistent pattern:

```go
type XxxHandler struct {
    createUseCase *xxx.CreateXxxUseCase
    listUseCase   *xxx.ListXxxUseCase
    // ... other use cases
}

func (h *XxxHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    var req xxx.CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        common.BadRequest(w, "Invalid request body")
        return
    }
    
    // 2. Execute use case
    resp, err := h.createUseCase.Execute(r.Context(), req)
    if err != nil {
        common.BadRequest(w, err.Error())
        return
    }
    
    // 3. Return response
    common.Created(w, resp)
}
```

### Handler Files

| File | Endpoints | Purpose |
|------|-----------|---------|
| `auth_handler.go` | 2 endpoints | User registration & login |
| `user_handler.go` | 4 endpoints | User CRUD operations |
| `project_handler.go` | 4 endpoints | Project CRUD operations |
| `role_handler.go` | 2 endpoints | Role assignment & revocation |
| `schema_handler.go` | 4 endpoints | Schema management |
| `config_handler.go` | 5 endpoints | Config CRUD + rollback |
| `read_handler.go` | 1 endpoint | Public client API |

**Total: 22 endpoints**

---

## Middleware Stack

### Global Middleware (All Routes)

Applied in order:

1. **RequestID** - Generates unique ID for each request
   - Checks `X-Request-ID` header
   - Generates UUID if missing
   - Adds to context and response header

2. **Recovery** - Catches panics and returns 500
   - Logs panic with stack trace
   - Prevents server crash
   - Returns JSON error response

3. **Logging** - Structured request/response logging
   - Logs request start (method, path, IP, user agent)
   - Logs response (status, size, duration)
   - Uses request ID for correlation

4. **CORS** - Cross-origin resource sharing
   - Allowed origins: `localhost:3000`, `localhost:8080`
   - Allowed methods: GET, POST, PUT, DELETE, OPTIONS
   - Credentials support: enabled

5. **Compress** - Response compression (gzip)
   - Compression level: 5
   - Reduces bandwidth usage

6. **Timeout** - Request timeout
   - Max duration: 60 seconds
   - Prevents long-running requests

7. **RateLimit** - IP-based rate limiting
   - Per-IP token bucket algorithm
   - Configurable RPS and burst
   - Auto-cleanup of stale limiters
   - Returns 429 Too Many Requests

### Protected Route Middleware

8. **Auth** - JWT token validation
   - Extracts Bearer token from Authorization header
   - Validates JWT signature and expiration
   - Adds user ID and email to context
   - Returns 401 Unauthorized if invalid

9. **Authorization** - Role-based access control
   - Checks user role in project context
   - Three levels: admin, editor, viewer
   - Hierarchical permissions
   - Returns 403 Forbidden if insufficient

---

## Authentication

### JWT Authentication

**Token Generation:**

```go
middleware.GenerateToken(
    userID,
    email,
    jwtSecret,
    24 * time.Hour, // 24h expiration
)
```

**Token Format:**

```json
{
  "user_id": "usr_123",
  "email": "user@example.com",
  "iss": "cfguardian",
  "iat": 1234567890,
  "exp": 1234654290
}
```

**Usage:**

```bash
curl -H "Authorization: Bearer <jwt_token>" \
     http://localhost:8080/api/v1/users
```

### API Key Authentication

**For Client Read API:**

```bash
# Method 1: In URL path
GET /api/v1/read/cfg_abc123def456/app-config

# Method 2: In header (future)
curl -H "X-API-Key: cfg_abc123def456" \
     http://localhost:8080/api/v1/read/app-config
```

---

## Authorization

### Role Levels

| Role | Permissions |
|------|-------------|
| **admin** | Full CRUD on all resources in project |
| **editor** | Read + Write configs, Read schemas |
| **viewer** | Read-only access to all resources |

### Permission Hierarchy

```
admin ⊃ editor ⊃ viewer
```

### Middleware Usage

```go
// Require admin role
r.With(middleware.RequireAdmin(authzCfg)).Delete("/", handler.Delete)

// Require editor role (or higher)
r.With(middleware.RequireEditor(authzCfg)).Put("/", handler.Update)

// Require viewer role (or higher)
r.With(middleware.RequireViewer(authzCfg)).Get("/", handler.Get)
```

### Context Access

```go
// In handlers, retrieve authenticated user
userID := middleware.GetUserID(r.Context())
email := middleware.GetUserEmail(r.Context())
requestID := middleware.GetRequestID(r.Context())
```

---

## Error Handling

### Standard Error Response

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "additional context"
  }
}
```

### HTTP Status Codes

| Code | Name | Usage |
|------|------|-------|
| 200 | OK | Successful GET/PUT |
| 201 | Created | Successful POST |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Invalid input |
| 401 | Unauthorized | Missing/invalid auth |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | **Version mismatch** |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |

### Error Helper Functions

```go
// In common package
common.BadRequest(w, "Invalid email format")
common.Unauthorized(w, "Token expired")
common.Forbidden(w, "Admin role required")
common.NotFound(w, "Config not found")
common.Conflict(w, "Version mismatch - concurrent modification")
common.InternalServerError(w, "Database error")
```

### Optimistic Locking Errors

```http
PUT /api/v1/projects/proj_123/configs/app-config
{
  "expected_version": 5,
  "content": {"port": 9090}
}

→ 409 Conflict
{
  "error": "version mismatch: expected 5, got 6",
  "code": "CONFLICT"
}
```

---

## Request Flow

### Example: Update Config with Optimistic Locking

```
1. Client sends PUT request
   ↓
2. RequestID middleware
   - Generates request ID: "req_xyz789"
   - Adds to context and response header
   ↓
3. Recovery middleware
   - Sets up panic recovery
   ↓
4. Logging middleware
   - Logs: "http request started" (method, path, IP)
   ↓
5. CORS middleware
   - Validates origin
   - Adds CORS headers
   ↓
6. Compress middleware
   - Sets up gzip compression
   ↓
7. Timeout middleware
   - Sets 60s deadline
   ↓
8. RateLimit middleware
   - Checks IP rate limit
   - Returns 429 if exceeded
   ↓
9. Auth middleware
   - Validates JWT token
   - Extracts user_id: "usr_123"
   - Adds to context
   ↓
10. Authorization middleware (RequireEditor)
    - Gets project_id from URL
    - Checks user role in project
    - Verifies editor+ permission
    - Returns 403 if insufficient
   ↓
11. Router
    - Matches route pattern
    - Extracts URL params (projectId, configKey)
    - Dispatches to handler
   ↓
12. ConfigHandler.Update()
    - Parses request body
    - Extracts expected_version
    - Calls UpdateConfigUseCase
    ↓
13. UpdateConfigUseCase.Execute()
    - Validates schema
    - Checks version (optimistic locking)
    - Applies to Raft (consensus)
    - Saves revision
    ↓
14. Raft FSM Apply
    - Replicates across cluster
    - Applies state change
    - Returns success/conflict
    ↓
15. Handler response
    - Returns 200 OK + updated config
    - OR 409 Conflict (version mismatch)
   ↓
16. Logging middleware
    - Logs: "http request completed" (status, duration)
   ↓
17. Response to client
    - JSON response
    - X-Request-ID header
    - Compressed (if applicable)
```

---

## Response Helpers

### Success Responses

```go
// 200 OK
common.OK(w, map[string]interface{}{
    "id": "123",
    "name": "example",
})

// 201 Created
common.Created(w, newResource)

// 204 No Content
common.NoContent(w)
```

### Error Responses

```go
// 400 Bad Request
common.BadRequest(w, "Invalid email format")

// 401 Unauthorized
common.Unauthorized(w, "Token expired")

// 403 Forbidden
common.Forbidden(w, "Admin role required")

// 404 Not Found
common.NotFound(w, "Resource not found")

// 409 Conflict (Optimistic Locking!)
common.Conflict(w, "Version mismatch - concurrent modification")

// 500 Internal Server Error
common.InternalServerError(w, "Database connection failed")
```

---

## Configuration

### Router Configuration

```go
cfg := http.RouterConfig{
    JWTSecret:      "your-secret-key",
    RateLimitRPS:   100,      // 100 requests per second
    RateLimitBurst: 200,      // Burst of 200 requests
    
    // Handlers
    AuthHandler:    authHandler,
    UserHandler:    userHandler,
    ProjectHandler: projectHandler,
    RoleHandler:    roleHandler,
    SchemaHandler:  schemaHandler,
    ConfigHandler:  configHandler,
    ReadHandler:    readHandler,
    
    // Authorization
    AuthorizationConfig: middleware.AuthorizationConfig{
        CheckPermission: checkPermissionUseCase,
    },
}

router := http.NewRouter(cfg)
```

### Starting the Server

```go
server := &http.Server{
    Addr:         ":8080",
    Handler:      router,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}

log.Info("Starting HTTP server", "addr", server.Addr)
if err := server.ListenAndServe(); err != nil {
    log.Error("Server failed", "error", err)
}
```

---

## Rate Limiting

### Configuration

```go
rateLimiter := middleware.NewRateLimiter(
    100,  // 100 requests per second
    200,  // Burst capacity of 200
)
```

### Behavior

- **Per-IP rate limiting** using token bucket algorithm
- **Automatic cleanup** of inactive limiters (1-minute intervals)
- **429 Too Many Requests** when limit exceeded
- **Thread-safe** with read/write locks

### Response on Rate Limit

```json
HTTP/1.1 429 Too Many Requests
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED"
}
```

---

## Security Features

### 🔐 Security Measures

1. **JWT Authentication**
   - HMAC-SHA256 signing
   - Token expiration validation
   - Secure token generation

2. **RBAC Authorization**
   - Project-scoped permissions
   - Hierarchical role levels
   - Permission checks before operations

3. **Rate Limiting**
   - Prevents DoS attacks
   - Per-IP enforcement
   - Configurable limits

4. **Panic Recovery**
   - Prevents server crashes
   - Logs stack traces
   - Returns safe error responses

5. **CORS**
   - Controlled origin access
   - Credential support
   - Preflight handling

6. **Request Timeout**
   - Prevents resource exhaustion
   - 60-second maximum
   - Graceful timeout handling

---

## Testing

### Handler Testing

```go
func TestConfigHandler_Update(t *testing.T) {
    // Setup
    mockUseCase := &MockUpdateConfigUseCase{}
    handler := handlers.NewConfigHandler(nil, nil, mockUseCase, nil, nil)
    
    // Create request
    body := `{"expected_version": 5, "content": {"port": 9090}}`
    req := httptest.NewRequest("PUT", "/api/v1/projects/proj_1/configs/app", strings.NewReader(body))
    rec := httptest.NewRecorder()
    
    // Execute
    handler.Update(rec, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

### Integration Testing

```bash
# Start server
make run

# Test endpoints
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123"}'

curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123"}'

# Use returned token
TOKEN="<jwt_token>"

curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/v1/projects
```

---

## Logging

### Log Format (JSON)

```json
{
  "time": "2024-12-02T10:30:00Z",
  "level": "INFO",
  "msg": "http request completed",
  "request_id": "req_xyz789",
  "method": "PUT",
  "path": "/api/v1/projects/proj_1/configs/app",
  "status": 200,
  "size": 256,
  "duration": "15.234ms",
  "user_id": "usr_123"
}
```

### Log Levels

- **INFO** - Normal operations (request start/complete)
- **WARN** - Rate limit exceeded, validation failures
- **ERROR** - Panics, database errors, auth failures

---

## Performance Considerations

### Request Processing Time

| Operation | Target | Notes |
|-----------|--------|-------|
| Health check | < 1ms | No DB access |
| Read config (client API) | < 10ms | Local FSM read |
| List configs | < 50ms | DB query |
| Update config | < 100ms | Raft consensus |
| Create config | < 100ms | Raft + DB write |

### Optimization Features

- ✅ **Connection pooling** - Reuses DB connections
- ✅ **Response compression** - Reduces bandwidth
- ✅ **Rate limiting** - Prevents abuse
- ✅ **Request timeout** - Prevents hanging
- ✅ **Efficient routing** - Chi radix tree
- ✅ **Local FSM reads** - No Raft consensus for reads

---

## Examples

### Create Project

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production App",
    "owner_user_id": "usr_123"
  }'

# Response: 201 Created
{
  "id": "proj_abc123",
  "name": "Production App",
  "api_key": "cfg_xyz789def456",
  "owner_user_id": "usr_123",
  "created_at": "2024-12-02T10:00:00Z",
  "updated_at": "2024-12-02T10:00:00Z"
}
```

### Update Config with Optimistic Locking

```bash
curl -X PUT http://localhost:8080/api/v1/projects/proj_123/configs/app-config \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "expected_version": 5,
    "content": {
      "port": 9090,
      "debug": false
    }
  }'

# Success: 200 OK
{
  "project_id": "proj_123",
  "key": "app-config",
  "version": 6,
  "content": {"port": 9090, "debug": false},
  "updated_at": "2024-12-02T10:05:00Z"
}

# Conflict: 409 Conflict
{
  "error": "version mismatch: expected 5, got 6",
  "code": "CONFLICT"
}
```

### Rollback Config

```bash
curl -X POST http://localhost:8080/api/v1/projects/proj_123/configs/app-config/rollback \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_version": 3,
    "expected_version": 6
  }'
```

### Read Config (Client API)

```bash
# No authentication needed - API key in path
curl http://localhost:8080/api/v1/read/cfg_xyz789def456/app-config

# Response
{
  "key": "app-config",
  "version": 6,
  "content": {
    "port": 8080,
    "debug": true
  }
}
```

---

## Files Structure

```
internal/adapters/inbound/http/
├── router.go                    # Main router setup
├── common/
│   └── response.go              # Response helpers
├── handlers/
│   ├── auth_handler.go          # Authentication
│   ├── user_handler.go          # User management
│   ├── project_handler.go       # Project management
│   ├── role_handler.go          # Role management
│   ├── schema_handler.go        # Schema management
│   ├── config_handler.go        # Config management (main!)
│   └── read_handler.go          # Client read API
└── middleware/
    ├── request_id.go            # Request ID generation
    ├── logging.go               # Structured logging
    ├── auth.go                  # JWT authentication
    ├── authorization.go         # RBAC authorization
    ├── cors.go                  # CORS headers
    ├── rate_limit.go            # Rate limiting
    └── recovery.go              # Panic recovery

api/
└── openapi.yaml                 # OpenAPI 3.0 specification
```

---

## OpenAPI Specification

**Location**: `api/openapi.yaml`

### Features

- ✅ 20+ endpoint definitions
- ✅ Request/response schemas
- ✅ Authentication schemes (JWT + API Key)
- ✅ Error response definitions
- ✅ Examples for all operations
- ✅ Security requirements per endpoint

### Usage

```bash
# Generate documentation
swagger-codegen generate -i api/openapi.yaml -l html -o docs/api

# Validate spec
swagger-cli validate api/openapi.yaml
```

---

## Dependencies

```go
// Router
github.com/go-chi/chi/v5          v5.2.3

// Middleware
github.com/go-chi/cors            v1.2.2
golang.org/x/time/rate            v0.14.0

// Authentication
github.com/golang-jwt/jwt/v5      v5.3.0

// Utilities
github.com/google/uuid            (for request IDs)
```

---

## Next Steps

**Phase 7: Observability**
- Enhanced metrics collection
- Distributed tracing
- Advanced health checks
- Prometheus integration

**Phase 8: Testing**
- Unit tests for handlers
- Integration tests for API
- Load testing
- Security testing

---

## Summary

✅ **HTTP API Layer Complete!**

- **22 endpoints** fully implemented
- **9-layer middleware stack** for security and reliability
- **OpenAPI 3.0 spec** for documentation
- **JWT + API Key authentication** for flexibility
- **RBAC authorization** for project-scoped access
- **Optimistic locking** with 409 Conflict responses
- **Rate limiting** for DoS protection
- **Structured logging** with request correlation
- **Panic recovery** for stability
- **Production-ready** error handling

The API is ready for integration and testing! 🚀

