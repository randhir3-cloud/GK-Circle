import { mockNuxtImport } from "@nuxt/test-utils/runtime";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getQuizImportJobAPIError,
  useQuizImportJobsApi,
} from "@/composables/quiz_import_jobs";

const fetchMock = vi.hoisted(() => vi.fn());

mockNuxtImport("useRuntimeConfig", () => () => ({
  public: { apiUrl: "http://api.test/api/v1" },
}));
mockNuxtImport("useRequestHeaders", () => () => ({ cookie: "session=test" }));
vi.stubGlobal("$fetch", fetchMock);

beforeEach(() => {
  fetchMock.mockReset();
});

describe("quiz_import_jobs composable", () => {
  it("creates preview jobs with multipart CSV upload", async () => {
    fetchMock.mockResolvedValue({
      data: { id: "job-1", status: "PREVIEWED", valid_row_count: 2 },
    });

    const api = useQuizImportJobsApi();
    const file = new File(["csv"], "bank.csv", { type: "text/csv" });
    const job = await api.createPreviewJob("quiz-1", file);

    expect(job.id).toBe("job-1");
    expect(fetchMock).toHaveBeenCalledWith(
      "http://api.test/api/v1/quizzes/quiz-1/questions/import-jobs",
      expect.objectContaining({
        method: "POST",
        body: expect.any(FormData),
      })
    );
  });

  it("commits preview jobs without a request body", async () => {
    fetchMock.mockResolvedValue({
      data: {
        id: "job-1",
        status: "COMMITTED",
        commit_result: { committed_count: 2, question_ids: ["q-1", "q-2"] },
      },
    });

    const api = useQuizImportJobsApi();
    const job = await api.commitPreviewJob("quiz-1", "job-1");

    expect(job.status).toBe("COMMITTED");
    expect(fetchMock).toHaveBeenCalledWith(
      "http://api.test/api/v1/quizzes/quiz-1/questions/import-jobs/job-1/commit",
      expect.objectContaining({ method: "POST" })
    );
    const options = fetchMock.mock.calls[0][1];
    expect(options.body).toBeUndefined();
  });

  it("maps API errors for preview and commit flows", () => {
    expect(
      getQuizImportJobAPIError(
        { data: { message: "import job not found" } },
        "fallback"
      )
    ).toBe("import job not found");
  });
});
