import { describe, expect, it, vi } from "vitest";
import {
  buildAutosaveBody,
  createAutosaveRunner,
  createPerQuestionSaveQueues,
} from "@/composables/attempt_autosave";
import { createInitialQuestionState } from "@/composables/attempt_player_state";
import { SAVE_STATUS } from "@/utils/attempt_player_constants";

describe("attempt_autosave", () => {
  it("builds clear and selection bodies", () => {
    expect(buildAutosaveBody([])).toEqual({
      clear: true,
      selected_options: [],
    });
    expect(buildAutosaveBody([2])).toEqual({
      clear: false,
      selected_options: [2],
    });
  });

  it("serialises per-question saves without blocking other questions", async () => {
    const states = {
      q1: createInitialQuestionState("q1"),
      q2: createInitialQuestionState("q2"),
    };
    const calls = [];
    const runner = createAutosaveRunner({
      queues: createPerQuestionSaveQueues(),
      getQuestionState: (id) => states[id],
      saveRequest: (questionId) => {
        calls.push(questionId);
        return Promise.resolve({ selected_options: [1] });
      },
    });

    await Promise.all([
      runner.saveSelection("q1", [1]),
      runner.saveSelection("q2", [1]),
    ]);

    expect(calls.sort()).toEqual(["q1", "q2"]);
    expect(states.q1.savedSelection).toEqual([1]);
    expect(states.q2.savedSelection).toEqual([1]);
  });

  it("marks failures and preserves draft selection", async () => {
    const state = createInitialQuestionState("q1");
    state.draftSelection = [2];
    const runner = createAutosaveRunner({
      queues: createPerQuestionSaveQueues(),
      getQuestionState: () => state,
      saveRequest: () => Promise.reject(new Error("network")),
    });

    const result = await runner.saveSelection("q1", [2]);
    expect(result.ok).toBe(false);
    expect(state.saveStatus).toBe(SAVE_STATUS.FAILED);
    expect(state.draftSelection).toEqual([2]);
    expect(state.savedSelection).toEqual([]);
  });

  it("queues repeated saves for the same question", async () => {
    const state = createInitialQuestionState("q1");
    const saveRequest = vi
      .fn()
      .mockResolvedValueOnce({ selected_options: [1] })
      .mockResolvedValueOnce({ selected_options: [2] });
    const runner = createAutosaveRunner({
      queues: createPerQuestionSaveQueues(),
      getQuestionState: () => state,
      saveRequest,
    });

    const first = runner.saveSelection("q1", [1]);
    const second = runner.saveSelection("q1", [2]);
    await Promise.all([first, second]);

    expect(saveRequest).toHaveBeenCalledTimes(2);
    expect(state.savedSelection).toEqual([2]);
  });

  it("supports close() to reject new saves and flushAll() to await pending saves", async () => {
    const queues = createPerQuestionSaveQueues();
    let resolveSave;
    const savePromise = new Promise((resolve) => {
      resolveSave = resolve;
    });

    const runSave = queues.enqueue("q1", () => savePromise);
    expect(queues.isClosed()).toBe(false);

    queues.close();
    expect(queues.isClosed()).toBe(true);

    const blockedSave = await queues.enqueue("q1", () => Promise.resolve());
    expect(blockedSave).toEqual({ ok: false, reason: "queue-closed" });

    let flushed = false;
    const flushPromise = queues.flushAll().then(() => {
      flushed = true;
    });

    expect(flushed).toBe(false);
    resolveSave({ ok: true });
    await runSave;
    await flushPromise;
    expect(flushed).toBe(true);
  });
});
