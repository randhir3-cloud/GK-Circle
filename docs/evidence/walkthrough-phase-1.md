# Phase 1 — Visual Identity Walkthrough

**Scope:** Replace all inherited jovVix visual assets with GK Circle brand assets. No backend, database, API, Docker, routing, or CSS-token changes.

---

## Changes Made

### New Scripts

| File | Purpose |
|---|---|
| [build-brand-assets.mjs](file:///e:/GK%20Circle%20v2/scripts/build-brand-assets.mjs) | Deterministic asset pipeline — builds all brand files from a pure SVG emblem using Sharp and SVG text (no AI text rendering) |
| [check-brand-assets.mjs](file:///e:/GK%20Circle%20v2/scripts/check-brand-assets.mjs) | Validation script — verifies dimensions, channels, alpha, and ICO frame contents using Sharp + icojs |

### New Brand Assets (generated, committed to public/)

All assets are derived from a single SVG master emblem — consistent shape, spacing, and colour at every size.

**Master mark** (`app/public/brand/`)

| File | Dimensions | Notes |
|---|---|---|
| [gk-circle-mark.svg](file:///e:/GK%20Circle%20v2/app/public/brand/gk-circle-mark.svg) | — | SVG source of truth |
| [gk-circle-mark.png](file:///e:/GK%20Circle%20v2/app/public/brand/gk-circle-mark.png) | 1024×1024 RGBA | PNG render of master |
| [gk-circle-wordmark.png](file:///e:/GK%20Circle%20v2/app/public/brand/gk-circle-wordmark.png) | 1024×256 RGBA | Emblem + "GK Circle" (real-font, no AI text) |
| [gk-circle-wordmark.webp](file:///e:/GK%20Circle%20v2/app/public/brand/gk-circle-wordmark.webp) | 1024×256 RGBA | WebP version of wordmark |

**Public assets** (`app/public/`)

| File | Dimensions | Use |
|---|---|---|
| [favicon.ico](file:///e:/GK%20Circle%20v2/app/public/favicon.ico) | 16+32+48 frames | Browser tab, bookmarks |
| [favicon-16x16.png](file:///e:/GK%20Circle%20v2/app/public/favicon-16x16.png) | 16×16 | Small favicon |
| [favicon-32x32.png](file:///e:/GK%20Circle%20v2/app/public/favicon-32x32.png) | 32×32 | Standard favicon |
| [favicon-48x48.png](file:///e:/GK%20Circle%20v2/app/public/favicon-48x48.png) | 48×48 | Windows taskbar |
| [favicon-180x180.png](file:///e:/GK%20Circle%20v2/app/public/favicon-180x180.png) | 180×180 | Apple touch icon |
| [favicon-256x256.png](file:///e:/GK%20Circle%20v2/app/public/favicon-256x256.png) | 256×256 | High-DPI favicon |
| [og-image.png](file:///e:/GK%20Circle%20v2/app/public/og-image.png) | 1200×630 | Open Graph / social share |
| [readme-logo.webp](file:///e:/GK%20Circle%20v2/app/public/readme-logo.webp) | 1024×256 | README header |

### Nuxt Config Changes ([nuxt.config.ts](file:///e:/GK%20Circle%20v2/app/nuxt.config.ts))

Added missing global OG metadata defaults (safe additions — do not conflict with `app.vue` SSR output):

```diff
+ { property: "og:image:width",  content: "1200" },
+ { property: "og:image:height", content: "630" },
+ { property: "og:image:alt",    content: "GK Circle - PCS Exam Preparation" },
+ { property: "og:type",         content: "website" },
```

Added Apple touch icon link:

```diff
+ { rel: "apple-touch-icon", sizes: "180x180", href: "/favicon-180x180.png" },
```

**Not added** (already emitted server-side by `app.vue` via `useSeoMeta`):
- `og:title`, `og:description`, `og:image`, `twitter:card`, `twitter:image`, `twitter:title`, `twitter:description`

---

## Visual Assets

![GK Circle OG Banner](file:///e:/GK%20Circle%20v2/app/public/og-image.png)

![GK Circle Wordmark](file:///e:/GK%20Circle%20v2/app/public/brand/gk-circle-wordmark.png)

---

## Checks Run

| Check | Result |
|---|---|
| `node scripts/check-brand-assets.mjs` | ✅ 13/13 passed |
| `node scripts/check-retired-terminology.mjs` | ✅ Passed |
| `cd app && npm run lint` | ✅ Passed (prettier auto-fixed 3 spacing issues) |
| `cd app && npm run build` | ✅ Passed — 8.64 MB output, no errors |
| `docker compose config --quiet` | ✅ Passed |

---

## Known Limitations / Follow-up Notes

- **Emblem design:** The SVG emblem is a clean programmatic recreation (navy ring, gold orbital arc, white chevron) that is recognisable at 16×16. The original AI-generated raster was discarded due to baked checkerboard pixels — the SVG source is the canonical master going forward.
- **Font rendering:** The wordmark uses the system's `Segoe UI` / Helvetica font stack via SVG. On the production server (Linux), Sharp will use the system Pango/freetype fonts. If exact font consistency is required across environments, consider bundling a licensed font and referencing it via SVG `@font-face`.
- **No code/content changes:** No page components, API routes, database migrations, or Compose services were modified.

---

## Breaking Changes: NO

## Database Migration Status: Not applicable (no DB changes)
