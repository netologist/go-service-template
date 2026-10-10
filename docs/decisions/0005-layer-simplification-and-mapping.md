# ADR-0005: Layer Simplification and Zero-Ceremony Mapping Architecture

## Status
Accepted

## Date
2026-10-08

## Context
In hexagonal or clean architecture codebases, teams frequently introduce boilerplate layers and mirror DTOs to preserve boundary isolation:

1. **Repetitive Model Mapping (Ceremony over Utility)**:
   A single entity (e.g., `Item`) was mapped across four distinct representations:
   `sqlc.Item` $\rightarrow$ `domain.Item` $\rightarrow$ `usecase.ItemOutput` / `CreateItemOutput` $\rightarrow$ `openapi.Item`.
   Here, `usecase.ItemOutput`, `CreateItemOutput`, and `CreateItemInput` were identical structural mirrors of `domain.Item` with no added semantic behavior or distinct contracts.

2. **Unnecessary Heap Allocations & Pointer Semantics**:
   - Repository interfaces returned pointer slices (`[]*domain.Item`), requiring individual pointer allocations per item.
   - Database columns defined as `NOT NULL` (`created_at`, `updated_at`) were mapped to `*time.Time` pointers in domain models.
   - Mapping functions (`toDomainItem`) took addresses of sqlc row fields (`&row.CreatedAt.Time`), creating unnecessary heap escapes.

3. **Driver-Specific Type Leakage (`pgtype`)**:
   By default, `sqlc` generates `pgtype.UUID` and `pgtype.Timestamptz` for PostgreSQL types. This forced adapter code to perform manual byte slice slicing and struct unwrapping (`uuid.UUID(row.ID.Bytes)`, `&row.CreatedAt.Time`).

4. **Unused Model Annotations**:
   `sqlc.yml` had `emit_json_tags: true` enabled, even though SQL models are internal persistence row representations and never serialized directly to JSON.

We needed to eliminate ceremony, reduce cognitive load and allocations, and maintain a boring, maintainable design without compromising architectural boundaries.

## Decision

We streamline internal layer boundaries and type representations through the following principles:

### 1. Remove Mirror DTOs in Use Cases
- Eliminate intermediate mirror structs (`ItemOutput`, `CreateItemOutput`, `CreateItemInput`) in `internal/usecase`.
- Use cases accept standard primitive/domain inputs (e.g., `CreateItem(ctx, name, description string, price int64)`) and return canonical `domain.Item` values.
- HTTP handlers map directly between OpenAPI request/response types (`openapi.CreateItemRequest`, `openapi.Item`) and `domain.Item`.

### 2. Value Semantics over Unnecessary Pointers
- In `domain.Item`, use non-pointer `time.Time` values for `CreatedAt` and `UpdatedAt` since both columns are `NOT NULL` in the schema.
- `domain.ItemRepository` signatures return and accept value types and slices (`[]domain.Item`, `domain.Item`) instead of pointer slices (`[]*domain.Item`).
- Avoid taking field addresses of row structs during mapping, eliminating heap escapes.

### 3. Native Type Overrides in `sqlc.yml`
- Configure `overrides` in `sqlc.yml` to bind database types directly to standard Go types:
  ```yaml
  overrides:
    - db_type: "uuid"
      go_type: "github.com/google/uuid.UUID"
    - db_type: "timestamptz"
      go_type: "time.Time"
  ```
- Remove `emit_json_tags: true` from `sqlc.yml`.
- `sqlc` generates models and query parameters directly using `uuid.UUID` and `time.Time`, eliminating all `pgtype` boilerplate.

```mermaid
graph LR
    subgraph "Before (4 Layers of Mapping)"
        S1["sqlc.Item (pgtype)"] -->|toDomainItem| D1["*domain.Item (*time.Time)"]
        D1 -->|toOutput| U1["usecase.ItemOutput"]
        U1 -->|toItemResponse| O1["openapi.Item"]

    subgraph "After (Boring & Direct)"
        S2["sqlc.Item (native uuid & time.Time)"] -->|toDomainItem (value copy)| D2["domain.Item (time.Time)"]
        D2 -->|toItemResponse| O2["openapi.Item"]
    end
```

## Consequences

### Positive
- **Reduced Ceremony & Code Volume**: Deleted redundant structs and mapper functions in the usecase layer.
- **Improved Performance**: Switched from pointer slices and field addresses to value semantics, avoiding heap allocations on query results.
- **Cleaner Persistence Adapter**: Database repository code maps directly between `sqlc.Item` and `domain.Item` using native Go types with zero driver-specific casting.
- **Lower Cognitive Load**: New developers only navigate `domain.Item` and `openapi.Item`, rather than three competing struct variants.

### Negative
- **Handler Direct Coupling to Domain Entity**: Handlers now import `domain` to read `domain.Item` returned by use cases. This is an accepted and standard practice in hexagonal Go services.

## Alternatives Considered

1. **Keep Usecase DTOs**:
   - *Rejected*: In CRUD and standard enterprise workflows where usecase outputs exactly mirror domain models, duplicate DTOs produce pure boilerplate without safety or abstraction advantages.
2. **Keep `*time.Time` in Domain**:
   - *Rejected*: Schema guarantees `NOT NULL`. Pointers introduce nil-dereference bugs and heap allocation overhead.

## References
- ADR-0002: OpenAPI Specification-First Architecture with oapi-codegen and Chi
- ADR-0003: PostgreSQL Persistence Layer Architecture with pgx/v5, sqlc, goose, and UUIDv7
- Review Findings: `docs/reviews/2026-10-08-implementation-review.md` (A1)
