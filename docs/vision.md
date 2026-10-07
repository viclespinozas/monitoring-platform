# Vision: Cloud-Native Distributed Systems Monitoring Platform

## Problem

Teams running distributed services usually find out something is broken from a
customer complaint, not from their own tooling. The signals exist — logs,
metrics, traces — but they live in different systems, aren't correlated, and
nobody gets paged until the damage is done. Small teams and side projects
rarely build this muscle at all: proper alerting, incident tracking, and
resilience patterns feel like "big company" problems until an outage proves
otherwise.

This project builds a monitoring platform that registers services, watches
their health, turns failures into incidents automatically, and demonstrates
the resilience patterns (retries, circuit breakers, rate limiting) that keep
one failing dependency from taking down everything around it.

## Personas

**SRE (Sam)**
Owns uptime across every registered service, not the code inside them. Sam
wants one place to see what's up, what's degraded, and what's currently on
fire — plus a trustworthy incident history to run postmortems against. Sam is
paged, so false positives (flapping alerts) are a direct quality-of-life cost.

**Service owner (Priya)**
Builds and runs one or two services. Priya wants to register her service in
under five minutes, define "what does healthy mean" for it, and trust that if
it breaks, the right people find out fast — without having to build her own
alerting from scratch.

## Goals

- Detect a service outage and notify someone within 60 seconds of it starting.
- Make "why did this alert fire" always answerable from stored data (metrics,
  the triggering events, the incident timeline) — no "it was probably fine."
- Prove resilience patterns actually work by demonstrating a live failure
  (e.g. a dependency going down) and showing the system degrade gracefully
  instead of cascading.
- Keep onboarding a new service to a single API call plus a health-check URL.

## Non-goals

- Replacing Prometheus/Grafana as general-purpose metrics storage and
  dashboards — this platform builds on them, not around them.
- Supporting arbitrary third-party services outside this platform's own demo
  services; this isn't a multi-tenant SaaS.
- A fully-featured on-call scheduling/paging product (e.g. PagerDuty). One
  webhook-based notification channel is enough to prove the pattern.
- High availability of the platform itself (no multi-region, no HA Postgres).
  It needs to run the demo convincingly, not survive a datacenter loss.

## MVP scope

The MVP is the walking skeleton that makes the core loop real end to end:

1. Register a service (Service Registry API).
2. Probe it on an interval and publish results as events (health checks →
   Kafka).
3. Consume those events idempotently and compute uptime.
4. Expose uptime and service status through a queryable API.

Everything in Sprint 1–2 of the backlog. No alerting, no UI, no Kubernetes
yet — just: register something, watch it get checked, see the result.

## Later

Deliberately deferred past the MVP, in roughly the order the backlog takes
them:

- RED metrics, Prometheus/Grafana dashboards, load testing (Sprint 3).
- Alert rules, non-flapping evaluation, incident lifecycle, notifications
  (Sprint 4).
- Resilience patterns: retries, circuit breaker, rate limiting, DLQ
  (Sprint 5).
- Distributed tracing and centralized structured logs (Sprint 6).
- Chaos injection + the scripted failure demo, and the admin frontend
  (Sprint 7).
- Kubernetes (Helm, kind) and AWS via Terraform (Sprint 8).
- Auth hardening, performance report, and showcase materials (Sprint 9).

## Success metrics (for the demo)

- **Detection speed:** a simulated outage is reflected as a firing alert and
  an open incident within 60 seconds.
- **No false alarms:** a single transient blip (one failed health check) does
  not open an incident — only a sustained breach does.
- **Idempotency proven:** replaying the same check-result event twice never
  creates a duplicate record or a duplicate incident.
- **Resilience proven on camera:** when a dependency is killed via chaos
  injection, the dependent service's circuit breaker opens, the caller fails
  fast instead of hanging, and the system auto-recovers when the dependency
  comes back — all visible in Grafana and the incident timeline.
