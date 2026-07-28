import { mockNuxtImport } from "@nuxt/test-utils/runtime";
import { mount, flushPromises } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach } from "vitest";
import FinalScoreBoard from "~/components/FinalScoreBoard.vue";
import ScoreBoardTable from "~/components/ScoreBoardTable.vue";

const routeQuery = {
  aqi: "quiz123",
  winner_ui: "true",
};

vi.mock("notivue", () => ({
  usePush: vi.fn(() => ({
    error: vi.fn(),
    success: vi.fn(),
  })),
}));

vi.mock("~~/composables/avatar", () => ({
  getAvatarUrlByName: vi
    .fn()
    .mockImplementation(
      (name) => `https://api.dicebear.com/9.x/bottts/svg?seed=${name || "Eden"}`
    ),
}));

mockNuxtImport("useRoute", () => () => ({
  query: routeQuery,
}));

mockNuxtImport("useRuntimeConfig", () => () => ({
  public: {
    apiUrl: "http://localhost:5001",
  },
}));

describe("FinalScoreBoard Test", () => {
  beforeEach(() => {
    routeQuery.aqi = "quiz123";
    routeQuery.winner_ui = "true";
    vi.stubGlobal(
      "$fetch",
      vi.fn().mockResolvedValue({
        data: [
          {
            id: 1,
            firstname: "User 1",
            username: "User 1",
            rank: 1,
            score: 1000,
            img_key: "Sophia",
          },
          {
            id: 2,
            firstname: "User 2",
            username: "User 2",
            rank: 2,
            score: 900,
            img_key: "Jude",
          },
        ],
      })
    );
  });

  const mountComponent = (props = {}) =>
    mount(FinalScoreBoard, {
      props: {
        userURL: "/scoreboard",
        isAdmin: true,
        ...props,
      },
      global: {
        stubs: {
          WinnerConfetti: true,
          ClientOnly: { template: "<div><slot /></div>" },
          NuxtLink: true,
          ScoreBoardTable: true,
          WinnerCard: true,
        },
        plugins: [],
      },
    });

  it("fetches scoreboard data for admin", async () => {
    const wrapper = mountComponent();
    wrapper.vm.requestPending = false;
    wrapper.vm.scoreboardData = [
      { id: 1, firstname: "User 1", username: "User 1", rank: 1, score: 1000 },
      { id: 2, firstname: "User 2", username: "User 2", rank: 2, score: 900 },
    ];
    await flushPromises();
    expect(wrapper.vm.scoreboardData).toHaveLength(2);
  });

  it("does not request a scoreboard when no active quiz is selected", async () => {
    routeQuery.aqi = "";
    mountComponent();
    await flushPromises();

    expect($fetch).not.toHaveBeenCalled();
  });

  it("stores analysis data for users", async () => {
    const wrapper = mountComponent({ isAdmin: false });
    wrapper.vm.analysisData = [{ question: "Q1" }, { question: "Q2" }];
    await flushPromises();
    expect(wrapper.vm.analysisData).toHaveLength(2);
  });

  it("renders the winner UI for admin when winner_ui is true", async () => {
    const wrapper = mountComponent({ isAdmin: true });
    await flushPromises();
    expect(wrapper.vm.winnerUI).toBe("true");
    expect(wrapper.findComponent({ name: "WinnerConfetti" }).exists()).toBe(
      true
    );
  });

  it("renders the scoreboard table for non-admin users", async () => {
    const wrapper = mountComponent({ isAdmin: false });
    wrapper.vm.requestPending = false;
    wrapper.vm.scoreboardData = [
      { id: 1, firstname: "User 1", username: "User 1", rank: 1, score: 1000 },
    ];
    await flushPromises();
    await wrapper.vm.$nextTick();
    expect(wrapper.findComponent(ScoreBoardTable).exists()).toBe(true);
  });
});
