<script setup>
import { computed, onBeforeUnmount, ref, watch } from "vue";
import AttemptInstructionsPanel from "@/components/attempt/AttemptInstructionsPanel.vue";
import {
  getAssessmentAttemptAPIError,
  useAssessmentAttemptsApi,
} from "@/composables/assessment_attempts";
import { setUserDataStore } from "@/composables/auth";
import { useUsersStore } from "~~/store/users";

definePageMeta({ layout: "empty" });

const route = useRoute();
const router = useRouter();
const api = useAssessmentAttemptsApi();
const usersStore = useUsersStore();

const quizId = computed(() => String(route.params.quiz_id || ""));
const snapshotId = computed(() => String(route.query.snapshot_id || ""));

const instructions = ref(null);
const loading = ref(true);
const error = ref("");
const actionError = ref("");
const starting = ref(false);
const resuming = ref(false);
let requestGeneration = 0;

useSeoMeta({
  title: computed(() =>
    instructions.value?.quiz?.title
      ? `${instructions.value.quiz.title} - Start attempt - GK Circle`
      : "Start attempt - GK Circle"
  ),
  robots: "noindex, nofollow",
});

const loadInstructions = async () => {
  const generation = ++requestGeneration;
  instructions.value = null;
  error.value = "";
  actionError.value = "";
  loading.value = true;

  if (!quizId.value || !snapshotId.value) {
    error.value = "A quiz and snapshot are required to open instructions.";
    loading.value = false;
    return;
  }

  try {
    const result = await api.getInstructions(quizId.value, snapshotId.value);
    if (generation !== requestGeneration) return;
    instructions.value = result;
  } catch (requestError) {
    if (generation !== requestGeneration) return;
    error.value = getAssessmentAttemptAPIError(
      requestError,
      "Unable to load test instructions."
    );
  } finally {
    if (generation === requestGeneration) loading.value = false;
  }
};

const goToAttempt = (attemptId) => {
  router.push({
    path: `/attempt/quizzes/${encodeURIComponent(
      quizId.value
    )}/attempts/${encodeURIComponent(attemptId)}`,
  });
};

const startAttempt = async () => {
  starting.value = true;
  actionError.value = "";
  try {
    const attempt = await api.createAttempt(quizId.value, snapshotId.value);
    goToAttempt(attempt.id);
  } catch (requestError) {
    actionError.value = getAssessmentAttemptAPIError(
      requestError,
      "Unable to start this attempt."
    );
  } finally {
    starting.value = false;
  }
};

const resumeAttempt = async (attemptId) => {
  resuming.value = true;
  actionError.value = "";
  try {
    // Confirm ownership via resume contract before entering the shell.
    await api.resumeAttempt(quizId.value, attemptId);
    goToAttempt(attemptId);
  } catch (requestError) {
    actionError.value = getAssessmentAttemptAPIError(
      requestError,
      "Unable to resume this attempt."
    );
  } finally {
    resuming.value = false;
  }
};

watch(
  () => [quizId.value, snapshotId.value],
  async () => {
    try {
      await setUserDataStore(usersStore);
    } catch {
      // API auth still enforces Kratos; page shows API errors if unauthenticated.
    }
    await loadInstructions();
  },
  { immediate: true }
);

onBeforeUnmount(() => {
  requestGeneration += 1;
});
</script>

<template>
  <main class="attempt-launch">
    <p v-if="loading" class="attempt-launch__status" role="status">
      Loading instructions…
    </p>
    <p v-else-if="error" class="attempt-launch__error" role="alert">
      {{ error }}
    </p>
    <AttemptInstructionsPanel
      v-else-if="instructions"
      :instructions="instructions"
      :starting="starting"
      :resuming="resuming"
      :action-error="actionError"
      @start="startAttempt"
      @resume="resumeAttempt"
    />
  </main>
</template>

<style scoped>
.attempt-launch {
  min-height: 100vh;
  background: radial-gradient(
      circle at top right,
      rgba(15, 106, 90, 0.08),
      transparent 40%
    ),
    linear-gradient(180deg, #eef3f7 0%, #f8fafc 55%, #ffffff 100%);
}

.attempt-launch__status,
.attempt-launch__error {
  max-width: 40rem;
  margin: 0 auto;
  padding: 3rem 1.25rem;
  font-family: "Source Sans 3", "Segoe UI", sans-serif;
}

.attempt-launch__error {
  color: #8a2f1d;
}
</style>
