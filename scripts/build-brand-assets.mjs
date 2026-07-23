/**
 * scripts/build-brand-assets.mjs  v3 — "The Knowledge Arc"
 *
 * Icon concept:
 *   A 270° arc (bottom-to-right, clockwise) + horizontal crossbar creates the
 *   hidden letter G. An inner gold orbit ring reinforces "Circle". A gold
 *   nucleus anchors the composition. Together: intelligence, growth, circular
 *   learning, GK identity — inspired by Linear, Stripe, Arc Browser.
 *
 * Deliverables:
 *   brand/gk-circle-mark.svg            SVG source (dark, transparent bg)
 *   brand/gk-circle-mark.png            1024×1024 RGBA
 *   brand/gk-circle-mark-light.svg      SVG source (light, transparent bg)
 *   brand/gk-circle-mark-light.png      1024×1024 RGBA
 *   brand/gk-circle-wordmark.png        1280×280 RGBA (transparent bg)
 *   brand/gk-circle-wordmark.webp       same as above, WebP
 *   brand/gk-circle-wordmark-dark.png   1280×280, navy bg (dark mode use)
 *   brand/gk-circle-wordmark-dark.webp  same, WebP
 *   public/favicon-{16,32,48,180,256}x{}.png
 *   public/favicon.ico                  multi-res: 16, 32, 48
 *   public/og-image.png                 1200×630 (feature-rich banner)
 *   public/readme-logo.webp             1280×280, navy bg
 *
 * Usage:
 *   node scripts/build-brand-assets.mjs
 */

import { mkdir, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import sharp from "sharp";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(__dirname, "..");
const PUBLIC = path.join(ROOT, "app", "public");
const BRAND = path.join(PUBLIC, "brand");

// ── Brand tokens ──────────────────────────────────────────────────────────────
const NAVY   = "#1a2557";
const NAVY_MID = "#232f6e";   // slightly lighter navy for depth layering
const GOLD   = "#C9A84C";
const GOLD_L = "#DFB86A";     // lighter gold for hover / secondary text
const WHITE  = "#FFFFFF";
const OFFWHITE = "#F5F5F7";   // Apple-style off-white
const NAVY_R = 26, NAVY_G = 37, NAVY_B = 87;

// ── Mark geometry (1024 × 1024 canvas) ───────────────────────────────────────
//   Center C = (512,512), outer arc radius MR = 400
//
//   G-arc: 270° clockwise sweep
//     start = 6-o'clock  = (512, 912) = (C, C+MR)
//     end   = 3-o'clock  = (912, 512) = (C+MR, C)
//   Crossbar: horizontal from 3-o'clock leftward to x=452
//   Inner orbit: 300° clockwise from 60° to 0°
//     start = (627, 711), end = (742, 512)
//   Nucleus: filled circle at center, r=58

const C  = 512;   // canvas centre
const MR = 400;   // outer arc radius
const SW = 70;    // outer arc stroke width
const IR = 230;   // inner orbit radius
const ISW = 26;   // inner orbit stroke width
const DR = 58;    // nucleus (dot) radius

// Key coordinate objects
const P = {
  bottom : { x: C,       y: C + MR },   // 6 o'clock  (512, 912)
  right  : { x: C + MR,  y: C },        // 3 o'clock  (912, 512)
  barEnd : { x: C - 60,  y: C },        // crossbar endpoint (452, 512)
  iStart : { x: Math.round(C + IR * 0.5),    y: Math.round(C + IR * 0.866) }, // inner arc start at 60° (627, 711)
  iEnd   : { x: C + IR,  y: C },        // inner arc end at 0° (742, 512)
};

// ── SVG generators ────────────────────────────────────────────────────────────

/**
 * Master mark SVG.
 * @param {string} ringColor  Colour for the G-arc and crossbar (NAVY or WHITE).
 */
function buildMarkSvg(ringColor) {
  const { bottom: S, right: E, barEnd: B, iStart: I1, iEnd: I2 } = P;
  return `<svg xmlns="http://www.w3.org/2000/svg"
     width="1024" height="1024" viewBox="0 0 1024 1024">

  <!--
    GK Circle "Knowledge Arc" mark
    Hidden G: the 270° arc + crossbar reveals the letter G on close inspection.
    Inner orbit: reinforces the "Circle" concept and suggests continuous learning.
    Gold nucleus: the knowledge centre.
  -->

  <!-- G-arc: 270° clockwise from 6-o'clock (bottom) to 3-o'clock (right) -->
  <path d="M ${S.x},${S.y} A ${MR},${MR} 0 1,1 ${E.x},${E.y}"
        fill="none" stroke="${ringColor}" stroke-width="${SW}"
        stroke-linecap="round"/>

  <!-- G crossbar: horizontal from 3-o'clock endpoint inward -->
  <line x1="${E.x}" y1="${E.y}" x2="${B.x}" y2="${B.y}"
        stroke="${ringColor}" stroke-width="${SW}"
        stroke-linecap="round"/>

  <!-- Inner orbit arc: 300° clockwise (60° → 0°), gold -->
  <path d="M ${I1.x},${I1.y} A ${IR},${IR} 0 1,1 ${I2.x},${I2.y}"
        fill="none" stroke="${GOLD}" stroke-width="${ISW}"
        stroke-linecap="round" opacity="0.92"/>

  <!-- Nucleus: gold filled circle -->
  <circle cx="${C}" cy="${C}" r="${DR}" fill="${GOLD}"/>
</svg>`;
}

// ── ICO builder (manual, pure-PNG ICO format) ─────────────────────────────────
function buildIco(frames) {
  const nFrames = frames.length;
  const HEADER = 6;
  const DIR_ENTRY = 16;
  let offset = HEADER + DIR_ENTRY * nFrames;

  const header = Buffer.alloc(HEADER);
  header.writeUInt16LE(0, 0);        // reserved
  header.writeUInt16LE(1, 2);        // type: ICO
  header.writeUInt16LE(nFrames, 4);

  const dirs = frames.map((f) => {
    const entry = Buffer.alloc(DIR_ENTRY);
    const sz = f.size >= 256 ? 0 : f.size;
    entry.writeUInt8(sz, 0);
    entry.writeUInt8(sz, 1);
    entry.writeUInt8(0, 2);
    entry.writeUInt8(0, 3);
    entry.writeUInt16LE(1, 4);
    entry.writeUInt16LE(32, 6);
    entry.writeUInt32LE(f.buf.length, 8);
    entry.writeUInt32LE(offset, 12);
    offset += f.buf.length;
    return entry;
  });

  return Buffer.concat([header, ...dirs, ...frames.map((f) => f.buf)]);
}

// ── Helper: resize with Lanczos3 ─────────────────────────────────────────────
function resizeMark(src, size, bg = { r: 0, g: 0, b: 0, alpha: 0 }) {
  return sharp(src)
    .resize(size, size, { kernel: sharp.kernel.lanczos3, fit: "contain", background: bg })
    .png()
    .toBuffer();
}

// ═════════════════════════════════════════════════════════════════════════════
// BUILD PIPELINE
// ═════════════════════════════════════════════════════════════════════════════

await mkdir(BRAND, { recursive: true });

// ── STEP 1: Master marks ──────────────────────────────────────────────────────
console.log("\n▸ Step 1 — Master marks");

const darkMarkSvg = buildMarkSvg(NAVY);
await writeFile(path.join(BRAND, "gk-circle-mark.svg"), darkMarkSvg);
const darkMarkBuf = await sharp(Buffer.from(darkMarkSvg)).ensureAlpha().png().toBuffer();
await writeFile(path.join(BRAND, "gk-circle-mark.png"), darkMarkBuf);
console.log("  ✓ gk-circle-mark.svg + .png  1024×1024  (navy/gold, transparent bg)");

const lightMarkSvg = buildMarkSvg(WHITE);
await writeFile(path.join(BRAND, "gk-circle-mark-light.svg"), lightMarkSvg);
const lightMarkBuf = await sharp(Buffer.from(lightMarkSvg)).ensureAlpha().png().toBuffer();
await writeFile(path.join(BRAND, "gk-circle-mark-light.png"), lightMarkBuf);
console.log("  ✓ gk-circle-mark-light.svg + .png  1024×1024  (white/gold, transparent bg)");

// ── STEP 2: Favicons (derived from dark mark) ─────────────────────────────────
console.log("\n▸ Step 2 — Favicon PNGs");

const FAVICON_SIZES = [256, 180, 48, 32, 16];
const faviconBufs = {};

for (const size of FAVICON_SIZES) {
  const buf = await resizeMark(darkMarkBuf, size);
  faviconBufs[size] = buf;
  await writeFile(path.join(PUBLIC, `favicon-${size}x${size}.png`), buf);
  console.log(`  ✓ favicon-${size}x${size}.png`);
}

// ── STEP 3: favicon.ico ───────────────────────────────────────────────────────
console.log("\n▸ Step 3 — favicon.ico");

const icoFrames = await Promise.all(
  [16, 32, 48].map(async (s) => ({ size: s, buf: await resizeMark(darkMarkBuf, s) }))
);
await writeFile(path.join(PUBLIC, "favicon.ico"), buildIco(icoFrames));
console.log("  ✓ favicon.ico  (16 + 32 + 48 frames)");

// ── STEP 4: Horizontal wordmark ───────────────────────────────────────────────
// Layout (1280 × 280):
//   [Mark 232×232] [GK bold navy] [thin gold divider] [Circle light gold]
// ─────────────────────────────────────────────────────────────────────────────
console.log("\n▸ Step 4 — Horizontal wordmark");

const WM_W = 1280, WM_H = 280;
const WM_MARK = 232;
const WM_MARK_X = 24;
const WM_MARK_Y = Math.round((WM_H - WM_MARK) / 2);

// Scale mark to wordmark height
const wmMarkBuf = await resizeMark(darkMarkBuf, WM_MARK);

// Typography metrics (Segoe UI ExtraBold / Light — approximate for sans-serif stack):
// "GK" at 148px bold ≈ 220px wide; divider 2px; "Circle" at 148px light ≈ 490px wide
const TEXT_X  = WM_MARK_X + WM_MARK + 22;   // 278
const GK_W    = 220;                           // approx width of "GK"
const DIV_X   = TEXT_X + GK_W + 14;           // divider x position
const CIRC_X  = DIV_X + 18;                   // "Circle" x position
const BASELINE = Math.round(WM_H / 2 + 58);   // vertical baseline (184)

const wmTextSvg = `<svg xmlns="http://www.w3.org/2000/svg"
     width="${WM_W}" height="${WM_H}" viewBox="0 0 ${WM_W} ${WM_H}">
  <!-- GK — extra bold, navy -->
  <text x="${TEXT_X}" y="${BASELINE}"
        font-family="'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
        font-size="148" font-weight="800"
        fill="${NAVY}" letter-spacing="-4">GK</text>
  <!-- Thin vertical gold divider -->
  <line x1="${DIV_X}" y1="36" x2="${DIV_X}" y2="${WM_H - 36}"
        stroke="${GOLD}" stroke-width="2.5" stroke-linecap="round" opacity="0.75"/>
  <!-- Circle — light weight, gold -->
  <text x="${CIRC_X}" y="${BASELINE}"
        font-family="'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
        font-size="148" font-weight="300"
        fill="${GOLD}" letter-spacing="-2">Circle</text>
</svg>`;

const wmTextBuf = await sharp(Buffer.from(wmTextSvg)).ensureAlpha().png().toBuffer();

// Composite on transparent canvas: mark + text
const wmBuf = await sharp({
  create: { width: WM_W, height: WM_H, channels: 4, background: { r: 0, g: 0, b: 0, alpha: 0 } },
})
  .composite([
    { input: wmMarkBuf, top: WM_MARK_Y, left: WM_MARK_X },
    { input: wmTextBuf, top: 0, left: 0, blend: "over" },
  ])
  .png()
  .toBuffer();

await writeFile(path.join(BRAND, "gk-circle-wordmark.png"), wmBuf);
const wmWebp = await sharp(wmBuf).webp({ lossless: true }).toBuffer();
await writeFile(path.join(BRAND, "gk-circle-wordmark.webp"), wmWebp);
console.log("  ✓ gk-circle-wordmark.png + .webp  1280×280  (transparent bg)");

// Dark-mode wordmark: navy background, light mark, white GK + gold Circle
const wmTextDarkSvg = `<svg xmlns="http://www.w3.org/2000/svg"
     width="${WM_W}" height="${WM_H}" viewBox="0 0 ${WM_W} ${WM_H}">
  <text x="${TEXT_X}" y="${BASELINE}"
        font-family="'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
        font-size="148" font-weight="800"
        fill="${WHITE}" letter-spacing="-4">GK</text>
  <line x1="${DIV_X}" y1="36" x2="${DIV_X}" y2="${WM_H - 36}"
        stroke="${GOLD}" stroke-width="2.5" stroke-linecap="round" opacity="0.75"/>
  <text x="${CIRC_X}" y="${BASELINE}"
        font-family="'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
        font-size="148" font-weight="300"
        fill="${GOLD}" letter-spacing="-2">Circle</text>
</svg>`;
const wmTextDarkBuf = await sharp(Buffer.from(wmTextDarkSvg)).ensureAlpha().png().toBuffer();
const wmLightMarkBuf = await resizeMark(lightMarkBuf, WM_MARK);

const wmDarkBuf = await sharp({
  create: { width: WM_W, height: WM_H, channels: 4, background: { r: NAVY_R, g: NAVY_G, b: NAVY_B, alpha: 255 } },
})
  .composite([
    { input: wmLightMarkBuf, top: WM_MARK_Y, left: WM_MARK_X },
    { input: wmTextDarkBuf, top: 0, left: 0, blend: "over" },
  ])
  .png()
  .toBuffer();

await writeFile(path.join(BRAND, "gk-circle-wordmark-dark.png"), wmDarkBuf);
const wmDarkWebp = await sharp(wmDarkBuf).webp({ quality: 92 }).toBuffer();
await writeFile(path.join(BRAND, "gk-circle-wordmark-dark.webp"), wmDarkWebp);
await writeFile(path.join(PUBLIC, "readme-logo.webp"), wmDarkWebp);
console.log("  ✓ gk-circle-wordmark-dark.png + .webp  1280×280  (navy bg, for dark mode)");
console.log("  ✓ readme-logo.webp  (copy of dark wordmark)");

// ── STEP 5: Open Graph banner (1200 × 630) ────────────────────────────────────
// Layout & Hierarchy (Pentagram / Stripe / Linear minimal aesthetic):
//   Left (x=60..420): Large Master Mark (centred at x=240, y=315)
//   Right (x=460..1140):
//     1. Brand Name: "GK Circle" (font-size 92)
//     2. Primary Headline: "One Circle. Every Aspirant. Unlimited Possibilities."
//     3. Supporting Line: "India's AI-Powered Learning Community for Competitive Exams"
//     4. Thin Gold Divider Line
//     5. Categories: "UPSC · State PCS · Banking · Railways · Defence · Law · Teaching & More"
//     6. Philosophy: "Learn Together · Compete Together · Grow Together"
// Background: Deep navy with ambient radial lighting & low-opacity orbital rings.
// ─────────────────────────────────────────────────────────────────────────────
console.log("\n▸ Step 5 — Refined Open Graph banner (1200×630)");

const OG_MARK_SIZE = 330;
const OG_MARK_X   = Math.round(235 - OG_MARK_SIZE / 2);   // 70
const OG_MARK_Y   = Math.round(315 - OG_MARK_SIZE / 2);   // 150

// Scale light mark for OG banner
const ogMarkBuf = await resizeMark(lightMarkBuf, OG_MARK_SIZE);

// Background + typography layer (pure SVG, vector precision)
const OG_TX  = 460;   // text column start
const ogSvg = `<svg xmlns="http://www.w3.org/2000/svg"
     width="1200" height="630" viewBox="0 0 1200 630">

  <!-- ─ Base Navy Canvas ─ -->
  <rect width="1200" height="630" fill="${NAVY}"/>

  <!-- ─ Soft Ambient Radial Glow Behind Mark ─ -->
  <radialGradient id="markGlow" cx="235" cy="315" r="320" gradientUnits="userSpaceOnUse">
    <stop offset="0%" stop-color="#2a3a8a" stop-opacity="0.55"/>
    <stop offset="60%" stop-color="#1f2c6c" stop-opacity="0.25"/>
    <stop offset="100%" stop-color="${NAVY}" stop-opacity="0"/>
  </radialGradient>
  <rect width="1200" height="630" fill="url(#markGlow)"/>

  <!-- ─ Subtle Orbital Geometric Rings ─ -->
  <!-- Left mark cluster -->
  <circle cx="235" cy="315" r="215" fill="none" stroke="${GOLD}" stroke-width="1.5" opacity="0.12"/>
  <circle cx="235" cy="315" r="280" fill="none" stroke="${WHITE}" stroke-width="1" opacity="0.05"/>
  <circle cx="235" cy="315" r="360" fill="none" stroke="${WHITE}" stroke-width="0.75" opacity="0.03"/>

  <!-- Top-Right Orbit Cluster -->
  <circle cx="1180" cy="-20" r="340" fill="none" stroke="${GOLD}" stroke-width="1.5" opacity="0.09"/>
  <circle cx="1180" cy="-20" r="520" fill="none" stroke="${GOLD}" stroke-width="1" opacity="0.05"/>
  <circle cx="1180" cy="-20" r="700" fill="none" stroke="${WHITE}" stroke-width="0.5" opacity="0.03"/>

  <!-- Bottom-Left Ambient Arc -->
  <circle cx="-50" cy="680" r="320" fill="none" stroke="${GOLD}" stroke-width="1" opacity="0.06"/>

  <!-- ─ Delicate Vertical Divider ─ -->
  <line x1="435" y1="90" x2="435" y2="540"
        stroke="${WHITE}" stroke-width="1" opacity="0.08"/>

  <!-- ─ Bottom Accent Gold Bar ─ -->
  <rect x="0" y="622" width="1200" height="8" fill="${GOLD}" opacity="0.9"/>

  <!-- ─ Typography Panel (Right Column — Strategic Brand Hierarchy) ─ -->

  <!-- Level 1: Brand Title ("GK Circle") -->
  <text x="${OG_TX}" y="152"
        font-family="'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
        font-size="80" font-weight="800"
        fill="${WHITE}" letter-spacing="-3">GK</text>
  <text x="${OG_TX + 122}" y="152"
        font-family="'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
        font-size="80" font-weight="300"
        fill="${GOLD}" letter-spacing="-1"> Circle</text>

  <!-- Level 2: Universal Brand Position ("India's AI-Powered Learning Community") -->
  <text x="${OG_TX}" y="210"
        font-family="'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
        font-size="22" font-weight="400"
        fill="${GOLD_L}" opacity="0.94" letter-spacing="0.4">India's AI-Powered Learning Community</text>

  <!-- Level 3: Brand Slogan ("One Circle. Every Aspirant. Unlimited Possibilities.") -->
  <text x="${OG_TX}" y="272"
        font-family="'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
        font-size="30" font-weight="700"
        fill="${WHITE}" letter-spacing="-0.3">One Circle. Every Aspirant. Unlimited Possibilities.</text>

  <!-- Thin Gold Rule (Delicate Separator) -->
  <rect x="${OG_TX}" y="318" width="690" height="1.5"
        fill="${GOLD}" opacity="0.35" rx="1"/>

  <!-- Level 4: Learning Domains (Modular & Future-Proof) -->
  <text x="${OG_TX}" y="372"
        font-family="'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
        font-size="14.5" font-weight="600"
        fill="${WHITE}" opacity="0.52" letter-spacing="3.5">LEARNING DOMAINS</text>

  <text x="${OG_TX}" y="412"
        font-family="'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
        font-size="17.5" font-weight="400"
        fill="${WHITE}" opacity="0.90" letter-spacing="0.25">Government  ·  Engineering  ·  Medical  ·  Management  ·  Law  ·  Teaching  ·  Professional  ·  &amp; More</text>

  <!-- Level 5: Community-First Philosophy ("CONNECT · LEARN · GROW · SUCCEED") -->
  <text x="${OG_TX}" y="496"
        font-family="'Segoe UI', 'Helvetica Neue', Arial, sans-serif"
        font-size="17.5" font-weight="500"
        fill="${GOLD}" opacity="0.88" letter-spacing="3">CONNECT  ·  LEARN  ·  GROW  ·  SUCCEED</text>

</svg>`;

// Rasterise bg + text SVG
const ogBgBuf = await sharp(Buffer.from(ogSvg))
  .flatten({ background: { r: NAVY_R, g: NAVY_G, b: NAVY_B } })
  .png()
  .toBuffer();

// Composite the light mark on top
const ogFinalBuf = await sharp(ogBgBuf)
  .composite([{ input: ogMarkBuf, top: OG_MARK_Y, left: OG_MARK_X, blend: "over" }])
  .png()
  .toBuffer();

// Verify dimensions
const ogMeta = await sharp(ogFinalBuf).metadata();
if (ogMeta.width !== 1200 || ogMeta.height !== 630) {
  throw new Error(`OG image is ${ogMeta.width}×${ogMeta.height}, expected 1200×630`);
}

await writeFile(path.join(PUBLIC, "og-image.png"), ogFinalBuf);
console.log(`  ✓ og-image.png  ${ogMeta.width}×${ogMeta.height}  (Refined hero banner)`);

// ── Done ──────────────────────────────────────────────────────────────────────
console.log(`
✅  GK Circle brand assets built successfully.
   Mark sources : ${BRAND}
   Public assets: ${PUBLIC}
`);
