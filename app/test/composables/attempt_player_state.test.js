import { describe, expect, it } from "vitest";
import {
  applySaveResult,
  beginSave,
  buildOrderedSnapshotItems,
  createInitialQuestionState,
  derivePaletteStatus,
  hydrateQuestionStatesFromResume,
  toggleOptionSelection,
} from "@/composables/attempt_player_state";
import {
  PALETTE_STATUS,
  QUESTION_TYPE_SINGLE,
  QUESTION_TYPE_SURVEY,
  SAVE_STATUS,
} from "@/utils/attempt_player_constants";

const sampleResume = {
  status: "IN_PROGRESS",
  question_order: ["q-2", "q-1"],
  snapshot: {
    items: [
      {
        question_id: "q-1",
        position: 1,
        question: "First?",
        type: QUESTION_TYPE_SINGLE,
        options: { 1: "A", 2: "B" },
      },
      {
        question_id: "q-2",
        position: 0,
        question: "Second?",
        type: QUESTION_TYPE_SURVEY,
        options: { 1: "X", 2: "Y", 3: "Z" },
      },
    ],
  },
  answers: [{ question_id: "q-1", selected_options: [2] }],
};

describe("attempt_player_state", () => {
  it("orders items from frozen question_order", () => {
    const ordered = buildOrderedSnapshotItems(sampleResume);
    expect(ordered.map((item) => item.question_id)).toEqual(["q-2", "q-1"]);
  });

  it("hydrates saved answers without answer keys", () => {
    const { states } = hydrateQuestionStatesFromResume(sampleResume);
    expect(states["q-1"].savedSelection).toEqual([2]);
    expect(states["q-2"].savedSelection).toEqual([]);
  });

  it("derives palette statuses", () => {
    const state = createInitialQuestionState("q-1", [1]);
    expect(derivePaletteStatus(state)).toBe(PALETTE_STATUS.NOT_VISITED);
    state.visited = true;
    expect(derivePaletteStatus(state)).toBe(PALETTE_STATUS.ANSWERED);
    state.savedSelection = [];
    expect(derivePaletteStatus(state)).toBe(PALETTE_STATUS.VISITED_UNANSWERED);
    state.saveStatus = SAVE_STATUS.SAVING;
    expect(derivePaletteStatus(state)).toBe(PALETTE_STATUS.SAVING);
    state.saveStatus = SAVE_STATUS.FAILED;
    expect(derivePaletteStatus(state)).toBe(PALETTE_STATUS.SAVE_FAILED);
  });

  it("toggles single and survey selections", () => {
    expect(toggleOptionSelection(QUESTION_TYPE_SINGLE, [], 1)).toEqual([1]);
    expect(toggleOptionSelection(QUESTION_TYPE_SINGLE, [1], 1)).toEqual([]);
    expect(toggleOptionSelection(QUESTION_TYPE_SURVEY, [1], 2)).toEqual([1, 2]);
    expect(toggleOptionSelection(QUESTION_TYPE_SURVEY, [1, 2], 2)).toEqual([1]);
  });

  it("ignores stale save responses", () => {
    const state = createInitialQuestionState("q-1", []);
    beginSave(state);
    applySaveResult(state, 1, [1], false);
    beginSave(state);
    const stale = applySaveResult(state, 1, [9], false);
    expect(stale.stale).toBe(true);
    expect(state.savedSelection).toEqual([1]);
  });
});
