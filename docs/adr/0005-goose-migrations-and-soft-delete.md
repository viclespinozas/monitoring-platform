# ADR-0005: goose for migrations, soft delete for services

**Status:** Accepted
**Date:** 2026-10-14

## Context

Schema changes must be versioned and reproducible: any developer, CI and later
the cloud environment must reach the same schema from an empty database
([ADR-0004](0004-postgres-as-system-of-record.md)). Two related choices came
up while creating the first table (`services`).

**Migration tool.** goose and golang-migrate are both mature. golang-migrate
is slightly more common; goose supports SQL and Go migrations in one tool,
keeps `Up` and `Down` in a single file, and has a simple CLI.

**Deleting services.** Incidents, uptime history and check results will
reference a service. Removing the row would either cascade-delete that
history or leave dangling references.

## Decision

- Use **goose**, with plain SQL files in `migrations/`. It is run with a pinned
  `go run github.com/pressly/goose/v3/cmd/goose@<version>` from the Makefile,
  not added to `go.mod`: installing it as a module tool pulled dozens of cloud
  and database driver dependencies into the module for a CLI we only run.
- **Soft delete** services with a nullable `deleted_at` column. Name
  uniqueness is a partial unique index (`WHERE deleted_at IS NULL`), so a
  deleted service's name can be registered again without losing the old row.
- Every migration must have a working `Down`. CI applies all migrations, rolls
  them back and applies them again on a clean Postgres.

## Consequences

- Queries must filter `deleted_at IS NULL`. Forgetting that leaks deleted rows,
  so the repository layer owns this filter and handlers never touch it.
- Deleted rows accumulate. A purge job can be added later if it matters.
- The CLI version is pinned in the Makefile and must be bumped by hand. The
  first `make migrate-*` run downloads it.
- Rolling back destructive migrations loses data. `Down` is meant for local
  development and CI, not for production rollbacks.
- Revisit the tool if migrations need Go logic that goose handles awkwardly.
