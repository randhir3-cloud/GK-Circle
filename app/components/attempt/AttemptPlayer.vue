<script setup>
import { computed, onBeforeUnmount, reactive, ref, watch } from "vue";
import AttemptQuestionPalette from "@/components/attempt/AttemptQuestionPalette.vue";
import AttemptQuestionPanel from "@/components/attempt/AttemptQuestionPanel.vue";
import AttemptSubmitDialog from "@/components/attempt/AttemptSubmitDialog.vue";
import AttemptTimer from "@/components/attempt/AttemptTimer.vue";
import {
  createAutosaveRunner,
  createPerQuestionSaveQueues,
} from "@/composables/attempt_autosave";
import {
  derivePaletteStatus,
  hydrateQuestionStatesFromResume,
  markQuestionVisited,
  setDraftSelection,
  toggleOptionSelection,
} from "@/composables/attempt_player_state";
import {
  getAssessmentAttemptAPIError,
  useAssessmentAttemptsApi,
} from "@/composables/assessment_attempts";
import {
  ATTEMPT_STATUS_AUTO_SUBMITTED,
  ATTEMPT_STATUS_IN_PROGRESS,
  ATTEMPT_STATUS_SUBMITTED,
  SUBMIT_PHASE,
} from "@/utils/attempt_player_constants";

const props = defineProps({
  resume: {
    type: Object,
    required: true,
  },
  quizId: {
    type: String,
    required: true,
  },
  attemptId: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(["submitted"]);

const api = useAssessmentAttemptsApi();
const currentIndex = ref(0);
const terminalMessage = ref("");
const showSubmitDialog = ref(false);
const submitPhase = ref(SUBMIT_PHASE.IDLE);
let onlineListener = null;

const playerDisabled = computed(
  () =>
    props.resume?.status !== ATTEMPT_STATUS_IN_PROGRESS ||
    submitPhase.value !== SUBMIT_PHASE.IDLE
);

const hydrated = hydrateQuestionStatesFromResume(props.resume);
const orderedItems = ref(hydrated.items);
const questionStates = reactive(hydrated.states);

const saveQueues = createPerQuestionSaveQueues();
const autosave = createAutosaveRunner({
  queues: saveQueues,
  getQuestionState: (questionId) => questionStates[questionId],
  saveRequest: async (questionId, body) =>
    api.autosaveAnswer(props.quizId, props.attemptId, questionId, body),
});

const currentItem = computed(
  () => orderedItems.value[currentIndex.value] || null
);
const currentState = computed(() =>
  currentItem.value ? questionStates[currentItem.value.question_id] : null
);

const paletteStatusFor = (questionId) =>
  derivePaletteStatus(questionStates[questionId]);

const answeredCount = computed(
  () =>
    Object.values(questionStates).filter(
      (state) => (state.savedSelection || []).length > 0
    ).length
);

const unansweredCount = computed(
  () => orderedItems.value.length - answeredCount.value
);

const progressLabel = computed(
  () => `${answeredCount.value} of ${orderedItems.value.length} answered`
);

const ensureVisited = (index) => {
  const item = orderedItems.value[index];
  if (!item) return;
  markQuestionVisited(questionStates[item.question_id]);
};

const goToIndex = (index) => {
  if (index < 0 || index >= orderedItems.value.length) return;
  ensureVisited(index);
  currentIndex.value = index;
};

const goPrevious = () => goToIndex(currentIndex.value - 1);
const goNext = () => goToIndex(currentIndex.value + 1);

const persistSelection = async (questionId, selection) => {
  if (playerDisabled.value) {
    terminalMessage.value = "This attempt is no longer in progress.";
    return;
  }
  const result = await autosave.saveSelection(questionId, selection);
  if (!result.ok) {
    const message = getAssessmentAttemptAPIError(
      result.error,
      "Autosave failed"
    );
    questionStates[questionId].saveError = message;
    if (message.toLowerCase().includes("not in progress")) {
      terminalMessage.value = message;
    }
  }
};

const handleToggleOption = async (optionKey) => {
  const item = currentItem.value;
  const state = currentState.value;
  if (!item || !state || playerDisabled.value) return;
  ensureVisited(currentIndex.value);
  const nextSelection = toggleOptionSelection(
    item.type,
    state.draftSelection,
    optionKey
  );
  setDraftSelection(state, nextSelection);
  await persistSelection(item.question_id, nextSelection);
};

const handleClearAnswer = async () => {
  const item = currentItem.value;
  const state = currentState.value;
  if (!item || !state || playerDisabled.value) return;
  ensureVisited(currentIndex.value);
  setDraftSelection(state, []);
  await persistSelection(item.question_id, []);
};

const handleRetrySave = async () => {
  const item = currentItem.value;
  const state = currentState.value;
  if (!item || !state) return;
  await persistSelection(item.question_id, state.draftSelection);
};

const removeOnlineListener = () => {
  if (onlineListener && typeof window !== "undefined") {
    window.removeEventListener("online", onlineListener);
    onlineListener = null;
  }
};

const executeSubmit = async () => {
  showSubmitDialog.value = false;
  submitPhase.value = SUBMIT_PHASE.SUBMIT_REQUESTED;

  // 1. Close queues so no new answer changes are enqueued
  saveQueues.close();
  submitPhase.value = SUBMIT_PHASE.QUEUE_CLOSING;

  // 2. Await in-flight saves
  await saveQueues.flushAll();

  // 3. Post submission
  submitPhase.value = SUBMIT_PHASE.SUBMITTING;
  try {
    const result = await api.submitAttempt(props.quizId, props.attemptId);
    submitPhase.value = SUBMIT_PHASE.SUBMITTED;
    removeOnlineListener();
    emit("submitted", result);
  } catch (submitError) {
    const isNetworkError =
      !navigator.onLine ||
      submitError?.message?.includes("fetch") ||
      submitError?.name === "AbortError" ||
      !submitError?.status;

    if (isNetworkError) {
      submitPhase.value = SUBMIT_PHASE.OFFLINE_EXPIRED;
      terminalMessage.value =
        "Connection lost during submission. Waiting to reconnect to complete submission.";
      if (typeof window !== "undefined" && !onlineListener) {
        onlineListener = () => {
          terminalMessage.value = "Reconnected! Retrying submission…";
          executeSubmit();
        };
        window.addEventListener("online", onlineListener);
      }
    } else {
      submitPhase.value = SUBMIT_PHASE.ERROR;
      terminalMessage.value = getAssessmentAttemptAPIError(
        submitError,
        "Failed to submit attempt. Please retry."
      );
    }
  }
};

const handleTimerExpired = () => {
  if (submitPhase.value !== SUBMIT_PHASE.IDLE) return;
  terminalMessage.value = "Time expired! Submitting attempt automatically…";
  executeSubmit();
};

const handleTerminalStatus = (status) => {
  if (
    status === ATTEMPT_STATUS_SUBMITTED ||
    status === ATTEMPT_STATUS_AUTO_SUBMITTED
  ) {
    submitPhase.value = SUBMIT_PHASE.SUBMITTED;
    emit("submitted", { id: props.attemptId, status });
  }
};

watch(
  () => props.resume,
  (resume) => {
    const next = hydrateQuestionStatesFromResume(resume);
    orderedItems.value = next.items;
    for (const key of Object.keys(questionStates)) {
      delete questionStates[key];
    }
    Object.assign(questionStates, next.states);
    currentIndex.value = 0;
    terminalMessage.value = "";
    submitPhase.value = SUBMIT_PHASE.IDLE;
  },
  { deep: true }
);

onBeforeUnmount(() => {
  removeOnlineListener();
});

ensureVisited(0);
</script>

<template>
  <div class="attempt-player">
    <header class="attempt-player__header">
      <div>
        <p class="attempt-player__eyebrow">Self-paced attempt</p>
        <h1 class="attempt-player__title">Answer questions</h1>
      </div>
      <div class="attempt-player__header-actions">
        <AttemptTimer
          :expires-at="resume?.expires_at"
          :quiz-id="quizId"
          :attempt-id="attemptId"
          @expired="handleTimerExpired"
          @terminal-status="handleTerminalStatus"
        />
        <p class="attempt-player__progress" aria-live="polite">
          {{ progressLabel }}
        </p>
      </div>
    </header>

    <p v-if="terminalMessage" class="attempt-player__terminal" role="alert">
      {{ terminalMessage }}
    </p>

    <p
      v-if="submitPhase === SUBMIT_PHASE.OFFLINE_EXPIRED"
      class="attempt-player__offline-banner"
      role="alert"
    >
      Time expired. Waiting to reconnect to complete submission.
    </p>

    <AttemptQuestionPalette
      :items="orderedItems"
      :current-index="currentIndex"
      :palette-status-for="paletteStatusFor"
      @select-index="goToIndex"
    />

    <AttemptQuestionPanel
      v-if="currentItem && currentState"
      :item="currentItem"
      :index="currentIndex"
      :total="orderedItems.length"
      :draft-selection="currentState.draftSelection"
      :save-status="currentState.saveStatus"
      :save-error="currentState.saveError"
      :disabled="playerDisabled"
      @toggle-option="handleToggleOption"
      @clear-answer="handleClearAnswer"
      @retry-save="handleRetrySave"
    />

    <nav
      class="attempt-player__nav"
      aria-label="Question navigation and submission"
    >
      <button
        type="button"
        class="attempt-player__nav-button"
        :disabled="currentIndex === 0 || playerDisabled"
        @click="goPrevious"
      >
        Previous
      </button>

      <button
        type="button"
        class="attempt-player__nav-button attempt-player__nav-button--submit"
        :disabled="playerDisabled"
        @click="showSubmitDialog = true"
      >
        {{
          submitPhase === SUBMIT_PHASE.SUBMITTING ||
          submitPhase === SUBMIT_PHASE.QUEUE_CLOSING
            ? "Submitting…"
            : "Submit Attempt"
        }}
      </button>

      <button
        type="button"
        class="attempt-player__nav-button attempt-player__nav-button--primary"
        :disabled="currentIndex >= orderedItems.length - 1 || playerDisabled"
        @click="goNext"
      >
        Next
      </button>
    </nav>

    <AttemptSubmitDialog
      v-if="showSubmitDialog"
      :answered-count="answeredCount"
      :unanswered-count="unansweredCount"
      :total-questions="orderedItems.length"
      :submitting="submitPhase !== SUBMIT_PHASE.IDLE"
      @cancel="showSubmitDialog = false"
      @confirm="executeSubmit"
    />
  </div>
</template>

<style scoped>
.attempt-player {
  display: grid;
  gap: 1.25rem;
  max-width: 48rem;
  margin: 0 auto;
  padding: 1.25rem 1rem 2rem;
  color: #12263a;
  font-family: "Source Sans 3", "Segoe UI", sans-serif;
}

.attempt-player__header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: flex-start;
  flex-wrap: wrap;
}

.attempt-player__header-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.5rem;
}

.attempt-player__eyebrow {
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 0.75rem;
  color: #4a5d73;
}

.attempt-player__title {
  margin: 0.35rem 0 0;
  font-family: "Literata", "Georgia", serif;
  font-size: clamp(1.35rem, 4vw, 1.8rem);
}

.attempt-player__progress {
  margin: 0;
  font-weight: 600;
  color: #4a5d73;
}

.attempt-player__terminal {
  margin: 0;
  padding: 0.75rem 0.9rem;
  border-radius: 0.35rem;
  background: #fdecea;
  color: #8a2f1d;
}

.attempt-player__offline-banner {
  margin: 0;
  padding: 0.85rem 1rem;
  border-radius: 0.35rem;
  background: #fff8e1;
  border: 1px solid #ffe082;
  color: #896b00;
  font-weight: 600;
}

.attempt-player__nav {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.attempt-player__nav-button {
  appearance: none;
  border: 1px solid #b8c6d4;
  border-radius: 0.35rem;
  background: #fff;
  color: #12263a;
  font: inherit;
  font-weight: 600;
  padding: 0.7rem 1rem;
  cursor: pointer;
}

.attempt-player__nav-button--submit {
  background: #27ae60;
  border-color: #27ae60;
  color: #ffffff;
}

.attempt-player__nav-button--submit:hover:not(:disabled) {
  background: #219653;
}

.attempt-player__nav-button--primary {
  background: #0f6a5a;
  border-color: #0f6a5a;
  color: #fff;
}

.attempt-player__nav-button:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.attempt-player__nav-button:focus-visible {
  outline: 2px solid #12263a;
  outline-offset: 2px;
}

@media (max-width: 480px) {
  .attempt-player__header-actions {
    align-items: flex-start;
    width: 100%;
  }

  .attempt-player__nav {
    flex-direction: column;
  }

  .attempt-player__nav-button {
    width: 100%;
  }
}
</style>
