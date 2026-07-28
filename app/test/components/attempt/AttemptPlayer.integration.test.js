import { flushPromises, mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import AttemptPlayer from "@/components/attempt/AttemptPlayer.vue";
import {
  assertLearnerSafePayload,
  derivePaletteStatus,
  hydrateQuestionStatesFromResume,
} from "@/composables/attempt_player_state";
import {
  PALETTE_STATUS,
  QUESTION_TYPE_SINGLE,
  QUESTION_TYPE_SURVEY,
  SAVE_STATUS,
} from "@/utils/attempt_player_constants";

const autosaveAnswer = vi.fn();
const submitAttempt = vi.fn();
const getAttemptStatus = vi.fn();

vi.mock("@/composables/assessment_attempts", () => ({
  useAssessmentAttemptsApi: () => ({
    autosaveAnswer,
    submitAttempt,
    getAttemptStatus,
  }),
  getAssessmentAttemptAPIError: (error, fallback) =>
    error?.response?.data?.data || error?.message || fallback,
}));

const buildResume = (overrides = {}) => ({
  status: "IN_PROGRESS",
  attempt_id: "att-1",
  quiz_id: "quiz-1",
  test_snapshot_id: "snap-1",
  question_order: ["q-1", "q-2"],
  snapshot: {
    items: [
      {
        question_id: "q-1",
        position: 0,
        question: "Capital of India?",
        type: QUESTION_TYPE_SINGLE,
        options: { 1: "Delhi", 2: "Mumbai" },
        options_media: "text",
      },
      {
        question_id: "q-2",
        position: 1,
        question: "Select applicable topics",
        type: QUESTION_TYPE_SURVEY,
        options: { 1: "History", 2: "Polity" },
        options_media: "text",
      },
    ],
  },
  answers: [],
  ...overrides,
});

const mountPlayer = (resume = buildResume()) =>
  mount(AttemptPlayer, {
    props: {
      resume,
      quizId: "quiz-1",
      attemptId: "att-1",
    },
  });

const paletteButtonFor = (wrapper, index) =>
  wrapper.findAll(".attempt-palette__button")[index];

describe("AttemptPlayer integration flow", () => {
  beforeEach(() => {
    autosaveAnswer.mockReset();
  });

  it("selects an answer, autosaves, and marks palette answered only after success", async () => {
    let resolveSave;
    autosaveAnswer.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveSave = resolve;
        })
    );

    const wrapper = mountPlayer();
    await wrapper.find('input[type="radio"][value="1"]').setValue(true);
    await flushPromises();

    expect(autosaveAnswer).toHaveBeenCalledWith("quiz-1", "att-1", "q-1", {
      clear: false,
      selected_options: [1],
    });
    expect(paletteButtonFor(wrapper, 0).attributes("aria-label")).toContain(
      "Saving"
    );

    resolveSave({ selected_options: [1] });
    await flushPromises();

    expect(paletteButtonFor(wrapper, 0).attributes("aria-label")).toContain(
      "Answered"
    );
    expect(wrapper.text()).toContain("1 of 2 answered");
  });

  it("restores a saved answer after remount (reload)", async () => {
    const wrapper = mountPlayer(
      buildResume({
        answers: [{ question_id: "q-1", selected_options: [2] }],
      })
    );
    await flushPromises();

    const selected = wrapper.find('input[type="radio"][value="2"]');
    expect(selected.element.checked).toBe(true);
    expect(wrapper.text()).toContain("1 of 2 answered");
  });

  it("retries after save failure while preserving local draft", async () => {
    autosaveAnswer
      .mockRejectedValueOnce(new Error("Network error"))
      .mockResolvedValueOnce({ selected_options: [1] });

    const wrapper = mountPlayer();
    await wrapper.find('input[type="radio"][value="1"]').setValue(true);
    await flushPromises();

    expect(wrapper.text()).toContain("Network error");
    expect(wrapper.find(".attempt-question__retry").exists()).toBe(true);
    expect(paletteButtonFor(wrapper, 0).attributes("aria-label")).toContain(
      "Save failed"
    );

    await wrapper.find(".attempt-question__retry").trigger("click");
    await flushPromises();

    expect(autosaveAnswer).toHaveBeenCalledTimes(2);
    expect(paletteButtonFor(wrapper, 0).attributes("aria-label")).toContain(
      "Answered"
    );
  });

  it("retains draft selection when navigating while a save is pending", async () => {
    let resolveSave;
    autosaveAnswer.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveSave = resolve;
        })
    );

    const wrapper = mountPlayer();
    await wrapper.find('input[type="radio"][value="1"]').setValue(true);
    await flushPromises();

    await wrapper.get(".attempt-player__nav-button--primary").trigger("click");
    await flushPromises();
    expect(wrapper.text()).toContain("Select applicable topics");

    await paletteButtonFor(wrapper, 0).trigger("click");
    await flushPromises();
    expect(wrapper.find('input[type="radio"][value="1"]').element.checked).toBe(
      true
    );

    resolveSave({ selected_options: [1] });
    await flushPromises();
    expect(paletteButtonFor(wrapper, 0).attributes("aria-label")).toContain(
      "Answered"
    );
  });

  it("does not let a stale slower response overwrite a newer selection", async () => {
    const resolvers = [];
    autosaveAnswer.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolvers.push(resolve);
        })
    );

    const wrapper = mountPlayer();
    await wrapper.find('input[type="radio"][value="1"]').setValue(true);
    await flushPromises();
    expect(resolvers).toHaveLength(1);

    // Second selection bumps in-flight version while the first request is still open.
    await wrapper.find('input[type="radio"][value="2"]').setValue(true);
    await flushPromises();
    expect(resolvers).toHaveLength(1);

    resolvers[0]({ selected_options: [1] });
    await flushPromises();
    expect(resolvers).toHaveLength(2);

    resolvers[1]({ selected_options: [2] });
    await flushPromises();

    expect(wrapper.find('input[type="radio"][value="2"]').element.checked).toBe(
      true
    );
    expect(wrapper.find('input[type="radio"][value="1"]').element.checked).toBe(
      false
    );
    expect(paletteButtonFor(wrapper, 0).attributes("aria-label")).toContain(
      "Answered"
    );
  });

  it("clears an answer and remounts without a restored selection", async () => {
    autosaveAnswer.mockResolvedValue({ selected_options: [] });
    const wrapper = mountPlayer(
      buildResume({
        answers: [{ question_id: "q-1", selected_options: [1] }],
      })
    );
    await flushPromises();

    await wrapper.get(".attempt-question__secondary").trigger("click");
    await flushPromises();
    expect(autosaveAnswer).toHaveBeenCalledWith("quiz-1", "att-1", "q-1", {
      clear: true,
      selected_options: [],
    });

    const reloaded = mountPlayer(buildResume({ answers: [] }));
    await flushPromises();
    expect(
      reloaded.find('input[type="radio"][value="1"]').element.checked
    ).toBe(false);
    expect(reloaded.text()).toContain("0 of 2 answered");
  });

  it("surfaces terminal-attempt autosave rejection and disables further edits", async () => {
    autosaveAnswer.mockRejectedValue({
      response: { data: { data: "Attempt is not in progress" } },
    });

    const wrapper = mountPlayer();
    await wrapper.find('input[type="radio"][value="1"]').setValue(true);
    await flushPromises();

    expect(wrapper.find(".attempt-player__terminal").text()).toContain(
      "not in progress"
    );
  });

  it("keeps answer-key fields out of player state derived from resume", () => {
    const resume = buildResume({
      answers: [{ question_id: "q-1", selected_options: [1] }],
    });
    expect(() => assertLearnerSafePayload(resume)).not.toThrow();

    const { states } = hydrateQuestionStatesFromResume(resume);
    expect(() => assertLearnerSafePayload(states)).not.toThrow();
    expect(JSON.stringify(states)).not.toMatch(
      /official_answer|authoritative_answer|is_correct|"score"/
    );
    expect(states["q-1"].savedSelection).toEqual([1]);
    expect(states["q-1"].saveStatus).toBe(SAVE_STATUS.IDLE);
    states["q-1"].visited = true;
    expect(derivePaletteStatus(states["q-1"])).toBe(PALETTE_STATUS.ANSWERED);
  });

  it("opens submit dialog and submits attempt after flushing pending saves", async () => {
    submitAttempt.mockResolvedValue({ id: "att-1", status: "SUBMITTED" });
    const wrapper = mountPlayer();

    const submitBtn = wrapper.find(".attempt-player__nav-button--submit");
    expect(submitBtn.exists()).toBe(true);

    await submitBtn.trigger("click");
    await flushPromises();

    expect(wrapper.find('[role="dialog"]').exists()).toBe(true);

    await wrapper.find(".submit-dialog__button--confirm").trigger("click");
    await flushPromises();

    expect(submitAttempt).toHaveBeenCalledWith("quiz-1", "att-1");
    expect(wrapper.emitted("submitted")).toBeDefined();
    expect(wrapper.emitted("submitted")[0][0]).toEqual({
      id: "att-1",
      status: "SUBMITTED",
    });
  });
});
