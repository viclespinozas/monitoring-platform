# CONTEXT.md

Current state of the project. Update after every finished card.

_Last updated: 2026-10-15_

## Sprint 1 (2026-10-05 to 2026-10-16): walking skeleton

All eight cards are implemented, merged into `main`, and moved to Done in Trello.

| Card | Where it landed |
|------|-----------------|
| US-01 Product vision, personas & MVP scope | `docs/vision.md` |
| US-02 Monorepo layout & Go module structure | go.mod, Makefile, ADR-0001 |
| US-03 CI pipeline on GitHub Actions | `.github/workflows/ci.yml`, branch protection on `main` |
| US-04 Local stack with docker compose | `docker-compose.yml` (PR #1) |
| US-05 ADR process + first ADRs | `docs/adr/` ADRs 0002 to 0004 (PR #2) |
| US-07 Postgres schema & migrations | `migrations/`, `internal/platform/postgres`, ADR-0005 (PR #3, landed via #6) |
| US-06 Service Registry CRUD API | `cmd/service-registry`, `internal/registry`, `api/openapi.yaml`, ADR-0006 (PR #4, landed via #6) |
| US-08 Integration test harness | `internal/testutil/postgrestest`, `*_integration_test.go` (PR #5, landed via #6). Postgres only; a Kafka helper comes with the first consumer |

## Working mode
Sprint 1: Claude implemented every card. From Sprint 2 (2026-10-19) the user writes the Go code and Claude hints, reviews and explains, see CLAUDE.md.

## Repo state
- Public repo: https://github.com/viclespinozas/monitoring-platform, branch `main` protected (requires `build-lint-test`, branch up to date; admins not enforced).
- Module path: `github.com/viclespinozas/monitoring-platform`.
- Go code: `cmd/service-registry`, `internal/registry`, `internal/platform/{postgres,httpmw}`, `internal/testutil/postgrestest`. CI jobs: build-lint-test (required), migrations, integration; all green on `main`.
- Tooling installed: Go 1.27.1, `gh` (logged in as `viclespinozas`), Docker via OrbStack (must be running; `open -a OrbStack`). Not installed: `golangci-lint` (CI only).

## Sprint 2 (2026-10-19 to 2026-10-30): health checks and ingestion

Goal: health checks and uptime flowing through Kafka, idempotently. About 23 points (target velocity 20 to 23). Hint mode: the user writes the code.

| Card | Pts | Notes |
|------|-----|-------|
| US-11 Spike: Kafka client and event schema format | 2 | Do first. Compare franz-go, confluent-kafka-go, segmentio/kafka-go; JSON vs Protobuf; write ADR-0007. Timebox 1 day |
| US-09 Health-check prober worker | 5 | Interval, timeout, expected status from the registry; bounded worker pool; graceful shutdown; jitter |
| US-10 Publish check results to Kafka | 5 | Topic `check-results`, key = service_id, acks=all, idempotent producer, event has event_id, occurred_at, schema version |
| US-12 Processing service: idempotent consumer | 8 | Manual offset commit after DB write, dedup with `ON CONFLICT DO NOTHING` on event_id, redelivery test, lag metric. Needs a Kafka test helper |
| US-13 Uptime query API | 3 | `GET /api/v1/services/{id}/uptime?window=24h`; efficient query; edge-case tests |

Trello: the five Sprint 2 cards were moved out of Product Backlog on 2026-10-15 into the list still named "🎯 Sprint 1 Backlog" (the board has no Sprint 2 list and the connector cannot create or rename lists). The user may rename that list to "Sprint 2 Backlog". Cards have no acceptance-criteria checklist read yet; read it with trelloReadChecklist when starting each card.

## Open items
- Naming: history notes live in `docs/history/` (gitignored) because `docs/` is already committed. Confirm or rename.
- No Kafka code or Kafka test helper yet.

## Next suggested
Sprint 2 (US-09 to US-13: health-check prober, Kafka publishing, Kafka client spike, idempotent consumer, uptime API) starts 2026-10-19 in hint mode. The spike (US-11) should come before US-10 and US-12, since it decides the Kafka client and event format.
