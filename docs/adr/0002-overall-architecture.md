# ADR-0002: Event-driven architecture with small independent services

**Status:** Accepted
**Date:** 2026-10-14

## Context

The platform has to register services, probe their health, turn results into
uptime numbers, evaluate alert rules, open incidents and notify people (see
[vision](../vision.md)). Two broad shapes were possible:

1. A single deployable application where probing, processing and alerting
   call each other in-process.
2. Several small services that communicate through a message broker and
   share a database only where they have to.

The platform's own purpose is to demonstrate how distributed systems fail and
recover (retries, circuit breakers, idempotent consumers, a chaos demo). A
single process cannot show a dependency failing independently of the rest.

## Decision

Build several small Go services, all in one repository and one module
([ADR-0001](0001-monorepo-single-go-module.md)):

- **service-registry**: REST API and source of truth for registered services.
- **health-prober**: probes registered services on an interval and publishes
  check results as events.
- **processor**: consumes events idempotently, stores results, computes
  uptime and (later) alert signals.
- Later: alerting/incident components, and demo services used for chaos.

Services talk asynchronously through Kafka for high-volume facts (check
results, alerts) and synchronously over REST where a caller needs an answer
(for example the public API). Kafka is chosen in
[ADR-0003](0003-kafka-as-event-backbone.md), Postgres in
[ADR-0004](0004-postgres-as-system-of-record.md).

## Consequences

- Probing and processing are decoupled: a slow or failing processor does not
  stop probes, and events are buffered in Kafka until it catches up.
- Delivery is at-least-once, so every consumer has to be idempotent. That is
  extra work, and it is also one of the things this project exists to show.
- More moving parts than a monolith: more containers locally, more to
  observe, harder debugging. Metrics, tracing and centralized logs (Sprints
  3 and 6) are therefore required, not optional.
- Services share one repo and one deploy pipeline, so there is no
  compiler-enforced boundary between them. Boundaries rely on package
  discipline and review.
- Revisit if the operational cost of running several services outweighs the
  demonstration value, or if one service needs a very different scaling or
  release profile.
