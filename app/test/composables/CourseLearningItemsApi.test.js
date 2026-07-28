import { mockNuxtImport } from "@nuxt/test-utils/runtime";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getCourseAdminAPIError,
  useCourseLearningItemsApi,
} from "@/composables/course_learning_items";

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

describe("admin Course LearningItem transport", () => {
  it("uses the existing Course, node, and CRUD endpoints with encoded IDs", async () => {
    const api = useCourseLearningItemsApi();
    const createBody = {
      title: "New item",
      item_type: "ARTICLE",
      publish_state: "DRAFT",
    };
    const updateBody = {
      title: "Updated",
      item_type: "PDF",
      publish_state: "PUBLISHED",
      description: null,
    };

    await api.listCourses();
    await api.createCourse({ title: "PCS" });
    await api.updateCourse("course id", { status: "PUBLISHED" });
    await api.getTree("course id");
    await api.createNode("course id", {
      title: "History",
      node_type: "SUBJECT",
    });
    await api.listRootNodes("course id");
    await api.listChildren("course id", "node/id");
    await api.listItems("course id", "node/id");
    await api.createItem("course id", "node/id", createBody);
    await api.updateItem("course id", "node/id", "item id", updateBody);
    await api.deleteItem("course id", "node/id", "item id");

    const base = "http://api.test/api/v1/admin/courses";
    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      base,
      expect.objectContaining({ credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      base,
      expect.objectContaining({ method: "POST", body: { title: "PCS" } })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      `${base}/course%20id`,
      expect.objectContaining({
        method: "PATCH",
        body: { status: "PUBLISHED" },
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      4,
      `${base}/course%20id/nodes/tree`,
      expect.objectContaining({ credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      5,
      `${base}/course%20id/nodes`,
      expect.objectContaining({
        method: "POST",
        body: { title: "History", node_type: "SUBJECT" },
      })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      6,
      `${base}/course%20id/nodes`,
      expect.objectContaining({ credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      7,
      `${base}/course%20id/nodes/node%2Fid/children`,
      expect.objectContaining({ credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      8,
      `${base}/course%20id/nodes/node%2Fid/learning-items`,
      expect.objectContaining({ credentials: "include" })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      9,
      `${base}/course%20id/nodes/node%2Fid/learning-items`,
      expect.objectContaining({ method: "POST", body: createBody })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      10,
      `${base}/course%20id/nodes/node%2Fid/learning-items/item%20id`,
      expect.objectContaining({ method: "PATCH", body: updateBody })
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      11,
      `${base}/course%20id/nodes/node%2Fid/learning-items/item%20id`,
      expect.objectContaining({ method: "DELETE" })
    );
  });

  it("unwraps JSend data without transforming or reordering it", async () => {
    const payload = Object.freeze([
      Object.freeze({ id: "item-z", position: 9 }),
      Object.freeze({ id: "item-a", position: 1 }),
    ]);
    fetchMock.mockResolvedValueOnce({ status: "success", data: payload });

    const result = await useCourseLearningItemsApi().listItems(
      "course",
      "node"
    );

    expect(result).toBe(payload);
  });

  it("extracts fail, error, ordinary Error, and fallback messages", () => {
    expect(
      getCourseAdminAPIError(
        { data: { data: "validation failed" } },
        "fallback"
      )
    ).toBe("validation failed");
    expect(
      getCourseAdminAPIError({ data: { message: "server failed" } }, "fallback")
    ).toBe("server failed");
    expect(
      getCourseAdminAPIError(new Error("network failed"), "fallback")
    ).toBe("network failed");
    expect(getCourseAdminAPIError({}, "fallback")).toBe("fallback");
  });
});
