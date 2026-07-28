import { Page } from "@playwright/test";
import { highlightForObservation } from "./browser-observation.helper";

export interface CourseData {
  id: string;
  title: string;
  description: string;
  category: string;
  language: string;
  status: string;
}

export async function createCourseUI(
  page: Page,
  baseUrl: string,
  runId: string
): Promise<CourseData> {
  const title = `UPSC Civil Services General Studies Foundation — ${runId}`;
  const description = `Comprehensive operator acceptance-test course covering UPSC Preliminary and Main General Studies domains — ${runId}`;

  await page.goto(`${baseUrl}/admin/courses`, { waitUntil: "networkidle" });

  const titleInput = page.locator("#new-course-title");
  await highlightForObservation(titleInput);
  await titleInput.fill(title);

  const createBtn = page.locator('[data-testid="create-course-button"]');
  await highlightForObservation(createBtn);
  await Promise.all([
    page.waitForResponse(
      (resp) =>
        resp.url().includes("/api/v1/courses") &&
        resp.request().method() === "POST" &&
        resp.status() === 201
    ),
    createBtn.click(),
  ]);

  const courseSelect = page.locator("#builder-course");
  await courseSelect.waitFor({ state: "visible" });

  // Select the newly created course
  const options = await courseSelect.locator("option").allInnerTexts();
  const targetOption = options.find((opt) => opt.includes(runId));
  if (!targetOption) {
    throw new Error(
      `Created course ${runId} did not appear in the visible course selector.`
    );
  }
  await courseSelect.selectOption({ label: targetOption });
  const courseId = await courseSelect.inputValue();
  if (!courseId) {
    throw new Error(
      `Created course ${runId} has no visible course identifier.`
    );
  }

  return {
    id: courseId,
    title,
    description,
    category: "UPSC Civil Services Examination",
    language: "English",
    status: "Published",
  };
}
