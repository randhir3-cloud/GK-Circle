import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import ListUserAnswerd from "~/components/Quiz/ListUserAnswered.vue";
import { createTestingPinia } from "@pinia/testing";
import { useUserThatSubmittedAnswer } from "~/store/userSubmittedAnswer";

vi.mock("notivue", () => ({
  usePush: vi.fn(() => ({
    error: vi.fn(),
  })),
}));

vi.mock("~~/composables/avatar", () => ({
  getAvatarUrlByName: vi
    .fn()
    .mockImplementation(
      (name) => `https://api.dicebear.com/9.x/bottts/svg?seed=${name || "Eden"}`
    ),
}));

describe("ListUserAnswered Test", () => {
  let pinia;
  let usersThatSubmittedAnswer;

  beforeEach(() => {
    pinia = createTestingPinia({ createSpy: vi.fn });
    usersThatSubmittedAnswer = useUserThatSubmittedAnswer(pinia);
    usersThatSubmittedAnswer.usersSubmittedAnswers = [];
    vi.stubGlobal("useNuxtApp", () => ({
      $Fail: "fail",
      $GetQuestion: "send_question",
    }));
  });

  const mountComponent = (props = {}) =>
    mount(ListUserAnswerd, {
      props: {
        data: { status: "success", data: "" },
        runningQuizJoinUser: 10,
        ...props,
      },
      global: {
        plugins: [pinia],
      },
    });

  it("renders correctly when no answers are submitted", () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("No one answered till now");
  });

  it("renders correctly when answers are submitted", () => {
    usersThatSubmittedAnswer.usersSubmittedAnswers = [
      {
        UserId: 1,
        img_key: "Sophia",
        first_name: "Alice",
        username: "alice123",
      },
    ];
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain("1");
    expect(wrapper.text()).toContain("players answered");
    expect(wrapper.text()).toContain("Alice");
    expect(wrapper.text()).toContain("@alice123");
  });

  it("handles new user join event correctly", async () => {
    const wrapper = mountComponent();
    await wrapper.setProps({
      data: {
        event: "send_question",
        data: { totalJoinUser: 12 },
      },
    });
    expect(wrapper.vm.totalUser).toBe(12);
  });

  it("emits autoSkip when all users have answered", async () => {
    usersThatSubmittedAnswer.usersSubmittedAnswers = [
      {
        UserId: 1,
        img_key: "Sophia",
        first_name: "Alice",
        username: "alice123",
      },
      {
        UserId: 2,
        img_key: "Jude",
        first_name: "Bob",
        username: "bob321",
      },
    ];

    const wrapper = mountComponent({
      data: {
        event: "send_question",
        data: { totalJoinUser: 2 },
      },
      runningQuizJoinUser: 2,
    });

    expect(wrapper.emitted("autoSkip")).toBeTruthy();
  });

  it("displays the correct number of answered user cards", () => {
    usersThatSubmittedAnswer.usersSubmittedAnswers = [
      {
        UserId: 1,
        img_key: "Sophia",
        first_name: "Alice",
        username: "alice123",
      },
      {
        UserId: 2,
        img_key: "Jude",
        first_name: "Bob",
        username: "bob321",
      },
    ];
    const wrapper = mountComponent({ runningQuizJoinUser: 2 });
    const cards = wrapper.findAll("li");
    expect(cards.length).toBe(2);
    expect(cards[0].text()).toContain("Alice");
    expect(cards[1].text()).toContain("Bob");
  });
});
