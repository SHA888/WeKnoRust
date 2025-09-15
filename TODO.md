# Migration Plan: Move `internal/` from Go to Rust

This document outlines the plan to migrate the backend code under `internal/` from Go to Rust while preserving functionality and API compatibility. The strategy prioritizes safety, incremental migration, and observability.

## Goals
- [ ] Maintain feature parity with the current Go backend during migration.
- [ ] Keep external API contracts stable for the frontend and clients.
- [ ] Improve performance and observability where appropriate.
- [ ] Enable incremental migration with the ability to run mixed Go/Rust services during transition.

## High-Level Strategy
- [ ] Strangler Fig approach: replace components gradually behind stable HTTP/gRPC APIs.
- [ ] Introduce a Rust workspace alongside existing Go services.
- [ ] Start with leaf subsystems and stateless services, then core services.
- [ ] Keep a single source of truth for database schema and migrations.
- [ ] Maintain observability parity (tracing/logging/metrics).

## Target Rust Tech Stack
- [ ] Web framework: `axum` (async, tower-based middleware ecosystem). Alternative: `actix-web`.
- [ ] Async runtime: `tokio`.
- [ ] Config: `config` + `serde` + `dotenvy`.
- [ ] Logging/Tracing: `tracing`, `tracing-subscriber` (JSON output), `opentelemetry`.
- [ ] Database: `sqlx` (async, MySQL/Postgres support) or `sea-orm`. We currently use GORM + MySQL/Postgres (ParadeDB); `sqlx` recommended for flexibility and performance.
- [ ] Redis: `redis`/`deadpool-redis`.
- [ ] HTTP client: `reqwest`.
- [ ] JSON: `serde_json`.
- [ ] gRPC (if needed): `tonic` + `prost`.
- [ ] Testing: `tokio::test`, `insta` (snapshots), `proptest` (where helpful).
- [ ] Chinese word segmentation: `jieba-rs` or bind to a tokenizer via service call if parity with gojieba is required.

## Current Go Components in `internal/` and Rust Mapping
- [ ] `config/` (Go + Viper) -> Rust `config` + `serde` structs mirroring Go types; env override support.
- [ ] `logger/` (logrus + custom formatter) -> `tracing` with JSON formatter; map levels and fields.
- [ ] `middleware/` (Gin middlewares: auth, error) -> `tower` middlewares for `axum`.
- [ ] `router/` (Gin routes) -> `axum` routers (keep endpoints unchanged).
- [ ] `application/service/` (business logic) -> split into Rust services/modules by bounded context.
- [ ] `application/repository/` (GORM repos) -> `sqlx` queries and repository traits + impl (MySQL & Postgres support).
- [ ] `models/*` (LLM, Embedding, Rerank, Chat) -> Rust clients using `reqwest` and typed structs; keep Ollama/remote API compatibility.
- [ ] `stream/` (memory/redis stream manager) -> trait-based StreamManager + `redis` impl with TTL.
- [ ] `common/`, `types/` -> Rust `types` crate with shared models (serde) and utilities.

## Data and API Compatibility
- [ ] Keep JSON shapes identical. Generate OpenAPI from Rust, or validate against existing docs in `docs/API*.md`.
- [ ] Reuse existing proto definitions if we add gRPC edges; place `.proto` files in a shared folder and generate Go/Rust stubs.
- [ ] Preserve environment variables and defaults; ensure `.env` parity.

## Database & Migrations
- [ ] Continue to use existing SQL migrations in `migrations/`.
- [ ] Add a Rust migration runner (e.g., `sqlx migrate` or `refinery`) using the same SQL scripts.
- [ ] Validate schema compatibility in CI for both MySQL and Postgres (ParadeDB).

## Observability
- [ ] Map log fields: request_id, tenant_id, knowledge_base_id, etc.
- [ ] Add `tracing` spans for service entry points and DB calls; keep OTEL attributes already used.
- [ ] Maintain metrics surface if present; optionally add `metrics` + `prometheus`.

## Proposed Rust Workspace Layout
```
rust/
  Cargo.toml                   # workspace
  Cargo.lock
  crates/
    config/                    # config structs + loader
    logger/                    # tracing setup
    types/                     # shared models (serde)
    repos/                     # DB repositories (sqlx)
    services/                  # business logic
    models/                    # LLM/Embedding/Rerank clients
    stream/                    # stream manager (memory/redis)
    web/                       # axum router + middlewares
  apps/
    api/                       # main binary hosting HTTP APIs
```

## Migration Phases & Tasks

### Phase 0: Preparation
- [x] Create Rust workspace and base crates.
- [ ] Set up CI for Rust (fmt, clippy, tests, sqlx schema checks).
- [ ] Add shared `.env` and config files mirroring Go’s `config/config.yaml`.

### Phase 1: Foundations
- [ ] Implement `config` crate mapping current Go config structs.
- [ ] Implement `logger` crate with `tracing` JSON formatter compatible with logrus fields.
- [ ] Implement `types` crate covering core serde models (`Tenant`, `Knowledge*`, `Chunk`, etc.).

### Phase 2: Infrastructure
- [ ] Implement DB connection pool(s) and `repos` with `sqlx`.
- [ ] Port `tenant` repository: CRUD, `AdjustStorageUsed` with pessimistic lock.
- [ ] Port `stream` manager (memory + redis) with TTL semantics.

### Phase 3: Web Layer
- [ ] Implement `web` crate: `axum` router, middlewares (auth, error handler), request context propagation.
- [ ] Mirror route structure from `internal/router/router.go` and handlers.
- [ ] Ensure JSON response shapes match Go versions.

### Phase 4: Services
- [ ] Port `knowledgebase` service and endpoints.
- [ ] Port `knowledge` service incl. async document processing orchestration (delegate to existing docreader service over gRPC/HTTP).
- [ ] Port `models` clients (Embedding, Rerank, LLM, VLM/Ollama) using `reqwest`.
- [ ] Port `graph` builder if enabled (consider feature flag).

### Phase 5: Incremental Cutover
- [ ] Run Rust API on a separate port; add Nginx/Envoy route rules (or feature flag) to gradually shift traffic.
- [ ] Keep Go services running for unported endpoints.
- [ ] Validate parity via integration tests and golden responses.

### Phase 6: Decommission Go `internal/`
- [ ] After all endpoints and services are ported and validated, remove Go `internal/` from build.
- [ ] Update Dockerfiles and docker-compose to point to Rust binary.

## Testing Strategy
- [ ] Unit tests for repos/services using an ephemeral DB (dockerized MySQL/Postgres).
- [ ] Integration tests against the HTTP API; compare responses to Go reference where possible.
- [ ] Property tests for text utilities; snapshot tests for API responses (ensure stable JSON ordering where needed).

## Risks & Mitigations
- [ ] DB ORM differences: prefer explicit `sqlx` queries; add migration compatibility tests.
- [ ] Tokenization differences (gojieba vs jieba-rs): align dictionaries or fallback to calling existing Go tokenizer via sidecar service.
- [ ] Performance regressions: benchmark critical paths (retrieval, chunk processing).
- [ ] Third-party API differences: ensure identical headers/timeouts/retries.

## Tooling & CI
- [ ] Add `cargo fmt` / `clippy` checks.
- [ ] Add `sqlx` offline mode and schema verification jobs.
- [ ] Run integration tests in CI with services (docreader/minio/redis) via docker-compose.

## Milestones
- [ ] M0: Rust workspace + scaffolding + CI green.
- [ ] M1: Config/Logger/Types done; Rust API health endpoint.
- [ ] M2: Tenant repo/service + auth middleware; basic CRUD parity.
- [ ] M3: KnowledgeBase + Knowledge services ported; hybrid search via existing retriever engine(s).
- [ ] M4: Models clients ported; end-to-end RAG happy path.
- [ ] M5: Cutover plan executed; Go disabled in production build.

## Developer Notes
- [ ] Keep environment variable names unchanged.
- [ ] Use feature flags to toggle components (e.g., `ENABLE_GRAPH_RAG`).
- [ ] Preserve error message semantics where user-facing.
- [ ] Document any intentional deviations in `docs/`.
