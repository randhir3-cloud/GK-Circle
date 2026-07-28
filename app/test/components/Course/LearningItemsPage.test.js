import { flushPromises } from "@vue/test-utils";
import { mountSuspended } from "@nuxt/test-utils/runtime";
import { beforeEach, describe, expect, it, vi } from "vitest";
import LearningItemsPage from "@/pages/admin/courses/learning-items.vue";

const testState = vi.hoisted(() => ({
  api: {
    listCourses: vi.fn(),
    listRootNodes: vi.fn(),
    listChildren: vi.fn(),
    listItems: vi.fn(),
    createItem: vi.fn(),
    updateItem: vi.fn(),
    deleteItem: vi.fn(),
  },
  userStore: {
    userData: { role: "admin-user", canCreatePublicQuiz: true },
  },
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock("@/composables/course_learning_items", () => ({
  useCourseLearningItemsApi: () => testState.api,
  getCourseAdminAPIError: (error, fallback) =>
    error?.data?.data || error?.data?.message || error?.message || fallback,
}));

vi.mock("@/composables/auth", () => ({
  setUserDataStore: vi.fn(),
}));

vi.mock("~~/store/users", () => ({
  useUsersStore: () => testState.userStore,
}));

vi.mock("notivue", () => ({
  usePush: () => testState.toast,
}));

const EditorStub = {
  props: ["modelValue", "mode", "item", "saving", "error"],
  emits: ["update:modelValue", "save"],
  template: `
    <div v-if="modelValue" data-testid="editor-stub">
      <span>{{ error }}</span>
      <button
        data-testid="editor-save"
        @click="$emit('save', mode === 'create'
          ? { title: 'New item', item_type: 'ARTICLE', publish_state: 'DRAFT' }
          : { title: 'Updated item', item_type: 'VIDEO', publish_state: 'PUBLISHED', description: null })"
      >save</button>
      <button data-testid="editor-cancel" @click="$emit('update:modelValue', false)">
        cancel
      </button>
    </div>
  `,
};

const DeleteStub = {
  props: ["modelValue", "item", "deleting", "error"],
  emits: ["update:modelValue", "confirm"],
  template: `
    <div v-if="modelValue" data-testid="delete-stub">
      <span>{{ error }}</span>
      <button data-testid="delete-confirm" @click="$emit('confirm')">confirm</button>
      <button data-testid="delete-cancel" @click="$emit('update:modelValue', false)">
        cancel
      </button>
    </div>
  `,
};

const course = { id: "course-1", title: "PCS Course" };
const rootNode = {
  id: "node-root",
  title: "History",
  node_type: "SUBJECT",
};
const firstItem = {
  id: "item-1",
  title: "Introduction",
  item_type: "ARTICLE",
  description: null,
  position: 10,
  publish_state: "DRAFT",
};
const secondItem = {
  id: "item-2",
  title: "Foundation",
  item_type: "VIDEO",
  description: "Watch first",
  position: 2,
  publish_state: "PUBLISHED",
};

const mountPage = async () => {
  const wrapper = await mountSuspended(LearningItemsPage, {
    global: {
      stubs: {
        LearningItemEditorDialog: EditorStub,
        LearningItemDeleteDialog: DeleteStub,
      },
    },
  });
  await flushPromises();
  return wrapper;
};

const selectRoot = async (wrapper) => {
  await wrapper.get('[data-testid="course-selector"]').setValue(course.id);
  await flushPromises();
  await wrapper.get('[data-testid="node-selector-0"]').setValue(rootNode.id);
  await flushPromises();
};

beforeEach(() => {
  vi.clearAllMocks();
  testState.userStore.userData = {
    role: "admin-user",
    canCreatePublicQuiz: true,
  };
  testState.api.listCourses.mockResolvedValue([course]);
  testState.api.listRootNodes.mockResolvedValue([rootNode]);
  testState.api.listChildren.mockResolvedValue([]);
  testState.api.listItems.mockResolvedValue([]);
  testState.api.createItem.mockResolvedValue({});
  testState.api.updateItem.mockResolvedValue({});
  testState.api.deleteItem.mockResolvedValue({});
});

describe("Course LearningItems admin page", () => {
  it("loads Courses and starts with explicit selection prompts", async () => {
    const wrapper = await mountPage();

    expect(testState.api.listCourses).toHaveBeenCalledOnce();
    expect(wrapper.text()).toContain("Select a Course");
    expect(wrapper.get('[data-testid="node-prompt"]').text()).toContain(
      "Select a Course"
    );
    expect(wrapper.get('[data-testid="item-prompt"]').text()).toContain(
      "Select a CourseNode"
    );
  });

  it("shows the empty state only after a selected node returns an empty list", async () => {
    const wrapper = await mountPage();
    expect(wrapper.find('[data-testid="learning-item-empty"]').exists()).toBe(
      false
    );

    await selectRoot(wrapper);

    expect(testState.api.listItems).toHaveBeenCalledWith(
      course.id,
      rootNode.id
    );
    expect(wrapper.get('[data-testid="learning-item-empty"]').text()).toContain(
      "No Learning Items yet."
    );
  });

  it("renders items in the exact order returned by the server", async () => {
    testState.api.listItems.mockResolvedValue([firstItem, secondItem]);
    const wrapper = await mountPage();
    await selectRoot(wrapper);

    const rows = wrapper.findAll('[data-testid^="learning-item-"]');
    expect(rows.map((row) => row.attributes("data-testid"))).toEqual([
      "learning-item-table",
      "learning-item-item-1",
      "learning-item-item-2",
    ]);
    expect(rows[1].text()).toContain("10");
    expect(rows[2].text()).toContain("2");
  });

  it("drills through direct-child APIs and loads items on a deep node", async () => {
    const child1 = { id: "node-1", title: "Level 1", node_type: "SECTION" };
    const child2 = { id: "node-2", title: "Level 2", node_type: "TOPIC" };
    testState.api.listChildren.mockImplementation((_courseId, nodeId) => {
      if (nodeId === rootNode.id) return Promise.resolve([child1]);
      if (nodeId === child1.id) return Promise.resolve([child2]);
      return Promise.resolve([]);
    });

    const wrapper = await mountPage();
    await selectRoot(wrapper);
    await wrapper.get('[data-testid="node-selector-1"]').setValue(child1.id);
    await flushPromises();
    await wrapper.get('[data-testid="node-selector-2"]').setValue(child2.id);
    await flushPromises();

    expect(testState.api.listChildren).toHaveBeenNthCalledWith(
      1,
      course.id,
      rootNode.id
    );
    expect(testState.api.listChildren).toHaveBeenNthCalledWith(
      2,
      course.id,
      child1.id
    );
    expect(testState.api.listItems).toHaveBeenLastCalledWith(
      course.id,
      child2.id
    );
  });

  it("creates, updates, and deletes through existing APIs then refreshes", async () => {
    testState.api.listItems.mockResolvedValue([firstItem]);
    const wrapper = await mountPage();
    await selectRoot(wrapper);

    await wrapper.get("button").trigger("click");
    await wrapper.get('[data-testid="editor-save"]').trigger("click");
    await flushPromises();
    expect(testState.api.createItem).toHaveBeenCalledWith(
      course.id,
      rootNode.id,
      {
        title: "New item",
        item_type: "ARTICLE",
        publish_state: "DRAFT",
      }
    );

    await wrapper
      .get('button[aria-label="Edit Introduction"]')
      .trigger("click");
    await wrapper.get('[data-testid="editor-save"]').trigger("click");
    await flushPromises();
    expect(testState.api.updateItem).toHaveBeenCalledWith(
      course.id,
      rootNode.id,
      firstItem.id,
      {
        title: "Updated item",
        item_type: "VIDEO",
        publish_state: "PUBLISHED",
        description: null,
      }
    );

    await wrapper
      .get('button[aria-label="Delete Introduction"]')
      .trigger("click");
    await wrapper.get('[data-testid="delete-confirm"]').trigger("click");
    await flushPromises();
    expect(testState.api.deleteItem).toHaveBeenCalledWith(
      course.id,
      rootNode.id,
      firstItem.id
    );
    expect(testState.api.listItems.mock.calls.length).toBe(4);
  });

  it("cancels deletion without calling the API", async () => {
    testState.api.listItems.mockResolvedValue([firstItem]);
    const wrapper = await mountPage();
    await selectRoot(wrapper);
    await wrapper
      .get('button[aria-label="Delete Introduction"]')
      .trigger("click");
    await wrapper.get('[data-testid="delete-cancel"]').trigger("click");

    expect(testState.api.deleteItem).not.toHaveBeenCalled();
    expect(wrapper.find('[data-testid="delete-stub"]').exists()).toBe(false);
  });

  it("surfaces Course, item, and mutation API errors", async () => {
    testState.api.listCourses.mockRejectedValueOnce({
      data: { message: "Course service unavailable" },
    });
    let wrapper = await mountPage();
    expect(wrapper.text()).toContain("Course service unavailable");

    testState.api.listCourses.mockResolvedValue([course]);
    testState.api.listItems.mockRejectedValueOnce({
      data: { data: "Item list failed" },
    });
    wrapper = await mountPage();
    await selectRoot(wrapper);
    expect(wrapper.text()).toContain("Item list failed");

    testState.api.listItems.mockResolvedValue([firstItem]);
    testState.api.createItem.mockRejectedValueOnce({
      data: { data: "Create rejected" },
    });
    wrapper = await mountPage();
    await selectRoot(wrapper);
    await wrapper.get("button").trigger("click");
    await wrapper.get('[data-testid="editor-save"]').trigger("click");
    await flushPromises();
    expect(wrapper.text()).toContain("Create rejected");
  });

  it("discards stale node results when the Course changes", async () => {
    const otherCourse = { id: "course-2", title: "Other Course" };
    const otherRoot = {
      id: "other-root",
      title: "Other Root",
      node_type: "TOPIC",
    };
    let resolveFirstRoots;
    testState.api.listCourses.mockResolvedValue([course, otherCourse]);
    testState.api.listRootNodes
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveFirstRoots = resolve;
          })
      )
      .mockResolvedValueOnce([otherRoot]);

    const wrapper = await mountPage();
    await wrapper.get('[data-testid="course-selector"]').setValue(course.id);
    await wrapper
      .get('[data-testid="course-selector"]')
      .setValue(otherCourse.id);
    await flushPromises();
    resolveFirstRoots([rootNode]);
    await flushPromises();

    expect(wrapper.get('[data-testid="node-selector-0"]').text()).toContain(
      "Other Root"
    );
    expect(wrapper.get('[data-testid="node-selector-0"]').text()).not.toContain(
      "History"
    );
  });
});
