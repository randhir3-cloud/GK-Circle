import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import ListJoinUser from "~/components/ListJoinUser.vue";
import { createTestingPinia } from "@pinia/testing";
import { useListUserstore } from "~/store/userlist";

vi.mock("~~/composables/avatar", () => ({
  getAvatarUrlByName: vi
    .fn()
    .mockImplementation(
      (name) => `https://api.dicebear.com/9.x/bottts/svg?seed=${name || "Eden"}`
    ),
}));

const mockListUsers = [
  { UserId: 1, UserName: "Alice", Avatar: "Alice" },
  { UserId: 2, UserName: "Bob", Avatar: "Bob" },
];

describe("ListJoinUser test", () => {
  let pinia;
  let listUserStore;

  beforeEach(() => {
    pinia = createTestingPinia({ createSpy: vi.fn });
    listUserStore = useListUserstore(pinia);
    listUserStore.listUsers = [...mockListUsers];
  });

  const mountComponent = () =>
    mount(ListJoinUser, {
      global: {
        plugins: [pinia],
      },
    });

  it("shows the participant count when listUsers is not empty", () => {
    const wrapper = mountComponent();
    expect(wrapper.find("h5").text()).toBe("2 Participants");
  });

  it("renders user cards with correct data", () => {
    const wrapper = mountComponent();
    const cards = wrapper.findAll(".jv-card");
    expect(cards.length).toBe(2);
    expect(cards[0].text()).toContain("Alice");
    expect(cards[0].find("img").attributes("src")).toContain("seed=Alice");
    expect(cards[1].text()).toContain("Bob");
    expect(cards[1].find("img").attributes("src")).toContain("seed=Bob");
  });

  it("shows waiting copy when listUsers is empty", async () => {
    listUserStore.listUsers = [];
    const wrapper = mountComponent();
    await wrapper.vm.$nextTick();
    expect(wrapper.find("h5").text()).toBe("Waiting for Participants...");
    expect(wrapper.findAll(".jv-card").length).toBe(0);
  });
});
