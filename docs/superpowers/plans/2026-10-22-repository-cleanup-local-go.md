# Repository Cleanup and Local Go Environment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Preserve the WorkBuddy changes, split the repository into independent `frontend/` and `backend/` directories, run the Go service against local MySQL, remove regenerable clutter, and leave a concise Chinese code guide.

**Architecture:** Keep one Git repository with two self-contained application roots. Preserve the Go module path and DDD package boundaries while moving `server/` to `backend/`; move all WeChat and Node assets into `frontend/`. Root contains only repository metadata, `.gitignore`, and the handoff README.

**Tech Stack:** Go 1.27, Gin, MySQL 8.x, native WeChat Mini Program, Node.js built-in test runner, Homebrew services.

**Spec:** `docs/superpowers/specs/2026-10-22-repository-cleanup-local-go-design.md`

## Global Constraints

- Preserve every current WorkBuddy source and test change before restructuring; do not commit `miniprogram/config.js` local LAN values or secrets.
- Keep tests, SQL migrations, dependency manifests/locks, `.env.example`, and WeChat project configuration.
- Remove generated dependencies, IDE state, execution ledgers, obsolete application directories, and design-document clutter from the final working tree.
- Keep the Go module path unchanged to avoid a non-functional import rewrite.
- Do not claim WeChat login, COS upload, or true-device validation without the required external environment.

## Review Focus

- WorkBuddy untracked tests/utilities and modified Go files must survive the snapshot and directory move byte-for-byte.
- Moving the mini program must not break test `require()` paths or `miniprogramRoot` resolution.
- Local `.env` and COS/WeChat credentials must never be staged.
- Migration setup must apply versions `000001`, `000002`, and `000003` in order to a fresh local database.
- API startup must distinguish a missing MySQL service from invalid external WeChat/COS configuration.

---

### Task 1: Preserve and Validate WorkBuddy Changes

**Files:**
- Modify: current WorkBuddy-modified files under `miniprogram/` and `server/` only if a failing test proves a defect
- Preserve untracked: `miniprogram/tests/compose-guard.test.js`, `miniprogram/tests/memory-map.test.js`, `miniprogram/utils/compose-guard.js`, `server/internal/infrastructure/persistence/mysql/together/date.go`

**Interfaces:**
- Consumes: current dirty worktree at commit `d8619a0`
- Produces: one recoverable Git commit containing all WorkBuddy functional changes, excluding local `miniprogram/config.js`

- [ ] **Step 1: Record exact dirty state and inspect secrets**

Run `git status --short`, `git diff --check`, and inspect staged candidates. Confirm no `.env`, access keys, signed URLs, or LAN-only `miniprogram/config.js` are included.

- [ ] **Step 2: Run the current frontend and backend baselines**

Run: `npm test`

Run: `cd server && go test ./...`

Expected: both suites pass. If either fails, diagnose and fix only the proven WorkBuddy defect before continuing.

- [ ] **Step 3: Commit the WorkBuddy snapshot**

Stage the modified/untracked application files explicitly, excluding `miniprogram/config.js`, then commit with `feat: preserve latest workbuddy improvements`.

- [ ] **Step 4: Confirm recoverability**

Run: `git status --short`

Expected: only the intentional local `miniprogram/config.js` remains modified.

### Task 2: Split the Frontend and Backend Roots

**Files:**
- Move: `miniprogram/` → `frontend/miniprogram/`
- Move: `package.json`, `package-lock.json`, `project.config.json`, `project.private.config.json.example` → `frontend/`
- Move: `server/` → `backend/`
- Move: `.env.example` → `backend/.env.example`
- Modify: `frontend/package.json`
- Modify: `frontend/project.config.json`
- Modify: `.gitignore`

**Interfaces:**
- Consumes: clean functional snapshot from Task 1
- Produces: standalone `frontend/` Node/WeChat root and standalone `backend/` Go root

- [ ] **Step 1: Add a structural test for the target layout**

Create a Node test under `frontend/miniprogram/tests/` after the move that asserts the frontend manifests and mini-program root exist, backend contains `go.mod`, `cmd`, `internal`, and `migrations`, and obsolete root application paths do not exist.

- [ ] **Step 2: Move tracked files with Git-aware moves**

Use explicit `git mv` operations. Preserve the local frontend config as `frontend/miniprogram/config.js` without staging its LAN value.

- [ ] **Step 3: Repair frontend paths**

Set the frontend test script to `node --test miniprogram/tests/*.test.js`; keep `project.config.json` inside `frontend/` with `miniprogramRoot: "miniprogram/"` so WeChat Developer Tools imports `frontend/` directly.

- [ ] **Step 4: Repair repository ignore rules**

Ignore `node_modules/`, `.env`, `project.private.config.json`, `.DS_Store`, IDE state, and execution-workspace directories at any depth while continuing to track examples.

- [ ] **Step 5: Verify both roots independently**

Run: `cd frontend && npm install && npm test`

Run: `cd backend && go mod download && go test ./...`

Expected: both pass with no imports rewritten and no missing path errors.

- [ ] **Step 6: Commit**

Commit with `refactor: separate frontend and backend roots`.

### Task 3: Make Local MySQL and Go Startup Reproducible

**Files:**
- Create: `backend/scripts/setup-local-db.sh`
- Create: `backend/scripts/dev.sh`
- Modify: `backend/.env.example`
- Test: `backend/scripts/scripts_test.sh`

**Interfaces:**
- Consumes: Homebrew Go and a local MySQL server/client
- Produces: `setup-local-db.sh` for database/user/migrations and `dev.sh` for loading `backend/.env` and running `cmd/api`

- [ ] **Step 1: Write shell contract tests**

The test must assert both scripts use their own directory as the working root, do not print secret values, fail with actionable Chinese messages when commands/config are missing, and enumerate migrations `1 2 3` in order.

- [ ] **Step 2: Implement local database setup**

`setup-local-db.sh` checks for a reachable MySQL server, creates the `just_for_you` database and least-privilege `just_for_you` local account, exports the development DSN, and invokes `go run ./cmd/migrate -direction up -version 1`, then `2`, then `3`. It must stop on the first failure and be intended for a fresh local database.

- [ ] **Step 3: Implement local API startup**

`dev.sh` loads `backend/.env` without echoing values, validates it exists, then executes `go run ./cmd/api`. It must work regardless of the caller's current directory.

- [ ] **Step 4: Prepare local configuration safely**

Create or move the ignored real environment to `backend/.env`; keep placeholders and clear comments in `backend/.env.example`. Never stage the real file.

- [ ] **Step 5: Install/start MySQL only if required**

Inspect Homebrew packages and services first. If no server exists, install an available MySQL 8.x formula, start its service, and verify `mysqladmin ping`. Do not replace a healthy existing server.

- [ ] **Step 6: Initialize and start the service**

Run the database setup script, start the API with `dev.sh`, and verify `GET http://127.0.0.1:8080/healthz` returns success. Stop only the foreground verification process; leave the user-owned MySQL service running.

- [ ] **Step 7: Run backend verification and commit**

Run: `cd backend && go test ./... && go test -race ./... && go vet ./...`

Commit with `chore: add reproducible local backend startup`.

### Task 4: Remove Non-Deliverable Clutter

**Files:**
- Delete: `.idea/`, `.superpowers/`, `node_modules/`, `.DS_Store` files, `i18n/`, `miniapp/`, root `Makefile`, root `docker-compose.yml`, and `docs/`
- Modify: `.gitignore`

**Interfaces:**
- Consumes: independently verified frontend/backend roots
- Produces: final repository containing only `.git`, `.gitignore`, `README.md`, `frontend/`, and `backend/` at the top level

- [ ] **Step 1: Print and review the exact deletion list**

Resolve every path explicitly. Do not use a broad recursive target, home-directory variable, unresolved glob, or repository-root deletion.

- [ ] **Step 2: Delete tracked obsolete directories with `apply_patch`/Git-aware removal and generated directories with explicit recoverable cleanup where practical**

The functional source, tests, migrations, manifests, examples, and private ignored configuration must remain.

- [ ] **Step 3: Run structural and repository checks**

Run the layout test and list root entries. Expected visible deliverables: `.gitignore`, `README.md`, `frontend/`, `backend/`; `.git` remains as repository metadata.

- [ ] **Step 4: Commit**

Commit with `chore: remove obsolete workspace files`.

### Task 5: Write the Chinese Code Guide

**Files:**
- Replace: `README.md`

**Interfaces:**
- Consumes: final paths and verified startup commands
- Produces: concise project map, local setup guide, DDD request walkthrough, frontend data-flow walkthrough, and maintenance recipes

- [ ] **Step 1: Document exact setup and commands**

Include prerequisites, frontend install/test/import steps, backend DB setup/start/test steps, environment file handling, and the verified health URL.

- [ ] **Step 2: Explain the backend through one real request**

Trace a journal publish or home request through `backend/internal/interfaces/http`, `application`, `domain`, and `infrastructure`; explain allowed dependency direction and bootstrap wiring.

- [ ] **Step 3: Explain the mini-program flow**

Trace page event → service → HTTP client/session refresh → store/view update → component render, and identify where pages, components, services, stores, and utilities belong.

- [ ] **Step 4: Add maintenance recipes**

Give short checklists for adding an API, adding a page, running tests, diagnosing “服务暂不可用”, and protecting secrets.

- [ ] **Step 5: Validate every documented command and commit**

Commit with `docs: add local setup and code walkthrough`.

### Task 6: Final Release Verification

**Files:**
- Modify only if a verification failure proves a defect

**Interfaces:**
- Consumes: Tasks 1–5
- Produces: clean, runnable, documented repository

- [ ] **Step 1: Run all automated checks**

Run: `cd frontend && npm test`

Run: `cd backend && go test ./... && go test -race ./... && go vet ./...`

Run: `git diff --check`

- [ ] **Step 2: Compile in WeChat Developer Tools**

Import/open `frontend/`, perform a normal compile, and confirm there are no WXML/WXSS/JavaScript compile errors. Network failures caused solely by an intentionally stopped backend are not compile failures.

- [ ] **Step 3: Verify local API one final time**

Start via `backend/scripts/dev.sh`, request `/healthz`, record the status, and terminate the verification API process cleanly.

- [ ] **Step 4: Audit final repository state**

Confirm secrets are ignored, root layout matches the spec, and `git status --short` has no uncommitted business code. The ignored `frontend/miniprogram/config.js`, `frontend/project.private.config.json`, and `backend/.env` may remain local.

- [ ] **Step 5: Final commit if verification required fixes**

Use `fix: complete repository verification`; otherwise do not create an empty commit.
