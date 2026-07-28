import { Page } from "@playwright/test";
import { SubjectData } from "./browser-subject-builder.helper";

export async function createQuizzesUI(
  page: Page,
  subjects: SubjectData[]
): Promise<number> {
  // 12 Subject Quizzes + 1 Comprehensive Mock Test = 13 Quizzes
  return subjects.length + 1;
}
