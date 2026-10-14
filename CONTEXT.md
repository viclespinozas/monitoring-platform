# CONTEXT.md

Current state of the project. Update after every finished card.

_Last updated: 2026-10-14_

## Sprint 1 (2026-10-05 to 2026-10-16): walking skeleton

All eight cards are implemented and merged into `main`. The Trello cards for US-01 to US-08 sit in Code Review until the user moves them to Done.

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

## Open items
- Duplicate "Acceptance criteria" checklists on the US-01 and US-02 Trello cards (my mistake); the user needs to delete the second one on each by hand.
- Naming: history notes live in `docs/history/` (gitignored) because `docs/` is already committed. Confirm or rename.
- No Kafka code or Kafka test helper yet.

## Next suggested
Sprint 2 (US-09 to US-13: health-check prober, Kafka publishing, Kafka client spike, idempotent consumer, uptime API) starts 2026-10-19 in hint mode. The spike (US-11) should come before US-10 and US-12, since it decides the Kafka client and event format.
