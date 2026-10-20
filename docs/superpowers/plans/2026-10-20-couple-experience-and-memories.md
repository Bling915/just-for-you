# Couple Experience and Shared Memories Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver the approved first- and second-stage product experience: real couple sharing and presence, editable memories with better uploads, shared planning tools, calendar/map/monthly retrospective views, and subscription-notification infrastructure.

**Architecture:** Keep one Go modular monolith and one MySQL database. Add `companionship` and `together` bounded contexts under the existing global DDD layers, extend `moment` for occurrence time/location/editing and read models, and expose small native-mini-program services/pages that consume stable `/v1` APIs.

**Tech Stack:** Go 1.22+, Gin, MySQL 8, Tencent COS SDK, native WeChat Mini Program JavaScript/WXML/WXSS, Node.js built-in test runner.

**Spec:** `docs/superpowers/specs/2026-10-20-couple-experience-and-memories-design.md`

## Global Constraints

- Preserve `domain / application / infrastructure / interfaces` package names and dependency direction.
- Keep one deployable Go service and one MySQL database; do not introduce microservices.
- Keep existing user, relationship, moment, and photo data compatible.
- Keep the three-tab navigation: 首页、时光、我的.
- Only relationship members may access couple data; only an author may edit or delete their moment.
- Treat `occurred_at` in `Asia/Shanghai` for calendar and monthly grouping.
- Do not place WeChat, COS, database, or signing secrets in source control or the mini-program bundle.
- Use TDD for every behavior change; run Go and mini-program suites before every task completion.
- Do not push, publish a mini-program build, or deploy to the server until the final verification task and explicit external-side-effect checkpoint.

## Review Focus

- An ended relationship must immediately lose access to presence, nudges, moments, planning data, and reports; pin this in Tasks 2, 3, 5, and 6.
- Dates around midnight, month boundaries, leap day, and UTC conversion must group under `Asia/Shanghai`; pin this in Tasks 5 and 6.
- Editing a moment concurrently must return a stable conflict and preserve the newer version; pin this in Task 4.
- Partial upload failures and compression failures must preserve recoverable local state without publishing missing photo keys; pin this in Task 4.
- Missing template IDs, denied subscription permission, and WeChat API failure must not fail the primary write; pin this in Task 7.

---

## File Structure

New backend bounded contexts:

```text
server/internal/domain/companionship/*
server/internal/application/companionship/*
server/internal/infrastructure/persistence/mysql/companionship/*
server/internal/interfaces/http/companionship/*

server/internal/domain/together/*
server/internal/application/together/*
server/internal/infrastructure/persistence/mysql/together/*
server/internal/interfaces/http/together/*
```

Extensions stay in their current owners:

```text
server/internal/domain/moment/*
server/internal/application/moment/*
server/internal/infrastructure/persistence/mysql/moment/*
server/internal/interfaces/http/moment/*
server/internal/domain/relationship/*
server/internal/application/relationship/*
server/internal/interfaces/http/relationship/*
server/internal/bootstrap/*
```

New mini-program capabilities:

```text
miniprogram/services/companionship.js
miniprogram/services/together.js
miniprogram/services/notifications.js
miniprogram/utils/image-compression.js
miniprogram/utils/calendar.js
miniprogram/pages/status-editor/*
miniprogram/pages/moment-edit/*
miniprogram/pages/calendar/*
miniprogram/pages/day-moments/*
miniprogram/pages/memory-map/*
miniprogram/pages/anniversaries/*
miniprogram/pages/todos/*
miniprogram/pages/wishes/*
miniprogram/pages/monthly-report/*
miniprogram/pages/notification-settings/*
```

## Task 1: Preserve the Timeline Refresh Fix and Establish Schema Evolution

**Files:**
- Modify: `miniprogram/pages/timeline/index.js`
- Create: `server/migrations/000002_couple_experience.up.sql`
- Create: `server/migrations/000002_couple_experience.down.sql`
- Create: `server/internal/architecture/migration_test.go`

**Interfaces:**
- Produces the columns and tables consumed by Tasks 2–7.
- Produces `moments.occurred_at`, `moments.updated_at`, `moments.latitude`, and `moments.longitude`.

- [ ] **Step 1: Write failing migration contract tests**

Add tests asserting the up migration creates the approved tables, backfills occurrence/update timestamps, defines relationship/time indexes, and that the down migration removes only version-2 objects.

- [ ] **Step 2: Run the focused test and verify RED**

Run: `cd server && go test ./internal/architecture -run TestCoupleExperienceMigration -v`

Expected: FAIL because migration `000002` does not exist.

- [ ] **Step 3: Add the compatible up/down migration**

Create `presences`, `nudges`, `notification_preferences`, `notification_deliveries`, `anniversaries`, `todos`, and `wishes`; extend `moments`; backfill before applying non-null constraints; add foreign keys and indexes from the spec.

- [ ] **Step 4: Verify the timeline page refreshes on every show**

Keep the existing working-tree fix that awaits `refresh()` after relationship validation. Confirm the page no longer relies on a one-time `initialized` flag.

- [ ] **Step 5: Run verification**

Run: `cd server && go test ./...`

Run: `npm test`

Expected: all tests PASS.

- [ ] **Step 6: Commit**

```bash
git add server/migrations server/internal/architecture/migration_test.go miniprogram/pages/timeline/index.js
git commit -m "feat: prepare shared experience data model"
```

## Task 2: Invitation Sharing and Partner Profile Read Model

**Files:**
- Modify: `server/internal/application/relationship/ports.go`
- Modify: `server/internal/application/relationship/get_current.go`
- Modify: `server/internal/infrastructure/persistence/mysql/relationship/mysql_repository.go`
- Modify: `server/internal/interfaces/http/relationship/handler.go`
- Modify: `server/internal/application/relationship/use_cases_test.go`
- Modify: `server/internal/interfaces/http/relationship/handler_test.go`
- Modify: `miniprogram/pages/invitation/index.js`
- Modify: `miniprogram/pages/invitation/index.wxml`
- Modify: `miniprogram/pages/home/index.js`
- Modify: `miniprogram/pages/home/index.wxml`
- Modify: `miniprogram/pages/bind/index.js`
- Modify: `miniprogram/services/relationships.js`
- Create: `miniprogram/utils/pending-invitation.js`
- Create: `miniprogram/tests/pending-invitation.test.js`

**Interfaces:**
- Produces `GET /v1/couples/me` response with `partner`, `boundAt`, and `soloTest`.
- Produces share path `/pages/welcome/index?invite=<CODE>` and pending-invite storage consumed after login.

- [ ] **Step 1: Write failing Go application and HTTP tests**

Cover real partner data, solo-test identity, ended relationship denial, and response omission of private OpenID data.

- [ ] **Step 2: Run focused Go tests and verify RED**

Run: `cd server && go test ./internal/application/relationship ./internal/interfaces/http/relationship -v`

Expected: FAIL because partner profile fields are absent.

- [ ] **Step 3: Implement the relationship read model**

Add repository projection and DTO mapping without expanding the relationship aggregate with presentation-only fields.

- [ ] **Step 4: Write failing mini-program pending-invite tests**

Cover capture from launch query, uppercase normalization, consume-on-bind, and persistence across login redirect.

- [ ] **Step 5: Run mini-program test and verify RED**

Run: `node --test miniprogram/tests/pending-invitation.test.js`

Expected: FAIL because the helper is absent.

- [ ] **Step 6: Implement sharing and partner presentation**

Add `onShareAppMessage`, pending-invite handling, bind-page prefill, partner card, relationship day count, and visible solo-test badge.

- [ ] **Step 7: Run complete suites and commit**

Run: `cd server && go test ./...`

Run: `npm test`

```bash
git add server miniprogram
git commit -m "feat: improve couple invitation and partner profile"
```

## Task 3: Presence and Lightweight Couple Interactions

**Files:**
- Create: `server/internal/domain/companionship/presence.go`
- Create: `server/internal/domain/companionship/nudge.go`
- Create: `server/internal/domain/companionship/errors.go`
- Create: `server/internal/domain/companionship/*_test.go`
- Create: `server/internal/application/companionship/ports.go`
- Create: `server/internal/application/companionship/update_presence.go`
- Create: `server/internal/application/companionship/send_nudge.go`
- Create: `server/internal/application/companionship/get_snapshot.go`
- Create: `server/internal/application/companionship/use_cases_test.go`
- Create: `server/internal/infrastructure/persistence/mysql/companionship/repository.go`
- Create: `server/internal/infrastructure/persistence/mysql/companionship/repository_test.go`
- Create: `server/internal/interfaces/http/companionship/handler.go`
- Create: `server/internal/interfaces/http/companionship/handler_test.go`
- Create: `server/internal/bootstrap/companionship.go`
- Modify: `server/internal/bootstrap/application.go`
- Modify: `server/internal/interfaces/http/shared/router.go`
- Create: `miniprogram/services/companionship.js`
- Create: `miniprogram/pages/status-editor/index.{js,json,wxml,wxss}`
- Modify: `miniprogram/pages/home/index.{js,wxml,wxss}`
- Modify: `miniprogram/app.json`
- Create: `miniprogram/tests/companionship.test.js`

**Interfaces:**
- Produces `PUT /v1/presence`, `GET /v1/presence`, and `POST /v1/nudges`.
- `PresenceInput { preset, customText, expiresAt }` and `NudgeInput { kind }`.
- Produces companionship snapshot for the home-page read model.

- [ ] **Step 1: Write domain tests and verify RED**

Test preset/custom validation, 20-rune limit, future expiry, expiry rendering, allowed nudge kinds, and receiver derivation from active relationship.

Run: `cd server && go test ./internal/domain/companionship -v`

Expected: FAIL because package is absent.

- [ ] **Step 2: Implement domain entities and verify GREEN**

- [ ] **Step 3: Write application/repository/HTTP failing tests**

Cover upsert, active relationship enforcement, ended relationship denial, latest effective presence, nudge throttling, and stable error codes.

Run: `cd server && go test ./internal/application/companionship ./internal/infrastructure/persistence/mysql/companionship ./internal/interfaces/http/companionship -v`

Expected: FAIL because use cases and adapters are absent.

- [ ] **Step 4: Implement ports, use cases, MySQL repository, handlers, and bootstrap wiring**

- [ ] **Step 5: Write failing client/service tests**

Test request payload normalization and expired-status view state.

Run: `node --test miniprogram/tests/companionship.test.js`

Expected: FAIL because service/helpers are absent.

- [ ] **Step 6: Implement status editor and home quick interactions**

Use accessible buttons, optimistic tap feedback, rollback on failure, and lavender visual tokens already in the app.

- [ ] **Step 7: Run full suites and commit**

Run: `cd server && go test ./...`

Run: `npm test`

```bash
git add server miniprogram
git commit -m "feat: add couple presence and nudges"
```

## Task 4: Editable Moments, Backdating, Coordinates, and Resilient Uploads

**Files:**
- Modify: `server/internal/domain/moment/moment.go`
- Modify: `server/internal/domain/moment/errors.go`
- Modify: `server/internal/domain/moment/moment_test.go`
- Modify: `server/internal/application/moment/ports.go`
- Create: `server/internal/application/moment/update_moment.go`
- Modify: `server/internal/application/moment/use_cases_test.go`
- Modify: `server/internal/infrastructure/persistence/mysql/moment/repository.go`
- Modify: `server/internal/infrastructure/persistence/mysql/moment/repository_test.go`
- Modify: `server/internal/interfaces/http/moment/handler.go`
- Modify: `server/internal/interfaces/http/moment/handler_test.go`
- Modify: `miniprogram/services/moments.js`
- Modify: `miniprogram/services/uploads.js`
- Modify: `miniprogram/utils/upload-queue.js`
- Create: `miniprogram/utils/image-compression.js`
- Create: `miniprogram/pages/moment-edit/index.{js,json,wxml,wxss}`
- Modify: `miniprogram/pages/compose/index.{js,wxml,wxss}`
- Modify: `miniprogram/components/photo-picker/index.{js,wxml,wxss}`
- Modify: `miniprogram/components/moment-card/index.js`
- Modify: `miniprogram/app.json`
- Modify: `miniprogram/tests/upload-queue.test.js`
- Create: `miniprogram/tests/image-compression.test.js`

**Interfaces:**
- Produces `PATCH /v1/moments/:id` with `{ text, status, place, latitude, longitude, photoKeys, occurredAt, version }`.
- Returns `updatedAt` as the next optimistic-lock version.
- Upload queue items expose `{ status, progress, errorCode, objectKey }`.

- [ ] **Step 1: Write failing moment domain tests**

Cover author-only update, future/past occurrence bounds, coordinate pairs and ranges, photo reorder, and immutable relationship/author identity.

Run: `cd server && go test ./internal/domain/moment -v`

Expected: FAIL because update behavior is absent.

- [ ] **Step 2: Implement domain update behavior**

- [ ] **Step 3: Write failing application, repository, and HTTP tests**

Cover successful update, non-author denial, ended relationship denial, optimistic conflict, atomic photo/grant update, and DTO timestamps.

Run: `cd server && go test ./internal/application/moment ./internal/infrastructure/persistence/mysql/moment ./internal/interfaces/http/moment -v`

Expected: FAIL for missing update use case.

- [ ] **Step 4: Implement backend update path**

Use a transaction and `WHERE updated_at = ?` optimistic check. Preserve old record when any new upload grant is invalid.

- [ ] **Step 5: Write failing upload/compression tests**

Cover compression success/fallback, request-task progress mapping, per-file retry, partial failures, and preventing publish until all selected photos succeed.

Run: `node --test miniprogram/tests/image-compression.test.js miniprogram/tests/upload-queue.test.js`

Expected: FAIL on missing compression/progress behavior.

- [ ] **Step 6: Implement mini-program edit and upload experience**

Use `wx.compressImage`, `wx.chooseLocation`, request progress callbacks, date/time selectors, conflict recovery, and draft preservation.

- [ ] **Step 7: Run complete suites and commit**

Run: `cd server && go test ./...`

Run: `npm test`

```bash
git add server miniprogram
git commit -m "feat: edit and backdate shared moments"
```

## Task 5: Anniversaries, Shared Todos, and Wishes

**Files:**
- Create: `server/internal/domain/together/anniversary.go`
- Create: `server/internal/domain/together/todo.go`
- Create: `server/internal/domain/together/wish.go`
- Create: `server/internal/domain/together/errors.go`
- Create: `server/internal/domain/together/*_test.go`
- Create: `server/internal/application/together/ports.go`
- Create: `server/internal/application/together/anniversaries.go`
- Create: `server/internal/application/together/todos.go`
- Create: `server/internal/application/together/wishes.go`
- Create: `server/internal/application/together/use_cases_test.go`
- Create: `server/internal/infrastructure/persistence/mysql/together/repository.go`
- Create: `server/internal/infrastructure/persistence/mysql/together/repository_test.go`
- Create: `server/internal/interfaces/http/together/handler.go`
- Create: `server/internal/interfaces/http/together/handler_test.go`
- Create: `server/internal/bootstrap/together.go`
- Modify: `server/internal/bootstrap/application.go`
- Modify: `server/internal/interfaces/http/shared/router.go`
- Create: `miniprogram/services/together.js`
- Create: `miniprogram/pages/anniversaries/index.{js,json,wxml,wxss}`
- Create: `miniprogram/pages/todos/index.{js,json,wxml,wxss}`
- Create: `miniprogram/pages/wishes/index.{js,json,wxml,wxss}`
- Modify: `miniprogram/pages/home/index.{js,wxml,wxss}`
- Modify: `miniprogram/pages/me/index.{js,wxml,wxss}`
- Modify: `miniprogram/app.json`
- Create: `miniprogram/tests/together.test.js`

**Interfaces:**
- Produces CRUD `/v1/anniversaries`, `/v1/todos`, and `/v1/wishes` endpoints from the spec.
- Produces next-anniversary summary consumed by home.

- [ ] **Step 1: Write domain tests and verify RED**

Cover title limits, creator/delete permissions, shared completion/reopen, wish-to-moment validation, leap-day anniversary calculation, past deadlines, and ended relationships.

Run: `cd server && go test ./internal/domain/together -v`

Expected: FAIL because package is absent.

- [ ] **Step 2: Implement domain entities and calculations**

- [ ] **Step 3: Write failing application/repository/HTTP tests**

Cover isolation by relationship, system binding anniversary immutability, default sorting, idempotent state transitions, and stable errors.

Run: `cd server && go test ./internal/application/together ./internal/infrastructure/persistence/mysql/together ./internal/interfaces/http/together -v`

Expected: FAIL because adapters are absent.

- [ ] **Step 4: Implement backend bounded context and wire it**

- [ ] **Step 5: Write failing mini-program service/view-model tests**

Run: `node --test miniprogram/tests/together.test.js`

Expected: FAIL because service/view helpers are absent.

- [ ] **Step 6: Implement the three pages and home/me entry cards**

Keep forms small, completion tactile, completed items collapsible, and empty states actionable.

- [ ] **Step 7: Run full suites and commit**

Run: `cd server && go test ./...`

Run: `npm test`

```bash
git add server miniprogram
git commit -m "feat: add shared anniversaries todos and wishes"
```

## Task 6: Calendar, Map, and Monthly Retrospective Read Models

**Files:**
- Create: `server/internal/application/moment/get_calendar.go`
- Create: `server/internal/application/moment/get_map.go`
- Create: `server/internal/application/moment/get_monthly_report.go`
- Modify: `server/internal/application/moment/ports.go`
- Modify: `server/internal/application/moment/use_cases_test.go`
- Modify: `server/internal/infrastructure/persistence/mysql/moment/repository.go`
- Modify: `server/internal/infrastructure/persistence/mysql/moment/repository_test.go`
- Modify: `server/internal/interfaces/http/moment/handler.go`
- Modify: `server/internal/interfaces/http/moment/handler_test.go`
- Create: `miniprogram/utils/calendar.js`
- Create: `miniprogram/pages/calendar/index.{js,json,wxml,wxss}`
- Create: `miniprogram/pages/day-moments/index.{js,json,wxml,wxss}`
- Create: `miniprogram/pages/memory-map/index.{js,json,wxml,wxss}`
- Create: `miniprogram/pages/monthly-report/index.{js,json,wxml,wxss}`
- Modify: `miniprogram/pages/timeline/index.{js,wxml,wxss}`
- Modify: `miniprogram/services/moments.js`
- Modify: `miniprogram/app.json`
- Create: `miniprogram/tests/calendar.test.js`
- Create: `miniprogram/tests/monthly-report.test.js`

**Interfaces:**
- Produces `GET /v1/moments/calendar?month=YYYY-MM`.
- Produces `GET /v1/moments/map?from=&to=`.
- Produces `GET /v1/reports/monthly?month=YYYY-MM`.
- Consumes together summaries through an application read port, never a MySQL cross-package repository.

- [ ] **Step 1: Write failing read-model tests**

Cover Shanghai month boundaries, leap day, empty month, photo counts, common-status tie-breaking, place de-duplication, completed todo/wish summaries, and ended relationship denial.

Run: `cd server && go test ./internal/application/moment ./internal/infrastructure/persistence/mysql/moment ./internal/interfaces/http/moment -v`

Expected: FAIL because queries are absent.

- [ ] **Step 2: Implement repository projections, use cases, and endpoints**

Use bounded result sizes and signed thumbnail URLs. Do not return private moment bodies in the map summary unless the requesting member opens that moment.

- [ ] **Step 3: Write failing calendar/report helper tests**

Test month grid generation, today/selected markers, day grouping, report empty state, and deterministic highlights.

Run: `node --test miniprogram/tests/calendar.test.js miniprogram/tests/monthly-report.test.js`

Expected: FAIL because helpers are absent.

- [ ] **Step 4: Implement calendar, day list, map, and monthly report pages**

Use native `map`, lazy-load markers, cluster nearby coordinates in the page model, and render a vertically shareable report card without leaking data to unauthenticated viewers.

- [ ] **Step 5: Run full suites and commit**

Run: `cd server && go test ./...`

Run: `npm test`

```bash
git add server miniprogram
git commit -m "feat: add calendar map and monthly memories"
```

## Task 7: Subscription Notification Preferences and Delivery Adapter

**Files:**
- Create: `server/internal/application/companionship/notification_ports.go`
- Create: `server/internal/application/companionship/notification_preferences.go`
- Modify: `server/internal/application/companionship/update_presence.go`
- Modify: `server/internal/application/companionship/send_nudge.go`
- Modify: `server/internal/application/moment/publish_moment.go`
- Create: `server/internal/infrastructure/notification/wechat/client.go`
- Create: `server/internal/infrastructure/notification/wechat/client_test.go`
- Create: `server/internal/infrastructure/persistence/mysql/companionship/notification_repository.go`
- Modify: `server/internal/interfaces/http/companionship/handler.go`
- Modify: `server/internal/infrastructure/config/config.go`
- Modify: `server/internal/infrastructure/config/config_test.go`
- Modify: `server/internal/bootstrap/companionship.go`
- Modify: `server/internal/bootstrap/moment.go`
- Create: `miniprogram/services/notifications.js`
- Create: `miniprogram/pages/notification-settings/index.{js,json,wxml,wxss}`
- Modify: `miniprogram/pages/me/index.{js,wxml}`
- Modify: `miniprogram/app.json`
- Create: `miniprogram/tests/notifications.test.js`
- Modify: `.env.example`

**Interfaces:**
- Produces `GET/PUT /v1/notification-preferences`.
- Defines `NotificationSender.Send(ctx, recipient, scene, data) error` and a no-op implementation when templates are absent.
- Adds environment variables for three approved template IDs without committing their values.

- [ ] **Step 1: Write failing application tests**

Cover disabled preference, missing template, denied authorization, delivery failure, and proof that the primary presence/nudge/moment write still succeeds.

Run: `cd server && go test ./internal/application/companionship ./internal/application/moment -v`

Expected: FAIL because notification ports are absent.

- [ ] **Step 2: Implement preferences and best-effort dispatch**

Persist an auditable delivery result without retrying synchronously inside the user request.

- [ ] **Step 3: Write failing WeChat adapter tests**

Use an HTTP test server to cover access-token retrieval, timeout, non-2xx responses, WeChat business error codes, and template payload mapping.

Run: `cd server && go test ./internal/infrastructure/notification/wechat -v`

Expected: FAIL because adapter is absent.

- [ ] **Step 4: Implement adapter, config, repositories, handlers, and bootstrap**

- [ ] **Step 5: Write failing mini-program notification tests**

Cover template availability, `wx.requestSubscribeMessage` accept/reject result normalization, and preference update behavior.

Run: `node --test miniprogram/tests/notifications.test.js`

Expected: FAIL because service/helper is absent.

- [ ] **Step 6: Implement notification settings page with graceful unavailable state**

- [ ] **Step 7: Run full suites and commit**

Run: `cd server && go test ./...`

Run: `npm test`

```bash
git add server miniprogram .env.example
git commit -m "feat: add subscription notification infrastructure"
```

## Task 8: Integrated Navigation, Visual Polish, and Regression Coverage

**Files:**
- Modify: `miniprogram/pages/home/index.{js,wxml,wxss}`
- Modify: `miniprogram/pages/timeline/index.{js,wxml,wxss}`
- Modify: `miniprogram/pages/me/index.{js,wxml,wxss}`
- Modify: `miniprogram/app.wxss`
- Create or modify: `miniprogram/tests/app-state.test.js`
- Create: `miniprogram/tests/navigation-model.test.js`
- Modify: `README.md`
- Create: `docs/testing/phase-one-two-manual-checklist.md`

**Interfaces:**
- Consumes all earlier mini-program services/pages.
- Produces the final three-tab information architecture and manual acceptance checklist.

- [ ] **Step 1: Write failing navigation/state tests**

Cover logout and relationship-end clearing every new private store/cache, home shortcuts, time-view entries, me-page entries, and no dead page paths.

Run: `node --test miniprogram/tests/app-state.test.js miniprogram/tests/navigation-model.test.js`

Expected: FAIL until all private state and routes are registered.

- [ ] **Step 2: Implement integrated navigation and state cleanup**

- [ ] **Step 3: Apply consistent lavender visual styling and restrained motion**

Reuse existing tokens and components; respect reduced-animation behavior by keeping all motion nonessential.

- [ ] **Step 4: Document local/server configuration and manual acceptance**

Document migration, environment variables, two-account testing, solo-mode testing, location permission, and notification-template limitations without including secrets.

- [ ] **Step 5: Run complete automated verification**

Run: `cd server && go test ./...`

Run: `npm test`

Run: `git diff --check`

Expected: all PASS with no whitespace errors.

- [ ] **Step 6: Compile and manually inspect in WeChat Developer Tools**

Verify no compile errors and complete the checklist using the current solo relationship. Then test invitation/binding boundaries that do not require sending an external message.

- [ ] **Step 7: Commit**

```bash
git add miniprogram README.md docs/testing
git commit -m "feat: integrate shared couple experience"
```

## Task 9: Whole-Branch Review and Deployment Preparation

**Files:**
- Modify only files required by review findings.
- Update: `docs/testing/phase-one-two-manual-checklist.md`

**Interfaces:**
- Validates all tasks as one release candidate.
- Does not push, deploy, or upload a mini-program build without the explicit external-side-effect checkpoint.

- [ ] **Step 1: Run the complete verification matrix**

Run: `cd server && go test -race ./...`

Run: `npm test`

Run: `git diff --check`

Run: `git status --short`

- [ ] **Step 2: Review architecture and security-sensitive boundaries**

Check DDD dependency rules, relationship isolation, ownership, secret handling, signed URLs, notification best-effort behavior, and migration reversibility.

- [ ] **Step 3: Perform one review-fix pass under TDD**

For every Critical or Important finding, add a failing regression test, verify RED, apply the smallest fix, and rerun the complete suites. Record Minor findings without silently expanding scope.

- [ ] **Step 4: Validate migration against a disposable MySQL database**

Apply `000001` then `000002`, run representative queries, apply the `000002` down migration, and confirm version-1 data remains intact.

- [ ] **Step 5: Prepare deployment notes and stop at the external-side-effect checkpoint**

List required environment variables, migration command, binary/service restart steps, WeChat request/upload/download domains, and subscription-template IDs. Ask before pushing to GitHub, running production migration, deploying the server, or uploading the mini-program build.

