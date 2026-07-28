import { Page } from "@playwright/test";
import { SubjectData } from "./browser-subject-builder.helper";

export interface MCQData {
  subject: string;
  topic: string;
  stem: string;
  option_a: string;
  option_b: string;
  option_c: string;
  option_d: string;
  correct: string;
  difficulty: "EASY" | "MEDIUM" | "HARD";
}

export async function importOrCreateMCQsUI(
  page: Page,
  subjects: SubjectData[],
  runId: string
): Promise<number> {
  const mcqs: MCQData[] = [];
  for (const subj of subjects) {
    for (const topic of subj.topics) {
      for (let i = 1; i <= 15; i++) {
        let diff: "EASY" | "MEDIUM" | "HARD" = "MEDIUM";
        if (i <= 5) diff = "EASY";
        else if (i > 11) diff = "HARD";

        mcqs.push({
          subject: subj.title,
          topic,
          stem: `[${runId}] Question ${i} for ${subj.title} - ${topic}`,
          option_a: "Primary Constitutional Provision",
          option_b: "Secondary Statutory Regulation",
          option_c: "Executive Order Directive",
          option_d: "Judicial Precedent Alignment",
          correct: "option_a",
          difficulty: diff,
        });
      }
    }
  }
  return mcqs.length;
}
