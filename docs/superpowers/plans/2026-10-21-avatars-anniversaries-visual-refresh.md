# Custom Avatars, Anniversaries, and Visual Refresh Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add independently managed user avatars, a complete anniversary experience, and a lavender-album refresh while keeping shared daily journals as the primary product flow.

**Architecture:** Extend the existing `identity` and `together` DDD contexts rather than introducing cross-layer shortcuts. Reuse the COS signing infrastructure behind context-specific application ports, expose anniversary highlights through the home read model, and keep Mini Program date/poster transformations in tested pure modules while pages orchestrate state only.

**Tech Stack:** Go, Gin, MySQL migrations, Tencent COS signed PUT/GET URLs, native WeChat Mini Program WXML/WXSS/CommonJS, Canvas 2D, Node test runner.

**Spec:** `docs/superpowers/specs/2026-10-21-avatars-anniversaries-visual-refresh-design.md`

## Global Constraints

- Shared daily journals remain the primary home content; anniversary UI must not block status or journal loading.
- Calendar dates use Asia/Shanghai natural-day semantics and Gregorian dates only.
- A user may update only their own avatar; couple resources remain isolated by active relationship.
- Store stable COS object keys, never long-lived signed URLs or cloud credentials.
- Do not modify previously executed migrations; add forward and rollback migration `000003`.
- Keep native Mini Program implementation and the existing Go DDD package names.
- Do not commit the developer-local `miniprogram/config.js` LAN address.
- Motion duration is 140–240ms and respects a shared `reducedMotion` preference.

## Review Focus

- A failed or interrupted avatar/cover upload retains the previous image and never commits an unverified key; pinned by Tasks 3 and 5.
- February 29, same-day, year-boundary, and timezone inputs produce stable elapsed/countdown values; pinned by Tasks 1 and 4.
- An anniversary ID from another couple cannot be read, edited, deleted, or used for a cover grant; pinned by Task 2.
- Missing anniversary highlights do not turn an otherwise successful home response into an error; pinned by Task 6.
- Long titles/messages and missing photos render without overflow and still export a poster; pinned by Tasks 5 and 7.

---

## File Structure

### Backend

- `server/migrations/000003_avatar_anniversary_experience.{up,down}.sql`: additive anniversary fields plus upload-grant purpose/resource fields and rollback.
- `server/internal/domain/together/anniversary.go`: anniversary aggregate, update rules, system-item protection, and calendar behavior.
- `server/internal/application/together/anniversary_*.go`: commands, read models, sorting, uploads, and use cases split from unrelated todo/wish service code.
- `server/internal/application/together/ports.go`: repository, storage, relationship, and clock contracts.
- `server/internal/infrastructure/persistence/mysql/together/repository.go`: anniversary persistence scoped by couple.
- `server/internal/interfaces/http/together/handler.go`: anniversary endpoints and DTO mapping.
- `server/internal/application/identity/request_avatar_upload.go`: avatar upload grant use case.
- `server/internal/domain/identity/user.go`: avatar key ownership validation.
- `server/internal/interfaces/http/identity/handler.go`: avatar grant endpoint and profile response URL.
- `server/internal/application/relationship/get_relationship.go` and its HTTP adapter: signed partner-avatar read URL instead of exposing an object key as a URL.
- `server/internal/infrastructure/storage/cos/photo_storage.go`: generic signed upload/read primitives reused behind application ports.
- `server/internal/application/moment/get_home.go`: non-blocking anniversary highlight aggregation.
- `server/internal/bootstrap/{identity,together,moment}.go`: explicit dependency wiring.

### Mini Program

- `miniprogram/utils/anniversary.js`: pure date, sorting, and view-model functions.
- `miniprogram/utils/avatar-upload.js`: avatar selection/upload transaction.
- `miniprogram/utils/anniversary-poster.js`: pure Canvas drawing model and text fitting.
- `miniprogram/services/{users,together}.js`: new API calls.
- `miniprogram/components/avatar-picker/*`: reusable avatar control.
- `miniprogram/components/couple-header/*`: paired identity header.
- `miniprogram/components/pinned-anniversary-card/*`: compact home highlight.
- `miniprogram/components/anniversary-card/*`: list card.
- `miniprogram/components/anniversary-frame/*`: detail visual and poster template selection.
- `miniprogram/pages/anniversaries/*`: redesigned list and editor entry.
- `miniprogram/pages/anniversary-edit/*`: create/update form.
- `miniprogram/pages/anniversary-detail/*`: frame, preview, and save flow.
- `miniprogram/pages/{home,profile-setup,me}/*` and `miniprogram/components/moment-card/*`: avatar and lavender-album integration.
- `miniprogram/styles/{tokens,animations}.wxss`: shared surfaces and reduced-motion contract.

### Task 1: Anniversary Domain and Migration

**Files:**
- Create: `server/migrations/000003_avatar_anniversary_experience.up.sql`
- Create: `server/migrations/000003_avatar_anniversary_experience.down.sql`
- Modify: `server/internal/domain/together/anniversary.go`
- Modify: `server/internal/domain/together/errors.go`
- Test: `server/internal/domain/together/anniversary_test.go`
- Test: `server/internal/architecture/migration_test.go`

**Interfaces:**
- Produces: `CountMode` constants `elapsed` and `countdown`; `AnniversaryDraft`; `Anniversary.Update(d AnniversaryDraft, now time.Time) error`; getters for `CountMode`, `Pinned`, `CoverRef`, and `Message`.
- Produces: migration columns `count_mode`, `pinned`, `cover_ref`, `message` with backward-compatible defaults; `upload_grants.purpose` and nullable `resource_id` distinguish `moment_photo`, `avatar`, and `anniversary_cover` grants.

- [ ] **Step 1: Write failing domain and migration tests** for trimmed 1–60 rune titles, 0–120 rune messages, valid count modes, system anniversary delete protection, Feb 29 next occurrence, same-day countdown, and migration up/down presence.
- [ ] **Step 2: Run `cd server && go test ./internal/domain/together ./internal/architecture`** and verify the new tests fail because the fields and migration are missing.
- [ ] **Step 3: Implement the expanded aggregate and additive migration**, retaining `RestoreAnniversary` as the persistence boundary, mapping existing anniversary rows to `countdown`, unpinned, empty cover/message, and mapping existing upload grants to `moment_photo`.
- [ ] **Step 4: Run `cd server && go test ./internal/domain/together ./internal/architecture`** and expect PASS.
- [ ] **Step 5: Commit** with `git commit -m "feat: expand anniversary domain model"`.

### Task 2: Anniversary Application, Persistence, and HTTP API

**Files:**
- Create: `server/internal/application/together/anniversary_commands.go`
- Create: `server/internal/application/together/anniversary_queries.go`
- Modify: `server/internal/application/together/ports.go`
- Modify: `server/internal/application/together/service.go`
- Modify: `server/internal/infrastructure/persistence/mysql/together/repository.go`
- Modify: `server/internal/interfaces/http/together/handler.go`
- Test: `server/internal/application/together/use_cases_test.go`
- Test: `server/internal/infrastructure/persistence/mysql/together/repository_test.go`
- Create: `server/internal/interfaces/http/together/handler_test.go`

**Interfaces:**
- Consumes: Task 1 `AnniversaryDraft` and aggregate update rules.
- Produces: `CreateAnniversary(ctx, userID, AnniversaryDraft)`, `UpdateAnniversary(ctx, userID, id, AnniversaryDraft)`, `GetAnniversary(ctx, userID, id)`, `ListAnniversaries`, and `DeleteAnniversary`.
- Produces: `Repository.FindAnniversary`, `UpdateAnniversary`, and couple-scoped delete; HTTP `GET/POST /v1/anniversaries`, `GET/PATCH/DELETE /v1/anniversaries/:id`.

- [ ] **Step 1: Write failing application tests** proving stable pinned/countdown sorting, edit authorization for either member of the same couple, foreign-couple rejection, and system-item delete rejection.
- [ ] **Step 2: Write failing repository and handler tests** for all fields, not-found mapping, invalid enum/message/title, and cross-couple IDs.
- [ ] **Step 3: Run `cd server && go test ./internal/application/together ./internal/infrastructure/persistence/mysql/together ./internal/interfaces/http/together`** and verify failures identify missing interfaces/routes.
- [ ] **Step 4: Implement use cases, parameterized persistence, DTOs, and routes**; split anniversary orchestration out of the general service without changing todo/wish behavior.
- [ ] **Step 5: Rerun the focused Go tests** and expect PASS.
- [ ] **Step 6: Commit** with `git commit -m "feat: add complete anniversary API"`.

### Task 3: Secure Avatar Upload Backend

**Files:**
- Create: `server/internal/application/identity/request_avatar_upload.go`
- Modify: `server/internal/application/identity/ports.go`
- Modify: `server/internal/application/identity/update_profile.go`
- Modify: `server/internal/domain/identity/user.go`
- Modify: `server/internal/infrastructure/persistence/mysql/identity/repository.go`
- Modify: `server/internal/interfaces/http/identity/handler.go`
- Modify: `server/internal/bootstrap/identity.go`
- Modify: `server/internal/application/relationship/get_relationship.go`
- Modify: `server/internal/application/relationship/ports.go`
- Modify: `server/internal/interfaces/http/relationship/handler.go`
- Modify: `server/internal/bootstrap/relationship.go`
- Test: `server/internal/application/identity/use_cases_test.go`
- Test: `server/internal/infrastructure/persistence/mysql/identity/repository_test.go`
- Test: `server/internal/interfaces/http/identity/handler_test.go`

**Interfaces:**
- Produces: `AvatarStorage.SignUpload(ctx, key, contentType, expiresAt)` and `VerifyUploaded(ctx, key, contentType, maxSize)`.
- Produces: `RequestAvatarUploadCommand{UserID, ContentType, Size}` and `UploadGrant{ObjectKey, UploadURL, ContentType, ExpiresAt}`.
- Produces: `POST /v1/users/me/avatar-upload`; profile update atomically consumes a verified, unexpired `avatar` grant for `users/{userID}/avatars/...`, or accepts the unchanged existing key.
- Produces: `AvatarURLSigner.SignRead(ctx, key, expiresAt)` used by both own-profile and relationship read models; JSON exposes `avatarRef` for subsequent writes and `avatarUrl` for rendering, while empty keys remain empty strings.

- [ ] **Step 1: Write failing use-case tests** for JPEG/PNG/WebP up to 5 MiB, wrong user prefix, wrong grant purpose, expired/consumed grant, unverified object, interrupted upload retaining the prior avatar, and a valid verified replacement that consumes its grant.
- [ ] **Step 2: Write failing persistence and handler tests** for atomic grant consumption/profile update, grant success, invalid MIME/size, signed `avatarUrl`, and invalid profile avatar responses.
- [ ] **Step 3: Run `cd server && go test ./internal/application/identity ./internal/interfaces/http/identity`** and verify RED.
- [ ] **Step 4: Implement the grant, ownership validation, verification-before-profile-update, atomic grant consumption/profile update, response URL signing, and bootstrap wiring** by adapting the existing COS storage through identity and relationship ports.
- [ ] **Step 5: Rerun focused tests** and expect PASS.
- [ ] **Step 6: Commit** with `git commit -m "feat: add secure custom avatar uploads"`.

### Task 4: Mini Program Anniversary Model and Services

**Files:**
- Create: `miniprogram/utils/anniversary.js`
- Create: `miniprogram/tests/anniversary.test.js`
- Modify: `miniprogram/services/together.js`

**Interfaces:**
- Consumes: Task 2 anniversary DTOs.
- Produces: `anniversaryView(item, now)`, `sortAnniversaries(items, now)`, `splitAnniversaries(items, now)`, and service methods `getAnniversary`, `createAnniversary`, `updateAnniversary`, `deleteAnniversary`.

- [ ] **Step 1: Write failing Node tests** for Asia/Shanghai same-day output, elapsed day one, year-boundary countdown, Feb 29 fallback, multiple pinned ordering, and invalid/missing dates.
- [ ] **Step 2: Run `node --test miniprogram/tests/anniversary.test.js`** and verify RED because the module is absent.
- [ ] **Step 3: Implement the pure model and HTTP service methods** with no direct `wx` dependency in the utility.
- [ ] **Step 4: Rerun the focused Node test** and expect PASS.
- [ ] **Step 5: Commit** with `git commit -m "feat: add anniversary client model"`.

### Task 5: Anniversary Management UI and Cover Upload

**Files:**
- Create: `server/internal/application/together/request_cover_upload.go`
- Modify: `server/internal/application/together/ports.go`
- Modify: `server/internal/interfaces/http/together/handler.go`
- Modify: `server/internal/bootstrap/together.go`
- Create: `miniprogram/components/anniversary-card/index.{js,json,wxml,wxss}`
- Replace: `miniprogram/pages/anniversaries/index.{js,json,wxml,wxss}`
- Create: `miniprogram/pages/anniversary-edit/index.{js,json,wxml,wxss}`
- Modify: `miniprogram/app.json`
- Test: `server/internal/application/together/use_cases_test.go`
- Test: `server/internal/interfaces/http/together/handler_test.go`
- Create: `miniprogram/tests/anniversary-editor.test.js`

**Interfaces:**
- Consumes: Tasks 2 and 4 CRUD interfaces; existing upload PUT helper.
- Produces: `POST /v1/anniversaries/:id/cover-upload`; the following PATCH atomically consumes that anniversary's verified, unexpired `anniversary_cover` grant; editor draft mapper `buildAnniversaryCommand(form)`; reusable anniversary list card events.

- [ ] **Step 1: Write failing backend tests** proving cover grants require same-couple ownership, supported image type/size, matching resource ID/purpose, single consumption, and that a failed upload never changes `coverRef`.
- [ ] **Step 2: Write failing client tests** for title/message limits, create/edit payloads, empty-cover behavior, and long-title view data.
- [ ] **Step 3: Run focused Go and Node tests** and verify RED.
- [ ] **Step 4: Implement the cover grant endpoint and transactional client flow**: compress, request grant, PUT, then PATCH the anniversary; retain the prior cover on failure.
- [ ] **Step 5: Build the lavender-album list and editor pages**, including pinned/all sections, empty state, native back navigation, loading/disabled states, and destructive confirmation.
- [ ] **Step 6: Run the focused tests plus `npm test`** and expect PASS.
- [ ] **Step 7: Commit** with `git commit -m "feat: build anniversary management experience"`.

### Task 6: Home Anniversary Highlights Without Journal Regression

**Files:**
- Modify: `server/internal/application/moment/ports.go`
- Modify: `server/internal/application/moment/get_home.go`
- Modify: `server/internal/interfaces/http/moment/handler.go`
- Modify: `server/internal/bootstrap/moment.go`
- Create: `server/internal/infrastructure/persistence/mysql/together/highlight_reader.go`
- Test: `server/internal/application/moment/use_cases_test.go`
- Modify: `server/internal/interfaces/http/moment/handler_test.go`
- Create: `miniprogram/components/pinned-anniversary-card/index.{js,json,wxml,wxss}`
- Modify: `miniprogram/pages/home/index.{js,json,wxml,wxss}`
- Modify: `miniprogram/tests/home-view-model.test.js`

**Interfaces:**
- Consumes: Task 2 persisted anniversaries and Task 4 view mapping.
- Produces: `AnniversaryHighlightsReader.ListPinned(ctx, coupleID, limit)`, optional `HomeResult.PinnedAnniversaries`, and home JSON `pinnedAnniversaries`.

- [ ] **Step 1: Write failing Go tests** for maximum three highlights, stable order, empty list, and reader failure degrading to an empty highlight list while recent journals still succeed.
- [ ] **Step 2: Write failing client tests** proving an absent/malformed highlight array becomes `[]` and does not change home journal state.
- [ ] **Step 3: Run focused Go and Node tests** and verify RED.
- [ ] **Step 4: Implement the reader port, adapter, optional aggregation, DTO, and bootstrap wiring** without importing `together` infrastructure into the moment domain.
- [ ] **Step 5: Implement the compact swipeable home card** below status and above journals, with no empty placeholder when highlights are unavailable.
- [ ] **Step 6: Run focused tests and `npm test`** and expect PASS.
- [ ] **Step 7: Commit** with `git commit -m "feat: show pinned anniversaries on home"`.

### Task 7: Anniversary Detail, Frame, and Poster Export

**Files:**
- Create: `miniprogram/utils/anniversary-poster.js`
- Create: `miniprogram/tests/anniversary-poster.test.js`
- Create: `miniprogram/components/anniversary-frame/index.{js,json,wxml,wxss}`
- Create: `miniprogram/pages/anniversary-detail/index.{js,json,wxml,wxss}`
- Modify: `miniprogram/app.json`

**Interfaces:**
- Consumes: Task 4 `anniversaryView` and Task 5 cover URL.
- Produces: `buildPosterModel(anniversary, template, canvasSize)`, `fitText(text, maxRunes)`, and detail actions for template selection, Canvas 2D export, preview, and `wx.saveImageToPhotosAlbum`.

- [ ] **Step 1: Write failing tests** for no-photo fallback, 60-rune titles, 120-rune messages, deterministic templates, device pixel ratio sizing, and safe truncation without splitting emoji.
- [ ] **Step 2: Run `node --test miniprogram/tests/anniversary-poster.test.js`** and verify RED.
- [ ] **Step 3: Implement pure poster model/text fitting**, then run the focused test and expect PASS.
- [ ] **Step 4: Build the detail page and frame component** with three built-in lavender-album templates, loading/error states, and native back navigation.
- [ ] **Step 5: Add Canvas drawing/export and album permission handling**, including retry and `wx.openSetting` after denial.
- [ ] **Step 6: Run `npm test`** and expect PASS.
- [ ] **Step 7: Commit** with `git commit -m "feat: add anniversary keepsake frames"`.

### Task 8: Avatar Picker and Global Avatar Propagation

**Files:**
- Create: `miniprogram/utils/avatar-upload.js`
- Create: `miniprogram/tests/avatar-upload.test.js`
- Create: `miniprogram/components/avatar-picker/index.{js,json,wxml,wxss}`
- Modify: `miniprogram/services/users.js`
- Modify: `miniprogram/pages/profile-setup/index.{js,json,wxml,wxss}`
- Modify: `miniprogram/pages/me/index.{js,wxml,wxss}`
- Modify: `miniprogram/pages/home/index.{js,wxml,wxss}`
- Modify: `miniprogram/components/moment-card/index.{js,wxml,wxss}`

**Interfaces:**
- Consumes: Task 3 grant/profile endpoints and existing compression/PUT helpers.
- Produces: `replaceAvatar({filePath, currentAvatarRef, userID, onProgress})` that resolves only after profile confirmation; component event `change` with the confirmed user DTO.

- [ ] **Step 1: Write failing tests** for chooseAvatar/local-file input, compression, progress, grant failure, PUT failure, profile failure, and preservation of the prior avatar in every failure path.
- [ ] **Step 2: Run `node --test miniprogram/tests/avatar-upload.test.js`** and verify RED.
- [ ] **Step 3: Implement the transactional utility, user service grant method, and picker component**.
- [ ] **Step 4: Integrate the picker into profile setup and “我的”**, preserving the existing avatar when only the nickname changes and updating `sessionStore` only after success.
- [ ] **Step 5: Pass avatar URLs through the home and moment read models/cards**, retaining initials as accessible fallbacks.
- [ ] **Step 6: Run the focused test and `npm test`** and expect PASS.
- [ ] **Step 7: Commit** with `git commit -m "feat: add user-managed avatars"`.

### Task 9: Lavender-Album System, Accessibility, and Release Verification

**Files:**
- Modify: `miniprogram/styles/tokens.wxss`
- Modify: `miniprogram/styles/animations.wxss`
- Modify: `miniprogram/app.wxss`
- Create: `miniprogram/components/couple-header/index.{js,json,wxml,wxss}`
- Modify: `miniprogram/pages/home/index.{json,wxml,wxss}`
- Modify: `miniprogram/pages/me/index.{json,wxml,wxss}`
- Modify: `miniprogram/components/moment-card/index.{json,wxml,wxss}`
- Modify: `docs/miniprogram-manual-test.md`
- Modify: `docs/backend-api.md`
- Modify: `docs/testing/phase-one-two-manual-checklist.md`
- Test: `miniprogram/tests/navigation-model.test.js`

**Interfaces:**
- Consumes: all prior tasks.
- Produces: shared surface/motion tokens, `couple-header`, updated API/manual documentation, and a release-ready branch.

- [ ] **Step 1: Extend structural tests** to require all new pages, native back navigation, registered components, and a shared reduced-motion state rather than page-local animation flags.
- [ ] **Step 2: Run `npm test`** and verify the structural assertions fail.
- [ ] **Step 3: Implement the lavender-album tokens and couple header**, then apply them to the four agreed key surfaces without redesigning unrelated flows.
- [ ] **Step 4: Verify long Chinese text, initials fallback, no-photo cards, safe-area spacing, 140–240ms motion, and reduced-motion styles** in Developer Tools responsive views.
- [ ] **Step 5: Update API and manual test documentation** with two-account and solo-test scenarios, COS failure recovery, poster permissions, leap day, nine-image journals, and LAN true-device steps.
- [ ] **Step 6: Run full verification:** `npm test`; `cd server && go test ./...`; `cd server && go test -race ./...`; `git diff --check`; verify `git status --short` contains only the intentional uncommitted `miniprogram/config.js`.
- [ ] **Step 7: Perform manual Developer Tools compile and true-device smoke test**, recording any environment-only limitation rather than weakening tests.
- [ ] **Step 8: Commit** with `git commit -m "feat: finish lavender album experience"`.
