# CONTEXT.md

Current state of the project. Update after every finished card.

_Last updated: 2026-10-13_

## Sprint 1 (2026-10-05 to 2026-10-16): walking skeleton

| Card | Status |
|------|--------|
| US-01 Product vision, personas & MVP scope | In Code Review (`docs/vision.md`) |
| US-02 Monorepo layout & Go module structure | In Code Review (commit `cdcfd37`) |
| US-03 CI pipeline on GitHub Actions | In Progress. `.github/workflows/ci.yml` written (uncommitted, never run). Blocked on: GitHub repo + push, then branch protection |
| US-04 Local stack with docker compose | Not started |
| US-05 ADR process + first ADRs | Not started (template and ADR-0001 exist; still needs architecture, Kafka and Postgres ADRs) |
| US-06 Service Registry CRUD API | Not started |
| US-07 Postgres schema & migrations | Not started |
| US-08 Integration test harness (testcontainers-go) | Not started |

## Repo state
- Git repo on `main`, one commit. No remote yet, no GitHub repo created.
- Uncommitted: `CLAUDE.md`, `CONTEXT.md`, `.gitignore` change (ignores `docs/history/`).
- No Go source yet, so `make build/vet/test` report "no packages". Expected until US-06.

## Open items
- Duplicate "Acceptance criteria" checklists on the US-01 and US-02 Trello cards (my mistake); the user needs to delete the second one on each by hand.
- `gh` CLI not installed; needed for repo creation and CI work (US-03).
- `golangci-lint` not installed.
- Naming: the user asked for a gitignored folder "called docs"; `docs/` is already committed, so history lives in `docs/history/`. Confirm or rename.

## Next suggested
US-03 (CI) and US-04 (docker compose), after the user finishes reviewing US-01 and US-02.
