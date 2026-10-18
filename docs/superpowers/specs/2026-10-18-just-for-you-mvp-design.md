# Just for You MVP Design

## 1. Product intent

Just for You is a private WeChat Mini Program for couples. Its first users are the developer and their partner. It should make small daily moments easy to share, make a long-distance partner feel present, and gradually turn those moments into a shared timeline.

The project also serves as an end-to-end development exercise covering product definition, Mini Program UI, Go API design, authentication, database design, object storage, deployment, testing, and release.

## 2. First milestone

Within seven days, deliver an experience build that works on two real phones for the developer and their partner. Passing WeChat's public release review is not required for this milestone because domain registration, infrastructure setup, and platform review may take longer.

The milestone succeeds when two users can:

1. Log in with WeChat and create profiles.
2. Bind into one private couple using a one-time six-character invitation code.
3. Publish a small moment with text, a quick status, an optional photo, and an optional place name.
4. See the partner's latest status on the home page.
5. Browse the same private shared timeline.
6. Delete only their own posts.
7. Unbind and immediately lose access to the former shared timeline.

## 3. Scope

### Included

- WeChat login and local user profile
- One active couple relationship per user
- Short-lived, single-use six-character invitation code
- Quick statuses such as `In class`, `Miss you`, and `A little tired`
- Short text posts
- Up to one photo per post
- Optional human-readable place name
- Partner's latest status and recent moments on the home page
- Reverse-chronological shared timeline
- Deletion of the current user's own posts
- Explicit confirmation before unbinding

### Excluded

- Chat and direct messaging
- Likes and comments
- Continuous or precise location tracking
- Push notifications
- Editing published moments
- Multiple partners or group relationships
- Public content and discovery
- Payments, subscriptions, and other monetization
- Historical-data export or permanent deletion workflow

These exclusions keep the first build focused on validating whether couples repeatedly use lightweight status sharing and shared memory capture.

## 4. User experience

### Login and profile

The user signs in through WeChat, then chooses a display name and avatar. The Mini Program exchanges the temporary WeChat login code with the Go API and stores the API's application token.

### Couple binding

An unbound user creates a short-lived six-character invitation code. A second unbound user enters it, reviews the inviter's identity, and confirms. The server performs the binding atomically. Each invitation is single-use, and each user can belong to only one active relationship. QR-code invitations are deferred until after the MVP.

### Home page

The home page shows the partner's latest quick status and its update time. A prominent `Record this moment` action opens the composer. A short list of recent moments provides immediate context.

### Moment composer

The user may select a quick status, write a short message, attach one photo, and add a place name. All fields other than the post itself are optional, but the server requires at least one meaningful content field. The UI prevents accidental duplicate submission.

### Shared timeline

The timeline displays both partners' moments in reverse chronological order. Each card identifies the author and may show text, status, photo, place name, and timestamp.

### Unbinding

Unbinding requires explicit confirmation. Once complete, both users immediately lose access to the former relationship and its moments. Historical data remains frozen in storage for the MVP but is not visible to either party. Export and deletion policies are deferred to a later design.

## 5. Architecture

The system uses a simple monolithic architecture:

```text
WeChat Mini Program
        |
        | HTTPS / JSON
        v
Go API (Gin)
  |-- authentication
  |-- users
  |-- couples and invitations
  |-- moments
  `-- object-storage integration
        |
        |-- MySQL: relational application data
        `-- Object storage: private photo objects
```

### Technology choices

- Client: native WeChat Mini Program using WXML, WXSS, and JavaScript
- Server: Go with Gin in a single deployable service
- Database: MySQL
- Data access: Go's `database/sql` with focused handwritten queries; no ORM in the MVP
- Photos: private object storage such as Tencent Cloud COS
- Edge: Nginx for TLS termination and reverse proxying
- Deployment: one cloud server for the Go service and MySQL during the private MVP
- Repository: one Git repository containing `miniprogram/`, `server/`, deployment configuration, and documentation

The design intentionally avoids microservices, event buses, and a separate admin application.

## 6. Server modules

- `auth`: WeChat code exchange, application access tokens, refresh tokens, and authentication middleware
- `user`: profile creation and updates
- `couple`: invitation creation, invitation acceptance, relationship lookup, and unbinding
- `moment`: moment creation, timeline queries, latest status, and deletion
- `storage`: photo upload authorization and temporary read URLs

Each module exposes a narrow service interface and owns its validation rules. HTTP handlers translate requests and responses but do not implement relationship authorization themselves.

## 7. Data model

### `users`

- Internal user ID
- Stable WeChat identifier
- Display name
- Avatar URL or object key
- Created and updated timestamps

### `couples`

- Relationship ID
- Both user IDs
- Status: active or ended
- Bound timestamp
- Ended timestamp

Database constraints and a transaction prevent either participant from entering two active relationships.

### `invite_codes`

- Invitation ID and unguessable code
- Creator user ID
- Expiration timestamp
- Consumed timestamp and consuming user ID

### `moments`

- Moment ID
- Relationship ID
- Author user ID
- Optional text
- Optional quick-status value
- Optional place name
- Optional photo object key
- Created timestamp
- Deleted timestamp for soft deletion if used

### `refresh_tokens`

- Token identifier or hash
- User ID
- Expiration and revocation timestamps

Raw refresh tokens and WeChat session credentials are never stored in logs.

## 8. Authorization and privacy

- Every protected endpoint derives the current user from the validated application token.
- Timeline and moment lookups derive the active relationship on the server. A relationship ID supplied by the client is never sufficient authorization.
- Only an active relationship participant may view its moments.
- Only the author may delete a moment.
- Invitation codes are unguessable, expire quickly, and can be consumed once.
- The invitation acceptance operation locks and validates both users and the invitation in one database transaction.
- Photos use random object keys and private buckets. The server issues short-lived upload or read authorization.
- The MVP stores a user-entered place name, not a continuous location history or precise movement trail.
- Unbinding revokes access immediately. Frozen historical data cannot be queried by either former participant.

## 9. API behavior and errors

The API returns a consistent JSON envelope containing either data or a stable application error code with a user-safe message. Expected failures include expired login, invalid or consumed invitations, existing relationships, forbidden moment access, invalid content, failed photo authorization, and transient server errors.

The Mini Program:

- prompts for a new login when authentication expires;
- shows actionable messages for invitation and validation failures;
- distinguishes photo-upload failure from post-creation failure;
- disables the submit action while a request is in flight; and
- allows safe retry after transient network errors.

The server applies field-length and file-type limits, request-size limits, authentication checks, and structured logging. Logs may contain request IDs and internal user IDs but not login codes, tokens, or WeChat session secrets.

## 10. Testing and acceptance

### Automated tests

- Unit tests for invitation expiration, single-use behavior, active-relationship rules, unbinding, and moment authorization
- Handler tests for validation and stable error responses
- Integration tests against MySQL for concurrent invitation acceptance and relationship-filtered timeline queries

### Manual acceptance

Use two distinct WeChat accounts on two real phones to verify:

1. Login and profile creation
2. Invitation creation and successful binding
3. Rejection of expired, reused, and invalid invitations
4. Text-only, status-only, and photo moments
5. Identical and correctly ordered timelines on both phones
6. Visibility of the partner's latest status
7. Deletion restricted to the author
8. No access from an unbound third account
9. Immediate access loss after unbinding
10. Clear behavior during network and photo-upload failures

## 11. Deferred product questions

Later iterations must explicitly design:

- permanent deletion, retention, and export of memories after unbinding;
- notification consent and frequency;
- richer memory formats and anniversary review;
- onboarding for public users;
- abuse reporting and account recovery; and
- monetization based on validated user behavior rather than speculative features.
