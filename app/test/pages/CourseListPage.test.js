import { flushPromises } from "@vue/test-utils";
import { mountSuspended } from "@nuxt/test-utils/runtime";
import { beforeEach, describe, expect, it, vi } from "vitest";
import CourseListPage from "@/pages/admin/courses/list.vue";

const testState = vi.hoisted(() => ({
  api: {
    listCourses: vi.fn(),
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

const courses = [
  { id: "course-1", title: "PCS Foundation", status: "DRAFT" },
  { id: "course-2", title: "Indian Geography", status: "PUBLISHED" },
];

const mountPage = async () => {
  const wrapper = await mountSuspended(CourseListPage);
  await flushPromises();
  return wrapper;
};

beforeEach(() => {
  vi.clearAllMocks();
  testState.api.listCourses.mockResolvedValue(courses);
  testState.api.updateCourse.mockResolvedValue({});
});

describe("Courses management page", () => {
  it("searches created Courses and provides visible management actions", async () => {
    const wrapper = await mountPage();

    expect(wrapper.findAll('[data-testid^="course-card-"]')).toHaveLength(2);
    await wrapper.get('[data-testid="course-search"]').setValue("geography");

    expect(wrapper.findAll('[data-testid^="course-card-"]')).toHaveLength(1);
    const card = wrapper.get('[data-testid="course-card-course-2"]');
    expect(card.text()).toContain("Indian Geography");
    expect(
      card.get('a[href="/admin/courses?course=course-2"]').text()
    ).toContain("Edit structure");
    const contentLink = card.get(
      'a[href="/admin/courses/learning-items?course=course-2"]'
    );
    expect(contentLink.text()).toContain("Manage content");
    expect(contentLink.classes()).toContain("text-jv-ink");
  });

  it("renames a Course from the dedicated list", async () => {
    const wrapper = await mountPage();

    await wrapper
      .get('button[aria-label="Rename PCS Foundation"]')
      .trigger("click");
    await wrapper
      .get('input[name="course_title"]')
      .setValue("PCS Foundation Course");
    await wrapper
      .get('[data-testid="course-card-course-1"] form')
      .trigger("submit");
    await flushPromises();

    expect(testState.api.updateCourse).toHaveBeenCalledWith("course-1", {
      title: "PCS Foundation Course",
    });
    expect(testState.toast.success).toHaveBeenCalledWith(
      "Course title updated."
    );
  });

  it("changes publication state from the Course card", async () => {
    const wrapper = await mountPage();

    await wrapper
      .get('select[aria-label="Publication state for PCS Foundation"]')
      .setValue("PUBLISHED");
    await flushPromises();

    expect(testState.api.updateCourse).toHaveBeenCalledWith("course-1", {
      status: "PUBLISHED",
    });
    expect(testState.toast.success).toHaveBeenCalledWith(
      "Course status set to PUBLISHED."
    );
  });
});
