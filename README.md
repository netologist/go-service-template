# Go Service Template

A production-grade, battle-tested Go service template built with Hexagonal Architecture (ports & adapters), OpenAPI spec-first code generation, PostgreSQL persistence with UUIDv7, hermetic toolchains (Mise), and a zero-trust CI/CD pipeline. Includes a fully implemented, working reference domain (Items Catalog) demonstrating the architecture end-to-end.

## Architecture & Design Decisions

- [ADR-0001: Standardized Development Environment & Mise](docs/decisions/0001-development-environment.md)
- [ADR-0002: OpenAPI Spec-First Architecture (oapi-codegen)](docs/decisions/0002-openapi-spec-first-architecture.md)
- [ADR-0003: PostgreSQL Persistence with pgx/v5, sqlc, Goose, and UUIDv7](docs/decisions/0003-postgresql-persistence-layer.md)
- [ADR-0004: Configuration Management with caarlos0/env/v11](docs/decisions/0004-configuration-management.md)
- [ADR-0005: Layer Simplification & Direct Model Mapping](docs/decisions/0005-layer-simplification-and-mapping.md)
- [ADR-0006: Local Development Workflow with Air Live-Reload and VS Code Delve Debugging](docs/decisions/0006-local-development-air-delve.md)

## Documentation

- [Project Folder Structure](docs/folder-structure.md) - Detailed breakdown of hexagonal architecture layers, responsibilities, and directory layout.
- [Technology Stack & Tools](docs/tech-stack.md) - Rationale, ecosystem tools, runtime libraries, and operational stack.
- [CI/CD & Security Pipeline](docs/pipeline.md) - Zero-trust GitHub Actions pipeline, drift detection, security gates, and provenance.
- [Template Usage & Adoption Guide](docs/template-guide.md) - How to instantiate and tailor this repository as a production service template.

## Creating a New Service from this Template

You can instantiate a new microservice from this template using either the official Go `gonew` tool or GitHub's **Use this template** button. Because `gonew` only replaces Go source imports, use `mise run template:init` to ensure all configurations (YAML, TOML, CI workflows, and code generators) are synchronized.

### Option A: Using `gonew` + `mise run template:init` (CLI Workflow)

```bash
# 1. Install gonew (official Go project templating tool)
go install golang.org/x/tools/cmd/gonew@latest

# 2. Instantiate a new service directory from this template
gonew github.com/netologist/go-service-template github.com/your-org/your-service

# 3. Enter the new project directory
cd your-service

# 4. Install toolchain via mise
mise install

# 5. Run template:init to sync non-Go configurations (.golangci.yml, .mockery.yaml, CI, mise.toml)
mise run template:init github.com/your-org/your-service
```

### Option B: Using GitHub Web + `mise run template:init`

```bash
# 1. Click "Use this template" on GitHub and create your new repo
git clone https://github.com/your-org/your-service.git
cd your-service

# 2. Install toolchain via mise
mise install

# 3. Run template:init to replace all module paths, regenerate code, and verify tests
mise run template:init github.com/your-org/your-service
```

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
- Items API: `GET /api/v1/items`, `POST /api/v1/items`

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

# Documentation (MkDocs Material)
mise run docs:serve          # Local documentation server with live-reload
mise run docs:build          # Build production static documentation site

# Git hooks (managed by lefthook)
mise run hooks:install
```

## Production Container

```bash
# Build the minimal scratch-based production image
mise run docker:build
```
