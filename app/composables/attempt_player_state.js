import {
  ATTEMPT_STATUS_IN_PROGRESS,
  PALETTE_STATUS,
  QUESTION_TYPE_SURVEY,
  SAVE_STATUS,
} from "@/utils/attempt_player_constants";

const sortByPosition = (items) =>
  [...items].sort(
    (left, right) => Number(left.position) - Number(right.position)
  );

export const buildOrderedSnapshotItems = (resume) => {
  const items = resume?.snapshot?.items || [];
  if (items.length === 0) return [];
  const order = resume?.question_order || [];
  if (order.length === 0) return sortByPosition(items);
  const byId = new Map(items.map((item) => [item.question_id, item]));
  const ordered = [];
  for (const questionId of order) {
    const item = byId.get(questionId);
    if (item) ordered.push(item);
  }
  if (ordered.length === items.length) return ordered;
  const seen = new Set(ordered.map((item) => item.question_id));
  for (const item of sortByPosition(items)) {
    if (!seen.has(item.question_id)) ordered.push(item);
  }
  return ordered;
};

export const createInitialQuestionState = (
  questionId,
  savedSelection = []
) => ({
  questionId,
  visited: false,
  savedSelection: [...savedSelection],
  draftSelection: [...savedSelection],
  saveStatus: SAVE_STATUS.IDLE,
  saveError: "",
  saveVersion: 0,
  inFlightVersion: 0,
});

export const hydrateQuestionStatesFromResume = (resume) => {
  const items = buildOrderedSnapshotItems(resume);
  const answersByQuestion = new Map(
    (resume?.answers || []).map((answer) => [
      answer.question_id,
      answer.selected_options || [],
    ])
  );
  const states = {};
  for (const item of items) {
    states[item.question_id] = createInitialQuestionState(
      item.question_id,
      answersByQuestion.get(item.question_id) || []
    );
  }
  return { items, states };
};

export const derivePaletteStatus = (questionState) => {
  if (!questionState) return PALETTE_STATUS.NOT_VISITED;
  if (questionState.saveStatus === SAVE_STATUS.SAVING) {
    return PALETTE_STATUS.SAVING;
  }
  if (questionState.saveStatus === SAVE_STATUS.FAILED) {
    return PALETTE_STATUS.SAVE_FAILED;
  }
  if (!questionState.visited) return PALETTE_STATUS.NOT_VISITED;
  if ((questionState.savedSelection || []).length > 0) {
    return PALETTE_STATUS.ANSWERED;
  }
  return PALETTE_STATUS.VISITED_UNANSWERED;
};

export const isSelectionDirty = (questionState) => {
  const saved = [...(questionState.savedSelection || [])].sort((a, b) => a - b);
  const draft = [...(questionState.draftSelection || [])].sort((a, b) => a - b);
  if (saved.length !== draft.length) return true;
  return saved.some((value, index) => value !== draft[index]);
};

export const markQuestionVisited = (questionState) => {
  questionState.visited = true;
};

export const beginSave = (questionState) => {
  questionState.saveVersion += 1;
  questionState.inFlightVersion = questionState.saveVersion;
  questionState.saveStatus = SAVE_STATUS.SAVING;
  questionState.saveError = "";
  return questionState.inFlightVersion;
};

export const applySaveResult = (questionState, version, selection, failed) => {
  if (version < questionState.inFlightVersion) {
    return { applied: false, stale: true };
  }
  if (failed) {
    questionState.saveStatus = SAVE_STATUS.FAILED;
    return { applied: true, stale: false };
  }
  questionState.savedSelection = [...selection];
  questionState.draftSelection = [...selection];
  questionState.saveStatus = SAVE_STATUS.SAVED;
  questionState.saveError = "";
  return { applied: true, stale: false };
};

export const setDraftSelection = (questionState, selection) => {
  questionState.draftSelection = [...selection];
};

export const toggleOptionSelection = (
  questionType,
  currentSelection,
  optionKey
) => {
  const key = Number(optionKey);
  if (questionType === QUESTION_TYPE_SURVEY) {
    const set = new Set(currentSelection);
    if (set.has(key)) set.delete(key);
    else set.add(key);
    return [...set].sort((a, b) => a - b);
  }
  if (currentSelection.length === 1 && currentSelection[0] === key) {
    return [];
  }
  return [key];
};

export const canOpenPlayer = (resume) =>
  resume?.status === ATTEMPT_STATUS_IN_PROGRESS;

export const countAnsweredQuestions = (questionStates) =>
  Object.values(questionStates).filter(
    (state) => (state.savedSelection || []).length > 0
  ).length;

export const assertLearnerSafePayload = (payload) => {
  const serialized = JSON.stringify(payload || {});
  for (const field of [
    "official_answer",
    "authoritative_answer",
    "answer_review_status",
    '"is_correct"',
    '"score"',
  ]) {
    if (serialized.includes(field)) {
      throw new Error(`learner-unsafe field detected: ${field}`);
    }
  }
};
