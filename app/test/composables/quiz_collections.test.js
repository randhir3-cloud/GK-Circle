import { mockNuxtImport } from "@nuxt/test-utils/runtime";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getQuizCollectionAPIError,
  useQuizCollectionsApi,
} from "@/composables/quiz_collections";

const fetchMock = vi.hoisted(() => vi.fn());

mockNuxtImport("useRuntimeConfig", () => () => ({
  public: { apiUrl: "http://api.test/api/v1" },
}));
mockNuxtImport("useRequestHeaders", () => () => ({ cookie: "session=test" }));
vi.stubGlobal("$fetch", fetchMock);

beforeEach(() => {
  fetchMock.mockReset();
});

describe("quiz_collections composable", () => {
  it("lists collections for a quiz", async () => {
    fetchMock.mockResolvedValue({
      data: [{ id: "c-1", title: "Section A", kind: "STATIC" }],
    });

    const api = useQuizCollectionsApi();
    const rows = await api.listCollections("quiz-1");

    expect(rows).toHaveLength(1);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://api.test/api/v1/quizzes/quiz-1/collections",
      expect.objectContaining({
        credentials: "include",
      })
    );
  });

  it("creates STATIC and DYNAMIC collections", async () => {
    fetchMock.mockResolvedValue({
      data: { id: "c-2", kind: "DYNAMIC", title: "Pool" },
    });

    const api = useQuizCollectionsApi();
    await api.createCollection("quiz-1", {
      title: "Pool",
      kind: "DYNAMIC",
      position: 0,
      filter: { subject: "History" },
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://api.test/api/v1/quizzes/quiz-1/collections",
      expect.objectContaining({
        method: "POST",
        body: expect.objectContaining({
          kind: "DYNAMIC",
          filter: { subject: "History" },
        }),
      })
    );
  });

  it("replaces STATIC members and resolves collections", async () => {
    fetchMock
      .mockResolvedValueOnce({
        data: { id: "c-1", members: [{ question_id: "q-1", position: 0 }] },
      })
      .mockResolvedValueOnce({
        data: {
          collection_id: "c-1",
          kind: "STATIC",
          resolution_status: "RESOLVED",
          question_ids: ["q-1"],
        },
      });

    const api = useQuizCollectionsApi();
    await api.replaceMembers("quiz-1", "c-1", ["q-1"]);
    await api.resolveCollection("quiz-1", "c-1");

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "http://api.test/api/v1/quizzes/quiz-1/collections/c-1/members",
      expect.objectContaining({
        method: "PUT",
        body: { question_ids: ["q-1"] },
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "http://api.test/api/v1/quizzes/quiz-1/collections/c-1/resolve",
      expect.anything()
    );
  });

  it("maps API errors", () => {
    expect(
      getQuizCollectionAPIError(
        { data: { message: "question collection not found" } },
        "fallback"
      )
    ).toBe("question collection not found");
  });
});
