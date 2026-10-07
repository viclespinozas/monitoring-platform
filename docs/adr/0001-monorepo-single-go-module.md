# ADR-0001: Monorepo with a single Go module

**Status:** Accepted
**Date:** 2026-10-07

## Context

The platform is made of several independently-deployable services (service
registry, health-check prober, event processor, demo microservices, and
later an alerting engine). We need to decide two things at once:

1. One repo or many repos?
2. One Go module or a Go workspace (`go.work`) of several modules?

Go 1.18+ supports multi-module workspaces specifically for the case where
several modules need to be developed and versioned independently but built
together locally. That flexibility has a cost: separate `go.mod` files,
separate dependency versions to keep in sync, and `go.work` files that are
easy to forget to update when a new module is added.

This project is a solo-developer portfolio project, not a multi-team
organization where different services need independent release cadences or
different Go versions. Everything ships from one place, by one person, on
one CI pipeline.

## Decision

Single repo (monorepo), single Go module at the repo root
(`github.com/victorlespinozas/monitoring-platform`), with this layout:

- `cmd/<service>/main.go` — one entrypoint per deployable binary (e.g.
  `cmd/service-registry`, `cmd/health-prober`, `cmd/processor`). Thin: wiring
  and startup only.
- `internal/<domain>/` — all real logic, private to this module by Go's own
  enforcement (`internal/` cannot be imported from outside this repo). Most
  code lives here. Each service gets its own subpackage
  (`internal/registry`, `internal/healthcheck`, ...); code shared between
  services (e.g. a Postgres connection helper, an event envelope type) lives
  in a shared `internal/` package rather than being duplicated.
- `pkg/` — only for code we'd be comfortable with an external project
  importing (e.g. a generated API client). Expected to stay nearly empty;
  defaulting new code to `internal/` and promoting to `pkg/` only when there
  is a real external consumer avoids `pkg/` turning into a dumping ground.

## Consequences

- One `go.mod`, one `go.sum`, one Go version, one `go test ./...` for
  everything. CI stays a single simple pipeline (US-03) instead of a matrix
  over modules.
- Every `cmd/` binary can import any `internal/` package directly — no
  publishing or versioning step between "I changed shared code" and "every
  service sees the change." That is the main cost: there's no compiler-
  enforced boundary between services beyond directory convention and code
  review discipline. In a multi-team org this would be a real risk; solo,
  it's an acceptable trade for speed.
- If this project ever needs independently-versioned modules (e.g. extracting
  a client library for external use, or splitting into separately-owned
  repos), that's the trigger to revisit this ADR and introduce a Go
  workspace or split repos — not before.
