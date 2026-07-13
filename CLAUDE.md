# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this service is

`ecommerce-catalog-service` is the **write side / source of truth** in a CQRS system: it owns
products, categories, and attributes. It exposes Connect-RPC (gRPC-compatible) write and read
APIs, persists to MongoDB, and publishes domain events to Kafka via a **transactional outbox**.
Downstream query services project those events into their own read models — this service never
serves reads on their behalf.

This repo is one module of a larger multi-repo workspace (see the workspace-root `CLAUDE.md` one
directory up for cross-repo rules, `go.work`, and the local k3d/Tilt stack). Locally, dependencies
on `ecommerce-commons` and the `*-api` repos resolve through the root `go.work`, so changes in
those repos are visible here immediately with no release/bump.

## Common commands

Run from this directory. `make help` lists everything.

```bash
make run                 # go run ./cmd/main.go (reads configs/ + .env, APP_ENV=standalone)
make build               # binary into bin/, injects VERSION via -ldflags
make test                # all tests, -race + coverage.out
make test-unit           # -short (unit only, no containers)
make test-integration    # -tags=integration (spins up real Mongo via testcontainers)
make test-e2e            # -tags=e2e (full app in-process + Mongo + Redpanda containers)
make lint                # golangci-lint (config .golangci.yml)
make fmt                 # gofmt -s + goimports
make generate-mocks      # mockery v3 (see caveat below)
make vuln-check          # govulncheck
make check-all           # deps + fmt + lint + test + vuln-check (the CI pipeline)
make install-tools       # install golangci-lint, mockery, govulncheck, etc.
```

Run a single test:
```bash
go test ./internal/application/product/ -run TestCreateProduct -v
go test -tags=integration ./internal/infrastructure/outbound/mongo/ -run TestProductRepository -v
```

Integration and e2e tests need Docker (testcontainers). Unit tests do not.

## Architecture

**Hexagonal + CQRS-write + event-driven.** Read `cmd/main.go` first — it is the composition
root and contains no business logic, only `fx.Options` module wiring. Everything is assembled
through `go.uber.org/fx`; to add a component, provide it in the relevant package's `Module()`
rather than constructing it in `main.go`.

Layers (each of the three aggregates — `product`, `category`, `attribute` — repeats this shape):

- **`internal/application/<aggregate>/`** — domain model + use cases. The aggregate file
  (`product.go`) holds the domain struct and its invariants; constructors (`NewProduct`) and
  `Update` validate business rules, while `Reconstruct` rebuilds from persistence **without**
  validation. One file per use case (`create_product.go`, `get_product.go`), each defining a
  `...CommandHandler`/`...QueryHandler` interface + impl. `repository.go` is the **port**
  interface, `errors.go` holds sentinel errors, `event_factory.go` is the event-building port.
- **`internal/infrastructure/inbound/connect/`** — Connect-RPC handlers (the driving adapters).
  `module.go` wires handlers and, importantly, declares **per-procedure permission scopes** in
  `provideProcedurePermissions` (e.g. `products:write`, `categories:read`) enforced by the
  commons validation interceptor.
- **`internal/infrastructure/outbound/mongo/`** — repository implementations (driven adapters).
  Each aggregate has `*_entity.go` (BSON document), `*_mapper.go` (domain↔entity), and
  `*_repository.go`. Repos embed `commonsmongo.GenericRepository` and are built with
  `NewTenantRepository` (see multi-tenancy below); custom queries like `FindList` build `bson`
  filters on top.
- **`internal/infrastructure/outbound/kafka/`** — event factories that turn domain objects into
  protobuf outbox messages. Topic is resolved via `apiEvents.TopicFor(event)` from the `-api` repo.

### The write path (the core pattern)

A command handler like `createProductHandler.Handle` does: validate cross-aggregate refs →
build the domain object → build an outbox `Message` from the event factory → then
`persistAndPublish` wraps the DB insert **and** `outbox.Create` in a single
`mongo.WithTransaction`. The transactional outbox guarantees the event is durably queued in the
same transaction as the state change; the actual Kafka send (`res.Send(ctx)`) happens
best-effort **after** commit and its error is intentionally ignored (already logged by the outbox
relay). Preserve this ordering when adding new write use cases.

### Multi-tenancy (database-per-tenant)

Tenant modules from `ecommerce-commons/pkg/tenant` and `tenant-service-api` are wired in
`main.go`. Tenant-scoped collections (`product`, `category`, `attribute`) use
`commonsmongo.NewTenantRepository`, which resolves a per-tenant database from request context.
The transactional `outbox` collection is **not** tenant-scoped. Respect tenant context in any new
repository query.

## Conventions & gotchas

- **Mocks (`mockery` v3, `make generate-mocks`)** are generated **inline** in each application
  package — the `Repository` and `*EventFactory` mocks live beside their interface as
  `package product`/`category`/`attribute` (not a `mocks/` subpackage). Only the commons
  `Outbox`/`TxManager` mocks land in `internal/testutil/mocks/` (`package mocks`). Regenerate via
  the config rather than hand-editing `mock_*.go`.
- **API contracts live in `ecommerce-catalog-service-api`**, not here. Protobuf under that repo's
  `proto/` (`catalog/v1/` for RPCs, `catalog/events/v1/` for Kafka events) generates the
  `catalogv1connect` handlers and `eventsv1` types this service imports. To change an RPC or
  event schema, edit `.proto` there and `make generate` — never hand-edit `gen/`.
- **MongoDB indexes are declarative migrations** in `db/migrations/*.json` (golang-migrate
  command format), applied at startup by the commons persistence module. Add indexes there.
- **Config** is `configs/config.standalone.yaml` selected by `APP_ENV` (set in `.env`),
  overridable by env vars. Covers `mongo`, `kafka`, `security.jwks` (JWT via JWKS, Logto locally),
  `logger`, `observability` (OpenTelemetry).
- **CI** (`.github/workflows/ci.yml`) delegates to the shared `go-ci.yml` in
  `ecommerce-infrastructure`; coverage excludes mocks, `module.go`, `main.go`, and infrastructure
  glue. Editing `.go` files triggers auto-format hooks (see workspace-root `CLAUDE.md`).
