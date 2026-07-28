<script setup>
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import {
  getInstructorAnalyticsAPIError,
  useInstructorAnalyticsApi,
} from "@/composables/instructor_analytics";
import QuizCohortSummary from "@/components/instructor-analytics/QuizCohortSummary.vue";
import QuizAttemptTable from "@/components/instructor-analytics/QuizAttemptTable.vue";
import QuestionQualityTable from "@/components/instructor-analytics/QuestionQualityTable.vue";
import QuizEngagementPanel from "@/components/instructor-analytics/QuizEngagementPanel.vue";

definePageMeta({ layout: "empty" });
useSeoMeta({
  title: "Quiz Analytics - GK Circle",
  description:
    "Detailed quiz cohort performance, questions quality, and engagement statistics.",
});

const route = useRoute();
const api = useInstructorAnalyticsApi();

const quizId = route.params.quiz_id;

const activeTab = ref("attempts");
const loading = ref(true);
const error = ref("");

const summary = ref(null);

const attempts = ref([]);
const attemptCursor = ref("");
const attemptHasMore = ref(false);
const loadingAttempts = ref(false);

const questions = ref([]);
const questionCursor = ref("");
const questionHasMore = ref(false);
const loadingQuestions = ref(false);

const engagement = ref(null);

const loadSummary = async () => {
  try {
    summary.value = await api.getQuizSummary(quizId);
  } catch (err) {
    error.value = getInstructorAnalyticsAPIError(
      err,
      "Failed to load quiz summary"
    );
  }
};

const loadAttempts = async (isMore = false) => {
  loadingAttempts.value = true;
  try {
    const res = await api.getQuizAttempts(quizId, {
      cursor: isMore ? attemptCursor.value : "",
    });
    if (isMore) {
      attempts.value = [...attempts.value, ...(res.attempts || [])];
    } else {
      attempts.value = res.attempts || [];
    }
    attemptCursor.value = res.next_cursor || "";
    attemptHasMore.value = !!res.has_more;
  } catch (err) {
    error.value = getInstructorAnalyticsAPIError(
      err,
      "Failed to load quiz attempts"
    );
  } finally {
    loadingAttempts.value = false;
  }
};

const loadQuestions = async (isMore = false) => {
  loadingQuestions.value = true;
  try {
    const res = await api.getQuestionMetrics(quizId, {
      cursor: isMore ? questionCursor.value : "",
    });
    if (isMore) {
      questions.value = [...questions.value, ...(res.questions || [])];
    } else {
      questions.value = res.questions || [];
    }
    questionCursor.value = res.next_cursor || "";
    questionHasMore.value = !!res.has_more;
  } catch (err) {
    error.value = getInstructorAnalyticsAPIError(
      err,
      "Failed to load question quality metrics"
    );
  } finally {
    loadingQuestions.value = false;
  }
};

const loadEngagement = async () => {
  try {
    engagement.value = await api.getEngagement(quizId);
  } catch (err) {
    error.value = getInstructorAnalyticsAPIError(
      err,
      "Failed to load engagement metrics"
    );
  }
};

onMounted(async () => {
  loading.value = true;
  await Promise.all([
    loadSummary(),
    loadAttempts(false),
    loadQuestions(false),
    loadEngagement(),
  ]);
  loading.value = false;
});
</script>

<template>
  <div
    class="min-h-screen bg-background text-text p-4 sm:p-6 lg:p-8 min-w-0 max-w-full"
  >
    <div class="max-w-7xl mx-auto space-y-6 min-w-0 max-w-full">
      <!-- Header -->
      <div
        class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border pb-4"
      >
        <div>
          <h1 class="text-2xl font-bold text-text tracking-tight">
            {{ summary?.title || "Quiz Analytics" }}
          </h1>
          <p class="text-sm text-text-secondary mt-1">
            Per-quiz attempt analysis, question difficulty, and telemetry
            engagement.
          </p>
        </div>
        <NuxtLink
          to="/instructor/analytics"
          class="px-4 py-2 text-sm font-medium border border-border rounded-md hover:bg-surface-hover transition-colors inline-flex items-center gap-1.5 self-start sm:self-auto"
        >
          ← Portfolio Overview
        </NuxtLink>
      </div>

      <!-- Error Alert -->
      <div
        v-if="error"
        class="p-4 bg-danger/10 text-danger rounded-lg border border-danger/20 text-sm"
      >
        {{ error }}
      </div>

      <!-- Cohort Summary Header Cards -->
      <QuizCohortSummary :summary="summary" />

      <!-- Navigation Tabs -->
      <div
        class="flex items-center gap-2 border-b border-border overflow-x-auto min-w-0 max-w-full pb-1"
      >
        <button
          v-for="tab in [
            { key: 'attempts', label: 'Learner Attempts' },
            { key: 'questions', label: 'Question Quality' },
            { key: 'engagement', label: 'Telemetry Engagement' },
          ]"
          :key="tab.key"
          class="px-4 py-2 text-sm font-medium border-b-2 transition-colors whitespace-nowrap"
          :class="
            activeTab === tab.key
              ? 'border-primary text-primary font-semibold'
              : 'border-transparent text-text-secondary hover:text-text'
          "
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Tab Content -->
      <div v-if="activeTab === 'attempts'" class="min-w-0 max-w-full">
        <QuizAttemptTable
          :attempts="attempts"
          :loading="loadingAttempts"
          :has-more="attemptHasMore"
          @load-more="loadAttempts(true)"
        />
      </div>

      <div v-else-if="activeTab === 'questions'" class="min-w-0 max-w-full">
        <QuestionQualityTable
          :questions="questions"
          :loading="loadingQuestions"
          :has-more="questionHasMore"
          @load-more="loadQuestions(true)"
        />
      </div>

      <div v-else-if="activeTab === 'engagement'" class="min-w-0 max-w-full">
        <QuizEngagementPanel :engagement="engagement" />
      </div>
    </div>
  </div>
</template>
