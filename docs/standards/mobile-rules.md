# GK Circle Mobile Rules

Version: 1.0

Status: Mandatory

---

# Technology

React Native

Expo

TypeScript

---

# Mobile First Rule

Mobile is primary platform.

Desktop adapts.

Not vice versa.

---

# Offline Rule

Support:

Downloaded Lessons

Saved Notes

Bookmarks

Offline Revision

---

# Notification Rule

Support:

Test Reminders

Current Affairs

Mentorship Updates

Community Notifications

---

# Performance Rule

Target:

60 FPS

Fast Startup

Low Memory Usage

---

# Accessibility Rule

Support:

Screen Readers

Dynamic Text

Touch Targets

Dark Mode

---

# v0.5 Addendum — Architecture-Intent Standard

The v1.0 section above defines the principles (React Native + Expo, Mobile First, Offline, Notifications, Performance, Accessibility). This addendum makes those principles enforceable as code, configuration, and process. The structure mirrors the other v0.5 standards.

This is **architecture-intent**: the mobile app is not yet implemented. The rules below are the contract it MUST satisfy when built. The next revision (v1.0 of this file) will merge these rules into the body.

The mobile platform is the primary surface for GK Circle users. Most aspirants preparing for UPSC, SSC, banking, and similar exams in India are on mobile, often on mid-range Android devices with intermittent connectivity, on metered data, in regions with low-bandwidth cellular networks (2G/3G). The standards below are tuned for that reality.

---

# Dependency Rule (Mandatory)

This document does not replace AGENTS.md.

It supplements AGENTS.md.

If this document conflicts with AGENTS.md, AGENTS.md wins.

If this document conflicts with another standards file, the more specific standard wins.

If ambiguity exists, document the ambiguity in an ADR before implementation.

# Verification Requirement (Mandatory)

A rule is not considered satisfied because code, configuration, tests, or documentation exist. A rule is satisfied only when evidence exists. For mobile, evidence must include:

- A device farm test on at least: 1 low-end Android (e.g., Samsung A13, Android 12, 2GB RAM), 1 mid-range Android (e.g., Redmi Note 11, Android 13, 4GB RAM), 1 iPhone (iPhone 12, iOS 16), 1 iPad (iPad 8th gen, iPadOS 16)
- An airplane-mode test demonstrating that downloaded lessons, notes, and bookmarks are accessible offline
- A 2G/3G network test demonstrating that the app remains usable (page load < 3s on 3G)
- A battery test demonstrating that background sync consumes < 2% battery per hour on a 4000mAh device

---

# Offline-First (Mandatory)

The mobile app MUST be offline-first. A user who has never connected to the network should still be able to:

- View downloaded lessons
- Read saved notes
- Browse bookmarks
- Take a downloaded mock test (if cached)
- View a previously loaded leaderboard snapshot

## Offline Data Categories

The app distinguishes between four categories of data:

| Category | Description | Offline Behavior |
|---|---|---|
| `CRITICAL` | Auth token, user profile, downloaded lessons, downloaded tests, saved notes, bookmarks | Always available offline |
| `CACHED` | Recently viewed content, search results, leaderboard snapshots | Available offline for 7 days |
| `REALTIME` | Live test sessions, real-time chat, live leaderboard | NOT available offline; clear messaging when offline |
| `BACKGROUND` | Analytics, telemetry, non-urgent sync | Queued and synced on next online session |

## Offline Storage (Mandatory)

Offline data is stored in:

- **AsyncStorage** (or MMKV) for small key-value data: auth tokens, user preferences, last-seen timestamps
- **SQLite** (via `expo-sqlite` or `react-native-sqlite-storage`) for structured data: lessons, notes, bookmarks
- **File system** (via `expo-file-system`) for media: video lessons, audio, images, PDFs

The storage layer is wrapped behind a `MobileStorageClient` interface. The default implementation uses SQLite + file system, but the interface allows swapping to a different backend (e.g., WatermelonDB for sync, Realm for reactive queries).

## Storage Quota and Eviction (Mandatory)

Mobile devices have limited storage. The app MUST enforce quotas:

| Data Type | Default Quota | Eviction Policy |
|---|---|---|
| Downloaded lessons | 2 GB | LRU (least recently accessed) |
| Notes | 100 MB | Manual (user-managed) |
| Bookmarks | 10 MB | Manual (user-managed) |
| Cached content | 500 MB | LRU + 7-day TTL |
| Media cache | 500 MB | LRU + 30-day TTL |

The user is notified when the quota reaches 80% ("Storage almost full — old lessons will be auto-removed soon"). The user is prompted when the quota reaches 95% ("Storage full — please remove some downloaded lessons").

The user can manually pin lessons (preventing auto-removal) up to a hard cap of 5 GB.

## Offline Indicator (Mandatory)

A persistent offline indicator is shown in the app chrome when the device is offline. The indicator shows:

- A "You are offline" label
- The last sync timestamp ("Last synced 12 minutes ago")
- A "Retry sync" button (manual trigger)
- A count of pending sync operations ("3 items waiting to sync")

The indicator is dismissible but reappears on every screen if the device remains offline for ≥ 5 seconds.

---

# Sync Queue (Mandatory)

Every mutation made by the user (submit a test, add a note, bookmark a lesson, react to a post) is enqueued in a sync queue if the device is offline or the request fails.

## Sync Queue Implementation (Mandatory)

The sync queue is a persistent FIFO queue stored in SQLite. Each entry has:

```typescript
interface SyncQueueEntry {
  id: string;                      // UUID
  type: 'TEST_SUBMIT' | 'NOTE_CREATE' | 'NOTE_UPDATE' | 'BOOKMARK_ADD' | 'BOOKMARK_REMOVE' | 'REACTION_ADD' | 'REACTION_REMOVE' | 'COMMENT_CREATE' | 'PROFILE_UPDATE';
  payload: Record<string, any>;
  clientCreatedAt: string;         // ISO 8601
  attempts: number;
  lastAttemptAt?: string;
  lastError?: string;
  status: 'PENDING' | 'IN_FLIGHT' | 'COMPLETED' | 'FAILED';
  dependsOn?: string[];            // IDs of other entries that must complete first
}
```

## Sync Rules

1. The queue is processed in FIFO order, except for entries with `dependsOn` (which wait for their dependencies).
2. An entry is retried with exponential backoff: 1s, 5s, 30s, 2min, 10min, 1hr, 6hr. After 6hr, the entry is marked `FAILED` and the user is notified.
3. An entry is marked `COMPLETED` only when the server returns a `2xx` response. The server response is required.
4. The server response is stored alongside the entry for inspection (`serverResponse`).
5. The queue is bounded at 10,000 entries. Older entries are evicted (and the user is notified) if the queue exceeds this.

## Conflict Resolution

Conflicts (e.g., a note was edited on both web and mobile) are resolved by the server using last-writer-wins, with a `version` field on each entity. If the conflict is significant (e.g., different content), the user is shown both versions and asked to pick.

## Sync Triggers (Mandatory)

The sync queue is processed:

1. On app start (foreground)
2. When connectivity is restored (offline → online transition)
3. On a 5-minute timer (if the app is in the foreground and online)
4. On user-initiated "Sync now" button
5. On app backgrounding (best-effort; OS may kill the app)

## Sync Telemetry (Mandatory)

The app sends a heartbeat to the server every 60 seconds when in the foreground, indicating the queue depth. This is used to detect stuck devices.

---

# Media Download Strategy (Mandatory)

Lesson videos, audio, images, and PDFs are large. The download strategy MUST be bandwidth-aware and battery-aware.

## Download Modes

The app supports three download modes for lessons:

| Mode | Description | Use Case |
|---|---|---|
| `MANUAL` | User explicitly taps "Download" | Default for paid content |
| `AUTO_ON_WIFI` | Auto-downloaded on Wi-Fi | Default for free content the user has bookmarked |
| `AUTO_ALWAYS` | Auto-downloaded on any network | Opt-in only; user must explicitly enable |

The default for free content the user has bookmarked is `AUTO_ON_WIFI`. The default for paid content is `MANUAL` (the user has paid; they choose what to download).

## Download Resumption (Mandatory)

A download interrupted by network loss, app backgrounding, or OS kill MUST be resumable from the last byte received. The HTTP `Range` header is used. The server MUST support byte-range requests.

## Download Verification (Mandatory)

After download, the file's SHA-256 hash is verified against the server-provided hash. A mismatch deletes the file and re-downloads (one retry). A second mismatch marks the lesson as "download failed" and the user is notified.

## Storage of Downloaded Media (Mandatory)

Downloaded media is stored in the app's private documents directory (`FileSystem.documentDirectory` in Expo). Media is encrypted at rest using the device's secure enclave (iOS Keychain / Android Keystore) when available.

## Cache Invalidation (Mandatory)

When a lesson is updated on the server (new version published), the local copy is invalidated and the user is prompted to re-download. The old copy is deleted only after the new copy is fully downloaded.

---

# Background Sync (Mandatory)

The app supports background sync to keep local data fresh without requiring the user to open the app.

## Background Sync Triggers

Background sync is triggered by:

1. **Time-based**: every 6 hours (configurable; default 6am, 12pm, 6pm, 12am in the user's local timezone)
2. **Push notification**: a server-side trigger via FCM (Android) or APNs (iOS) that wakes the app
3. **Wi-Fi only**: when the device connects to Wi-Fi (if `syncOnWifiOnly: true`, default)
4. **Charging only**: when the device is charging (if `syncOnChargingOnly: true`, default for non-time-based syncs)

## Background Sync Behavior

When background sync runs:

- The app fetches deltas (not full data) for lessons, notes, bookmarks, and the user's profile
- New lesson content is downloaded only if it is in `AUTO_ON_WIFI` or `AUTO_ALWAYS` mode AND the device is on Wi-Fi
- Analytics and telemetry are batched and uploaded
- The auth token is refreshed if it is within 1 hour of expiration

Background sync MUST complete in < 30 seconds and consume < 1% battery per run. If it takes longer, the OS will kill the process; the app is built to handle this gracefully (the next foreground launch detects incomplete sync and resumes).

## Background Fetch Limitations

iOS and Android both limit background execution:

- iOS: `BGAppRefreshTask` is best-effort; the OS may delay or skip it
- Android: `WorkManager` constraints (charging, Wi-Fi, etc.) determine when the task runs

The app is built to be opportunistic. If background sync doesn't run, the user sees a stale-while-revalidate experience on next foreground.

## Push Notification-Based Sync

For high-priority updates (e.g., a live test is about to start), the server sends a silent push notification that wakes the app for a brief sync. The notification payload is empty; the side effect is the app fetching the latest data.

The silent push is rate-limited:

- ≤ 1 per user per 5 minutes
- ≤ 5 per user per hour

---

# Low-Bandwidth Mode (Mandatory)

GK Circle users frequently access the app on 2G/3G networks. The app MUST remain usable on these networks.

## Low-Bandwidth Detection

The app detects low-bandwidth via:

1. **`navigator.connection.effectiveType`**: `2g`, `3g`, `4g`, `slow-2g` (where supported)
2. **Effective bandwidth measurement**: download a 10KB file on app start and measure throughput
3. **User preference**: the user can manually set the bandwidth mode in settings

The detected bandwidth is stored in a `BandwidthProfile`:

```typescript
interface BandwidthProfile {
  effectiveType: '2g' | '3g' | '4g' | 'wifi' | 'unknown';
  measuredKbps: number;
  measuredAt: string;
  userOverride?: 'LOW' | 'AUTO' | 'HIGH';
}
```

## Low-Bandwidth Behavior (Mandatory)

When bandwidth is detected as `2g`, `slow-2g`, or the user has set `LOW` mode, the app:

- Disables autoplay of videos (videos show a poster + play button)
- Loads lower-resolution images (a `?w=480` query param to the image CDN)
- Skips prefetching of upcoming content
- Disables real-time features (live test joining, real-time chat) and shows a "Low bandwidth — some features are disabled" message
- Uses text-only view for community posts (no image previews)
- Disables video thumbnails in feeds
- Pauses background sync entirely (no point syncing on 2G)
- Disables analytics beacon (saves data)

When bandwidth improves, the app re-enables these features within 5 seconds.

## Image Loading

Images are loaded via a CDN that supports responsive sizing:

- `image-480.jpg` for low-bandwidth
- `image-720.jpg` for normal
- `image-1080.jpg` for high-bandwidth
- `image-2160.jpg` for retina desktop

The app requests the appropriate variant based on the `BandwidthProfile`.

## Lazy Loading (Mandatory)

Below-the-fold content is lazy-loaded. The app uses `IntersectionObserver` (web) or `FlatList` with `windowSize` tuned for the device (mobile: 3, tablet: 5).

---

# Battery-Aware Behavior (Mandatory)

The app MUST be battery-aware. Heavy operations (background sync, large downloads, video streaming) consume battery. The app detects battery state and adjusts behavior.

## Battery State Detection (Mandatory)

The app reads battery state via:

- `expo-battery` (returns `batteryLevel`, `batteryState`, `lowPowerMode`)
- iOS `UIApplication.shared.isLowPowerModeEnabled` (iOS only)
- Android `PowerManager.isPowerSaveMode` (Android only)

The `BatteryProfile` is:

```typescript
interface BatteryProfile {
  level: number;            // 0–1
  state: 'CHARGING' | 'DISCHARGING' | 'FULL' | 'UNKNOWN';
  lowPowerMode: boolean;
  updatedAt: string;
}
```

## Battery-Aware Rules (Mandatory)

| Battery State | Behavior |
|---|---|
| `level < 0.15` and `lowPowerMode: true` | Disable all background sync. Reduce animation. Disable autoplay. |
| `level < 0.20` | Pause auto-download. Reduce sync frequency to 1/hr. |
| `CHARGING` and `level > 0.80` | Allow heavy sync and auto-download. |
| `FULL` (charging complete) | Allow heavy operations. |
| `lowPowerMode: true` (regardless of level) | Same as `level < 0.15` behavior. |

The user is shown a "Battery saver mode is on — some features are reduced" banner when the app is in low-power mode. The banner is dismissible.

## Foreground vs Background

- **Foreground**: full functionality; user can opt into heavy operations
- **Background**: heavy operations are gated by battery, charging, and Wi-Fi constraints

The app does NOT use the camera, microphone, or GPS in the background unless the user is on a video call or in a live session.

## Battery Telemetry (Mandatory)

The app reports battery state to the server in a heartbeat (every 5 minutes, foreground only). This is used to correlate app behavior with battery impact and to surface battery-impacting issues in the admin panel.

---

# Notification Strategy (Mandatory)

The app uses push notifications for user engagement. The strategy is consent-first and value-driven.

## Notification Categories (Mandatory)

| Category | Default | Examples |
|---|---|---|
| `TEST_REMINDERS` | Opt-in | "Your daily test is ready", "UPSC Prelims mock test in 24 hours" |
| `CURRENT_AFFAIRS` | Opt-in | "Daily current affairs digest is ready" |
| `COMMUNITY` | Opt-in | "X replied to your post", "Y mentioned you" |
| `MENTORSHIP` | Opt-in | "Your mentor has scheduled a session", "New note from your mentor" |
| `LIVE_TEST` | Opt-out (always on) | "Live test starts in 15 minutes", "Your live test result is ready" |
| `SYSTEM` | Always on | "Password changed", "New device login", "Subscription updated" |

`LIVE_TEST` and `SYSTEM` are always on. All other categories are opt-in. The user is asked for permission on first use, with a clear explanation of what each category means.

## Notification Frequency Cap (Mandatory)

A user receives at most:

- 5 `TEST_REMINDERS` per day
- 3 `CURRENT_AFFAIRS` per day
- 10 `COMMUNITY` per day (with smart batching: "X and 3 others replied to your post")
- 3 `MENTORSHIP` per day

If a category would exceed its cap, the notifications are batched ("5 community updates in the last hour" with a deep link to the activity feed).

## Quiet Hours (Mandatory)

The user can set quiet hours (e.g., 10pm–7am). During quiet hours, only `SYSTEM` and `LIVE_TEST` (if a test starts in < 15 minutes) are delivered. All others are queued and delivered at the end of quiet hours.

## Deep Links (Mandatory)

Every notification has a deep link that takes the user to the relevant screen:

- A `TEST_REMINDERS` notification for "Daily test" → opens the test
- A `COMMUNITY` notification for "X replied" → opens the post
- A `MENTORSHIP` notification for "session scheduled" → opens the session detail

The deep link uses the app's URL scheme (`gkcircle://`) and is registered with both iOS and Android.

---

# Performance Targets (Mandatory)

| Metric | Target | Critical Threshold |
|---|---|---|
| App cold start (median) | < 1.5s | < 3s |
| App warm start (median) | < 0.5s | < 1s |
| Screen transition (median) | < 300ms | < 600ms |
| Scroll FPS (target) | 60 FPS | ≥ 50 FPS |
| JS bundle size (initial) | < 5 MB | < 8 MB |
| Native binary size | < 50 MB | < 80 MB |
| Memory usage (typical) | < 200 MB | < 400 MB |
| Crash-free sessions | > 99.5% | > 99.0% |

A regression in any of these by ≥ 20% blocks the release.

## Performance Telemetry (Mandatory)

The app reports the following metrics to the server on every screen transition:

- `screen_name`
- `screen_load_time_ms` (from navigation start to first contentful paint)
- `js_heap_size_mb`
- `fps_avg`
- `fps_min`

This is sampled at 10% to limit telemetry overhead.

## Native Module Bridge Budget

Native module calls (e.g., to AsyncStorage, SQLite) MUST complete in < 50ms p95. Calls that exceed this are a sign of misuse (e.g., large objects being serialized across the bridge) and MUST be flagged in code review.

---

# Accessibility (Mandatory)

The app supports the following accessibility features (in addition to the v1.0 list):

- **Screen reader support**: every interactive element has a meaningful `accessibilityLabel` and `accessibilityHint` (where applicable)
- **Dynamic type**: respects the user's system font size preference (iOS) or font scale (Android)
- **Touch targets**: minimum 44x44 pt (iOS) / 48x48 dp (Android)
- **Color contrast**: WCAG 2.1 AA minimum (4.5:1 for body text, 3:1 for large text)
- **Reduce motion**: respects the user's "reduce motion" preference
- **Dark mode**: supports both light and dark themes; auto-switches based on system preference
- **Focus management**: visible focus indicators on all interactive elements

Accessibility is tested with VoiceOver (iOS) and TalkBack (Android) before each release.

## Accessibility Audit (Mandatory)

Before each release, the app is run through an accessibility audit using:

- `axe` for React Native (via `react-native-a11y` or similar)
- Manual VoiceOver and TalkBack testing on at least 2 critical screens
- Color contrast checks on the design system

A regression in accessibility score by ≥ 10% blocks the release.

---

# Device and OS Support (Mandatory)

The app supports the following devices and OS versions:

| Platform | Minimum | Recommended |
|---|---|---|
| iOS | iOS 15.0 | iOS 17.0+ |
| Android | Android 9.0 (API 28) | Android 13+ |
| Screen size | 4.7" (iPhone SE) | 6.1"+ (iPhone 14) |
| RAM | 2 GB | 4 GB+ |
| Storage | 100 MB free | 1 GB+ free |

The app MUST function on a 2GB RAM Android device with no more than 200MB of memory usage during typical flows.

## Device Performance Tiering (Mandatory)

The app detects the device's performance tier and adjusts behavior:

| Tier | Characteristics | Behavior |
|---|---|---|
| `HIGH` | iPhone 12+, Pixel 6+, ≥ 6GB RAM | Full animations, all effects, full sync |
| `MID` | iPhone XR/11, mid-range Android, 3–6GB RAM | Reduced animations, normal sync |
| `LOW` | Older devices, 2–3GB RAM | Minimal animations, reduced sync, lower-res media |

The tier is detected on first launch and stored in `DeviceProfile`. The user can override this in settings.

---

# Security on Device (Mandatory)

The app implements on-device security:

- **Auth tokens**: stored in the iOS Keychain / Android Keystore (via `expo-secure-store`)
- **Biometric unlock**: optional; user can enable fingerprint/face unlock for app access
- **Screen recording prevention**: the app blocks screen recording on iOS (`UIScreen.captured`); on Android, the app shows a "Recording detected" warning and the user is asked to stop recording
- **Jailbreak/root detection**: the app detects jailbroken/rooted devices and warns the user (does not block access, but flags for review)
- **Certificate pinning**: the app pins the API server's TLS certificate to prevent MITM attacks
- **Clipboard**: the app does not read from or write to the clipboard except for explicitly user-initiated actions (e.g., "Copy coupon code")
- **Background screen blur**: when the app is backgrounded, a blur is applied to the last screen (prevents sensitive content from being visible in the app switcher)

## PII Handling (Mandatory)

The app does not log PII (name, email, phone) to console or to telemetry. The `__DEV__` console logs are stripped in production builds.

## App Transport Security (iOS) / Network Security Config (Android)

- iOS: ATS is enabled; HTTP traffic is blocked (`NSAllowsArbitraryLoads = false`). Exceptions are explicit and documented.
- Android: cleartext traffic is blocked by default (`usesCleartextTraffic="false"` in the manifest).

---

# Compatibility With Other Standards

This document defers to:

- [security-rules.md](docs/standards/security-rules.md) for auth and session management
- [architecture-rules.md](docs/standards/architecture-rules.md) for module boundaries and the ADR requirement (a mobile platform change requires an ADR)
- [rag-rules.md](docs/standards/rag-rules.md) for any AI features embedded in the app
- [live-exam-rules.md](docs/standards/live-exam-rules.md) for live test client behavior
- [admin-panel-rules.md](docs/standards/admin-panel-rules.md) for any admin actions available on mobile

---

# Sprint 1 Compliance Checklist

Since the mobile app is not yet implemented, this section will be expanded in v1.0. For Sprint 1:

- [ ] ADR written for: state management (Redux vs Zustand vs Context), offline storage (SQLite vs MMKV vs WatermelonDB), navigation (React Navigation vs Expo Router), push notification provider (FCM vs OneSignal vs Expo)
- [ ] `MobileStorageClient` interface and SQLite adapter
- [ ] Sync queue implementation with exponential backoff
- [ ] `BandwidthProfile` and `BatteryProfile` detection
- [ ] Notification permission flow with category opt-in
- [ ] Auth token storage in SecureStore
- [ ] Certificate pinning configured
- [ ] Sentry or equivalent for crash reporting and performance monitoring
- [ ] Deep link registration
- [ ] Device farm test on ≥ 3 devices

---

# Final Directive

A mobile app is the user's first touch with GK Circle.

A slow app is a deleted app.

A battery-draining app is a deleted app.

A data-hungry app is a deleted app.

A crashed app is a lost user.

Build mobile that is fast, offline-capable, battery-aware, low-bandwidth-friendly, and accessible.

Verify all five.
