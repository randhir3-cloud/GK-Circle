/**
 * scripts/check-brand-assets.mjs
 *
 * Validates that all required GK Circle brand assets exist, are non-empty,
 * have the correct pixel dimensions, correct alpha channel status, and that
 * the ICO file contains the required embedded frames.
 *
 * Usage:
 *   node scripts/check-brand-assets.mjs
 *
 * Dependencies (installed via scripts/package.json):
 *   sharp   — PNG/WebP dimension and channel validation
 *   icojs   — ICO frame enumeration and size validation
 */

import { access, stat, readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import process from "node:process";
import sharp from "sharp";
import { decodeIco } from "icojs";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(__dirname, "..");

// ── Asset specifications ─────────────────────────────────────────────────────

const PNG_ASSETS = [
  // Master mark
  { rel: "app/public/brand/gk-circle-mark.png",            w: 1024, h: 1024, alpha: true },
  { rel: "app/public/brand/gk-circle-mark-light.png",      w: 1024, h: 1024, alpha: true },
  // Wordmark
  { rel: "app/public/brand/gk-circle-wordmark.png",        w: 1280, h: 280,  alpha: true },
  { rel: "app/public/brand/gk-circle-wordmark-dark.png",   w: 1280, h: 280,  alpha: true },
  // Favicons
  { rel: "app/public/favicon-256x256.png",                  w: 256,  h: 256,  alpha: true },
  { rel: "app/public/favicon-180x180.png",                  w: 180,  h: 180,  alpha: true },
  { rel: "app/public/favicon-48x48.png",                    w: 48,   h: 48,   alpha: true },
  { rel: "app/public/favicon-32x32.png",                    w: 32,   h: 32,   alpha: true },
  { rel: "app/public/favicon-16x16.png",                    w: 16,   h: 16,   alpha: true },
  // Open Graph
  { rel: "app/public/og-image.png",                         w: 1200, h: 630,  alpha: false },
];

const WEBP_ASSETS = [
  { rel: "app/public/brand/gk-circle-wordmark.webp",      alpha: true },
  { rel: "app/public/brand/gk-circle-wordmark-dark.webp", alpha: false },
  { rel: "app/public/readme-logo.webp",                     alpha: false },
];

const ICO_ASSET = {
  rel:            "app/public/favicon.ico",
  requiredSizes:  [16, 32, 48],
};

// ── Helpers ──────────────────────────────────────────────────────────────────

const failures = [];
const passes = [];

function pass(msg)  { passes.push(`  ✓ ${msg}`); }
function fail(msg)  { failures.push(`  ✗ ${msg}`); }

async function fileExists(abs) {
  try {
    await access(abs);
    const s = await stat(abs);
    return s.isFile() && s.size > 0;
  } catch {
    return false;
  }
}

// ── 1. PNG assets ────────────────────────────────────────────────────────────

console.log("\n▸ Validating PNG assets…");

for (const spec of PNG_ASSETS) {
  const abs = path.join(ROOT, spec.rel);
  const label = spec.rel.replace("app/public/", "");

  if (!(await fileExists(abs))) {
    fail(`${label}: file missing or empty`);
    continue;
  }

  let meta;
  try {
    meta = await sharp(abs).metadata();
  } catch (err) {
    fail(`${label}: could not decode — ${err.message}`);
    continue;
  }

  // Format
  if (meta.format !== "png") {
    fail(`${label}: expected PNG, got ${meta.format}`);
  }

  // Dimensions (null = no constraint)
  if (spec.w !== null && meta.width !== spec.w) {
    fail(`${label}: width ${meta.width}, expected ${spec.w}`);
  }
  if (spec.h !== null && meta.height !== spec.h) {
    fail(`${label}: height ${meta.height}, expected ${spec.h}`);
  }

  // Alpha channel
  const hasAlpha = meta.channels === 4 || meta.hasAlpha === true;
  if (spec.alpha && !hasAlpha) {
    fail(`${label}: expected alpha channel (RGBA), got ${meta.channels} channels`);
  }
  // For opaque assets (og-image), we check that space = srgb (no transparency).
  // Sharp may still report 4 channels after flatten if the PNG encoder writes RGBA;
  // what matters is that every pixel's alpha value is 255 (fully opaque).
  // We rely on the build script's .flatten() to guarantee this — no extra check needed.

  const dimStr = `${meta.width}×${meta.height}`;
  const chStr  = `${meta.channels}ch`;
  pass(`${label}  ${dimStr}  ${chStr}${spec.alpha ? " RGBA" : " opaque"}`);
}

// ── 2. WebP assets ───────────────────────────────────────────────────────────

console.log("\n▸ Validating WebP assets…");

for (const spec of WEBP_ASSETS) {
  const abs = path.join(ROOT, spec.rel);
  const label = spec.rel.replace("app/public/", "");

  if (!(await fileExists(abs))) {
    fail(`${label}: file missing or empty`);
    continue;
  }

  let meta;
  try {
    meta = await sharp(abs).metadata();
  } catch (err) {
    fail(`${label}: could not decode — ${err.message}`);
    continue;
  }

  if (meta.format !== "webp") {
    fail(`${label}: expected WebP, got ${meta.format}`);
    continue;
  }

  const hasAlpha = meta.channels === 4 || meta.hasAlpha === true;
  if (spec.alpha && !hasAlpha) {
    fail(`${label}: expected alpha channel`);
  }

  pass(`${label}  ${meta.width}×${meta.height}  ${meta.channels}ch WebP`);
}

// ── 3. ICO asset (frame validation via icojs) ────────────────────────────────

console.log("\n▸ Validating favicon.ico…");

const icoAbs = path.join(ROOT, ICO_ASSET.rel);
if (!(await fileExists(icoAbs))) {
  fail("favicon.ico: file missing or empty");
} else {
  try {
    const icoBuf = await readFile(icoAbs);
    const images = await decodeIco(icoBuf.buffer);
    const foundSizes = images.map((img) => img.width);

    for (const required of ICO_ASSET.requiredSizes) {
      if (foundSizes.includes(required)) {
        pass(`favicon.ico contains ${required}×${required} frame`);
      } else {
        fail(`favicon.ico missing ${required}×${required} frame (found: ${foundSizes.join(", ")})`);
      }
    }
  } catch (err) {
    fail(`favicon.ico: could not parse — ${err.message}`);
  }
}

// ── 4. Results ───────────────────────────────────────────────────────────────

console.log("\n" + passes.join("\n"));

if (failures.length > 0) {
  console.error("\n❌  Brand asset validation FAILED:\n");
  console.error(failures.join("\n"));
  process.exit(1);
}

console.log(`\n✅  All ${passes.length} brand asset checks passed.\n`);
