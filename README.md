# Hotel Booking Platform

A clean, production-grade Go service for hotel management and reservation workflows, built with Hexagonal Architecture (ports & adapters), OpenAPI spec-first code generation, and PostgreSQL persistence with UUIDv7 primary keys.

## Architecture & Design Decisions

- [ADR-0001: Standardized Development Environment & Mise](docs/decisions/0001-development-environment.md)
- [ADR-0002: OpenAPI Spec-First Architecture (oapi-codegen)](docs/decisions/0002-openapi-spec-first-architecture.md)
- [ADR-0003: PostgreSQL Persistence with pgx/v5, sqlc, Goose, and UUIDv7](docs/decisions/0003-postgresql-persistence-layer.md)
- [ADR-0004: Configuration Management with caarlos0/env/v11](docs/decisions/0004-configuration-management.md)
- [ADR-0005: Layer Simplification & Direct Model Mapping](docs/decisions/0005-layer-simplification-and-mapping.md)
- [Implementation Review (2026-10-08)](docs/reviews/2026-10-08-implementation-review.md)

## Prerequisites

- [Mise](https://mise.jdx.dev/) (manages Go and CLI tools)
- [Docker](https://docs.docker.com/get-docker/) & Docker Compose (for local PostgreSQL)

## Quickstart

```bash
# 1. Install toolchain via mise
mise install

# 2. Copy the environment configuration template
cp .env.example .env

# 3. Start PostgreSQL
mise run docker:compose-up

# 4. Apply database migrations
mise run migrate:up

# 5. Run the server
go run ./cmd/server
```

The server starts on `http://localhost:8080`.

- Health check probe: `GET http://localhost:8080/healthz`
- Swagger UI (development only): `GET http://localhost:8080/docs/`
- OpenAPI Specification: `GET http://localhost:8080/openapi.json`
- Hotels API: `GET /api/v1/hotels`, `POST /api/v1/hotels`

## Development Commands

```bash
# Code generation (OpenAPI, SQLC, Mockery)
mise run generate

# Testing
mise run test               # Unit tests with race detector
mise run test:coverage      # Unit tests + coverage summary
mise run test:e2e           # Full end-to-end suite with compose

# Linting & Security
mise run lint               # golangci-lint (full)
mise run lint:fix           # Auto-fix linting issues
mise run format             # Format Go files with goimports
mise run secret:check       # Gitleaks detection
mise run vuln               # Govulncheck vulnerability scanner

# Git hooks (managed by lefthook)
mise run hooks:install
```

## Production Container

```bash
# Build the minimal scratch-based production image
mise run docker:build
```
