# ADR-0004: PostgreSQL as the system of record

**Status:** Accepted
**Date:** 2026-10-14

## Context

The platform stores registered services, check results, alert rules,
incidents and their timelines. Most of it is relational (a service has rules,
a rule has incidents, an incident has timeline events) and needs
transactions: for example, opening an incident and writing its notification
to an outbox table must succeed or fail together.

Alternatives considered:

- **MongoDB or another document store**: flexible schema, but the data here
  is relational and the transactional outbox pattern wants multi-row ACID
  transactions.
- **A time-series database (TimescaleDB, InfluxDB)**: good for raw metrics,
  but metrics already live in Prometheus. Check results at this scale fit in
  ordinary tables with the right indexes.
- **PostgreSQL**: mature, ACID, strong constraints (unique service names,
  `INSERT ... ON CONFLICT DO NOTHING` for idempotent consumers), good Go
  support through pgx, and managed on AWS (RDS).

## Decision

PostgreSQL is the system of record for all durable platform state. Schema
changes are versioned migrations. Redis is used only for ephemeral state
(rate limiting, sliding-window aggregation) and is never the source of
truth. Kafka carries events in flight but is not queried as a database.

## Consequences

- Idempotency is enforced by the database: a unique `event_id` plus
  `ON CONFLICT DO NOTHING` makes redelivered events harmless.
- Transactions let us use a transactional outbox for notifications.
- Check-result tables will grow quickly. Uptime queries need indexes or a
  rollup table, and old rows will need a retention policy (US-13).
- Postgres is a single write node here. High availability of the platform is
  an explicit non-goal in the vision, so this is accepted.
- Revisit if check-result volume outgrows plain tables; TimescaleDB is a
  Postgres extension, so that move would not change the data access layer
  much.
