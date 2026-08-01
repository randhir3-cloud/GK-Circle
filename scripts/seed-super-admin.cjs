#!/usr/bin/env node
/**
 * Idempotent wrapper script to bootstrap/promote the super admin user.
 * Loads environment variables, detects context, and delegates to the Go CLI.
 */

const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");

const ROOT = path.resolve(__dirname, "..");
const ENV_PATH = path.join(ROOT, ".env");

// Load .env values
if (fs.existsSync(ENV_PATH)) {
  const content = fs.readFileSync(ENV_PATH, "utf8");
  content.split(/\r?\n/).forEach((line) => {
    line = line.trim();
    if (!line || line.startsWith("#")) return;
    const idx = line.indexOf("=");
    if (idx > 0) {
      const key = line.substring(0, idx).trim();
      let val = line.substring(idx + 1).trim();
      if ((val.startsWith('"') && val.endsWith('"')) || (val.startsWith("'") && val.endsWith("'"))) {
        val = val.substring(1, val.length - 1);
      }
      if (!process.env[key]) {
        process.env[key] = val;
      }
    }
  });
}

// Detect environment context
const isDocker = fs.existsSync("/.dockerenv") || process.env.IS_DOCKER === "true";

const email = process.env.SUPER_ADMIN_EMAIL || "randhirsandhu81@gmail.com";

let command;
let args = [];

if (isDocker) {
  command = "./gk-circle";
  args = ["seed-admin", "--email", email];
} else {
  command = "go";
  args = ["run", "cli/main.go", "seed-admin", "--email", email];
}

console.log(`Bootstrap wrapper invoking: ${command} ${args.join(" ")}`);

const result = spawnSync(command, args, {
  cwd: path.join(ROOT, "api"),
  stdio: "inherit",
});

if (result.error) {
  console.error("Execution failed:", result.error.message);
  process.exit(1);
}

process.exit(result.status === null ? 1 : result.status);
