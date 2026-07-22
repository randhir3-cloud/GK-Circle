#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();

const retiredPatterns = [
  { name: 'retired aggregate name', regex: /\bProducts?\b/ },
  { name: 'retired lowercase aggregate name', regex: /\bproducts?\b/ },
  { name: 'retired permission namespace', regex: /\bproduct:[A-Za-z0-9_.-]+\b/ },
  { name: 'retired service symbol', regex: /\bProducts?Service\b/ },
  { name: 'retired controller symbol', regex: /\bProducts?Controller\b/ },
  { name: 'retired module symbol', regex: /\bProducts?Module\b/ },
  { name: 'retired DTO symbol', regex: /\bProducts?(?:Dto|DTO)\b/ },
  { name: 'retired Prisma delegate', regex: /\bprisma\.product\b/ },
  { name: 'retired camel-case id', regex: /\bproductId\b/ },
  { name: 'retired snake-case id', regex: /\bproduct_id\b/ },
];

const activeRoots = [
  'backend/src',
  'backend/test',
  'backend/prisma/schema.prisma',
  'backend/prisma/seed.ts',
  'backend/prisma/seed-admin.ts',
  'backend/prisma/seed-users.ts',
  'frontend/src',
  'frontend/playwright',
  'frontend/scripts',
  'scripts',
  '.github/workflows',
  'docs/standards',
];

const excludedParts = new Set([
  '.git',
  'node_modules',
  'dist',
  'build',
  '.next',
  'coverage',
  'playwright-report',
  'test-results',
  '.cache',
]);

const excludedFiles = new Set([
  path.normalize('scripts/check-retired-terminology.mjs'),
  // Historical ADR-019 certification intentionally asserts old route families
  // return the platform's normal not-found behavior.
  path.normalize('frontend/playwright/adr019-course-rename-certification.spec.ts'),
  // PHASE-C-012 final certification keeps the same negative-route assertion
  // so retired route families cannot reappear as active contracts.
  path.normalize('frontend/playwright/phase-c-012-certification.spec.ts'),
]);

const textExtensions = new Set([
  '.cjs',
  '.css',
  '.cts',
  '.env',
  '.html',
  '.js',
  '.json',
  '.jsx',
  '.md',
  '.mjs',
  '.prisma',
  '.ps1',
  '.sh',
  '.sql',
  '.ts',
  '.tsx',
  '.txt',
  '.yaml',
  '.yml',
]);

function exists(relativePath) {
  return fs.existsSync(path.join(root, relativePath));
}

function shouldSkip(relativePath) {
  const normalized = path.normalize(relativePath);
  if (excludedFiles.has(normalized)) return true;

  const parts = normalized.split(path.sep);
  return parts.some((part) => excludedParts.has(part));
}

function collectFiles(relativePath, output) {
  if (!exists(relativePath) || shouldSkip(relativePath)) return;

  const absolutePath = path.join(root, relativePath);
  const stat = fs.statSync(absolutePath);

  if (stat.isDirectory()) {
    for (const entry of fs.readdirSync(absolutePath)) {
      collectFiles(path.join(relativePath, entry), output);
    }
    return;
  }

  if (!stat.isFile()) return;

  const ext = path.extname(relativePath);
  if (!textExtensions.has(ext)) return;

  output.push(relativePath);
}

function findMatches(file) {
  const text = fs.readFileSync(path.join(root, file), 'utf8');
  const lines = text.split(/\r?\n/);
  const matches = [];

  lines.forEach((line, index) => {
    for (const pattern of retiredPatterns) {
      if (pattern.regex.test(line)) {
        matches.push({
          file,
          line: index + 1,
          type: pattern.name,
          text: line.trim(),
        });
      }
    }
  });

  return matches;
}

const files = [];
for (const activeRoot of activeRoots) {
  collectFiles(activeRoot, files);
}

const matches = files.flatMap(findMatches);

if (matches.length > 0) {
  console.error('Retired educational-domain terminology found in active files.');
  console.error('Historical ADRs, immutable migrations, archived evidence, dependencies, and build output are intentionally outside this active-source scan.');
  console.error('');

  for (const match of matches) {
    console.error(`${match.file}:${match.line} [${match.type}] ${match.text}`);
  }

  process.exit(1);
}

console.log(`Retired terminology check passed (${files.length} active files scanned).`);
