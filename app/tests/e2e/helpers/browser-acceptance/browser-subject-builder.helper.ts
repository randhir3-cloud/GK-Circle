import { Page } from "@playwright/test";
import { highlightForObservation } from "./browser-observation.helper";

export interface SubjectData {
  title: string;
  description: string;
  topics: string[];
}

export async function createSubjectsUI(
  page: Page,
  courseId: string,
  subjects: SubjectData[]
): Promise<number> {
  let createdCount = 0;
  for (const subj of subjects) {
    const titleInput = page.locator("#node-title");
    await highlightForObservation(titleInput);
    await titleInput.fill(subj.title);

    const typeSelect = page.locator("#node-type");
    await typeSelect.selectOption("SUBJECT");

    const parentSelect = page.locator("#node-parent");
    await parentSelect.selectOption("");

    const addBtn = page.locator('[data-testid="create-node-button"]');
    await highlightForObservation(addBtn);
    await Promise.all([
      page.waitForResponse(
        (resp) =>
          resp.url().includes(`/api/v1/courses/${courseId}/nodes`) &&
          resp.request().method() === "POST" &&
          resp.status() === 201
      ),
      addBtn.click(),
    ]);
    createdCount++;
  }
  return createdCount;
}
