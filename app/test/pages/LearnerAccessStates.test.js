import { flushPromises } from "@vue/test-utils";
import { mountSuspended } from "@nuxt/test-utils/runtime";
import { beforeEach, describe, expect, it, vi } from "vitest";
import CoursesPage from "@/pages/courses/index.vue";
import AnalyticsPage from "@/pages/analytics/index.vue";

const testState = vi.hoisted(() => ({
  authenticatedUser: null,
  coursesApi: {
    listPublishedCourses: vi.fn(),
  },
  analyticsApi: {
    getDashboard: vi.fn(),
    getSubjects: vi.fn(),
    getActivity: vi.fn(),
    getTrends: vi.fn(),
    getAttemptTimeline: vi.fn(),
  },
}));

vi.mock("@/composables/auth", () => ({
  setUserDataStore: vi.fn(() => Promise.resolve(testState.authenticatedUser)),
}));

vi.mock("@/composables/learner_learning_items", () => ({
  useLearnerLearningItemsApi: () => testState.coursesApi,
}));

vi.mock("@/composables/learner_analytics", () => ({
  useLearnerAnalyticsApi: () => testState.analyticsApi,
}));

vi.mock("~~/store/users", () => ({
  useUsersStore: () => ({ setUserData: vi.fn() }),
}));

beforeEach(() => {
  vi.clearAllMocks();
  testState.authenticatedUser = null;
  testState.coursesApi.listPublishedCourses.mockResolvedValue([]);
  testState.analyticsApi.getDashboard.mockResolvedValue({ total_attempts: 0 });
  testState.analyticsApi.getSubjects.mockResolvedValue({ subjects: [] });
  testState.analyticsApi.getActivity.mockResolvedValue({ items: [] });
  testState.analyticsApi.getTrends.mockResolvedValue({ buckets: [] });
});

describe("learner authentication states", () => {
  it("allows unauthenticated learners to browse published courses catalog", async () => {
    const wrapper = await mountSuspended(CoursesPage);
    await flushPromises();

    expect(wrapper.text()).toContain("Courses");
    expect(wrapper.text()).not.toContain("session_id");
    expect(wrapper.text()).not.toContain("kratos");
    expect(testState.coursesApi.listPublishedCourses).toHaveBeenCalled();
  });

  it("shows a sign-in state on Analytics without exposing or calling the API", async () => {
    const wrapper = await mountSuspended(AnalyticsPage);
    await flushPromises();

    expect(wrapper.text()).toContain(
      "Sign in to view your learning analytics."
    );
    expect(wrapper.text()).not.toContain("session_id");
    expect(wrapper.text()).not.toContain("kratos");
    expect(testState.analyticsApi.getDashboard).not.toHaveBeenCalled();
  });

  it("shows useful Analytics placeholders when a learner has no attempts", async () => {
    testState.authenticatedUser = { role: "user" };
    const wrapper = await mountSuspended(AnalyticsPage);
    await flushPromises();

    expect(wrapper.get('[data-testid="analytics-empty"]').text()).toContain(
      "No assessments completed yet."
    );
    expect(wrapper.text()).toContain("Accuracy");
    expect(wrapper.text()).toContain("Questions Attempted");
    expect(wrapper.text()).toContain("Weak Subjects");
  });
});
