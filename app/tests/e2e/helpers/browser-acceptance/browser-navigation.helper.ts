import { Page } from "@playwright/test";

export async function navigateToAdminDashboard(
  page: Page,
  baseUrl: string
): Promise<void> {
  await page
    .goto(`${baseUrl}/instructor/analytics`, { waitUntil: "networkidle" })
    .catch(() => null);
}

export async function navigateToCourseManagement(
  page: Page,
  baseUrl: string
): Promise<void> {
  await page
    .goto(`${baseUrl}/instructor/courses`, { waitUntil: "networkidle" })
    .catch(() => null);
}

export async function navigateToReports(
  page: Page,
  baseUrl: string
): Promise<void> {
  await page
    .goto(`${baseUrl}/instructor/reports`, { waitUntil: "networkidle" })
    .catch(() => null);
}
