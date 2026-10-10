# Project Folder Structure & Architecture Layout

This project implements **Hexagonal Architecture** (also known as **Ports and Adapters** / **Clean Architecture**) in Go. The architecture enforces strict layer separation, dependency inversion, and business logic isolation from delivery mechanisms and persistent storage engines.

---

## 1. Architectural Philosophy & Dependency Rule

The core architectural invariant is the **Dependency Inversion Principle**:
- **Domain** is at the center: it has **zero external dependencies** (no HTTP, no database drivers, no ORM annotations, no frameworks).
- **Use Cases** orchestrate domain operations and define application behavior.
- **Adapters** (HTTP, PostgreSQL, etc.) live at the outer rim and depend inward on ports (interfaces) declared by the domain or application layers.
- **cmd/server** acts as the *Composition Root*: it instantiates concrete implementations, injects dependencies, and configures the runtime lifecycle.

```mermaid
graph TD
    subgraph Driving_Adapters [Driving Adapters (Primary)]
        HTTP[HTTP REST API / Chi Router]
        Swagger[Swagger UI & OpenAPI Spec]
    end

    subgraph Core [Application Core]
        UC[Use Cases / Application Services]
        DOM[Domain Models, Invariants & Ports]
    end

    subgraph Driven_Adapters [Driven Adapters (Secondary)]
        PG[(PostgreSQL / pgx & sqlc)]
    end

    HTTP -->|Calls| UC
    UC -->|Operates on| DOM
    PG -->|Implements Ports of| DOM
    UC -->|Invokes Repository Ports| DOM
```

---

## 2. Directory Tree Overview

```text
.
├── .air.toml                      # Hot-reload configuration for local development
├── .dockerignore                  # Docker build context exclusions
├── .editorconfig                  # Uniform indentation and formatting across IDEs
├── .env.example                   # Local environment variable template
├── .gitattributes                 # Git line-ending and diff configurations
├── .gitignore                     # Git tracking exclusions
├── .golangci.yml                  # Static analysis and linter configuration (golangci-lint)
├── .mockery.yaml                  # Mockery configuration for interface mocking
├── Dockerfile                     # Multi-stage production container build (scratch-based)
├── Dockerfile.e2e                 # Dedicated container build for CI/local E2E suites
├── README.md                      # Project documentation and quickstart guide
├── docker-compose.yml             # Local service dependencies (PostgreSQL 16)
├── docker-compose.e2e.yml         # Isolated test environment for Ginkgo E2E suites
├── go.mod                         # Go module definition and tool dependencies
├── go.sum                         # Checksums for Go module dependencies
├── lefthook.yml                   # Git pre-commit and pre-push hooks
├── mise.toml                      # Hermetic toolchain and task runner orchestration
├── oapi-codegen.yml               # OpenAPI code generation rules and configuration
├── sqlc.yml                       # SQL-to-Go code generator configuration
│
├── api/                           # API contracts and specifications
│   ├── openapi.yaml               # OpenAPI 3.0 specification (Single Source of Truth)
│   └── requests.http              # HTTP client scenarios for manual/REST client testing
│
├── cmd/                           # Application entrypoints
│   └── server/                    # Main HTTP server binary
│       ├── main.go                # Composition root, dependency wiring, graceful shutdown
│       └── main_test.go           # Integration smoke tests for bootstrap and flags
│
├── database/                      # Database schema and persistence definitions
│   ├── migrations/                # Goose migration files (Raw SQL DDL)
│   │   └── 20261006111444_create_items.sql
│   └── queries/                   # Raw SQL queries compiled by sqlc
│       └── items.sql
│
├── docs/                          # Architecture, decisions, and system documentation
│   ├── decisions/                 # Architecture Decision Records (ADRs)
│   │   ├── 0001-development-environment.md
│   │   ├── 0002-openapi-spec-first-architecture.md
│   │   ├── 0003-postgresql-persistence-layer.md
│   │   ├── 0004-configuration-management.md
│   │   ├── 0005-layer-simplification-and-mapping.md
│   │   └── 0006-local-development-air-delve.md
│   ├── techdebts/                 # Tracked technical debts and resolutions
│   │   └── 0001-ci-cache-poisoning.md
│   ├── folder-structure.md        # Detailed directory breakdown (this document)
│   ├── tech-stack.md              # Technical stack, libraries, and tool justification
│   ├── pipeline.md                # CI/CD and container security pipeline architecture
│   └── template-guide.md          # Guide for using this repository as a service template
│
├── internal/                      # Private application code (not importable by external repos)
│   ├── adapter/                   # Outer adapters (infrastructure & transport)
│   │   ├── httpapi/               # HTTP REST transport adapter
│   │   │   ├── handler/           # HTTP handlers implementing OpenAPI StrictServer interface
│   │   │   │   ├── healthcheck.go
│   │   │   │   ├── healthcheck_test.go
│   │   │   │   ├── item.go
│   │   │   │   ├── item_test.go
│   │   │   │   └── helper_test.go
│   │   │   ├── problem/           # RFC 7807 / RFC 9457 Problem Details error formatter
│   │   │   │   ├── problem.go
│   │   │   │   └── problem_test.go
│   │   │   ├── swagger/           # Embedded, offline Swagger UI assets
│   │   │   │   ├── swagger.go
│   │   │   │   └── assets/        # HTML, JS, CSS files for Swagger UI
│   │   │   ├── router.go          # Chi router setup, middleware chaining, route mounting
│   │   │   ├── router_test.go     # Router middleware and routing tests
│   │   │   ├── security.go        # Security header and CORS definitions
│   │   │   └── server.go          # HTTP server wrapper with timeout configurations
│   │   └── postgres/              # PostgreSQL persistence adapter
│   │       ├── item.go            # ItemRepository implementation using pgxpool & sqlc
│   │       ├── item_test.go       # Repository unit and mapping tests
│   │       ├── postgres.go        # pgxpool connection lifecycle and health check
│   │       └── postgres_test.go   # Connection error and ping tests
│   │
│   ├── config/                    # Configuration management and validation
│   │   ├── config.go              # Env struct tags mapped with caarlos0/env/v11
│   │   ├── config_test.go         # Default value and validation failure tests
│   │   └── validate.go            # Custom struct validation using go-playground/validator
│   │
│   ├── domain/                    # Pure domain models and ports (interfaces)
│   │   ├── errors.go              # Sentinel domain errors (ErrInvalidItem, ErrItemAlreadyExists, etc.)
│   │   ├── item.go                # Domain entity, business invariants, and Repository/Service ports
│   │   └── item_test.go           # Domain constructor and invariant unit tests
│   │
│   ├── gen/                       # Generated code (NEVER edit manually)
│   │   ├── mocks/                 # Auto-generated unit test mocks (Mockery)
│   │   ├── openapi/               # Auto-generated OpenAPI models & server interfaces
│   │   └── sqlc/                  # Auto-generated type-safe PostgreSQL queries
│   │
│   └── usecase/                   # Application business use cases
│       ├── item.go                # ItemService implementation and orchestration
│       └── item_test.go           # Use case unit tests using domain mocks
│
├── tests/                         # Integration, Architecture, and End-to-End suites
│   ├── arch/                      # Hexagonal Architecture layer boundary enforcement tests
│   │   └── architecture_test.go   # Pure Go parser tests verifying layer isolation
│   └── e2e/                       # Ginkgo v2 & Gomega BDD end-to-end tests
│       ├── e2e_suite_test.go      # Ginkgo suite bootstrap and lifecycle hooks
│       ├── healthz_test.go        # Health probe verification
│       └── item_test.go           # CRUD lifecycle E2E scenarios against running service
│
└── .vscode/                       # VS Code workspace settings and debugging launches
    ├── launch.json                # Delve debug targets (server, unit tests, E2E)
    ├── settings.json              # Go extension and linter configurations
    └── extensions.json            # Recommended developer extensions
```

---

## 3. Layer Breakdown & Responsibilities

### 3.1. `cmd/server/` (Composition Root)
- The single entry point for the service.
- **Responsibilities**:
  1. Loads and validates environment variables via `internal/config`.
  2. Initializes structured logging (`slog`) and observability.
  3. Initializes the database connection pool (`internal/adapter/postgres`).
  4. Wires concrete adapters to use case implementations and use cases to HTTP handlers.
  5. Instantiates the HTTP server with strict read, write, and idle timeouts.
  6. Listens for OS termination signals (`SIGINT`, `SIGTERM`) and performs orderly graceful shutdown, draining active HTTP connections before closing database pools.

### 3.2. `internal/domain/` (The Core)
- Holds pure domain entities, value objects, and business rules.
- **Strict Invariants**:
  - Contains **zero third-party imports** (standard library only, except UUID if used as an identity primitive).
  - Domain constructors (`NewItem(...)`) validate business invariants (e.g., non-empty strings, string length boundaries, data trimming, non-negative price).
  - Declares **Ports** (interfaces):
    - `ItemRepository`: Contract for storage operations.
    - `ItemService`: Contract for business use cases.
  - Declares **Sentinel Domain Errors**: `ErrInvalidItem`, `ErrItemAlreadyExists`, `ErrItemNotFound`.

### 3.3. `internal/usecase/` (Application Layer)
- Implements application business logic and orchestrates domain operations.
- Interacts strictly with repository interfaces (`domain.ItemRepository`).
- Contains no HTTP or SQL details; unit-tested thoroughly using generated mocks from `internal/gen/mocks`.

### 3.4. `internal/adapter/httpapi/` (Driving HTTP Adapter)
- Exposes the service via HTTP REST endpoints.
- **Components**:
  - `router.go`: Configures Chi router with standard middleware stack (Request ID, Real IP, Structured Slog Logger, Panic Recoverer, CORS, Security Headers).
  - `handler/`: Implements the `StrictServerInterface` generated by `oapi-codegen`. Handlers map OpenAPI request types to domain/use case calls and translate domain errors into RFC 7807 / RFC 9457 Problem Details responses.
  - `problem/`: Standardized error response formatter producing `application/problem+json`.
  - `swagger/`: Serves embedded static Swagger UI assets and the raw `openapi.yaml` specification directly from binary memory (zero external CDN reliance).

### 3.5. `internal/adapter/postgres/` (Driven Persistence Adapter)
- Implements repository interfaces using `pgxpool.Pool` and `sqlc`-generated queries.
- Manages PostgreSQL error mapping (e.g., mapping Postgres error code `23505` to `domain.ErrConflict`).
- Maps raw database rows directly into domain entities (`toDomainItem`).

### 3.6. `internal/config/` (Configuration)
- Manages application configuration loaded strictly from environment variables using `caarlos0/env/v11`.
- Enforces runtime assertions on startup using `go-playground/validator/v10` (e.g., port ranges, valid database URI schemes, positive pool sizes).

### 3.7. `internal/gen/` (Generated Artifacts)
- Houses automatically generated code maintained under strict drift detection in CI:
  - `gen/openapi/`: Types, request/response models, and Chi HTTP server interfaces generated from `api/openapi.yaml`.
  - `gen/sqlc/`: Type-safe Go query methods generated from `database/queries/*.sql`.
  - `gen/mocks/`: Interface mocks generated by `mockery` for unit testing.

### 3.8. `database/` (Data Schemas)
- `migrations/`: Versioned SQL migration files managed by Goose.
- `queries/`: Clean, hand-crafted SQL queries annotated with sqlc tags (`-- name: ListItems :many`).

### 3.9. `tests/e2e/` (Black-Box Verification)
- Ginkgo v2 and Gomega BDD tests that treat the entire application as a black box.
- Runs against an active HTTP port and real PostgreSQL database, ensuring real-world correctness.
