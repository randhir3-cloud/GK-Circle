import { applySaveResult, beginSave } from "@/composables/attempt_player_state";

export const createPerQuestionSaveQueues = () => {
  const chains = new Map();
  let isClosed = false;

  const runChain = async (questionId, execute) => {
    if (isClosed) {
      return Promise.resolve({ ok: false, reason: "queue-closed" });
    }
    const previous = chains.get(questionId) || Promise.resolve();
    const next = previous.catch(() => undefined).then(() => execute());
    chains.set(
      questionId,
      next.finally(() => {
        if (chains.get(questionId) === next) chains.delete(questionId);
      })
    );
    return next;
  };

  return {
    enqueue(questionId, execute) {
      return runChain(questionId, execute);
    },
    close() {
      isClosed = true;
    },
    isClosed() {
      return isClosed;
    },
    flushAll() {
      return Promise.allSettled(Array.from(chains.values()));
    },
  };
};

export const buildAutosaveBody = (selection) => {
  if (!selection || selection.length === 0) {
    return { clear: true, selected_options: [] };
  }
  return { clear: false, selected_options: selection };
};

export const createAutosaveRunner = ({
  queues,
  saveRequest,
  getQuestionState,
}) => {
  const saveSelection = async (questionId, selection) => {
    const questionState = getQuestionState(questionId);
    if (!questionState) return { ok: false, reason: "missing-question" };

    const version = beginSave(questionState);
    const body = buildAutosaveBody(selection);

    return queues.enqueue(questionId, async () => {
      try {
        const response = await saveRequest(questionId, body);
        applySaveResult(
          questionState,
          version,
          response?.selected_options ?? selection,
          false
        );
        return { ok: true, version, response };
      } catch (error) {
        applySaveResult(
          questionState,
          version,
          questionState.savedSelection,
          true
        );
        return { ok: false, version, error };
      }
    });
  };

  return { saveSelection };
};
