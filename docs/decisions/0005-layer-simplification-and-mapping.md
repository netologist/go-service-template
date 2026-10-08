# ADR-0005: Layer Simplification and Zero-Ceremony Mapping Architecture

## Status
Accepted

## Date
2026-10-08

## Context
In hexagonal or clean architecture codebases, teams frequently introduce boilerplate layers and mirror DTOs to preserve boundary isolation:

1. **Repetitive Model Mapping (Ceremony over Utility)**:
   A single entity (e.g., `Hotel`) was mapped across four distinct representations:
   `sqlc.Hotel` $\rightarrow$ `domain.Hotel` $\rightarrow$ `usecase.HotelOutput` / `CreateHotelOutput` $\rightarrow$ `openapi.Hotel`.
   Here, `usecase.HotelOutput`, `CreateHotelOutput`, and `CreateHotelInput` were identical structural mirrors of `domain.Hotel` with no added semantic behavior or distinct contracts.

2. **Unnecessary Heap Allocations & Pointer Semantics**:
   - Repository interfaces returned pointer slices (`[]*domain.Hotel`), requiring individual pointer allocations per item.
   - Database columns defined as `NOT NULL` (`created_at`, `updated_at`) were mapped to `*time.Time` pointers in domain models.
   - Mapping functions (`toDomainHotel`) took addresses of sqlc row fields (`&row.CreatedAt.Time`), creating unnecessary heap escapes.

3. **Driver-Specific Type Leakage (`pgtype`)**:
   By default, `sqlc` generates `pgtype.UUID` and `pgtype.Timestamptz` for PostgreSQL types. This forced adapter code to perform manual byte slice slicing and struct unwrapping (`uuid.UUID(row.ID.Bytes)`, `&row.CreatedAt.Time`).

4. **Unused Model Annotations**:
   `sqlc.yml` had `emit_json_tags: true` enabled, even though SQL models are internal persistence row representations and never serialized directly to JSON.

We needed to eliminate ceremony, reduce cognitive load and allocations, and maintain a boring, maintainable design without compromising architectural boundaries.

## Decision

We streamline internal layer boundaries and type representations through the following principles:

### 1. Remove Mirror DTOs in Use Cases
- Eliminate intermediate mirror structs (`HotelOutput`, `CreateHotelOutput`, `CreateHotelInput`) in `internal/usecase`.
- Use cases accept standard primitive/domain inputs (e.g., `CreateHotel(ctx, name, address, city string)`) and return canonical `domain.Hotel` values.
- HTTP handlers map directly between OpenAPI request/response types (`openapi.CreateHotelRequest`, `openapi.Hotel`) and `domain.Hotel`.

### 2. Value Semantics over Unnecessary Pointers
- In `domain.Hotel`, use non-pointer `time.Time` values for `CreatedAt` and `UpdatedAt` since both columns are `NOT NULL` in the schema.
- `domain.HotelRepository` signatures return and accept value types and slices (`[]domain.Hotel`, `domain.Hotel`) instead of pointer slices (`[]*domain.Hotel`).
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
        S1["sqlc.Hotel (pgtype)"] -->|toDomainHotel| D1["*domain.Hotel (*time.Time)"]
        D1 -->|toOutput| U1["usecase.HotelOutput"]
        U1 -->|toHotelResponse| O1["openapi.Hotel"]
    end

    subgraph "After (Boring & Direct)"
        S2["sqlc.Hotel (native uuid & time.Time)"] -->|toDomainHotel (value copy)| D2["domain.Hotel (time.Time)"]
        D2 -->|toHotelResponse| O2["openapi.Hotel"]
    end
```

## Consequences

### Positive
- **Reduced Ceremony & Code Volume**: Deleted redundant structs and mapper functions in the usecase layer.
- **Improved Performance**: Switched from pointer slices and field addresses to value semantics, avoiding heap allocations on query results.
- **Cleaner Persistence Adapter**: Database repository code maps directly between `sqlc.Hotel` and `domain.Hotel` using native Go types with zero driver-specific casting.
- **Lower Cognitive Load**: New developers only navigate `domain.Hotel` and `openapi.Hotel`, rather than three competing struct variants.

### Negative
- **Handler Direct Coupling to Domain Entity**: Handlers now import `domain` to read `domain.Hotel` returned by use cases. This is an accepted and standard practice in hexagonal Go services.

## Alternatives Considered

1. **Keep Usecase DTOs**:
   - *Rejected*: In CRUD and standard enterprise workflows where usecase outputs exactly mirror domain models, duplicate DTOs produce pure boilerplate without safety or abstraction advantages.
2. **Keep `*time.Time` in Domain**:
   - *Rejected*: Schema guarantees `NOT NULL`. Pointers introduce nil-dereference bugs and heap allocation overhead.

## References
- ADR-0002: OpenAPI Specification-First Architecture with oapi-codegen and Chi
- ADR-0003: PostgreSQL Persistence Layer Architecture with pgx/v5, sqlc, goose, and UUIDv7
- Review Findings: `docs/reviews/2026-10-08-implementation-review.md` (A1)
