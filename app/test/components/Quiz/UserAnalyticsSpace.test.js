import { mount } from "@vue/test-utils";
import { describe, it, expect, vi } from "vitest";
import UserAnalyticsSpace from "~/components/Quiz/UserAnalyticsSpace.vue";

vi.mock("~~/composables/avatar", () => ({
  getAvatarUrlByName: vi
    .fn()
    .mockImplementation(
      (name) => `https://api.dicebear.com/9.x/bottts/svg?seed=${name || "Eden"}`
    ),
}));

const props = {
  data: [
    {
      firstname: "John",
      username: "john_doe",
      avatar: "Eden",
      rank: 1,
      total_score: 10,
    },
  ],
  userName: "john_doe",
};

describe("UserAnalyticsSpace test", () => {
  it("renders the component with provided props", () => {
    const wrapper = mount(UserAnalyticsSpace, { props });
    expect(wrapper.text()).toContain("John");
    expect(wrapper.text()).toContain("john_doe");
    expect(wrapper.find("img").attributes("src")).toContain("seed=Eden");
  });

  it("computes the correct avatar URL", () => {
    const wrapper = mount(UserAnalyticsSpace, { props });
    const avatarImg = wrapper.find("img");
    expect(avatarImg.attributes("src")).toBe(
      "https://api.dicebear.com/9.x/bottts/svg?seed=Eden"
    );
  });
});
