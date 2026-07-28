import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import WinnerCard from "@/components/WinnerCard.vue";

vi.mock("~~/composables/avatar", () => ({
  getAvatarUrlByName: vi
    .fn()
    .mockImplementation(
      (name) => `https://api.dicebear.com/9.x/bottts/svg?seed=${name || "Eden"}`
    ),
}));

const baseWinner = {
  rank: 1,
  img_key: "Sophia",
  firstname: "John",
  username: "john_doe",
  score: 8256,
};

const mountComponent = (winner = baseWinner) =>
  mount(WinnerCard, {
    props: { winner },
  });

describe("WinnerCard test", () => {
  it("renders the correct medal label for rank 1", () => {
    const wrapper = mountComponent({ ...baseWinner, rank: 1 });
    expect(wrapper.get('[role="article"]').attributes("aria-label")).toContain(
      "1st place"
    );
    expect(wrapper.text()).toContain("1st Place");
  });

  it("renders the correct medal label for rank 2", () => {
    const wrapper = mountComponent({ ...baseWinner, rank: 2 });
    expect(wrapper.text()).toContain("2nd Place");
  });

  it("renders the correct medal label for rank 3", () => {
    const wrapper = mountComponent({ ...baseWinner, rank: 3 });
    expect(wrapper.text()).toContain("3rd Place");
  });

  it("renders the correct avatar image", () => {
    const wrapper = mountComponent();
    const avatarImg = wrapper.find('img[alt="Avatar for John"]');
    expect(avatarImg.exists()).toBe(true);
    expect(avatarImg.attributes("src")).toContain("seed=Sophia");
  });

  it("renders winner details correctly", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("John");
    expect(wrapper.text()).toContain("@john_doe");
    expect(wrapper.text()).toContain("8256");
  });
});
