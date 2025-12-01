# Subcult - AI Agent Instructions

## Project Overview

**Subcult** is a privacy-first platform for mapping underground music communities, combining real-time location awareness, live audio streaming, and trust-based discovery. The mission is to empower grassroots music scenes while respecting user privacy and enabling direct artist support.

## Core Principles

### 1. Privacy First

- **Geographic Privacy**: All public coordinates use geohash-based jitter (deterministic noise) to prevent precise location tracking
- **Consent Management**: Users explicitly opt-in for precise location sharing; default is always jittered
- **Minimal Data Collection**: Only collect what's necessary; avoid tracking user movements
- **Privacy-Aware Queries**: All location-based queries must respect consent flags and apply jitter by default

### 2. Trust-Based Discovery

- **Alliance System**: Users form alliances with role multipliers (organizer, artist, promoter, etc.)
- **Trust Scores**: Composite scores based on alliance strength and role weights
- **Ranking Integration**: Search results weighted by trust graph (feature-flagged for safe rollout)
- **No Centralized Curation**: Discovery driven by peer trust, not algorithmic feeds

### 3. Direct Artist Support

- **Stripe Connect**: Artists onboard with Express accounts for direct payments
- **Application Fees**: Transparent platform fee on transactions
- **Zero Platform Lock-in**: Artists own their payment relationships

### 4. Real-Time & Live

- **AT Protocol Ingestion**: Real-time commit streaming via Jetstream WebSocket
- **LiveKit Streaming**: Low-latency audio streaming for live performances
- **Immediate Availability**: Content appears on map within seconds of posting

## Tech Stack

### Backend

- **Language**: Go 1.22+
- **Router**: chi
- **Database**: Neon Postgres 16 with PostGIS
- **Config**: koanf
- **Logging**: structured logging with slog
- **Auth**: JWT access + refresh tokens with dual-key rotation

### Frontend

- **Framework**: Vite + React + TypeScript
- **Build**: SWC for fast compilation
- **Styling**: Tailwind CSS with dark mode support
- **State**: Zustand or Redux
- **Routing**: react-router
- **i18n**: i18next

### Infrastructure

- **Containerization**: Docker Compose for local dev
- **Reverse Proxy**: Caddy
- **Streaming**: LiveKit Cloud
- **Storage**: Cloudflare R2 for media uploads
- **Payments**: Stripe Connect Express
- **Maps**: MapLibre + MapTiler
- **Deployment**: VPS with zero-downtime rollout

### Integrations

- **AT Protocol**: Jetstream for real-time ingestion
- **LiveKit**: Audio streaming with participant tracking
- **Stripe**: Payment processing and Connect onboarding
- **Cloudflare R2**: Signed URLs for secure uploads

## Architecture Decisions

### Geographic Privacy Implementation

```text
User coordinates → Geohash (precision 7) → Deterministic jitter → Public display
Consent flag gates precise coordinates for trusted relationships only
```

### Trust Graph Flow

```text
Alliance creation → Role assignment → Weight calculation → Trust score update
Background job recomputes scores periodically with adaptive scheduling
```

### Search Ranking Formula

```text
composite_score = (text_relevance * 0.4) + (proximity_score * 0.3) + (recency * 0.2) + (trust_weight * 0.1)
Feature flag controls trust_weight inclusion; fallback to 0.0 when disabled
```

### Streaming Session Flow

```text
1. Client requests token → Backend generates LiveKit JWT
2. Client joins room → LiveKit cloud establishes connection
3. Session persisted in DB with metadata
4. Participant events tracked for analytics
5. Latency metrics collected (publish → playback)
```

## Development Workflow

### Issue Structure

- **Epics**: High-level features tracked in issues #2-#24
- **Tasks**: Granular implementation issues (#25+) linked to epics
- **Roadmap**: Master issue #1 links all epics and phases

### Code Organization

```text
cmd/              - Entry points (api, backfill)
internal/         - Private application code
  ├── api/        - HTTP handlers
  ├── auth/       - JWT, session management
  ├── db/         - Database access layer
  ├── geo/        - Geohash, jitter utilities
  ├── ingest/     - Jetstream client
  ├── ranking/    - Search ranking module
  ├── validate/   - Input validation
  └── ...
pkg/              - Reusable packages
web/              - Frontend React app
  ├── src/
  │   ├── components/
  │   ├── hooks/
  │   ├── services/
  │   └── stores/
migrations/       - Database migrations
scripts/          - Automation, dev tools
docs/             - Architecture, guides
configs/          - Config templates
perf/             - Performance baselines
```

### Testing Requirements

- **Backend**: >80% coverage; use testutil helpers
- **Frontend**: >70% coverage; React Testing Library
- **E2E**: Playwright smoke tests for critical paths
- **Load**: k6 scenarios for API and streaming
- **Integration**: Harness with mocked external services

### Performance Budgets

- **API Latency**: p95 <300ms
- **Stream Join**: <2s
- **Map Render**: <1.2s
- **FCP**: <1.0s
- **Trust Recompute**: <5m

### Security Practices

- **Input Validation**: All user input sanitized via validation layer
- **Rate Limiting**: Per-endpoint buckets with Redis backend
- **CORS**: Strict allowlist; no wildcard origins
- **CSP**: Report-only → enforce progression
- **Audit Logging**: Hash chain for tamper detection
- **Secret Rotation**: Dual JWT key support for zero-downtime rotation
- **Dependency Scanning**: govulncheck + npm audit in CI
- **SSRF Protection**: URL allowlist for external fetches

## Key Patterns & Conventions

### Error Handling

```go
// Use structured errors with context
return fmt.Errorf("failed to create scene: %w", err)

// Log errors with relevant fields
slog.Error("scene creation failed", "user_id", userID, "error", err)
```

### Geographic Operations

```go
// Always apply jitter for public display
jitteredCoords := geo.ApplyJitter(coords, consentLevel)

// Check consent before returning precise coordinates
if !user.HasPreciseLocationConsent(requesterID) {
    coords = geo.ApplyJitter(coords, geo.DefaultPrecision)
}
```

### Feature Flags

```go
// Gate experimental features
if config.Features.TrustRanking {
    score += trustWeight
}
```

### Metrics

```go
// Instrument critical paths
metrics.APILatency.Observe(elapsed.Seconds())
metrics.StreamJoinCount.Inc()
```

### Frontend State

- Use hooks for component-local state
- Zustand/Redux for global app state
- React Query for server state caching
- Minimize prop drilling; prefer context for cross-cutting concerns

### Styling with Tailwind

- Use Tailwind utilities for all styling
- Extend theme config for brand colors
- Dark mode via `dark:` prefix
- Prettier sorts classes automatically

## Common Tasks

### Adding a New API Endpoint

1. Define handler in `internal/api/`
2. Add route in router setup
3. Implement validation schema
4. Add database queries if needed
5. Write unit tests
6. Update OpenAPI spec
7. Add metrics instrumentation

### Adding a New Database Table

1. Create migration in `migrations/`
2. Update schema documentation
3. Add model struct in `internal/db/`
4. Implement query functions
5. Add indexes for performance
6. Write migration tests
7. Update `schema_version` tracking

### Adding a New Frontend Component

1. Create component in `web/src/components/`
2. Use Tailwind for styling
3. Implement accessibility (ARIA, keyboard nav)
4. Add i18n for user-facing text
5. Write unit tests with React Testing Library
6. Add to Storybook if needed

### Implementing a Search Feature

1. Add FTS indexes to relevant tables
2. Implement query in search provider
3. Integrate ranking formula module
4. Add feature flag if experimental
5. Capture EXPLAIN baseline for regression detection
6. Add calibration metrics
7. Write parity tests

## Git Workflow

### Branch Naming

- `feature/issue-123-short-description`
- `fix/issue-456-bug-description`
- `docs/issue-789-update-readme`

### Commit Messages

```text
type(scope): brief description

Longer explanation if needed.

Fixes #123
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### Pull Requests

- Link to issue: "Closes #123"
- Follow PR template (will be in #149)
- Ensure CI passes
- Request Copilot review for automated feedback
- Minimum one human review for production code

## Resources

- **Master Roadmap**: Issue #1
- **Architecture Docs**: Will be in `docs/ARCHITECTURE.md` (Issue #150)
- **API Reference**: Will be generated from OpenAPI (Issue #151)
- **Style Guide**: Will be in `docs/STYLE_GUIDE.md` (Issue #152)
- **Contributing Guide**: Will be in `CONTRIBUTING.md` (Issue #149)

## Current Status

**Phase**: Foundation & Initial Scaffolding  
**First Task**: #198 - Initial Project Scaffolding & Directory Structure  
**Total Issues**: 200 tasks across 24 epics  
**Active Development**: Beginning infrastructure setup

---

## Agent-Specific Guidance

When working on this codebase:

1. **Always prioritize privacy**: Check consent flags, apply jitter, avoid logging PII
2. **Follow security best practices**: Validate input, parameterize queries, check CORS
3. **Instrument everything**: Add metrics, structured logs, tracing spans
4. **Test thoroughly**: Unit, integration, and E2E coverage for all new features
5. **Document decisions**: Use ADR process for architectural choices (Issue #156)
6. **Check dependencies**: Reference linked issues before starting work
7. **Maintain performance**: Verify against budgets, capture EXPLAIN plans
8. **Respect feature flags**: Gate experimental features, provide fallbacks

**Question everything that might compromise user privacy or trust.**
