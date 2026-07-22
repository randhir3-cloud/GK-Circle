import { readFileSync, readdirSync, statSync } from "node:fs";
import { extname, join, normalize, relative } from "node:path";

const root = normalize(join(import.meta.dirname, ".."));
const roots = ["api", "app", "scripts", "deploy"];
const rootFiles = [
  "docker-compose.yaml",
  "docker-compose.override.yml",
  "docker-compose.nuc.yml",
  ".env.example",
];
const extensions = new Set([
  ".go", ".js", ".mjs", ".ts", ".vue", ".json", ".yaml", ".yml",
  ".toml", ".ps1", ".sh", ".conf", ".md",
]);
const excludedParts = new Set(["node_modules", ".nuxt", ".output", "coverage", "go.sum"]);
const forbidden = [
  { label: "legacy product name", pattern: /\bjovvix\b/i },
  { label: "wrong backend framework", pattern: /\bNestJS\b/i },
  { label: "wrong frontend framework", pattern: /\bNext\.js\b/i },
  { label: "wrong ORM", pattern: /\bPrisma\b/i },
];

function collect(path) {
  if (excludedParts.has(path.split(/[\\/]/).at(-1))) return [];
  if (statSync(path).isDirectory()) {
    return readdirSync(path).flatMap((entry) => collect(join(path, entry)));
  }
  return extensions.has(extname(path)) || path.endsWith("Dockerfile") ? [path] : [];
}

const files = [
  ...roots.flatMap((entry) => collect(join(root, entry))),
  ...rootFiles.map((entry) => join(root, entry)),
];
const findings = [];

for (const file of files) {
  const display = relative(root, file);
  const lines = readFileSync(file, "utf8").split(/\r?\n/);
  lines.forEach((line, index) => {
    for (const rule of forbidden) {
      if (rule.pattern.test(line)) findings.push(`${display}:${index + 1} ${rule.label}`);
    }
  });
}

if (findings.length) {
  console.error(findings.join("\n"));
  process.exit(1);
}

console.log("Project identity check passed.");
