<script setup>
import { computed, onMounted, ref, watch } from "vue";
import {
  getLearnerAnalyticsAPIError,
  useLearnerAnalyticsApi,
} from "@/composables/learner_analytics";
import { setUserDataStore } from "@/composables/auth";
import { useUsersStore } from "~~/store/users";
import AnalyticsSummaryCard from "@/components/analytics/AnalyticsSummaryCard.vue";
import StudyTimeCard from "@/components/analytics/StudyTimeCard.vue";
import PerformanceTrendChart from "@/components/analytics/PerformanceTrendChart.vue";
import SubjectPerformanceTable from "@/components/analytics/SubjectPerformanceTable.vue";
import RecentActivityTable from "@/components/analytics/RecentActivityTable.vue";
import AttemptTimeline from "@/components/analytics/AttemptTimeline.vue";

definePageMeta({ layout: "empty" });
useSeoMeta({
  title: "Analytics - GK Circle",
  description: "Review your assessment performance and study activity.",
});

const api = useLearnerAnalyticsApi();
const usersStore = useUsersStore();

const loading = ref(true);
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
  } catch (err) {
    error.value = getLearnerAnalyticsAPIError(
      err,
      "Unable to load learner analytics."
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
  } catch (err) {
    error.value = getLearnerAnalyticsAPIError(
      err,
      "Unable to load more activity."
    );
  } finally {
    loadingMore.value = false;
  }
};

const selectAttempt = async (attemptId) => {
  try {
    timeline.value = await api.getAttemptTimeline(attemptId);
  } catch (err) {
    error.value = getLearnerAnalyticsAPIError(
      err,
      "Unable to load attempt timeline."
    );
  }
};

watch(granularity, async () => {
  try {
    await loadTrends();
  } catch (err) {
    error.value = getLearnerAnalyticsAPIError(err, "Unable to load trends.");
  }
});

onMounted(async () => {
  try {
    await setUserDataStore(usersStore);
  } catch {
    /* ignore */
  }
  await loadDashboard();
});
</script>

<template>
  <div
    class="min-h-screen overflow-x-hidden bg-jv-cream px-4 py-6 text-jv-ink sm:px-6"
  >
    <header class="mx-auto w-full max-w-6xl">
      <h1 class="font-headings text-3xl sm:text-4xl">Your analytics</h1>
      <p class="mt-2 text-sm font-bold text-jv-muted">
        Read-only insights from your assessment attempts.
        <span v-if="summary?.resolved_timezone">
          Timezone: {{ summary.resolved_timezone }}
        </span>
      </p>
    </header>

    <p v-if="loading" class="mx-auto mt-6 max-w-6xl font-bold">Loading…</p>
    <p
      v-else-if="error"
      class="mx-auto mt-6 max-w-6xl font-bold text-red-700"
      role="alert"
      data-testid="analytics-error"
    >
      {{ error }}
    </p>

    <div
      v-else
      class="mx-auto mt-6 grid grid-cols-1 w-full min-w-0 max-w-6xl gap-4 [&>*]:min-w-0"
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
