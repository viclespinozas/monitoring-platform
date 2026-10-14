# ADR-0003: Kafka as the event backbone

**Status:** Accepted
**Date:** 2026-10-14

## Context

Health-check results and alert events need to flow between services
([ADR-0002](0002-overall-architecture.md)). Requirements that matter here:

- Many small events per second, produced continuously.
- Ordering per service (a service's checks must be processed in order).
- Consumers must be able to fall behind, restart and replay without losing
  events.
- Failed or poison messages must be isolatable (retry topics, dead-letter
  queue) without blocking healthy traffic.
- The project should build skills that transfer to real jobs.

Alternatives considered:

- **RabbitMQ**: mature broker with flexible routing and per-message acks.
  Messages are removed once consumed, so replaying history means extra
  machinery; ordering across competing consumers is weaker.
- **NATS (JetStream)**: lightweight and fast, simple to run. Smaller
  ecosystem and fewer Go examples of the retry/DLQ and consumer-group
  patterns this project wants to demonstrate.
- **Kafka**: a durable, partitioned, replayable log. Key-based partitioning
  gives ordering per service; consumer groups give scaling and offset
  control.

## Decision

Use Apache Kafka, running in KRaft mode (no ZooKeeper), with events keyed by
`service_id` so each service's events land on one partition and stay ordered.
Locally it runs as a single node in docker compose (US-04). The Go client
library and the event schema format are decided separately in a spike in
Sprint 2.

## Consequences

- Replay is cheap: a new or fixed consumer can re-read history, which makes
  idempotent-consumer and DLQ-replay demonstrations straightforward.
- Kafka is heavier to run than the alternatives. Locally that is one more
  container; on AWS it is the largest cost driver (self-managed in-cluster
  via Strimzi to avoid MSK pricing, see Sprint 8).
- Delivery is at-least-once. Consumers must deduplicate by `event_id` and
  commit offsets only after the database write succeeds.
- A single partition per key limits per-service throughput. Acceptable at
  this scale; revisit if one service's event rate outgrows a partition.
- Revisit if operating Kafka starts costing more time than the project
  gains, in which case NATS JetStream is the fallback.
