# GK Circle Live Exam Rules

Version: 1.0

Status: Mandatory

---

# Purpose

Govern Live Tests and Live Exam Sessions.

---

# Single Engine Rule

Use ONE test engine.

Do not create separate engines for:

Practice

Mock

Live

Quiz

Poll

---

# Live Session Architecture

Host

↓

Session

↓

Questions

↓

Participants

↓

Answers

↓

Results

---

# Real Time Rule

Use:

Socket.IO

or equivalent realtime architecture.

---

# Live Analytics Rule

Show:

Participants Count

Answer Distribution

Correct %

Response Time

Leaderboard

---

# Anti Cheating Rule

Support:

Tab Switching Detection

Device Tracking

Session Validation

Activity Logs

---

# AI Rule

AI should generate:

Live Explanations

Answer Analysis

Post-Test Insights

Performance Reports

---

# Completion Rule

A Live Exam is complete only when:

Questions

↓

Realtime Participation

↓

Results

↓

Analytics

↓

AI Review

all function correctly.

---

# v0.5 Addendum — Architecture-Intent Standard

The v1.0 section above defines the principles (Single Engine, Real Time via Socket.IO, Live Analytics, Anti-Cheating, AI). This addendum makes those principles enforceable as code, configuration, and process. The structure mirrors the other v0.5 standards.

This is **architecture-intent**: the live exam engine is not yet implemented. The rules below are the contract it MUST satisfy when built. The next revision (v1.0 of this file) will merge these rules into the body.

---

# Dependency Rule (Mandatory)

This document does not replace AGENTS.md.

It supplements AGENTS.md.

If this document conflicts with AGENTS.md, AGENTS.md wins.

If this document conflicts with another standards file, the more specific standard wins.

If ambiguity exists, document the ambiguity in an ADR before implementation.

# Verification Requirement (Mandatory)

A rule is not considered satisfied because code, configuration, tests, or documentation exist. A rule is satisfied only when evidence exists. For live exams, evidence must include:

- A load test demonstrating 10,000 concurrent users in a single test with < 200ms p95 message latency
- A chaos test demonstrating that timer accuracy is preserved when 30% of clients disconnect randomly
- A replay test demonstrating that a disconnected user can rejoin and continue from the correct question with the correct remaining time
- A penetration test confirming that no client can manipulate the timer, the score, or the leaderboard

---

# Real-Time Synchronization (Mandatory)

Live exam interactions MUST propagate to all participants within 200ms p95 (target: 100ms p95) of the originating event.

## Transport (Mandatory)

The primary transport is **WebSocket** over TLS (Socket.IO, per the v1.0 Real Time Rule). Long-polling and SSE are NOT acceptable for live exams.

## Message Types

WebSocket messages MUST be versioned. Every message has:

```typescript
interface WSMessage<T = unknown> {
  v: number;             // Protocol version (currently 1)
  type: string;          // See catalog below
  sessionId: string;
  userId: string;        // Omitted in server-to-client
  seq: number;           // Monotonic per-session sequence
  ts: number;            // Unix ms when generated
  payload: T;
  signature?: string;    // HMAC of payload, used for sensitive messages
}
```

## Message Type Catalog (v1)

| `type` | Direction | Purpose |
|---|---|---|
| `session.joined` | server→client | Confirmed join; sends full session state |
| `session.state` | server→client | Session state update (e.g., `STARTING` → `LIVE` → `ENDED`) |
| `session.ended` | server→client | Final state; no further messages accepted |
| `question.next` | server→client | Show next question |
| `question.show` | server→client | Display a question |
| `question.hide` | server→client | Hide question (during question transition) |
| `answer.submit` | client→server | User submits an answer |
| `answer.accepted` | server→client | Server confirms receipt; does NOT confirm correctness |
| `answer.rejected` | server→client | Submission rejected (e.g., already submitted) |
| `timer.tick` | server→client | 1Hz heartbeat with remaining time |
| `timer.snapshot` | server→client | Authoritative remaining time on (re)connect |
| `leaderboard.update` | server→client | Leaderboard delta (top-20 + the user's rank) |
| `participant.joined` | server→client | Another user joined (visible rosters) |
| `participant.left` | server→client | A user left or disconnected |
| `moderation.flag` | server→client | Moderation action broadcast to the user |
| `error.protocol` | server→client | Protocol-level error (client should reconnect) |
| `error.session` | server→client | Session-level error (e.g., session already ended) |

## Sequencing (Mandatory)

Every message carries a monotonic `seq` per session. Clients MUST reject out-of-order messages. Clients MUST drop messages whose `seq` is older than the last seen. The `seq` is also used for resumability after reconnect.

## Server-to-Client Batching

For high-frequency updates (e.g., leaderboard), the server MAY batch messages with the same `ts` bucket. Clients MUST apply batched updates atomically.

## Latency Budget (Mandatory)

| Operation | p50 | p95 | p99 |
|---|---|---|---|
| `answer.accepted` (from `answer.submit`) | 50ms | 150ms | 300ms |
| `leaderboard.update` (from any score-changing event) | 100ms | 300ms | 600ms |
| `question.show` (from server's planned time) | 50ms | 200ms | 500ms |
| `timer.snapshot` (from `join` request) | 100ms | 300ms | 600ms |

A regression of ≥ 30% on any of these for ≥ 5 minutes pages the on-call.

## Clock Skew Rule (Mandatory)

The server's clock is the source of truth. Client clocks MUST NOT be trusted for any event timing. Clients display a server-relative timer computed from `serverNow` deltas (see Timer Authority on Server below).

---

# Reconnect Handling (Mandatory)

Users WILL disconnect. A user who disconnects MUST be able to reconnect and continue the test without losing progress.

## Reconnect State Restoration

On reconnect, the server sends a `session.state` and `timer.snapshot`. The client uses these to:

1. Re-render the current question
2. Reset the timer to the server's authoritative remaining time
3. Re-sync the leaderboard (full snapshot, not delta)
4. Display a "Reconnected" indicator to the user

## Idempotent Joins (Mandatory)

A reconnecting user joins the same `LiveExamParticipation` record. The server MUST:

- Return the user's existing `currentQuestionIndex`
- Return the user's existing `answers` (so they are not lost)
- Return the user's existing `score` (so it is not double-counted)
- NOT reset the timer; the timer is per-session, not per-user

## Sequence Resume (Mandatory)

After reconnect, the client sends the last `seq` it received. The server replays all messages with `seq > lastSeenSeq` from a server-side buffer. The buffer holds the last 5 minutes of messages per session. After 5 minutes, the client receives a full `session.state` and `timer.snapshot` instead.

## Maximum Reconnect Window (Mandatory)

A user can disconnect for a maximum of `MAX_DISCONNECT_DURATION` (default: 2 minutes). After this, the user is marked `ABANDONED` and cannot reconnect. The session continues for other users. Abandoned submissions are scored if they were submitted before the disconnect; abandoned-but-not-submitted questions count as unanswered.

If the session is in `RESULT_FINALIZATION` or `ENDED` state, reconnect is not allowed.

## WebSocket Reconnection (Mandatory)

The client uses exponential backoff with jitter:

- Attempt 1: 250ms + jitter
- Attempt 2: 500ms + jitter
- Attempt 3: 1s + jitter
- Attempt 4: 2s + jitter
- Attempt 5: 4s + jitter
- Attempt 6+: 8s + jitter (cap at 8s)

The cap on total reconnect attempts is `MAX_DISCONNECT_DURATION`. After that, the client shows a "You have been disconnected" screen with a "Try to rejoin" button (which may be denied if the user is past `MAX_DISCONNECT_DURATION`).

## Session Resume Token (Mandatory)

The `LiveExamParticipation.sessionResumeToken` is a signed JWT issued on `session.joined`. The client stores it in `localStorage` and presents it on reconnect. The token is bound to `userId`, `sessionId`, optional `deviceId`, `issuedAt`, and `expiresAt = min(now + MAX_DISCONNECT_DURATION, sessionEndTime)`. Signed with the session's signing key and verified on reconnect.

## Suspicious Reconnect Detection

If a user reconnects from a different IP, browser fingerprint, or device in a way that exceeds heuristics, the session MAY require admin approval to resume. Heuristics:

- ≥ 3 distinct IPs within a session (excluding expected carrier-grade NAT)
- ≥ 2 distinct device fingerprints
- Reconnect from a country not matching the user's profile (within 1 hour of last activity)

The user is shown a "Please contact support to resume this session" message and the participation is flagged for review.

---

# Timer Authority on Server (Mandatory)

The timer is the most important piece of state in a live exam. The client MUST NEVER be the source of truth for timing.

## Server-Authoritative Timer (Mandatory)

The `LiveExamSession.serverEndsAt` (UTC `DateTime`) is the single source of truth for remaining time. The client displays a remaining time computed from `serverEndsAt - serverNow`. The client clock is never trusted for any event timing. The server emits a `timer.tick` at 1Hz; the client uses the most recent tick to estimate server clock skew and re-renders the displayed time from that estimate. If the client misses ≥ 3 ticks, it requests a `timer.snapshot` on resume. The `serverEndsAt` is immutable after session start; pause/resume recomputes it from the stored `pausedRemainingMs`. Anti-tampering: any client-side attempt to modify, skip, or locally pause the timer is detected by integrity checks and may result in disqualification.

## Server-Authoritative Scoring (Mandatory)

The score for a `LiveExamAnswer` is NEVER computed on the client. The server holds the correct answer in the database (encrypted at rest; not exposed to clients). The `LiveExamParticipation.score` is updated server-side in two paths:

1. **Live correctness** (if `liveCorrectness: true` on the test): the server computes and updates the score on every `answer.submit`. The user sees the score in `answer.accepted`.
2. **Deferred correctness** (default for Live Tests): the server does NOT update the score during the test. The user sees only `answer.accepted` (confirmation of receipt, not correctness). The score is computed in the finalization transaction at session end (see Result Finalization below).

The client NEVER decides "this was correct" or "you got N points." The user cannot influence the scoring formula. The scoring formula (correct = +P, incorrect = -N, unanswered = 0, with per-question overrides) is applied identically to every user and stored in the `ScoreComputationLog`.

## Server-Authoritative Submission Status (Mandatory)

The submission status (`PENDING`, `ACCEPTED`, `REJECTED`, `SCORED`, `FINAL`) is determined by the server and only the server. The client MUST NOT assume a submission is `ACCEPTED` based on local state; the only signal that matters is the `answer.accepted` WebSocket message from the server. If the client sends an `answer.submit` and does not receive `answer.accepted` or `answer.rejected` within 5 seconds, the client retries with the same idempotency key; the server detects the duplicate and returns the original response. A submission is `FINAL` only after result finalization completes. The client UI does not show "submitted" as a final state — it shows "submitted (pending review)" until the server confirms.

## Timer State Lives on the Server

The `LiveExamSession` has a `serverEndsAt` field (UTC `DateTime`) computed at session start and immutable thereafter. The client displays a remaining time computed from `serverEndsAt - serverNow` (the server's current time).

## Timer Tick (Mandatory)

The server emits a `timer.tick` message at 1Hz:

```typescript
interface TimerTick {
  serverNow: number;
  remainingMs: number;
  pausedRemainingMs?: number;
}
```

## Client Timer Display

The client computes the display time as:

```
displayed_remaining_ms = max(0, serverNow + (client_estimate_of_server_clock_skew) - serverEndsAt)
```

Where `client_estimate_of_server_clock_skew` is the difference between the client's local clock and `serverNow` measured on the last `timer.tick`. This estimate is updated on every tick.

If the client misses ≥ 3 consecutive ticks (e.g., backgrounded), the client MUST request a `timer.snapshot` immediately on resume.

## Pause and Resume (Mandatory)

A session can be paused by an admin (e.g., network incident at the venue) or the system (e.g., detected server overload). When paused:

- The server sets `pausedAt` and `pausedRemainingMs = serverEndsAt - pausedAt`
- A `session.state` with `state: "PAUSED"` is broadcast
- The timer stops decrementing on clients
- A "Session paused" UI is shown to all users
- The session can be resumed by an admin or automatically after a system pause resolves

When resumed:

- The server sets `serverEndsAt = now + pausedRemainingMs`
- A `session.state` with `state: "LIVE"` is broadcast
- The timer resumes from `pausedRemainingMs`

## Server-Clock Drift

The server uses NTP-synchronized time. If the server's clock drifts by > 500ms from the NTP source, a Sentry alert is fired. Sessions in flight at the time of drift are NOT retroactively adjusted.

## Per-Question Timer (Optional)

A live exam MAY have a per-question timer. The same rules apply: server-authoritative. The server sends `question.show` with `questionEndsAt`. The client displays the per-question timer computed from this. If the user is on question 5 of 50 and runs out of time, the server auto-advances.

## Anti-Tampering Rule (Mandatory)

The client MUST NOT:

- Attempt to modify the timer
- Attempt to skip the timer
- Attempt to pause the timer locally

Any client-side attempts are detected by integrity checks (see Anti-Cheating Rules below) and may result in disqualification.

---

# Anti-Cheating Rules (Mandatory)

Live exams are the highest-value target for cheating. The system MUST defend against:

- Tab/window switching
- Multi-accounting (a single user on multiple devices)
- Collusion (users sharing answers in real time)
- Network manipulation (e.g., a user joining twice to see two perspectives)
- Bot/automated answer submission
- Question leakage (users capturing and sharing questions)
- Time manipulation (users with manipulated clocks)

## Tab/Window Switching Detection (Mandatory)

The client emits a `visibilitychange` event. The server tracks:

- `tabSwitches`: count of visibility changes
- `tabSwitchedAt`: timestamp of last switch
- `totalTabSwitchedMs`: total time the tab was hidden

Heuristics:

- ≥ 3 tab switches in 60 seconds → warn the user
- ≥ 5 tab switches total → flag the session for moderator review
- Total time hidden ≥ 30 seconds → flag for moderator review

A `moderation.flag` message is sent to the user with a warning. The session is marked `FLAGGED` and the user is shown a "Please stay on the test page" message.

## Multi-Accounting Detection (Mandatory)

A user can have at most one active `LiveExamParticipation` per session. Attempts to join from a second device result in:

- The new device receives `error.session: "ALREADY_JOINED"`
- The existing device is notified via `participant.joined: false` (the existing join remains)
- The new IP is logged in `LiveExamParticipation.attemptedJoinIPs`

If a user is detected joining from ≥ 2 distinct IPs within a 30-second window, the session is flagged for admin review.

## Collusion Detection (Mandatory)

The leaderboard and time-to-answer distributions are monitored in real time. Two users with anomalously similar response patterns (e.g., > 90% answer agreement, identical timing) are flagged for review. The flag is reviewed by a human moderator before any action is taken. This is a signal, not a verdict; false positives are common.

## Bot/Automated Submission Detection (Mandatory)

Submission timing is monitored:

- Submissions arriving in < 200ms after `question.show` for ≥ 3 consecutive questions are flagged
- Submissions with sub-50ms response times are flagged (humans cannot do this)
- Submissions from clients with no `mousemove`, `keydown`, or `touchstart` events are flagged

Flagged users are moved to a "review pool" and may be asked to complete a CAPTCHA.

## Question Leakage Detection (Mandatory)

If a question appears in a public forum, social media, or messaging app within 24 hours of the test, the question is auto-flagged. GK Circle uses:

- A hash of every question's text
- Periodic crawls of known public sources (not exhaustive)
- A user-report mechanism (e.g., "I saw this question leaked")

Leaked questions are removed from the active pool and replaced with a backup question. Users who received the leaked question are re-graded (the leaked question is removed from their score).

## Time Manipulation Detection (Mandatory)

The client clock is not trusted. The server uses `serverNow` from the heartbeat. Submission timestamp = the time the message reached the server. If a submission's `clientTs` differs from `serverTs` by > 5 seconds, the submission is logged with a `clockAnomaly` flag.

## Device Integrity Checks (Optional but Recommended)

For high-stakes exams, the client may run an integrity check at session start:

- Browser fingerprint
- OS and browser version
- Screen size
- WebGL renderer string
- Audio context fingerprint
- Camera and microphone availability (for proctored sessions)

The fingerprint is hashed and stored with the participation. Anomalies are flagged for review.

## Penalties

| Violation | First | Repeat |
|---|---|---|
| Tab switching | Warning | Session flagged for review |
| Multi-accounting | Disqualification | Account suspension |
| Collusion | Session flagged for review | Disqualification + account review |
| Bot/automated | CAPTCHA + flag | Disqualification |
| Question leakage | (Action against leaker, not test-taker) | |
| Time manipulation | Submission rejected | Session flagged |

Disqualification is final and cannot be appealed through the in-app flow. Appeals go to a human review board.

## Audit Trail (Mandatory)

Every anti-cheating detection and action is logged to `AntiCheatLog`:

```typescript
interface AntiCheatLog {
  id: string;
  sessionId: string;
  userId: string;
  detection: 'TAB_SWITCH' | 'MULTI_ACCOUNT' | 'COLLUSION' | 'BOT' | 'CLOCK_ANOMALY' | 'FINGERPRINT_MISMATCH';
  severity: 'INFO' | 'WARNING' | 'CRITICAL';
  metadata: Record<string, any>;
  action: 'WARN' | 'FLAG' | 'CAPTCHA' | 'DISQUALIFY' | 'REVIEW';
  createdAt: string;
}
```

The log is immutable; entries are never deleted or modified.

---

# WebSocket Security (Mandatory)

The WebSocket layer is the primary attack surface for live exams. The standards below are non-negotiable.

## Authentication (Mandatory)

Every WebSocket connection MUST be authenticated before the `session.joined` message is sent. The authentication is via:

1. A short-lived JWT (≤ 60 seconds) obtained from a REST endpoint (`POST /live-exams/:id/join`)
2. The JWT contains `userId`, `sessionId`, and `expiresAt`
3. The client passes the JWT as a query parameter to the WebSocket upgrade request
4. The server verifies the JWT signature, expiration, and that the user is eligible to join (enrolled, paid, age-appropriate, etc.)
5. After the upgrade, the server generates a `sessionResumeToken` for reconnects

Anonymous WebSocket connections are forbidden.

## Authorization (Mandatory)

After authentication, the server checks:

- The user is enrolled in the Course containing this live exam (or the exam is free/public)
- The user has not already been disqualified
- The user has not exceeded their `MAX_CONCURRENT_SESSIONS` (default: 1)
- The user is on the allowlist for invite-only exams
- The session is in a state that allows joining (`SCHEDULED` is allowed within a join window; `LIVE` is allowed; `ENDED` is not)

A failed authorization returns a `403` close code with a JSON payload explaining the reason.

## Origin and CSRF Protection (Mandatory)

The server checks the `Origin` header on the WebSocket upgrade request. Only allowlisted origins are accepted (the same CORS allowlist as REST). A missing or disallowed origin returns a `403` close code.

The sessionResumeToken is bound to the user, so CSRF is mitigated. Defense in depth: the WebSocket library is configured to reject requests without a valid JWT regardless of origin.

## Rate Limiting (Mandatory)

WebSocket messages are rate-limited per connection:

- ≥ 100 messages per 10 seconds → drop the connection
- ≥ 5 invalid messages (e.g., malformed JSON) → drop the connection
- ≥ 3 authentication failures → ban the IP for 1 hour

Rate limit violations are logged to Sentry and an alert is fired on a sudden spike.

## Message Validation (Mandatory)

Every incoming message is validated against a schema (Zod or class-validator). A message that fails validation is dropped and the failure is logged. The schema is versioned with the protocol version.

## Server-Initiated Disconnect (Mandatory)

The server MAY disconnect a client for: protocol violations, anti-cheating violations, admin action, session ending, inactivity (no message in 5 minutes). Disconnects use a close code from the registered IANA range. The payload is a JSON object with a `reason` field.

## TLS (Mandatory)

WebSocket connections MUST use `wss://`. Plain `ws://` is forbidden in production. Development environments may use `ws://` over localhost only.

## PII and Logging (Mandatory)

WebSocket messages are NOT logged in plaintext. The server logs `type`, `seq`, `userId`, and latency, but NOT the payload. The exception is `answer.submit`, where the server logs the answer choice (without question text). The `question.show` payload is NOT logged.

---

# Result Finalization (Mandatory)

A live exam's result is final once the session ends. There is no "unsubmit," "resubmit," or "change answer" after the deadline.

## End-of-Session

The session ends in one of three ways:

1. **Normal completion**: `serverEndsAt` is reached. The server transitions to `RESULT_FINALIZATION`.
2. **Admin early termination**: An admin ends the session manually. Reason is required and logged.
3. **System termination**: A critical error forces the session to end. The session is marked `INVALIDATED` and no scores are recorded.

## Result Finalization Process

When the session ends, the server executes the finalization in a single transaction:

```
BEGIN TRANSACTION
  1. Lock all LiveExamParticipation rows for this session
  2. For each participation:
     a. Mark unanswered questions as `UNANSWERED`
     b. Apply scoring rules (correct, incorrect, unanswered, negative marking)
     c. Compute total score, accuracy, percentile
  3. Update LiveExamSession.scoreComputed = true
  4. Generate a SessionResult record
  5. Insert AntiCheatLog entries for any flagged users
  6. Trigger leaderboard update
  7. Send result notifications (email, push)
  8. Update user stats (XP, ranks)
COMMIT
```

The transaction MUST complete in < 5 seconds for sessions with up to 10,000 participants. If it does not, an alert is fired.

## Result Idempotency (Mandatory)

The finalization process is idempotent. If it is retried (e.g., after a transient error), the result is the same. This is achieved by:

- A `sessionResultId` keyed on `sessionId`
- UPSERT semantics for the `SessionResult` row
- UPSERT semantics for the `LiveExamParticipation` row updates

## Result Notification

Once finalization is complete, users receive a `session.ended` message with their personal result (score, rank, accuracy). The full leaderboard is fetched on a separate REST call (not over WebSocket, to keep the broadcast small).

## Result Disputes

A user MAY dispute their result within 7 days by contacting support. The dispute creates a `ResultDispute` record. An admin reviews the dispute and may:

- Re-grade the session if a bug is found (the `SessionResult` is versioned)
- Refund the entry fee if the session was `INVALIDATED`
- Uphold the result

The original result is preserved; the corrected result is a new version.

## Result Audit Trail (Mandatory)

Every score computation is recorded in `ScoreComputationLog`:

```typescript
interface ScoreComputationLog {
  id: string;
  sessionId: string;
  participationId: string;
  computedAt: string;
  rulesVersion: string;
  input: { answers: AnswerSubmission[], questions: TestQuestion[] };
  output: { totalScore: number, accuracy: number, perQuestion: PerQuestionScore[] };
  flagged: boolean;
}
```

This is the legal record of the score. It is immutable.

---

# Leaderboard Consistency (Mandatory)

## Leaderboard Data Structure (Mandatory)

The leaderboard is a sorted set (ZSET in Redis) keyed on `(sessionId, userId)` with score = `totalScore`. The score is updated atomically with each scored answer.

## Update Path

When a user submits an answer:

```
1. Server receives answer.submit
2. Server validates the answer (within question window, no duplicate)
3. Server scores the answer (correct, incorrect, with negative marking)
4. Server updates Redis ZSET: ZADD sessionId:leaderboard <newScore> <userId>
5. Server emits leaderboard.update with the user's new rank and the top-20
6. Server emits answer.accepted to the user
```

## Top-K Selection (Mandatory)

The leaderboard message sent to clients contains:

- The top 20 users (rank 1–20)
- The current user's rank and score (even if outside top 20)
- The total number of participants
- The user's percentile (computed from rank / total)

If the user is in the top 20, only the top 20 is sent. If the user is outside, the top 20 is still sent (so the user can see what they are chasing), plus the user's own rank/score.

## Consistency Guarantees (Mandatory)

- **Per-session atomicity**: All updates to a session's leaderboard happen in order. No two updates for the same session can race.
- **No lost updates**: If the user submits two answers in quick succession, both are scored and the leaderboard reflects both.
- **Eventual consistency across regions**: If the system runs in multiple regions, the leaderboard is eventually consistent within 5 seconds p99. The primary region writes; secondary regions read.

## Tie-Breaking Rule (Mandatory)

Ties in score are broken by:

1. **Earlier completion time**: the user who completed the test in less time wins
2. **Faster first-correct-answer time**: the user who got their first question correct in less time wins
3. **User ID (lexicographic)**: the lower user ID wins (deterministic)

The tie-breaking is computed once at result finalization and stored in the `SessionResult`.

## Anti-Leaderboard-Manipulation (Mandatory)

- Self-submissions (a user submitting from a second account) are detected by IP and fingerprint matching
- Coordinated inflation is detected by pattern analysis
- Score changes are auditable; any retroactive change (e.g., admin override) is logged in `LeaderboardAdjustmentLog`

## Display Caching (Mandatory)

The top-20 view is cached in Redis with a 1-second TTL. The user's own rank is not cached. This balances freshness (user's rank updates within 1 second) with backend load.

---

# Negative Marking Support (Mandatory)

Many competitive exams (UPSC, SSC, Banking) use negative marking. The system MUST support configurable negative marking per question or per test.

## Configuration (Mandatory)

The `Test.negativeMarking` field is a `Decimal(4,2)` representing the negative marks per incorrect answer. Common values:

| Exam | Negative Marking |
|---|---|
| UPSC Prelims | 0.33 |
| SSC | 0.25 |
| Banking (IBPS PO Mains) | 0.25 |
| Banking (RBI Grade B) | 0.25 |
| Teaching (TET) | 0 (typically) |

A value of 0 disables negative marking.

## Per-Question Override

A `TestQuestion.negativeMarkingOverride` field allows overriding the test-level rule for a specific question. Default: inherit from test.

## Scoring Formula (Mandatory)

For a question with positive marks `P` and negative marking `N`:

| Outcome | Score |
|---|---|
| Correct | `+P` |
| Incorrect | `-N` |
| Unanswered | `0` |
| Marked for review (if unanswered) | `0` |
| Marked for review (if answered) | score of the answer |

The total score is the sum of per-question scores. There is no cap or minimum (a user can have a negative total score if they get many wrong).

## UI Behavior

- The negative marking value MUST be shown to the user before the test starts and on every question
- The current score (with negative marking applied) MAY be shown during the test (configurable per test)
- A "mark for review" option is available if the test allows it

## Anti-Negative-Marking-Gaming

A pattern of "selective answering" (always leaving hard questions unanswered) is detected and may be flagged as suspicious but is not penalized (it is a legitimate strategy).

---

# State Machine (Mandatory)

The live exam session is a state machine:

```
SCHEDULED → STARTING → LIVE → RESULT_FINALIZATION → ENDED
    ↓           ↓        ↓              ↓
CANCELLED   CANCELLED PAUSED         INVALIDATED
                ↑      ↓
                └── LIVE
```

## State Definitions

| State | Description |
|---|---|
| `SCHEDULED` | Session is created but not yet started. Users can join within the join window (default: 5 minutes before start). |
| `STARTING` | Session is about to start. Final checks. Users are connected and ready. |
| `LIVE` | Session is running. Timer is active. Users can submit answers. |
| `PAUSED` | Session is paused. Timer is frozen. Users cannot submit answers. |
| `RESULT_FINALIZATION` | Session has ended. Server is computing results. Users cannot submit answers. |
| `ENDED` | Results are final. Users can view their results. Leaderboard is published. |
| `CANCELLED` | Session was cancelled before it started. Users are refunded. |
| `INVALIDATED` | Session ended in error. No scores recorded. Users are refunded. |

## Transition Rules (Mandatory)

- `SCHEDULED → STARTING`: server-initiated, automatic, at `startsAt`
- `SCHEDULED → CANCELLED`: admin-initiated or system-initiated (e.g., too few participants)
- `STARTING → LIVE`: server-initiated, automatic, after all users are ready or 30 seconds elapse
- `LIVE → PAUSED`: admin-initiated or system-initiated
- `PAUSED → LIVE`: admin-initiated or system-initiated
- `LIVE → RESULT_FINALIZATION`: server-initiated, automatic, at `serverEndsAt`
- `LIVE → INVALIDATED`: system-initiated, on critical error
- `RESULT_FINALIZATION → ENDED`: server-initiated, automatic, after finalization completes

## State Transition Side Effects

Each transition produces:

- A `session.state` WebSocket broadcast
- An `AuditLog` entry
- A Sentry breadcrumb

State transitions are atomic; if any side effect fails, the transition is rolled back (or compensated).

## State Validation

A user action is only valid in certain states. Submitting an answer in `PAUSED` is rejected. The client UI hides the submit button in non-`LIVE` states; the server also rejects it as defense in depth.

---

# Question and Answer Flow (Mandatory)

The flow of a single question:

```
[1] Server: question.show (with question, options, questionEndsAt)
[2] User: thinks, may mark for review, may submit
[3] User: answer.submit (or timer expires)
[4] Server: validates submission (within window, not duplicate)
[5] Server: scores the answer (deferred to result finalization if scoring is hidden)
[6] Server: answer.accepted (does NOT confirm correctness)
[7] Server: leaderboard.update (if scoring is shown live)
[8] [Optional] Server: question.hide + question.next (move to next question)
```

## Question Visibility (Mandatory)

A question is sent to a user exactly once, in `question.show`. The question is NOT sent again unless the user reconnects (in which case the in-flight question is re-sent as part of `session.state`).

The question text and options are NOT logged on the server (see WebSocket Security). The `questionId` is logged for grading purposes.

## Answer Submission Rules (Mandatory)

- An answer is accepted if and only if:
  - The current state is `LIVE`
  - The current question is still within the per-question time window
  - The user has not already submitted an answer for this question
  - The submitted option is one of the valid options for the question
- An answer that fails any check is rejected with `answer.rejected` and a reason
- The `answer.accepted` message does NOT confirm correctness; the correctness is computed during result finalization (or revealed live if the test is configured for live-correctness)

## Live Correctness (Optional)

A test MAY be configured to reveal correctness immediately after submission (`liveCorrectness: true`). In that case:

- `answer.accepted` includes the correctness and the correct option
- The leaderboard updates immediately on `answer.submit`
- This is the default for Quiz Battles but not for Live Tests

## Mark for Review

A user MAY mark a question for review (`markForReview: true`). The flag is stored on the `LiveExamAnswer` row. A flagged question is highlighted in the question palette. At the end of the test (or during the test, if enabled), the user can revisit flagged questions.

If the user has not submitted an answer for a question by the per-question deadline, the question is marked `UNANSWERED` regardless of the review flag.

## Question Navigation

In Live Test mode, questions are auto-advanced (no user navigation). In Quiz Battle mode, the user navigates explicitly. The flow differs per mode; the system MUST NOT conflate them.

---

# Failure Modes and Recovery (Mandatory)

| Failure | Detection | Recovery |
|---|---|---|
| Server crash during LIVE | Health check fails | A new server picks up the session from Redis. The session is restored. Clients reconnect. |
| WebSocket message loss | Client does not receive `answer.accepted` within 5 seconds | Client retries with idempotency key. Server detects duplicate and returns the original `answer.accepted`. |
| Redis unavailable | Health check fails | Circuit breaker opens. The session continues in degraded mode (no leaderboard updates). The leaderboard is rebuilt from the database when Redis recovers. |
| Database slow | Slow query log | Read replicas are promoted. The session continues. |
| Network partition (client) | No message received in 30 seconds | Client attempts reconnect. After `MAX_DISCONNECT_DURATION`, user is marked `ABANDONED`. |
| LLM API rate limit (if used) | 429 response | The system queues the AI step or skips it. The exam continues. |

## State Recovery (Mandatory)

Every `LiveExamSession` has a recovery point stored in Redis: `sessionId`, `currentQuestionIndex`, `serverEndsAt`, `state`, `participationCount`. A new server instance reads this on startup. If the recovery point is inconsistent with the database, the database is the source of truth and Redis is updated.

## Disaster Recovery

Live exams are time-sensitive. The RPO is 30 seconds; the RTO is 60 seconds. The session state is stored in both PostgreSQL (durable) and Redis (fast). On full region failure, the session is restored from PostgreSQL in a different region.

---

# Telemetry and Observability (Mandatory)

## Per-Session Telemetry

Every live exam session produces a `LiveExamSessionMetric` record:

```typescript
interface LiveExamSessionMetric {
  sessionId: string;
  startedAt: string;
  endedAt: string;
  duration: number;
  totalParticipants: number;
  completedParticipants: number;
  abandonedParticipants: number;
  disqualifiedParticipants: number;
  averageScore: number;
  medianScore: number;
  p50LatencyMs: number;
  p95LatencyMs: number;
  p99LatencyMs: number;
  totalMessages: number;
  flaggedSessions: number;
  errors: number;
}
```

## Per-User Telemetry

Every participation produces a `LiveExamParticipationMetric` record:

```typescript
interface LiveExamParticipationMetric {
  participationId: string;
  userId: string;
  connectedDurationMs: number;
  disconnectedDurationMs: number;
  tabSwitchCount: number;
  messagesSent: number;
  messagesReceived: number;
  firstAnswerLatencyMs: number;
  averageAnswerLatencyMs: number;
  reconnectCount: number;
  flags: AntiCheatDetection[];
}
```

## Sentry Integration (Mandatory)

Every WebSocket message is wrapped in a Sentry span with op `live_exam.ws_message` and tags `live_exam.type`, `live_exam.session_id`, `live_exam.user_id`. Server-side, the span is grouped by `type` to keep the trace volume manageable.

## Dashboards (Mandatory)

A live exam dashboard in Grafana shows, in real time:

- Active sessions count
- Active participants count
- p50/p95/p99 message latency
- Reconnect rate
- Error rate
- Disqualification rate
- Anti-cheating flags

The dashboard is auto-launched 5 minutes before any scheduled live exam.

---

# Compatibility With Other Standards

This document defers to:

- [security-rules.md](docs/standards/security-rules.md) for auth, rate limiting, and audit logging
- [architecture-rules.md](docs/standards/architecture-rules.md) for module boundaries and the ADR requirement (a live exam mode change requires an ADR)
- [backend-rules.md](docs/standards/backend-rules.md) for service patterns and Sentry instrumentation
- [Course-rules.md](docs/standards/Course-rules.md) for Course ownership and audit requirements
- [testing-rules.md](docs/standards/testing-rules.md) for load testing and chaos testing
- [admin-panel-rules.md](docs/standards/admin-panel-rules.md) for moderation actions and audit log access

---

# Sprint 1 Compliance Checklist

Since the live exam engine is not yet implemented, this section will be expanded in v1.0. For Sprint 1:

- [ ] ADR written for: WebSocket library (Socket.IO vs native), timer storage (Redis vs in-memory), scoring engine location (in-session vs post-session)
- [ ] `LiveExamSession` and `LiveExamParticipation` models defined in `schema.prisma`
- [ ] `SessionState` enum and state-machine transitions implemented
- [ ] WebSocket gateway with JWT authentication
- [ ] Server-authoritative timer with `timer.tick` at 1Hz
- [ ] Reconnect flow with `sessionResumeToken` and sequence replay
- [ ] Anti-cheating detection for tab switching, multi-accounting, and bot submission
- [ ] Leaderboard with ZSET, top-20 + user rank, tie-breaking rules
- [ ] Negative marking scoring formula
- [ ] Result finalization transaction with idempotency
- [ ] Sentry spans for every WebSocket message and finalization step
- [ ] Load test harness for 10,000 concurrent users
- [ ] Chaos test harness for 30% random disconnects
- [ ] Disaster recovery runbook

---

# Final Directive

A live exam is a contract with the user.

The user trusts that:

- The timer is fair
- The questions are right
- The leaderboard is honest
- The result is final

A broken live exam is a broken promise.

Build live exams that are timed correctly, scored fairly, defended against cheating, and finalized with audit trails.

Verify all four.
