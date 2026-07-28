import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import ScoreBoardTable from "~/components/ScoreBoardTable.vue";

vi.mock("~~/composables/avatar", () => ({
  getAvatarUrlByName: vi
    .fn()
    .mockImplementation(
      (name) => `https://api.dicebear.com/9.x/bottts/svg?seed=${name || "Eden"}`
    ),
}));

describe("ScoreBoardTable.vue", () => {
  const scoreboardData = [
    {
      rank: 1,
      username: "john_doe",
      firstname: "John",
      score: 150,
      img_key: "Sophia",
    },
    {
      rank: 2,
      username: "jane_doe",
      firstname: "Jane",
      score: 140,
      img_key: "Jude",
    },
  ];

  const mountComponent = (props = {}) =>
    mount(ScoreBoardTable, {
      props: {
        scoreboardData,
        isAdmin: true,
        userName: "",
        ...props,
      },
    });

  it("renders the rankings with the correct data", () => {
    const wrapper = mountComponent();
    const rows = wrapper.findAll("li");
    expect(rows.length).toBe(2);
    expect(rows[0].text()).toContain("John");
    expect(rows[0].text()).toContain("150");
    expect(wrapper.get('[aria-label="First place"]').exists()).toBe(true);
  });

  it("does not highlight any row when the user is an admin", () => {
    const wrapper = mountComponent({ isAdmin: true, userName: "jane_doe" });
    expect(wrapper.html()).not.toContain("bg-jv-yellow-soft");
  });

  it("displays username for admins", () => {
    const wrapper = mountComponent({ isAdmin: true });
    expect(wrapper.text()).toContain("@john_doe");
  });

  it("highlights the current user row when not an admin", () => {
    const wrapper = mountComponent({
      isAdmin: false,
      userName: "jane_doe",
    });
    const highlighted = wrapper
      .findAll("li")
      .find((row) => row.classes().includes("bg-jv-yellow-soft"));
    expect(highlighted).toBeTruthy();
    expect(highlighted.text()).toContain("Jane");
    expect(highlighted.text()).toContain("You");
  });

  it("renders avatar URLs correctly", () => {
    const wrapper = mountComponent();
    const avatars = wrapper.findAll("img");
    expect(avatars[0].attributes("src")).toContain("seed=Sophia");
    expect(avatars[1].attributes("src")).toContain("seed=Jude");
  });
});
