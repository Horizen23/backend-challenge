# Backend Golang Coding Test Solution

This repository contains the complete implementation of the **User Management API & gRPC Services** and the comprehensive design document and Proof of Concept (POC) for the **Lottery Search System**.

The project structure strictly adopts clean architecture / DDD design standards (`domain`, `usecase`, `infrastructure`, `presentation` with Echo framework, DTOs, Mappers, Google Wire, and Mockery/Goverter code generation).

---

## Part 1: User Management API & gRPC Architecture

The project is structured as multi-service microservices using Go Workspaces (`go.work`), with a shared contract module (`apps/common/`) and clear separation of concerns (API Gateway / BFF Pattern where `gateway-api` communicates exclusively via gRPC without direct database access):

```
.
├── apps/
│   ├── common/                             # Shared Protobuf Contracts & Core Utilities
│   │   ├── pkg/logger/                     # Shared Zap logger
│   │   ├── pkg/ctxutils/                   # Shared context propagation
│   │   └── proto/
│   │       ├── pb/                         # Compiled Go protobuf & gRPC stubs
│   │       └── user.proto                  # Protobuf definition
│   │
│   ├── gateway-api/                        # HTTP REST API Microservice (Port 8080)
│   │   │                                   # NO DIRECT DB ACCESS: Calls user-service via gRPC Client
│   │   ├── cmd/api/                        # Bootstrap, Wire DI, graceful shutdown
│   │   ├── internal/                       # REST Clean Architecture
│   │   │   ├── config/                     # Configuration loader (GRPC_SERVER_ADDR, JWT)
│   │   │   ├── domain/user/                # User entity
│   │   │   ├── infrastructure/
│   │   │   │   ├── auth/                   # JWT Provider & bcrypt
│   │   │   │   └── grpc/client/            # gRPC Client connection to user-service
│   │   │   ├── presentation/               # Echo HTTP handlers, DTOs, mappers, Swagger, 10s worker
│   │   │   └── usecase/                    # REST User business usecases (delegating to gRPC)
│   │   ├── tests/mocks/                    # Mockery generated mock interfaces
│   │   ├── Taskfile.yml & Dockerfile       # Self-contained build & task definitions
│   │   └── go.mod                          # Standalone Go module
│   │
│   └── user-service/                       # gRPC Microservice (Port 50051)
│       │                                   # PERSISTENCE OWNER: Manages MongoDB connection & indexes
│       ├── cmd/grpc/                       # Bootstrap, Wire DI, graceful shutdown
│       ├── internal/                       # gRPC Clean Architecture
│       │   ├── config/                     # Configuration loader (MongoURI, DBName)
│       │   ├── domain/user/                # User entity & repository contract
│       │   ├── infrastructure/             # MongoDB repo & Goverter model mapper
│       │   ├── presentation/grpc/          # gRPC Server implementation & in-memory test
│       │   └── usecase/                    # gRPC-focused usecase
│       ├── tests/mocks/                    # Mockery generated mock interfaces
│       ├── Taskfile.yml & Dockerfile       # Self-contained build & task definitions
│       └── go.mod                          # Standalone Go module
│
├── docs/
│   └── lottery/
│       ├── system_design/                  # Lottery search system design proposal
│       │   └── README.md                   # Architecture & Event-Driven Diagrams
│       └── poc/                            # Executable Redis Bitmap & Lua Proof of Concept
│           ├── go.mod                      # POC module
│           ├── main.go                     # POC entrypoint
│           └── engine/                     # Bitmap matching, pagination & lock engine
│
├── docker-compose.yml                      # Runs mongodb, user-service (50051), and gateway-api (8080)
├── Taskfile.yml                            # Root automation orchestrator
├── go.work & go.work.sum                   # Go Workspace referencing common, gateway-api, user-service, poc
```

---

### Features & Design Decisions

1. **Clean Microservice Separation & Zero-Trust Internal Boundary**:
   - `gateway-api` (`apps/gateway-api`): Public API Gateway / BFF using `Echo v4`. Validates request schemas, handles JWT Bearer authentication, and communicates exclusively with `user-service` via gRPC without touching the database directly.
   - `user-service` (`apps/user-service`): Internal High-Throughput Core Service. Exclusive owner of MongoDB persistence, unique email indexing, and domain business rules.
   - **Internal Security & Context Propagation**: Because `user-service` resides within a private internal network, `gateway-api` authenticates public callers and propagates the verified `UserID`, `Email`, and `X-Request-ID` over gRPC Metadata headers. `user-service` extracts these identity attributes via `MetadataServerInterceptor` for audit logging and tracing, eliminating redundant cryptographic JWT validation overhead across internal service calls.
2. **Compile-Time Code Generators**:
   - `goverter`: Generates type-safe struct mappers without reflection (DTOs <-> Domain <-> Proto requests <-> Mongo model).
   - `mockery`: Type-safe mock interfaces with expectation assertions (`tests/mocks/`).
   - `wire`: Compile-time Dependency Injection without reflection.
3. **MongoDB Persistence**:
   - Official Mongo Go Driver (`go.mongodb.org/mongo-driver`).
   - Unique index auto-creation on `email`.
4. **Swagger OpenAPI Documentation**:
   - Integrated with Echo Swagger UI available at `http://localhost:8080/swagger/index.html` (under `-tags swagger`).
5. **Graceful Shutdown**:
   - OS signal trapping (`SIGINT`, `SIGTERM`) on both services with context timeout draining.

---

### Getting Started

#### Run with Docker Compose (Recommended)

Starts MongoDB, HTTP API (`:8080`), and gRPC service (`:50051`):
```bash
task docker:up
# or: docker compose up --build -d
```

Check container logs:
```bash
task docker:logs
```

Stop containers:
```bash
task docker:down
```

#### Run with Taskfile (`task`)

- `task test` : Run all unit tests with race detection (`-race`)
- `task build` : Build both `api` and `grpc-service` binaries
- `task generate` : Trigger all generators (`goverter`, `mockery`, `wire`)
- `task run:api` : Run HTTP API server locally on port `8080`
- `task run:swagger` : Run HTTP API server with Swagger UI enabled on `http://localhost:8080/swagger/index.html`
- `task run:grpc` : Run gRPC service locally on port `50051`
- `task clean` : Clean build artifacts

---

### API Endpoints & Samples

#### HTTP Endpoints (`:8080`)

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/auth/register` | Register new user and return JWT |
| `POST` | `/api/auth/login` | Authenticate user and return JWT |
| `GET` | `/healthz` | Health check endpoint |
| `POST` | `/api/users` | Create user (Protected) |
| `GET` | `/api/users` | List all users (Protected) |
| `GET` | `/api/users/:id` | Get user by ID (Protected) |
| `PUT` | `/api/users/:id` | Update user (Protected) |
| `DELETE` | `/api/users/:id` | Delete user (Protected) |

**Sample Auth Register:**
```bash
curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "mypassword123"
  }'
```

---

#### gRPC Service (`:50051`)

Supports `pb.UserService`:
- `CreateUser(CreateUserRequest) returns (UserResponse)`
- `GetUser(GetUserRequest) returns (UserResponse)`
- `ListUsers(ListUsersRequest) returns (ListUsersResponse)`
- `UpdateUser(UpdateUserRequest) returns (UserResponse)`
- `DeleteUser(DeleteUserRequest) returns (DeleteUserResponse)`
- `CountUsers(CountUsersRequest) returns (CountUsersResponse)`
- `VerifyPassword(VerifyPasswordRequest) returns (VerifyPasswordResponse)`

---

## Part 2: Lottery Search System Design & POC

The full technical design proposal and architecture diagrams are available in [`lottery_system_design/README.md`](lottery_system_design/README.md).

### Summary Highlights:
1. **Positional Bitmaps (60 Bitmaps)**:
   - Indexes 10M tickets using 60 bitmaps (10 digits * 6 positions) representing only **~75 MB** of RAM (100K tickets use only ~16 KB).
   - Any wildcard pattern (e.g. `****23`, `1****5`, `123***`) executes in **sub-millisecond (< 1ms)** via hardware-accelerated SIMD `BITOP AND`.
2. **Atomic Distribution & Anti-Collision**:
   - Single-step Redis Lua script execution combines matching with atomic leasing.
   - Automatically marks selected tickets as leased with a TTL (e.g. 5 minutes).
   - Guarantees zero duplicate allocations to concurrent users searching the exact same pattern simultaneously.
3. **Cursor-based Pagination**:
   - Instant count of matching tickets using `BITCOUNT`.
   - Streaming pagination using `BITPOS` starting from the cursor offset without full scans.

### Run the POC:
```bash
cd lottery_system_design/poc && go run main.go
```
