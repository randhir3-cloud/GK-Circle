import { flushPromises } from "@vue/test-utils";
import { mountSuspended, mockNuxtImport } from "@nuxt/test-utils/runtime";
import { nextTick, reactive } from "vue";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import LearningItemDetailPage from "@/pages/courses/[course_id]/nodes/[node_id]/learning-items/[item_id].vue";
import LearningItemListPage from "@/pages/courses/[course_id]/nodes/[node_id]/learning-items/index.vue";

const testState = vi.hoisted(() => ({
  route: null,
  api: {
    listItems: vi.fn(),
    getItem: vi.fn(),
    enroll: vi.fn(),
  },
}));

mockNuxtImport("useRoute", () => () => testState.route);

vi.mock("@/composables/learner_learning_items", () => ({
  useLearnerLearningItemsApi: () => testState.api,
  getLearnerLearningItemAPIError: (error, fallback) =>
    error?.data?.data || error?.data?.message || error?.message || fallback,
  isCourseEnrollmentRequiredError: (errorOrMessage) => {
    const message =
      typeof errorOrMessage === "string"
        ? errorOrMessage
        : errorOrMessage?.data?.data ||
          errorOrMessage?.data?.message ||
          errorOrMessage?.message ||
          "";
    return message === "course enrollment required";
  },
}));

const NuxtLinkStub = {
  props: ["to"],
  template: '<a :href="to"><slot /></a>',
};

const RendererStub = {
  props: ["metadata"],
  template: '<div data-testid="renderer-stub">{{ metadata?.marker }}</div>',
};

const deferred = () => {
  let resolve;
  let reject;
  const promise = new Promise((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
};

const mountedWrappers = [];

const trackMount = async (mountPromise) => {
  const wrapper = await mountPromise;
  mountedWrappers.push(wrapper);
  return wrapper;
};

const mountList = () =>
  trackMount(
    mountSuspended(LearningItemListPage, {
      global: { stubs: { NuxtLink: NuxtLinkStub } },
    })
  );

const mountDetail = () =>
  trackMount(
    mountSuspended(LearningItemDetailPage, {
      global: {
        stubs: {
          NuxtLink: NuxtLinkStub,
          LearningItemRenderer: RendererStub,
        },
      },
    })
  );

afterEach(() => {
  for (const wrapper of mountedWrappers) wrapper.unmount();
  mountedWrappers.length = 0;
});

beforeEach(() => {
  vi.clearAllMocks();
  testState.route = reactive({
    params: {
      course_id: "course-1",
      node_id: "node-1",
      item_id: "item-1",
    },
  });
  testState.api.listItems.mockResolvedValue([]);
  testState.api.enroll.mockResolvedValue({
    course_id: "course-1",
    enrolled: true,
  });
  testState.api.getItem.mockResolvedValue({
    learning_item: {
      id: "item-1",
      title: "First item",
      item_type: "ARTICLE",
      description: null,
      metadata: { marker: "first", version: 1, blocks: [] },
      publish_state: "PUBLISHED",
    },
    previous: null,
    next: null,
  });
});

describe("learner LearningItem list page", () => {
  it("renders a frozen list in the exact API order without sorting", async () => {
    const items = Object.freeze([
      Object.freeze({ id: "z", title: "Position 9", item_type: "ARTICLE" }),
      Object.freeze({ id: "a", title: "Position 1", item_type: "VIDEO" }),
      Object.freeze({ id: "m", title: "Position 4", item_type: "PDF" }),
    ]);
    testState.api.listItems.mockResolvedValue(items);

    const wrapper = await mountList();
    await flushPromises();

    expect(testState.api.listItems).toHaveBeenCalledWith("course-1", "node-1");
    expect(
      wrapper
        .findAll('[data-testid^="learning-item-"]')
        .map((entry) => entry.attributes("data-testid"))
    ).toEqual([
      "learning-item-list",
      "learning-item-z",
      "learning-item-a",
      "learning-item-m",
    ]);
  });

  it("shows an empty state only after a successful empty response", async () => {
    const wrapper = await mountList();
    await flushPromises();

    expect(wrapper.get("main").classes()).toEqual(
      expect.arrayContaining(["w-full", "flex-1"])
    );
    expect(wrapper.get("main > div").classes()).toContain("w-full");
    expect(wrapper.get('[data-testid="items-empty"]').text()).toBe(
      "No Learning Items available."
    );
  });

  it("displays unauthorized errors through the established API contract", async () => {
    testState.api.listItems.mockRejectedValue({
      data: { data: "unauthenticated" },
    });

    const wrapper = await mountList();
    await flushPromises();

    expect(wrapper.get('[data-testid="items-error"]').text()).toBe(
      "unauthenticated"
    );
  });

  it("enrolls through the real transport contract and reloads the list", async () => {
    testState.api.listItems
      .mockRejectedValueOnce({ data: { data: "course enrollment required" } })
      .mockResolvedValueOnce([
        { id: "published", title: "Published item", item_type: "ARTICLE" },
      ]);

    const wrapper = await mountList();
    await flushPromises();

    expect(wrapper.get('[data-testid="items-error"]').text()).toContain(
      "course enrollment required"
    );
    await wrapper.get('[data-testid="enroll-course"]').trigger("click");
    await flushPromises();

    expect(testState.api.enroll).toHaveBeenCalledOnce();
    expect(testState.api.enroll).toHaveBeenCalledWith("course-1");
    expect(testState.api.listItems).toHaveBeenCalledTimes(2);
    expect(wrapper.text()).toContain("Published item");
    expect(wrapper.find('[data-testid="items-error"]').exists()).toBe(false);
  });

  it("keeps the enrollment failure visible without showing false success", async () => {
    testState.api.listItems.mockRejectedValueOnce({
      data: { data: "course enrollment required" },
    });
    testState.api.enroll.mockRejectedValueOnce({
      data: { data: "enrollment unavailable" },
    });

    const wrapper = await mountList();
    await flushPromises();
    await wrapper.get('[data-testid="enroll-course"]').trigger("click");
    await flushPromises();

    expect(wrapper.get('[data-testid="items-error"]').text()).toBe(
      "enrollment unavailable"
    );
    expect(wrapper.find('[data-testid="items-empty"]').exists()).toBe(false);
    expect(wrapper.find('[data-testid^="learning-item-"]').exists()).toBe(
      false
    );
  });

  it("discards stale list responses when route parameters change", async () => {
    const oldRequest = deferred();
    testState.api.listItems
      .mockReturnValueOnce(oldRequest.promise)
      .mockResolvedValueOnce([
        { id: "current", title: "Current node", item_type: "ARTICLE" },
      ]);

    const wrapper = await mountList();
    testState.route.params.node_id = "node-2";
    await nextTick();
    await flushPromises();

    expect(wrapper.text()).toContain("Current node");
    oldRequest.resolve([
      { id: "stale", title: "Stale node", item_type: "ARTICLE" },
    ]);
    await flushPromises();

    expect(wrapper.text()).toContain("Current node");
    expect(wrapper.text()).not.toContain("Stale node");
  });
});

describe("learner LearningItem detail page", () => {
  it("uses API-provided previous and next IDs without deriving adjacency", async () => {
    testState.api.getItem.mockResolvedValue({
      learning_item: {
        id: "item-1",
        title: "Current",
        item_type: "ARTICLE",
        metadata: { marker: "current", version: 1, blocks: [] },
      },
      previous: { id: "api-prev-9", title: "Previous from API" },
      next: { id: "api-next-2", title: "Next from API" },
    });

    const wrapper = await mountDetail();
    await flushPromises();

    expect(testState.api.getItem).toHaveBeenCalledWith(
      "course-1",
      "node-1",
      "item-1"
    );
    expect(
      wrapper.get('[data-testid="previous-item"]').attributes("href")
    ).toBe("/courses/course-1/nodes/node-1/learning-items/api-prev-9");
    expect(wrapper.get('[data-testid="next-item"]').attributes("href")).toBe(
      "/courses/course-1/nodes/node-1/learning-items/api-next-2"
    );
  });

  it("renders no fake navigation links for null API values", async () => {
    const wrapper = await mountDetail();
    await flushPromises();

    expect(wrapper.find('[data-testid="previous-item"]').exists()).toBe(false);
    expect(wrapper.find('[data-testid="next-item"]').exists()).toBe(false);
    expect(
      wrapper.find('nav[aria-label="Learning Item navigation"]').exists()
    ).toBe(false);
  });

  it("clears stale visible content immediately on a route change", async () => {
    const wrapper = await mountDetail();
    await flushPromises();
    expect(wrapper.text()).toContain("First item");

    const pending = deferred();
    testState.api.getItem.mockReturnValueOnce(pending.promise);
    testState.route.params.item_id = "item-2";
    await nextTick();

    expect(wrapper.text()).not.toContain("First item");
    expect(wrapper.get('[data-testid="item-loading"]').exists()).toBe(true);

    pending.resolve({
      learning_item: {
        id: "item-2",
        title: "Second item",
        item_type: "ARTICLE",
        metadata: { marker: "second", version: 1, blocks: [] },
      },
      previous: null,
      next: null,
    });
    await flushPromises();
    expect(wrapper.text()).toContain("Second item");
  });

  it("prevents an older detail response from overwriting the current route", async () => {
    const oldRequest = deferred();
    testState.api.getItem
      .mockReturnValueOnce(oldRequest.promise)
      .mockResolvedValueOnce({
        learning_item: {
          id: "item-2",
          title: "Current detail",
          item_type: "ARTICLE",
          metadata: { marker: "current", version: 1, blocks: [] },
        },
        previous: null,
        next: null,
      });

    const wrapper = await mountDetail();
    testState.route.params.item_id = "item-2";
    await nextTick();
    await flushPromises();
    expect(wrapper.text()).toContain("Current detail");

    oldRequest.resolve({
      learning_item: {
        id: "item-1",
        title: "Stale detail",
        item_type: "ARTICLE",
        metadata: { marker: "stale", version: 1, blocks: [] },
      },
      previous: null,
      next: null,
    });
    await flushPromises();

    expect(wrapper.text()).toContain("Current detail");
    expect(wrapper.text()).not.toContain("Stale detail");
  });

  it("displays detail API failures without rendering stale content", async () => {
    testState.api.getItem.mockRejectedValue({
      data: { message: "Learning Item not found" },
    });

    const wrapper = await mountDetail();
    await flushPromises();

    expect(wrapper.get('[data-testid="item-error"]').text()).toBe(
      "Learning Item not found"
    );
    expect(wrapper.find('[data-testid="renderer-stub"]').exists()).toBe(false);
  });

  it("keeps a responsive single-column-to-two-column navigation contract", async () => {
    testState.api.getItem.mockResolvedValue({
      learning_item: {
        id: "item-1",
        title: "Responsive item",
        item_type: "ARTICLE",
        metadata: { marker: "responsive", version: 1, blocks: [] },
      },
      previous: { id: "previous", title: "Previous" },
      next: { id: "next", title: "Next" },
    });

    const wrapper = await mountDetail();
    await flushPromises();
    const navigation = wrapper.get(
      'nav[aria-label="Learning Item navigation"]'
    );

    expect(wrapper.get("main").classes()).toEqual(
      expect.arrayContaining(["w-full", "flex-1"])
    );
    expect(wrapper.get("main > div").classes()).toContain("w-full");
    expect(navigation.classes()).toEqual(
      expect.arrayContaining(["grid-cols-1", "sm:grid-cols-2"])
    );
  });
});
