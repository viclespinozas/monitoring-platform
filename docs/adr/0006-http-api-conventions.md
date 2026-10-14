# ADR-0006: HTTP API conventions

**Status:** Accepted
**Date:** 2026-10-14

## Context

The service registry is the first public API. The conventions chosen here
will be copied by every later endpoint (uptime, alert rules, incidents), so
they are worth fixing once.

## Decision

- **Router**: the standard library `net/http` `ServeMux` (Go 1.22+ method and
  path patterns such as `GET /api/v1/services/{id}`). No web framework.
- **Layering**: handler -> `Manager` (validation, defaults, business rules) ->
  `Repository` interface (SQL). Handlers never touch SQL; the repository never
  validates. `context.Context` is passed through every layer.
- **Errors**: always RFC 9457 `application/problem+json`. Mapping:
  malformed JSON, unknown fields or trailing data -> 400; validation -> 422 with
  an `errors` list of `{field, message}`; unknown id -> 404; duplicate live
  name -> 409; body over 1 MiB -> 413; anything else -> opaque 500, with the real
  error logged server-side only.
- **Updates**: `PUT` replaces the whole resource; omitted optional fields take
  their defaults. No `PATCH` until something needs it.
- **Lists**: `limit` (default 20, max 100, larger values clamped) and `offset`
  with filters as query parameters. The response carries `total`, `limit` and
  `offset` so clients can page.
- **Contract**: `api/openapi.yaml` is hand-written and committed.
- **Defaults via zero values**: an omitted or `0` interval, timeout or status
  means "use the default". These fields have no valid zero value, so the
  ambiguity is harmless and avoids pointer fields.

## Consequences

- Standard library only: fewer dependencies and the handlers are easy to test
  with `httptest`. We write a little more plumbing (middleware, JSON helpers)
  ourselves.
- Offset paging gets slow and can skip or repeat rows when data changes during
  paging. Fine at registry scale; high-volume endpoints (check results) should
  use keyset paging instead.
- The OpenAPI file can drift from the code because nothing generates it. Until
  a contract test exists, reviewers must check both together.
- Revisit the router choice if middleware chains or route groups get unwieldy.
