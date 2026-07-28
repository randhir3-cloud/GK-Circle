import { Page } from "@playwright/test";

export interface ConsoleErrorRecord {
  type: string;
  text: string;
  location?: string;
}

export function attachConsoleMonitor(page: Page): ConsoleErrorRecord[] {
  const errors: ConsoleErrorRecord[] = [];
  page.on("console", (msg) => {
    if (msg.type() === "error") {
      errors.push({ type: "console.error", text: msg.text() });
    }
  });
  page.on("pageerror", (err) => {
    errors.push({ type: "pageerror", text: err.message });
  });
  return errors;
}
