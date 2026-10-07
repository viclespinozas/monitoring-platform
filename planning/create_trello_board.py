#!/usr/bin/env python3
"""Creates the Scrum board for the Cloud-Native Monitoring Platform on Trello.

Usage:
    export TRELLO_KEY=...    # https://trello.com/power-ups/admin -> your Power-Up -> API key
    export TRELLO_TOKEN=...  # generated from the same page ("Token" link)
    python3 planning/create_trello_board.py

Stdlib only. Credentials are sent over HTTPS as query params (Trello's documented method).
"""
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from datetime import date, timedelta

API = "https://api.trello.com/1"
BOARD_NAME = "Monitoring Platform · Scrum"
SPRINT_1_START = date(2026, 10, 5)  # Monday
SPRINT_COUNT = 9


def sprint_dates(n):
    start = SPRINT_1_START + timedelta(days=14 * (n - 1))
    return start, start + timedelta(days=11)  # Mon .. Fri of week 2


# ---------------------------------------------------------------- labels
TYPE_LABELS = {
    "backend": ("Backend", "green"),
    "frontend": ("Frontend", "yellow"),
    "devops": ("DevOps / Infra", "orange"),
    "qa": ("QA / Testing", "lime"),
    "docs": ("Documentation", "blue"),
    "product": ("Product / Refinement", "pink"),
    "obs": ("Observability", "sky"),
    "sec": ("Security", "black"),
    "arch": ("Architecture / ADR", "purple"),
    "spike": ("Spike / Research", "purple_light"),
    "career": ("Career / Showcase", "pink_dark"),
    "bug": ("Bug", "red"),
    "blocked": ("Blocked", "red_dark"),
    "debt": ("Tech Debt", "black_light"),
}

EPICS = {
    "E1": ("Foundations & Dev Workflow", "green_light",
           "Monorepo, CI, local stack, conventions, ADRs. Everything else stands on this."),
    "E2": ("Service Registry & Query API", "yellow_light",
           "Register services, store metadata, expose a clean REST API to query them."),
    "E3": ("Health Checks & Uptime", "orange_light",
           "Probe registered services, publish results as events, compute uptime."),
    "E4": ("Event Ingestion Pipeline", "red_light",
           "Kafka contracts, producers, idempotent consumers, retry topics, DLQ."),
    "E5": ("Metrics, Tracing & Logs", "sky_light",
           "RED metrics, Prometheus/Grafana as code, OpenTelemetry tracing, centralized logs."),
    "E6": ("Alerting & Incidents", "purple_light",
           "Alert rules, non-flapping evaluation, incident lifecycle and history, notifications."),
    "E7": ("Resilience Patterns", "blue_light",
           "Timeouts, retries with jitter, circuit breaker, rate limiting, idempotency keys."),
    "E8": ("Chaos & Failure Simulation", "pink_light",
           "Demo services + chaos controller. The headline interview demo."),
    "E9": ("Frontend Dashboard", "lime_light",
           "Small admin UI: services, status, incidents, rules. Grafana covers the metrics."),
    "E10": ("Kubernetes & AWS", "black_light",
            "Containers, Helm, kind locally, Terraform EKS on AWS, CD with OIDC."),
    "E11": ("Security & Hardening", "green_dark",
            "AuthN for the API, secrets handling, log retention, least privilege."),
    "E12": ("Showcase & Career", "yellow_dark",
            "README, diagrams, demo video, blog post, interview stories."),
}

# ---------------------------------------------------------------- stories
# (id, sprint, epic, points, labels, title, story, acceptance criteria, notes)
S = [
    # ---------------- Sprint 1: walking skeleton
    ("US-01", 1, "E1", 2, ["product", "docs"], "Product vision, personas & MVP scope",
     "As the product owner, I want a one-page vision and an explicit MVP scope so that I can say no to scope creep.",
     ["docs/vision.md with problem, personas (SRE, service owner), goals and non-goals",
      "MVP scope and 'later' list written down",
      "Success metrics for the demo defined (e.g. alert within 60s of failure)"],
     "Interview angle: how you prioritize and cut scope."),
    ("US-02", 1, "E1", 2, ["arch", "backend"], "Monorepo layout & Go module structure",
     "As a developer, I want a clear repo layout so that each service is independent but shares tooling.",
     ["cmd/<service>, internal/, pkg/ convention decided and documented",
      "Makefile with build/test/lint/run targets",
      "Go 1.23+ workspace or single module decision recorded in an ADR"],
     "Avoid over-abstracting pkg/. Keep internal/ private."),
    ("US-03", 1, "E1", 3, ["devops", "qa"], "CI pipeline on GitHub Actions",
     "As a developer, I want every PR built, linted and tested so that main is always green.",
     ["Workflow runs go build, golangci-lint, go test -race -cover",
      "Branch protection on main requires CI green",
      "Build caching enabled, pipeline under 5 min"],
     ""),
    ("US-04", 1, "E1", 3, ["devops"], "Local stack with docker compose",
     "As a developer, I want Postgres, Kafka (KRaft) and Redis running locally with one command.",
     ["make up / make down work from a clean clone",
      "Health checks defined for each container",
      "Kafka UI (e.g. redpanda console / kafka-ui) available locally"],
     ""),
    ("US-05", 1, "E1", 2, ["arch", "docs"], "ADR process + first ADRs",
     "As an engineer, I want architectural decisions recorded so that I can defend them in interviews.",
     ["docs/adr/ with template (context, decision, consequences)",
      "ADR-001 overall architecture, ADR-002 Kafka over RabbitMQ/NATS, ADR-003 Postgres as system of record"],
     "German teams love written reasoning. ADRs = interview gold."),
    ("US-06", 1, "E2", 5, ["backend"], "Service Registry CRUD API",
     "As a service owner, I want to register, update, list and delete services so that the platform knows what to monitor.",
     ["POST/GET/PUT/DELETE /api/v1/services with validation",
      "Pagination + filtering by team/tag on list",
      "Consistent error model (RFC 9457 problem+json)",
      "OpenAPI spec committed"],
     "Handler -> service -> repository layering; context everywhere."),
    ("US-07", 1, "E2", 3, ["backend"], "Postgres schema & migrations",
     "As a developer, I want versioned migrations so that schema changes are reproducible.",
     ["goose or golang-migrate wired into make + CI",
      "services table with unique name, timestamps, soft delete decision",
      "pgx connection pool with sane limits"],
     ""),
    ("US-08", 1, "E1", 3, ["qa"], "Integration test harness with testcontainers-go",
     "As a developer, I want integration tests against real Postgres/Kafka so that I trust my repositories and consumers.",
     ["Repository tests run against a throwaway Postgres container",
      "Tests tagged (//go:build integration) and run in CI",
      "Table-driven unit tests for handlers"],
     ""),

    # ---------------- Sprint 2: health checks & ingestion
    ("US-09", 2, "E3", 5, ["backend"], "Health-check prober worker",
     "As an SRE, I want registered services probed on an interval so that I know if they are up.",
     ["HTTP probe with configurable interval, timeout, expected status",
      "Worker pool with bounded concurrency; graceful shutdown on SIGTERM",
      "Jitter on schedule to avoid thundering herd"],
     "Talk about goroutine leaks, context cancellation, errgroup."),
    ("US-10", 2, "E4", 5, ["backend"], "Publish check results to Kafka",
     "As the platform, I want check results published as events so that processing is decoupled from probing.",
     ["Topic check-results with key = service_id (ordering per service)",
      "Producer with acks=all, idempotent producer enabled",
      "Event has event_id, occurred_at, schema version"],
     ""),
    ("US-11", 2, "E4", 2, ["spike", "arch"], "Spike: Kafka client & event schema format",
     "As an engineer, I want to choose a Go Kafka client and schema format with evidence.",
     ["Compare franz-go vs confluent-kafka-go vs segmentio/kafka-go",
      "Decide JSON vs Protobuf (+ schema registry or not)",
      "ADR-004 written"],
     "Timebox: 1 day."),
    ("US-12", 2, "E4", 8, ["backend"], "Processing service: idempotent consumer",
     "As the platform, I want check results consumed and stored exactly once in effect, even if Kafka redelivers.",
     ["Consumer group with manual offset commit after DB write",
      "Dedup via unique event_id (INSERT ... ON CONFLICT DO NOTHING)",
      "Test proving redelivery does not create duplicates",
      "Consumer lag exposed as a metric"],
     "Core interview topic: at-least-once + idempotency = effectively-once."),
    ("US-13", 2, "E3", 3, ["backend"], "Uptime query API",
     "As a service owner, I want uptime % for 24h/7d/30d so that I can report availability.",
     ["GET /api/v1/services/{id}/uptime?window=24h",
      "Efficient query (index or rollup table), documented",
      "Unit tests for edge cases (no data, partial window)"],
     ""),

    # ---------------- Sprint 3: demo services & metrics
    ("US-14", 3, "E8", 5, ["backend"], "Demo microservices: order, inventory, payment",
     "As a demo presenter, I want a small realistic call chain so that failures propagate like production.",
     ["order -> inventory, order -> payment over HTTP (or gRPC)",
      "Each service has /healthz and /readyz",
      "Registered automatically in the Service Registry on startup"],
     "Keep business logic trivial. They exist to fail."),
    ("US-15", 3, "E5", 3, ["obs", "backend"], "RED metrics middleware",
     "As an SRE, I want rate, errors and duration for every endpoint so that I can see service health.",
     ["Prometheus histogram with sensible buckets + counters by route/status",
      "Low-cardinality labels (route template, not raw path)",
      "/metrics exposed on every service"],
     "Be ready to explain cardinality explosions."),
    ("US-16", 3, "E5", 3, ["obs", "devops"], "Prometheus + Grafana as code",
     "As an SRE, I want dashboards provisioned from the repo so that they are versioned and reproducible.",
     ["Prometheus scrape config for all services",
      "Grafana datasources and dashboards provisioned from JSON in repo",
      "Service overview dashboard: RPS, error %, p50/p95/p99"],
     ""),
    ("US-17", 3, "E1", 3, ["qa"], "Load generator with k6",
     "As a tester, I want repeatable traffic scenarios so that dashboards and alerts have realistic data.",
     ["k6 scripts: steady, spike, soak",
      "make load target",
      "Thresholds defined (p95 < X ms)"],
     ""),
    ("US-18", 3, "E6", 8, ["backend"], "Sliding-window latency & error-rate aggregation",
     "As the platform, I want per-service error rate and latency over sliding windows so that alerts have signals.",
     ["Consume request events, aggregate 1m/5m windows",
      "Redis used for window state; algorithm documented",
      "Late/out-of-order events handling documented",
      "Benchmarks (go test -bench) included"],
     "Discuss tumbling vs sliding windows, eventual consistency."),

    # ---------------- Sprint 4: alerting & incidents
    ("US-19", 4, "E6", 5, ["backend"], "Alert rule model & API",
     "As a service owner, I want to define alert rules (e.g. error rate > 5% for 2m) so that I get notified of real problems.",
     ["CRUD /api/v1/alert-rules with validation",
      "Rule: metric, operator, threshold, for-duration, severity",
      "Rules stored in Postgres, cached for evaluation"],
     ""),
    ("US-20", 4, "E6", 8, ["backend"], "Rule evaluation engine (no flapping)",
     "As an on-call engineer, I want alerts that fire after a sustained breach and resolve cleanly so that I am not spammed.",
     ["States: inactive -> pending -> firing -> resolved",
      "for-duration + hysteresis to prevent flapping",
      "Emits alert events to Kafka topic alerts",
      "Deterministic tests with a fake clock"],
     "Fake clock testing is a great senior signal."),
    ("US-21", 4, "E6", 5, ["backend"], "Incident lifecycle & history API",
     "As an SRE, I want incidents opened from alerts with a timeline so that I can do postmortems.",
     ["Incident opened on firing alert; dedup of repeated alerts",
      "Ack / resolve endpoints; timeline of events stored",
      "GET /api/v1/incidents with filters (service, status, time range)"],
     ""),
    ("US-22", 4, "E6", 5, ["backend"], "Notifications via transactional outbox",
     "As an on-call engineer, I want Slack/Discord/email notifications that are never lost nor duplicated.",
     ["Outbox table written in the same transaction as the incident",
      "Relay publishes outbox rows, marks sent, retries on failure",
      "Webhook notifier (Slack or Discord) + one more channel"],
     "Outbox pattern = classic distributed-systems interview question."),

    # ---------------- Sprint 5: resilience
    ("US-23", 5, "E7", 3, ["backend"], "Timeouts & retries with exponential backoff + jitter",
     "As a service, I want bounded retries with jitter so that transient failures heal without retry storms.",
     ["Retry only idempotent/retryable errors",
      "Full-jitter backoff, max attempts, respects context deadline",
      "Retry metrics exported"],
     ""),
    ("US-24", 5, "E7", 5, ["backend"], "Circuit breaker (own implementation)",
     "As the order service, I want a circuit breaker around inventory so that I fail fast when it is down.",
     ["Closed / open / half-open states with configurable thresholds",
      "State changes emitted as metrics and Kafka events",
      "Concurrency-safe; tests with -race; compared against sony/gobreaker in ADR"],
     "Write it yourself first, then compare. You'll explain it much better."),
    ("US-25", 5, "E7", 5, ["backend", "sec"], "Distributed rate limiting on the public API",
     "As the platform owner, I want per-client rate limits so that one client cannot starve others.",
     ["Token bucket (or GCRA) backed by Redis with Lua for atomicity",
      "429 with Retry-After and RateLimit headers",
      "Behaviour when Redis is down decided (fail open vs closed) + ADR"],
     ""),
    ("US-26", 5, "E4", 5, ["backend"], "Retry topics & dead-letter queue",
     "As the platform, I want poison messages isolated so that one bad event never blocks a partition.",
     ["Retry topics with delay tiers, then DLQ",
      "DLQ inspection + replay command",
      "Test with a deliberately malformed event"],
     ""),
    ("US-27", 5, "E7", 3, ["backend"], "Idempotency keys on write endpoints",
     "As an API client, I want to safely retry POSTs so that network failures never create duplicates.",
     ["Idempotency-Key header stored with response for 24h",
      "Same key + different body -> 422",
      "Concurrent same-key requests handled"],
     ""),
    ("US-28", 5, "E7", 2, ["docs", "arch"], "Resilience patterns write-up",
     "As a reader of the repo, I want a doc explaining which pattern is used where and why.",
     ["docs/resilience.md with diagrams",
      "Trade-offs of each pattern + failure modes"],
     ""),

    # ---------------- Sprint 6: tracing & logs (holiday sprint, lighter)
    ("US-29", 6, "E5", 8, ["obs", "backend"], "Distributed tracing with OpenTelemetry",
     "As an engineer, I want one trace across HTTP calls and Kafka hops so that I can find where latency comes from.",
     ["OTel SDK in all services; W3C tracecontext over HTTP",
      "Trace context propagated through Kafka headers",
      "Sampling strategy chosen and documented"],
     ""),
    ("US-30", 6, "E5", 3, ["obs", "devops"], "Tracing backend (Jaeger or Tempo)",
     "As an engineer, I want to browse traces and jump from Grafana to a trace.",
     ["OTel Collector + Jaeger/Tempo in compose",
      "Exemplars or trace links from Grafana panels"],
     ""),
    ("US-31", 6, "E5", 5, ["obs", "backend"], "Structured logs, centralized",
     "As an engineer, I want JSON logs with trace IDs searchable in one place.",
     ["log/slog JSON with trace_id, span_id, service",
      "Shipped to Loki (or OpenSearch) and queryable from Grafana",
      "Retention policy defined (GDPR-aware: no PII in logs)"],
     "Loki is cheaper/simpler; OpenSearch only if you want the search story."),

    # ---------------- Sprint 7: chaos + frontend
    ("US-32", 7, "E8", 5, ["backend"], "Chaos controller",
     "As a demo presenter, I want to inject latency, errors or outages into a service at runtime.",
     ["Admin endpoint/CLI: inject latency, error %, full outage, with duration",
      "Chaos state visible in metrics and Grafana annotations",
      "Safe by default: disabled unless CHAOS_ENABLED=true"],
     ""),
    ("US-33", 7, "E8", 5, ["qa"], "E2E scenario: 'Inventory Service becomes unavailable'",
     "As a demo presenter, I want a scripted scenario proving failure -> retry -> circuit breaker -> event -> alert -> recovery.",
     ["make chaos-inventory runs the scenario end to end",
      "Automated assertions: CB opens, alert fires < 60s, incident created, auto-resolves",
      "Runbook in docs/scenarios/inventory-outage.md"],
     "This is your interview demo. Rehearse it."),
    ("US-34", 7, "E6", 3, ["backend"], "Recovery detection & auto-resolve",
     "As an on-call engineer, I want incidents auto-resolved when the service recovers, with MTTR recorded.",
     ["Resolve after N healthy windows",
      "MTTD / MTTR computed per incident and exposed via API"],
     ""),
    ("US-35", 7, "E9", 5, ["frontend"], "Frontend: service list & status page",
     "As a service owner, I want to see all services and their current status at a glance.",
     ["React + TypeScript (Vite) or Go templates + HTMX. Decide in ADR",
      "Status badge, uptime %, last check, link to Grafana",
      "Uses the public API only; typed client from OpenAPI"],
     "Keep it small. Backend roles care little about CSS."),
    ("US-36", 7, "E9", 3, ["frontend"], "Frontend: incident timeline",
     "As an SRE, I want to view an incident and its timeline so that I can write a postmortem.",
     ["Incident list with filters", "Detail page with timeline events, ack/resolve buttons"],
     ""),

    # ---------------- Sprint 8: Kubernetes & AWS
    ("US-37", 8, "E10", 2, ["devops", "sec"], "Production-grade container images",
     "As an operator, I want small, secure images.",
     ["Multi-stage builds, distroless, non-root, read-only fs",
      "Image scanning (Trivy) in CI"],
     ""),
    ("US-38", 8, "E10", 8, ["devops"], "Kubernetes manifests (Helm) on kind",
     "As an operator, I want the whole platform deployable to Kubernetes with one command.",
     ["Helm chart per service (or umbrella chart)",
      "Liveness/readiness probes, resource requests/limits, PodDisruptionBudget",
      "HPA on processing service; graceful shutdown verified",
      "make k8s-up deploys to a local kind cluster"],
     ""),
    ("US-39", 8, "E10", 8, ["devops"], "Terraform: AWS infrastructure",
     "As an operator, I want the AWS environment defined as code and destroyable in one command.",
     ["Modules: VPC, EKS, RDS Postgres (Kafka in-cluster via Strimzi to save cost)",
      "Remote state (S3 + locking), tflint + terraform validate in CI",
      "AWS Budget alarm + make aws-destroy documented"],
     "Cost guardrail: spin up only for demos/recordings, then destroy."),
    ("US-40", 8, "E10", 5, ["devops", "sec"], "CD pipeline to EKS with OIDC",
     "As a developer, I want merges to main deployed automatically without static AWS keys.",
     ["GitHub Actions OIDC -> IAM role (no long-lived keys)",
      "Build -> push to ECR -> helm upgrade",
      "Rollback documented"],
     ""),

    # ---------------- Sprint 9: hardening & showcase
    ("US-41", 9, "E11", 3, ["sec", "backend"], "API authentication",
     "As the platform owner, I want only authenticated clients to use the API.",
     ["API keys (hashed at rest) or JWT. Decision in ADR",
      "Scopes: read vs write", "Secrets from K8s Secrets / AWS Secrets Manager"],
     ""),
    ("US-42", 9, "E1", 5, ["qa", "obs"], "Performance test & report",
     "As an engineer, I want measured numbers so that I can quote throughput and p99 in interviews.",
     ["Events/sec sustained, p99 end-to-end latency, consumer lag under load",
      "One bottleneck found, fixed and documented (before/after)",
      "docs/performance.md"],
     "'I measured, found X, fixed it' beats any buzzword."),
    ("US-43", 9, "E12", 3, ["docs"], "README, architecture diagrams, runbooks",
     "As a recruiter or hiring manager, I want to understand the project in 2 minutes.",
     ["README: what, why, architecture diagram, quickstart, demo GIF",
      "C4 context + container diagrams", "Runbooks for top 3 alerts"],
     ""),
    ("US-44", 9, "E12", 3, ["docs", "career"], "Demo video & blog post",
     "As a candidate, I want a 5-minute video and a write-up I can link in applications.",
     ["Video: chaos scenario end to end with narration",
      "Blog post (dev.to/Medium/LinkedIn) on one deep topic, e.g. idempotent consumers"],
     ""),
    ("US-45", 9, "E12", 3, ["career"], "Interview narratives & system design walkthrough",
     "As a candidate, I want STAR stories and a system design talk track based on this project.",
     ["5 STAR stories (trade-off, failure, debugging, scope cut, perf fix)",
      "30-min system design walkthrough rehearsed (in English; German summary optional)"],
     ""),
    ("US-46", 9, "E9", 3, ["frontend"], "Frontend: register service & manage alert rules",
     "As a service owner, I want to register services and edit alert rules from the UI.",
     ["Forms with client + server validation", "Optimistic UI or clear loading states"],
     ""),
]

AGREEMENTS = [
    ("📅 Sprint calendar & goals", "\n".join(
        [f"**Sprint {n}** · {sprint_dates(n)[0]:%a %d %b} → {sprint_dates(n)[1]:%a %d %b %Y} · {g}"
         for n, g in enumerate([
             "Walking skeleton: repo, CI, local stack, first service registered",
             "Health checks & uptime flowing through Kafka, idempotently",
             "Demo services + RED metrics visible in Grafana",
             "Error-rate breach opens an incident and notifies",
             "Resilience: services survive dependency failures",
             "Tracing & centralized logs (holiday sprint, lighter load)",
             "Chaos demo end to end + minimal UI",
             "Kubernetes locally, then AWS via Terraform",
             "Hardening, performance report, showcase & interview prep",
         ], start=1)]
    ) + "\n\nTarget velocity: ~20–23 pts/sprint (you're also job hunting). Adjust after Sprint 2 using real velocity."),
    ("🔁 Ceremonies (solo edition)", (
        "- **Planning** (Mon, 45 min): pick sprint goal, pull stories from Product Backlog into Sprint Backlog, break into tasks.\n"
        "- **Daily** (5 min, written): yesterday / today / blockers as a comment on the card you're working on.\n"
        "- **Refinement** (mid-sprint, 30 min): groom next sprint's stories, re-estimate, split anything > 8.\n"
        "- **Review** (last Fri): demo to yourself on video (reusable for the showcase).\n"
        "- **Retro** (last Fri): keep / drop / try, as a comment on the sprint calendar card.")),
    ("✅ Definition of Ready", (
        "- User story written (As a / I want / so that)\n- Acceptance criteria as checklist\n"
        "- Estimated (Fibonacci) and ≤ 8 pts, otherwise split\n- Dependencies known\n- Labels: type + epic")),
    ("🏁 Definition of Done", (
        "- Acceptance criteria met\n- Tests: unit + integration where relevant, CI green, `-race` clean\n"
        "- Code reviewed (ask Claude for a review after committing)\n- Metrics/logs added for new behaviour\n"
        "- Docs/ADR/OpenAPI updated\n- Merged to main via PR, card moved to Done")),
    ("🔢 Estimation guide", (
        "Fibonacci, relative to complexity + uncertainty, not hours:\n"
        "- **1**: trivial, config change\n- **2**: small, well-known\n- **3**: a solid half-day to day\n"
        "- **5**: a few days, some unknowns\n- **8**: most of a week, several unknowns (consider splitting)\n"
        "- **13**: too big, split it before planning")),
    ("🌿 Git & review workflow", (
        "- Trunk-based: short-lived branches `feat/US-12-idempotent-consumer`\n"
        "- Conventional Commits (`feat:`, `fix:`, `test:`, `docs:`, `chore:`)\n"
        "- One PR per story, story ID in PR title, link the PR in the card\n"
        "- After committing, ask Claude to review the branch or PR before merging")),
]

LISTS = ["📌 Team Agreements", "🗺️ Epics", "📥 Product Backlog", "🎯 Sprint 1 Backlog",
         "🔨 In Progress", "👀 Code Review", "✅ Done"]


# ---------------------------------------------------------------- API client
class Trello:
    def __init__(self, key, token):
        self.auth = urllib.parse.urlencode({"key": key.strip(), "token": token.strip()})

    def call(self, method, path, body=None):
        data = json.dumps(body).encode() if body is not None else None
        req = urllib.request.Request(f"{API}{path}?{self.auth}", data=data, method=method, headers={
            "Content-Type": "application/json", "Accept": "application/json"})
        for attempt in range(5):
            try:
                with urllib.request.urlopen(req) as resp:
                    time.sleep(0.12)  # stay under 100 req / 10 s
                    return json.loads(resp.read() or b"{}")
            except urllib.error.HTTPError as e:
                if e.code == 429 and attempt < 4:
                    time.sleep(2 ** attempt)
                    continue
                raise RuntimeError(f"{method} {path} -> {e.code}: {e.read().decode()[:300]}") from None


def main():
    key, token = os.environ.get("TRELLO_KEY"), os.environ.get("TRELLO_TOKEN")
    if not key or not token:
        sys.exit("Set TRELLO_KEY and TRELLO_TOKEN first (see docstring).")
    t = Trello(key, token)

    board = t.call("POST", "/boards/", {"name": BOARD_NAME, "defaultLists": False, "defaultLabels": False,
                                        "prefs_background": "blue",
                                        "desc": "Cloud-Native Distributed Systems Monitoring Platform · 2-week sprints · Oct 2026 → Feb 2027"})
    bid = board["id"]
    print("Board:", board["url"])

    lists = {name: t.call("POST", "/lists", {"name": name, "idBoard": bid, "pos": "bottom"})["id"] for name in LISTS}

    labels = {k: t.call("POST", "/labels", {"name": n, "color": c, "idBoard": bid})["id"]
              for k, (n, c) in TYPE_LABELS.items()}
    for eid, (name, color, _) in EPICS.items():
        labels[eid] = t.call("POST", "/labels", {"name": f"{eid} · {name}", "color": color, "idBoard": bid})["id"]

    points_field = None
    try:
        points_field = t.call("POST", "/customFields", {
            "idModel": bid, "modelType": "board", "name": "Story Points", "type": "number",
            "pos": "top", "display_cardFront": True})["id"]
    except RuntimeError as e:
        print("Custom field not available (points stay in card titles):", e)

    for title, desc in AGREEMENTS:
        t.call("POST", "/cards", {"idList": lists["📌 Team Agreements"], "name": title, "desc": desc, "pos": "bottom"})

    for eid, (name, _, goal) in EPICS.items():
        stories = [s for s in S if s[2] == eid]
        pts = sum(s[3] for s in stories)
        desc = f"**Goal:** {goal}\n\n**Total:** {pts} pts\n\n" + "\n".join(
            f"- {s[0]} ({s[3]}) {s[5]} · Sprint {s[1]}" for s in stories)
        t.call("POST", "/cards", {"idList": lists["🗺️ Epics"], "name": f"{eid} · {name}", "desc": desc,
                                  "idLabels": [labels[eid]], "pos": "bottom"})

    for sid, sprint, eid, pts, lbls, title, story, ac, notes in S:
        start, end = sprint_dates(sprint)
        desc = (f"**User story**\n{story}\n\n**Epic:** {eid} · {EPICS[eid][0]}\n"
                f"**Planned sprint:** {sprint} ({start:%d %b} → {end:%d %b})\n**Story points:** {pts}")
        if notes:
            desc += f"\n\n**Notes**\n{notes}"
        target = "🎯 Sprint 1 Backlog" if sprint == 1 else "📥 Product Backlog"
        card = t.call("POST", "/cards", {
            "idList": lists[target], "name": f"({pts}) {sid} · {title}", "desc": desc, "pos": "bottom",
            "due": f"{end.isoformat()}T17:00:00.000Z",
            "idLabels": [labels[l] for l in lbls] + [labels[eid]]})
        cl = t.call("POST", "/checklists", {"idCard": card["id"], "name": "Acceptance criteria"})
        for item in ac:
            t.call("POST", f"/checklists/{cl['id']}/checkItems", {"name": item})
        if points_field:
            t.call("PUT", f"/card/{card['id']}/customField/{points_field}/item", {"value": {"number": str(pts)}})
        print(f"  {sid} ({pts}) -> Sprint {sprint}")

    total = sum(s[3] for s in S)
    print(f"\nDone: {len(S)} stories, {total} pts across {SPRINT_COUNT} sprints. {board['url']}")


if __name__ == "__main__":
    main()
