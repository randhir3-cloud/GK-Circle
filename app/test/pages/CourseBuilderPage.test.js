import { flushPromises } from "@vue/test-utils";
import { mountSuspended } from "@nuxt/test-utils/runtime";
import { beforeEach, describe, expect, it, vi } from "vitest";
import CourseBuilderPage from "@/pages/admin/courses/index.vue";

const testState = vi.hoisted(() => ({
  api: {
    listCourses: vi.fn(),
    getTree: vi.fn(),
    createCourse: vi.fn(),
    createNode: vi.fn(),
    updateCourse: vi.fn(),
  },
  toast: {
    success: vi.fn(),
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
  useUsersStore: () => ({ userData: { role: "admin-user" } }),
}));

vi.mock("notivue", () => ({
  usePush: () => testState.toast,
}));

const course = {
  id: "course-1",
  title: "PCS Course",
  status: "DRAFT",
};
const subject = {
  id: "subject-1",
  title: "History",
  node_type: "SUBJECT",
};

const mountPage = async () => {
  const wrapper = await mountSuspended(CourseBuilderPage);
  await flushPromises();
  await wrapper
    .get('[data-testid="builder-course-selector"]')
    .setValue(course.id);
  await flushPromises();
  return wrapper;
};

beforeEach(() => {
  vi.clearAllMocks();
  testState.api.listCourses.mockResolvedValue([course]);
  testState.api.getTree.mockResolvedValue({ roots: [] });
  testState.api.createCourse.mockResolvedValue(course);
  testState.api.createNode.mockResolvedValue({});
  testState.api.updateCourse.mockResolvedValue(course);
});

describe("Course Builder node creation", () => {
  it("keeps the builder focused and links to the searchable Course list", async () => {
    const wrapper = await mountPage();

    expect(wrapper.find('[data-testid="created-courses"]').exists()).toBe(
      false
    );
    expect(wrapper.get('a[href="/admin/courses/list"]').text()).toContain(
      "Browse all Courses"
    );
  });

  it("appends a top-level Subject with the required position", async () => {
    const wrapper = await mountPage();

    await wrapper.get('[data-testid="node-title"]').setValue("History");
    await wrapper.get('[data-testid="node-type"]').setValue("SUBJECT");

    const parentOptions = wrapper
      .get('[data-testid="node-parent"]')
      .findAll("option");
    expect(parentOptions).toHaveLength(1);
    expect(parentOptions[0].text()).toContain("Top-level root");
    expect(wrapper.text()).toContain(
      "Subjects are always created at the Course root."
    );

    await wrapper.get('[data-testid="create-node-button"]').trigger("click");
    await flushPromises();

    expect(testState.api.createNode).toHaveBeenCalledWith(course.id, {
      title: "History",
      node_type: "SUBJECT",
      position: 0,
    });
  });

  it("appends a Topic after the selected parent's existing children", async () => {
    testState.api.getTree.mockResolvedValue({
      roots: [
        {
          node: subject,
          children: [
            {
              node: {
                id: "topic-1",
                title: "Ancient India",
                node_type: "TOPIC",
              },
              children: [],
            },
          ],
        },
      ],
    });
    const wrapper = await mountPage();

    await wrapper.get('[data-testid="node-title"]').setValue("Modern India");
    await wrapper.get('[data-testid="node-type"]').setValue("TOPIC");

    const addTopicButton = wrapper.get('[data-testid="create-node-button"]');
    expect(addTopicButton.attributes("disabled")).toBeDefined();
    const parentOptions = wrapper
      .get('[data-testid="node-parent"]')
      .findAll("option");
    expect(parentOptions.map((option) => option.text())).toEqual([
      "Select a Subject",
      "History (SUBJECT)",
    ]);
    expect(wrapper.text()).toContain(
      "Choose the Subject that will contain this Topic."
    );

    await wrapper.get('[data-testid="node-parent"]').setValue(subject.id);
    expect(addTopicButton.attributes("disabled")).toBeUndefined();
    await wrapper.get('[data-testid="create-node-button"]').trigger("click");
    await flushPromises();

    expect(testState.api.createNode).toHaveBeenCalledWith(course.id, {
      title: "Modern India",
      node_type: "TOPIC",
      position: 1,
      parent_id: subject.id,
    });

    const nestedTopic = wrapper
      .get('[data-testid="course-outline"]')
      .find('[data-depth="1"]');
    expect(nestedTopic.text()).toContain("TOPIC — Ancient India");
    expect(nestedTopic.attributes("style")).toContain("margin-left: 32px");
  });
});
