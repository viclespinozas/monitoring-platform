# CONTEXT.md

Current state of the project. Update after every finished card.

_Last updated: 2026-10-14_

Sprint 1 cards US-05, US-07, US-06, US-08 are built as a stack of branches, each branched from the previous one (adr-process -> migrations -> registry-api -> integration-tests). Each PR targets the previous branch; review and merge in that order.

## Sprint 1 (2026-10-05 to 2026-10-16): walking skeleton

| Card | Status |
|------|--------|
| US-01 Product vision, personas & MVP scope | In Code Review (`docs/vision.md`) |
| US-02 Monorepo layout & Go module structure | In Code Review |
| US-03 CI pipeline on GitHub Actions | In Code Review (workflow green, branch protection on) |
| US-04 Local stack with docker compose | Merged (PR #1) |
| US-05 ADR process + first ADRs | In Code Review. Branch `feature/adr-process` (ADRs 0002 to 0004 + index) |
| US-06 Service Registry CRUD API | In Code Review. Branch `feature/service-registry-api` (REST API, OpenAPI, ADR-0006) |
| US-07 Postgres schema & migrations | In Code Review. Branch `feature/db-migrations` (goose, `services` table, pgx pool, CI job) |
| US-08 Integration test harness (testcontainers-go) | Not started |

## Repo state
- Public repo: https://github.com/viclespinozas/monitoring-platform, branch `main` protected (requires `build-lint-test`, branch up to date; admins not enforced).
- Module path: `github.com/viclespinozas/monitoring-platform`.
- Go code exists: `cmd/service-registry`, `internal/registry`, `internal/platform/{postgres,httpmw}`. CI build, lint, test and the migrations job all pass.
- Tooling installed: Go 1.27.1, `gh` (logged in as `viclespinozas`), Docker via OrbStack (must be running; `open -a OrbStack`). Not installed: `golangci-lint` (CI only).

## Open items
- Duplicate "Acceptance criteria" checklists on the US-01 and US-02 Trello cards (my mistake); the user needs to delete the second one on each by hand.
- Naming: history notes live in `docs/history/` (gitignored) because `docs/` is already committed. Confirm or rename.
- From US-04 on: branch `feature/<name>`, PR into `main`, wait for CI.

## Next suggested
Commit and PR US-04 when the user asks. Then US-05 (more ADRs) or US-07 (Postgres migrations, can use `make up`).
