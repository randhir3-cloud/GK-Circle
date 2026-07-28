<script setup>
import { computed, onBeforeUnmount, ref, watch } from "vue";
import AttemptResultHeader from "@/components/attempt/AttemptResultHeader.vue";
import AttemptResultPending from "@/components/attempt/AttemptResultPending.vue";
import AttemptScoreCard from "@/components/attempt/AttemptScoreCard.vue";
import AttemptSummary from "@/components/attempt/AttemptSummary.vue";
import AttemptQuestionReview from "@/components/attempt/AttemptQuestionReview.vue";
import {
  getAssessmentAttemptAPIError,
  useAssessmentAttemptsApi,
} from "@/composables/assessment_attempts";
import { setUserDataStore } from "@/composables/auth";
import { useUsersStore } from "~~/store/users";

definePageMeta({ layout: "empty" });

const route = useRoute();
const api = useAssessmentAttemptsApi();
const usersStore = useUsersStore();

const quizId = computed(() => String(route.params.quiz_id || ""));
const attemptId = computed(() => String(route.params.attempt_id || ""));

const resultData = ref(null);
const loading = ref(true);
const error = ref("");
let requestGen = 0;

useSeoMeta({
  title: "Attempt Result - GK Circle",
  robots: "noindex, nofollow",
});

const instructionsPath = computed(() => {
  return `/attempt/quizzes/${encodeURIComponent(quizId.value)}`;
});

const loadResult = async () => {
  const gen = ++requestGen;
  resultData.value = null;
  error.value = "";
  loading.value = true;

  if (!quizId.value || !attemptId.value) {
    error.value = "Attempt details are incomplete.";
    loading.value = false;
    return;
  }

  try {
    const res = await api.getAttemptResult(quizId.value, attemptId.value);
    if (gen !== requestGen) return;
    resultData.value = res;
  } catch (errRes) {
    if (gen !== requestGen) return;
    error.value = getAssessmentAttemptAPIError(
      errRes,
      "Unable to load assessment result."
    );
  } finally {
    if (gen === requestGen) loading.value = false;
  }
};

watch(
  () => [quizId.value, attemptId.value],
  async () => {
    try {
      await setUserDataStore(usersStore);
    } catch {
      // API enforces authentication
    }
    await loadResult();
  },
  { immediate: true }
);

onBeforeUnmount(() => {
  requestGen++;
});
</script>

<template>
  <main class="result-page">
    <div class="result-page__container">
      <p v-if="loading" class="result-page__status" role="status">
        Loading your assessment results…
      </p>

      <section v-else-if="error" class="result-page__error-panel">
        <p class="result-page__error" role="alert">{{ error }}</p>
        <NuxtLink class="result-page__back" :to="instructionsPath">
          Back to instructions
        </NuxtLink>
      </section>

      <div v-else-if="resultData" class="result-page__content">
        <!-- Result Release Pending Screen -->
        <AttemptResultPending
          v-if="resultData.can_view_result === false"
          :attempt-id="resultData.attempt_id"
          :submitted-at="resultData.submitted_at || ''"
          :message="
            resultData.message ||
            'Results for this assessment have not been released yet.'
          "
          :instructions-path="instructionsPath"
        />

        <template v-else>
          <!-- Header -->
          <AttemptResultHeader
            :attempt-id="resultData.attempt_id"
            :status="resultData.status"
            :submitted-at="resultData.submitted_at || ''"
            :instructions-path="instructionsPath"
          />

          <!-- Score Card -->
          <AttemptScoreCard
            v-if="resultData.summary"
            :summary="resultData.summary"
            :can-show-score="resultData.can_show_score !== false"
            :can-show-pass-fail="resultData.can_show_pass_fail !== false"
            class="result-page__section"
          />

          <!-- Performance Summary Grid -->
          <AttemptSummary
            v-if="resultData.summary"
            :summary="resultData.summary"
            class="result-page__section"
          />

          <!-- Question Review (or Review Disabled Notice) -->
          <AttemptQuestionReview
            v-if="
              resultData.can_review_questions && resultData.review?.questions
            "
            :questions="resultData.review.questions"
            :can-show-correctness="resultData.can_show_correctness !== false"
            :can-show-explanations="resultData.can_show_explanations !== false"
            class="result-page__section"
          />

          <div
            v-else-if="!resultData.can_review_questions"
            class="result-page__review-disabled"
          >
            <p class="result-page__notice">
              Review is unavailable for this assessment.
            </p>
          </div>
        </template>
      </div>
    </div>
  </main>
</template>

<style scoped>
.result-page {
  min-height: 100vh;
  padding: 2rem 1rem;
  background: radial-gradient(
      circle at top right,
      rgba(15, 106, 90, 0.08),
      transparent 40%
    ),
    linear-gradient(180deg, #eef3f7 0%, #ffffff 70%);
}

.result-page__container {
  max-width: 56rem;
  margin: 0 auto;
}

.result-page__content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.result-page__status,
.result-page__error-panel {
  max-width: 40rem;
  margin: 4rem auto;
  padding: 2.5rem 1.25rem;
  font-family: "Source Sans 3", "Segoe UI", sans-serif;
  text-align: center;
}

.result-page__error {
  margin: 0;
  color: #8a2f1d;
  font-weight: 600;
}

.result-page__back {
  display: inline-block;
  margin-top: 1rem;
  color: #0f6a5a;
  font-weight: 600;
  text-decoration: none;
}

.result-page__withheld,
.result-page__review-disabled {
  padding: 2rem;
  background: #ffffff;
  border-radius: 0.75rem;
  border: 1px solid #e5e7eb;
  text-align: center;
}

.result-page__withheld-msg,
.result-page__notice {
  margin: 0;
  font-size: 1rem;
  color: #4b5563;
  font-weight: 500;
}
</style>
