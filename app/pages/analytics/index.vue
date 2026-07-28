<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useLearnerAnalyticsApi } from "@/composables/learner_analytics";
import { setUserDataStore } from "@/composables/auth";
import {
  getSafeAPIErrorMessage,
  isAuthenticationError,
} from "@/utils/api_error";
import { useUsersStore } from "~~/store/users";
import AnalyticsSummaryCard from "@/components/analytics/AnalyticsSummaryCard.vue";
import StudyTimeCard from "@/components/analytics/StudyTimeCard.vue";
import PerformanceTrendChart from "@/components/analytics/PerformanceTrendChart.vue";
import SubjectPerformanceTable from "@/components/analytics/SubjectPerformanceTable.vue";
import RecentActivityTable from "@/components/analytics/RecentActivityTable.vue";
import AttemptTimeline from "@/components/analytics/AttemptTimeline.vue";
import PageStateCard from "@/components/common/PageStateCard.vue";

definePageMeta({ layout: "empty" });
useSeoMeta({
  title: "Analytics - GK Circle",
  description: "Review your assessment performance and study activity.",
});

const api = useLearnerAnalyticsApi();
const usersStore = useUsersStore();

const loading = ref(true);
const authenticated = ref(false);
const error = ref("");
const summary = ref(null);
const subjects = ref([]);
const activityItems = ref([]);
const activityCursor = ref("");
const activityHasMore = ref(false);
const loadingMore = ref(false);
const granularity = ref("daily");
const trendBuckets = ref([]);
const timeline = ref(null);

const formatPct = (value) =>
  value == null || Number.isNaN(Number(value)) ? "—" : `${value}%`;

const hasAttempts = computed(() => (summary.value?.total_attempts || 0) > 0);

const summaryCards = computed(() => [
  {
    label: "Completion rate",
    value: formatPct(summary.value?.completion_rate),
    hint: `${summary.value?.completed_attempts || 0} of ${
      summary.value?.total_attempts || 0
    } attempts`,
  },
  {
    label: "Average score",
    value: formatPct(summary.value?.average_percentage),
    hint:
      summary.value?.pending_result_count > 0
        ? `${summary.value.pending_result_count} result(s) pending`
        : "Released attempts only",
  },
  {
    label: "Current streak",
    value: `${summary.value?.current_streak_days || 0} days`,
    hint: `Best ${summary.value?.best_streak_days || 0} days`,
  },
]);

const emptyAnalyticsCards = [
  { label: "Accuracy", value: "—", hint: "Complete an assessment" },
  { label: "Questions Attempted", value: "0", hint: "No answers recorded" },
  { label: "Time Spent", value: "0 min", hint: "Practice time" },
  { label: "Weekly Progress", value: "—", hint: "No weekly activity" },
  { label: "Streak", value: "0 days", hint: "Start practising today" },
  { label: "Weak Subjects", value: "—", hint: "More attempts required" },
];

const setSafeRequestError = (requestError, fallback) => {
  if (isAuthenticationError(requestError)) {
    authenticated.value = false;
    error.value = "";
    return;
  }
  error.value = getSafeAPIErrorMessage(requestError, fallback);
};

const loadTrends = async () => {
  const to = new Date();
  const from = new Date();
  if (granularity.value === "monthly") from.setMonth(from.getMonth() - 5);
  else if (granularity.value === "weekly")
    from.setDate(from.getDate() - 7 * 11);
  else from.setDate(from.getDate() - 29);
  const trends = await api.getTrends({
    granularity: granularity.value,
    from: from.toISOString(),
    to: to.toISOString(),
  });
  trendBuckets.value = trends?.buckets || [];
};

const loadDashboard = async () => {
  loading.value = true;
  error.value = "";
  try {
    const [dash, subjectData, activity] = await Promise.all([
      api.getDashboard(),
      api.getSubjects(),
      api.getActivity({ limit: 20 }),
    ]);
    await loadTrends();
    summary.value = dash;
    subjects.value = subjectData?.subjects || [];
    activityItems.value = activity?.items || [];
    activityCursor.value = activity?.next_cursor || "";
    activityHasMore.value = Boolean(activity?.has_more);
  } catch (requestError) {
    setSafeRequestError(
      requestError,
      "Analytics could not be loaded. Please try again."
    );
  } finally {
    loading.value = false;
  }
};

const loadMoreActivity = async () => {
  if (!activityHasMore.value || loadingMore.value) return;
  loadingMore.value = true;
  try {
    const activity = await api.getActivity({
      limit: 20,
      cursor: activityCursor.value,
    });
    activityItems.value = [...activityItems.value, ...(activity?.items || [])];
    activityCursor.value = activity?.next_cursor || "";
    activityHasMore.value = Boolean(activity?.has_more);
  } catch (requestError) {
    setSafeRequestError(
      requestError,
      "More activity could not be loaded. Please try again."
    );
  } finally {
    loadingMore.value = false;
  }
};

const selectAttempt = async (attemptId) => {
  try {
    timeline.value = await api.getAttemptTimeline(attemptId);
  } catch (requestError) {
    setSafeRequestError(
      requestError,
      "The attempt timeline could not be loaded."
    );
  }
};

watch(granularity, async () => {
  if (!authenticated.value) return;
  try {
    await loadTrends();
  } catch (requestError) {
    setSafeRequestError(requestError, "Trends could not be loaded.");
  }
});

onMounted(async () => {
  const user = await setUserDataStore(usersStore);
  if (!user) {
    loading.value = false;
    return;
  }
  authenticated.value = true;
  await loadDashboard();
});
</script>

<template>
  <div
    class="min-h-screen overflow-x-hidden bg-jv-cream px-4 py-6 text-jv-ink sm:px-6 lg:px-8"
  >
    <header class="mx-auto w-full max-w-7xl">
      <p class="text-xs font-black uppercase tracking-wide text-jv-coral">
        Learning intelligence
      </p>
      <h1 class="mt-1 font-headings text-3xl sm:text-5xl">Your analytics</h1>
      <p class="mt-2 max-w-3xl text-sm font-bold text-jv-muted sm:text-base">
        Understand your accuracy, speed, consistency, and subject-level
        progress.
        <span v-if="summary?.resolved_timezone">
          Timezone: {{ summary.resolved_timezone }}
        </span>
      </p>
    </header>

    <div
      v-if="loading"
      class="mx-auto mt-6 grid w-full max-w-7xl gap-4 sm:grid-cols-2 xl:grid-cols-3"
      aria-label="Loading analytics"
    >
      <div
        v-for="index in 6"
        :key="index"
        class="h-36 animate-pulse rounded-[12px] border-[2px] border-jv-ink/15 bg-jv-white"
      ></div>
    </div>

    <div v-else-if="!authenticated" class="mx-auto mt-6 w-full max-w-7xl">
      <PageStateCard
        eyebrow="Private learning insights"
        title="Sign in to view your learning analytics."
        description="Your accuracy, speed, progress, streaks, and weak subjects are available after you sign in."
        action-label="Sign in"
        action-to="/account/login"
      />
    </div>

    <div
      v-else-if="error"
      class="mx-auto mt-6 w-full max-w-7xl"
      role="alert"
      data-testid="analytics-error"
    >
      <PageStateCard
        eyebrow="Analytics unavailable"
        title="We could not load your analytics right now."
        :description="error"
      />
    </div>

    <div
      v-else-if="!hasAttempts"
      class="mx-auto mt-6 grid w-full max-w-7xl gap-4"
      data-testid="analytics-empty"
    >
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <AnalyticsSummaryCard
          v-for="card in emptyAnalyticsCards"
          :key="card.label"
          :label="card.label"
          :value="card.value"
          :hint="card.hint"
        />
      </div>
      <PageStateCard
        eyebrow="Build your first insight"
        title="No assessments completed yet."
        description="Start practising to unlock insights."
        action-label="Start practising"
        action-to="/#practice"
      />
    </div>

    <div
      v-else
      class="mx-auto mt-6 grid w-full min-w-0 max-w-7xl grid-cols-1 gap-4 [&>*]:min-w-0"
    >
      <div class="grid min-w-0 grid-cols-1 gap-4 sm:grid-cols-3 [&>*]:min-w-0">
        <AnalyticsSummaryCard
          v-for="card in summaryCards"
          :key="card.label"
          :label="card.label"
          :value="card.value"
          :hint="card.hint"
        />
      </div>

      <StudyTimeCard
        :assessment-duration-seconds="summary?.assessment_duration_seconds || 0"
        :engaged-question-time-seconds="
          summary?.engaged_question_time_seconds || 0
        "
      />

      <PerformanceTrendChart v-model="granularity" :buckets="trendBuckets" />
      <SubjectPerformanceTable :subjects="subjects" />

      <div class="grid min-w-0 grid-cols-1 gap-4 lg:grid-cols-2 [&>*]:min-w-0">
        <RecentActivityTable
          :items="activityItems"
          :has-more="activityHasMore"
          :loading-more="loadingMore"
          @load-more="loadMoreActivity"
          @select-attempt="selectAttempt"
        />
        <AttemptTimeline :timeline="timeline" />
      </div>
    </div>
  </div>
</template>
