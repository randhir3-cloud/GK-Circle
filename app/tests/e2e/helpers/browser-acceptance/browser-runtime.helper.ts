import { existsSync, readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const REQUIRED_RUNTIME_KEYS = [
  "E2E_BASE_URL",
  "E2E_ADMIN_EMAIL",
  "E2E_ADMIN_PASSWORD",
  "E2E_LEARNER_EMAIL",
  "E2E_LEARNER_PASSWORD",
  "E2E_MAILPIT_URL",
] as const;

type RequiredRuntimeKey = (typeof REQUIRED_RUNTIME_KEYS)[number];

export interface HumanTestRuntime {
  baseUrl: string;
  adminEmail: string;
  adminPassword: string;
  learnerEmail: string;
  learnerPassword: string;
  mailpitUrl: string;
}

function parseEnvironmentFile(filePath: string): Record<string, string> {
  if (!existsSync(filePath)) {
    return {};
  }

  const values: Record<string, string> = {};
  for (const rawLine of readFileSync(filePath, "utf8").split(/\r?\n/u)) {
    const line = rawLine.trim();
    if (!line || line.startsWith("#")) {
      continue;
    }

    const separator = line.indexOf("=");
    if (separator < 1) {
      continue;
    }

    const key = line.slice(0, separator).trim();
    let value = line.slice(separator + 1).trim();
    if (
      value.length >= 2 &&
      ((value.startsWith('"') && value.endsWith('"')) ||
        (value.startsWith("'") && value.endsWith("'")))
    ) {
      value = value.slice(1, -1);
    }
    values[key] = value;
  }
  return values;
}

function setWhenMissing(key: string, value: string | undefined): void {
  if (!process.env[key]?.trim() && value?.trim()) {
    process.env[key] = value.trim();
  }
}

export function loadHumanTestRuntimeEnvironment(): void {
  const helperDirectory = dirname(fileURLToPath(import.meta.url));
  const repositoryRoot = resolve(helperDirectory, "../../../../..");
  const localValues = parseEnvironmentFile(
    resolve(repositoryRoot, ".env.e2e.local")
  );

  for (const [key, value] of Object.entries(localValues)) {
    setWhenMissing(key, value);
  }

  setWhenMissing("E2E_BASE_URL", process.env.PLAYWRIGHT_BASE_URL);
  setWhenMissing("E2E_ADMIN_EMAIL", process.env.E2E_CREATOR_EMAIL);
  setWhenMissing("E2E_ADMIN_PASSWORD", process.env.E2E_TEST_PASSWORD);
  setWhenMissing("E2E_LEARNER_EMAIL", process.env.E2E_STUDENT_EMAIL);
  setWhenMissing("E2E_LEARNER_PASSWORD", process.env.E2E_TEST_PASSWORD);

  // Mailpit is a required real-system dependency in the local Compose stack.
  setWhenMissing("E2E_MAILPIT_URL", "http://localhost:8025");
}

function requireRuntimeValue(key: RequiredRuntimeKey): string {
  const value = process.env[key]?.trim();
  if (!value) {
    throw new Error(
      `Human-observed certification cannot start: required ${key} is missing.`
    );
  }
  return value;
}

export function getHumanTestRuntime(): HumanTestRuntime {
  const values = Object.fromEntries(
    REQUIRED_RUNTIME_KEYS.map((key) => [key, requireRuntimeValue(key)])
  ) as Record<RequiredRuntimeKey, string>;

  return {
    baseUrl: values.E2E_BASE_URL,
    adminEmail: values.E2E_ADMIN_EMAIL,
    adminPassword: values.E2E_ADMIN_PASSWORD,
    learnerEmail: values.E2E_LEARNER_EMAIL,
    learnerPassword: values.E2E_LEARNER_PASSWORD,
    mailpitUrl: values.E2E_MAILPIT_URL,
  };
}
