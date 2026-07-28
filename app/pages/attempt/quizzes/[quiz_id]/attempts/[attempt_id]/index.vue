<script setup>
import { computed, onBeforeUnmount, ref, watch } from "vue";
import AttemptPlayer from "@/components/attempt/AttemptPlayer.vue";
import AttemptSubmittedScreen from "@/components/attempt/AttemptSubmittedScreen.vue";
import {
  getAssessmentAttemptAPIError,
  useAssessmentAttemptsApi,
} from "@/composables/assessment_attempts";
import { canOpenPlayer } from "@/composables/attempt_player_state";
import { setUserDataStore } from "@/composables/auth";
import {
  ATTEMPT_STATUS_AUTO_SUBMITTED,
  ATTEMPT_STATUS_SUBMITTED,
} from "@/utils/attempt_player_constants";
import { useUsersStore } from "~~/store/users";

definePageMeta({ layout: "empty" });

const route = useRoute();
const api = useAssessmentAttemptsApi();
const usersStore = useUsersStore();

const quizId = computed(() => String(route.params.quiz_id || ""));
const attemptId = computed(() => String(route.params.attempt_id || ""));

const resume = ref(null);
const submittedResult = ref(null);
const loading = ref(true);
const error = ref("");
let requestGeneration = 0;

useSeoMeta({
  title: "Attempt - GK Circle",
  robots: "noindex, nofollow",
});

const instructionsPath = computed(() => {
  const snapshotId = resume.value?.test_snapshot_id;
  const base = `/attempt/quizzes/${encodeURIComponent(quizId.value)}`;
  return snapshotId
    ? `${base}?snapshot_id=${encodeURIComponent(snapshotId)}`
    : base;
});

const isSubmitted = computed(
  () =>
    Boolean(submittedResult.value) ||
    resume.value?.status === ATTEMPT_STATUS_SUBMITTED ||
    resume.value?.status === ATTEMPT_STATUS_AUTO_SUBMITTED
);

const submittedStatus = computed(
  () =>
    submittedResult.value?.status ||
    resume.value?.status ||
    ATTEMPT_STATUS_SUBMITTED
);

const submittedTimestamp = computed(
  () => submittedResult.value?.submitted_at || resume.value?.submitted_at || ""
);

const submittedSummary = computed(
  () =>
    submittedResult.value?.summary ||
    (resume.value?.progress
      ? {
          answered_count: resume.value.progress.answered_count,
          total_questions: resume.value.progress.total_questions,
        }
      : null)
);

const handlePlayerSubmitted = (result) => {
  submittedResult.value = result;
};

const loadResume = async () => {
  const generation = ++requestGeneration;
  resume.value = null;
  submittedResult.value = null;
  error.value = "";
  loading.value = true;

  if (!quizId.value || !attemptId.value) {
    error.value = "Attempt details are incomplete.";
    loading.value = false;
    return;
  }

  try {
    const result = await api.resumeAttempt(quizId.value, attemptId.value);
    if (generation !== requestGeneration) return;
    resume.value = result;
    if (
      result?.status === ATTEMPT_STATUS_SUBMITTED ||
      result?.status === ATTEMPT_STATUS_AUTO_SUBMITTED
    ) {
      // Handled by AttemptSubmittedScreen template branch
    } else if (!canOpenPlayer(result) && result?.status) {
      error.value = `This attempt is ${result.status.toLowerCase()} and can no longer be edited.`;
    }
  } catch (requestError) {
    if (generation !== requestGeneration) return;
    error.value = getAssessmentAttemptAPIError(
      requestError,
      "Unable to resume this attempt."
    );
  } finally {
    if (generation === requestGeneration) loading.value = false;
  }
};

watch(
  () => [quizId.value, attemptId.value],
  async () => {
    try {
      await setUserDataStore(usersStore);
    } catch {
      // API enforces authentication.
    }
    await loadResume();
  },
  { immediate: true }
);

onBeforeUnmount(() => {
  requestGeneration += 1;
});
</script>

<template>
  <main class="attempt-page">
    <p v-if="loading" class="attempt-page__status" role="status">
      Loading your attempt…
    </p>

    <AttemptSubmittedScreen
      v-else-if="isSubmitted"
      :attempt-id="attemptId"
      :status="submittedStatus"
      :submitted-at="submittedTimestamp"
      :summary="submittedSummary"
      :instructions-path="instructionsPath"
    />

    <section v-else-if="error" class="attempt-page__error-panel">
      <p class="attempt-page__error" role="alert">{{ error }}</p>
      <NuxtLink class="attempt-page__back" :to="instructionsPath">
        Back to instructions
      </NuxtLink>
    </section>

    <AttemptPlayer
      v-else-if="resume && canOpenPlayer(resume)"
      :resume="resume"
      :quiz-id="quizId"
      :attempt-id="attemptId"
      @submitted="handlePlayerSubmitted"
    />
  </main>
</template>

<style scoped>
.attempt-page {
  min-height: 100vh;
  background: radial-gradient(
      circle at top right,
      rgba(15, 106, 90, 0.08),
      transparent 40%
    ),
    linear-gradient(180deg, #eef3f7 0%, #ffffff 70%);
}

.attempt-page__status,
.attempt-page__error-panel {
  max-width: 40rem;
  margin: 0 auto;
  padding: 2.5rem 1.25rem;
  font-family: "Source Sans 3", "Segoe UI", sans-serif;
}

.attempt-page__error {
  margin: 0;
  color: #8a2f1d;
}

.attempt-page__back {
  display: inline-block;
  margin-top: 1rem;
  color: #0f6a5a;
  font-weight: 600;
  text-decoration: none;
}
</style>
