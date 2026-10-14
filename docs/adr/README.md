# Architecture Decision Records

An ADR records one significant technical decision: the context, what was
decided and the consequences. Decisions stay written down next to the code so
the reasoning outlives anyone's memory.

## Process

1. Copy [template.md](template.md) to `NNNN-short-title.md`, using the next free number.
2. Fill in context, decision and consequences. Name the alternatives that were rejected and why.
3. Set status to `Proposed`, move to `Accepted` when the PR is merged.
4. Never edit an accepted ADR to change the decision. Write a new ADR and mark the old one `Superseded by ADR-NNNN`.
5. Link related ADRs to each other.

## Index

| ADR | Title | Status |
|-----|-------|--------|
| [0001](0001-monorepo-single-go-module.md) | Monorepo with a single Go module | Accepted |
| [0002](0002-overall-architecture.md) | Event-driven architecture with small independent services | Accepted |
| [0003](0003-kafka-as-event-backbone.md) | Kafka as the event backbone | Accepted |
| [0004](0004-postgres-as-system-of-record.md) | PostgreSQL as the system of record | Accepted |
