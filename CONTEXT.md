# CONTEXT.md

Current state of the project. Update after every finished card.

_Last updated: 2026-10-13_

## Sprint 1 (2026-10-05 to 2026-10-16): walking skeleton

| Card | Status |
|------|--------|
| US-01 Product vision, personas & MVP scope | In Code Review (`docs/vision.md`) |
| US-02 Monorepo layout & Go module structure | In Code Review |
| US-03 CI pipeline on GitHub Actions | In Code Review (workflow green, branch protection on) |
| US-04 Local stack with docker compose | Not started. First card on a `feature/` branch |
| US-05 ADR process + first ADRs | Not started (template and ADR-0001 exist; still needs architecture, Kafka and Postgres ADRs) |
| US-06 Service Registry CRUD API | Not started |
| US-07 Postgres schema & migrations | Not started |
| US-08 Integration test harness (testcontainers-go) | Not started |

## Repo state
- Public repo: https://github.com/viclespinozas/monitoring-platform, branch `main` protected (requires `build-lint-test`, branch up to date; admins not enforced).
- Module path: `github.com/viclespinozas/monitoring-platform`.
- No Go source yet, so CI skips build/lint/test. The first real run of those steps happens with US-06.
- Tooling installed: Go 1.27.1, `gh` (logged in as `viclespinozas`). Not installed: `golangci-lint` (CI only), Docker status unchecked.

## Open items
- Duplicate "Acceptance criteria" checklists on the US-01 and US-02 Trello cards (my mistake); the user needs to delete the second one on each by hand.
- Naming: history notes live in `docs/history/` (gitignored) because `docs/` is already committed. Confirm or rename.
- From US-04 on: branch `feature/<name>`, PR into `main`, wait for CI.

## Next suggested
US-04 (docker compose: Postgres, Kafka KRaft, Redis, Kafka UI), on a `feature/local-stack` branch. Check Docker is installed first.
