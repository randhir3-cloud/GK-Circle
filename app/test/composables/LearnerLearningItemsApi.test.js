import { mockNuxtImport } from "@nuxt/test-utils/runtime";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getLearnerLearningItemAPIError,
  useLearnerLearningItemsApi,
} from "@/composables/learner_learning_items";

const fetchMock = vi.hoisted(() => vi.fn());

mockNuxtImport("useRuntimeConfig", () => () => ({
  public: { apiUrl: "http://api.test/api/v1" },
}));
mockNuxtImport("useRequestHeaders", () => () => ({ cookie: "session=test" }));
vi.stubGlobal("$fetch", fetchMock);

beforeEach(() => {
  fetchMock.mockReset();
  fetchMock.mockResolvedValue({ data: [] });
});

describe("learner LearningItem transport", () => {
  it("uses learner list, detail, and enrollment endpoints", async () => {
    const api = useLearnerLearningItemsApi();

    await api.listPublishedCourses();
    await api.getPublishedCourse("course id");
    await api.getPublishedOutline("course id");
    await api.listItems("course id", "node/id");
    await api.getItem("course id", "node/id", "item id");
    await api.getEnrollment("course id");
    await api.enroll("course id");
    await api.unenroll("course id");

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "http://api.test/api/v1/learner/courses",
      expect.objectContaining({ credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "http://api.test/api/v1/learner/courses/course%20id",
      expect.objectContaining({ credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "http://api.test/api/v1/learner/courses/course%20id/nodes/tree",
      expect.objectContaining({ credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      4,
      "http://api.test/api/v1/learner/courses/course%20id/nodes/node%2Fid/learning-items",
      expect.objectContaining({ credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      5,
      "http://api.test/api/v1/learner/courses/course%20id/nodes/node%2Fid/learning-items/item%20id",
      expect.objectContaining({ credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      6,
      "http://api.test/api/v1/learner/courses/course%20id/enrollment",
      expect.objectContaining({ credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      7,
      "http://api.test/api/v1/learner/courses/course%20id/enrollment",
      expect.objectContaining({ method: "POST", credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      8,
      "http://api.test/api/v1/learner/courses/course%20id/enrollment",
      expect.objectContaining({ method: "DELETE", credentials: "include" })
    );
  });

  it("unwraps JSend data without filtering or transforming it", async () => {
    const payload = Object.freeze([
      Object.freeze({ id: "second", publish_state: "PUBLISHED" }),
      Object.freeze({ id: "first", publish_state: "DRAFT" }),
    ]);
    fetchMock.mockResolvedValueOnce({ data: payload });

    const result = await useLearnerLearningItemsApi().listItems(
      "course",
      "node"
    );

    expect(result).toBe(payload);
  });

  it("extracts the established API error shapes", () => {
    expect(
      getLearnerLearningItemAPIError(
        { data: { data: "unauthenticated" } },
        "fallback"
      )
    ).toBe("unauthenticated");
    expect(
      getLearnerLearningItemAPIError(
        { data: { message: "request failed" } },
        "fallback"
      )
    ).toBe("request failed");
  });
});
