import { describe, it, expect } from "vitest";
import { mountSuspended } from "@nuxt/test-utils/runtime";
import UserGuide from "~/pages/docs/user-guide.vue";

describe("UserGuide Page", () => {
  it("renders header and main sections", async () => {
    const wrapper = await mountSuspended(UserGuide);

    expect(wrapper.text()).toContain("GK Circle User Guide");
    expect(wrapper.text()).toContain("1. Getting Started");
    expect(wrapper.text()).toContain("2. Structured Courses & Learning Items");
    expect(wrapper.text()).toContain("3. Practice Sets & Mock Examinations");
    expect(wrapper.text()).toContain("4. Timed Live Competitions");
    expect(wrapper.text()).toContain("5. Performance Analytics & Reports");
  });

  it("contains navigation links to home, courses, and join", async () => {
    const wrapper = await mountSuspended(UserGuide);

    const html = wrapper.html();
    expect(html).toContain('href="/"');
    expect(html).toContain('href="/courses"');
    expect(html).toContain('href="/join"');
  });
});
