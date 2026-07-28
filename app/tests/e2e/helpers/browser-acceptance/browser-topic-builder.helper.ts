import { Page } from "@playwright/test";
import { SubjectData } from "./browser-subject-builder.helper";
import { highlightForObservation } from "./browser-observation.helper";

export async function createTopicsUI(
  page: Page,
  subjects: SubjectData[]
): Promise<number> {
  let topicCount = 0;
  for (const subj of subjects) {
    const parentSelect = page.locator("#node-parent");
    const options = await parentSelect.locator("option").allInnerTexts();
    const parentOption = options.find(
      (opt) => opt.includes(subj.title) && opt.includes("SUBJECT")
    );

    if (!parentOption) {
      continue;
    }

    for (const topicTitle of subj.topics) {
      const titleInput = page.locator("#node-title");
      await highlightForObservation(titleInput);
      await titleInput.fill(topicTitle);

      const typeSelect = page.locator("#node-type");
      await typeSelect.selectOption("TOPIC");

      await parentSelect.selectOption({ label: parentOption });

      const addBtn = page.locator('[data-testid="create-node-button"]');
      await highlightForObservation(addBtn);
      await Promise.all([
        page.waitForResponse(
          (resp) =>
            resp.url().includes("/api/v1/courses/") &&
            resp.url().includes("/nodes") &&
            resp.request().method() === "POST" &&
            resp.status() === 201
        ),
        addBtn.click(),
      ]);
      topicCount++;
    }
  }
  return topicCount;
}
