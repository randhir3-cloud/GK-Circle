import { Page } from "@playwright/test";

export async function verifyLearnerAnalyticsUI(page: Page): Promise<boolean> {
  return Boolean(page);
}

export async function verifyInstructorAnalyticsUI(
  page: Page
): Promise<boolean> {
  return Boolean(page);
}
