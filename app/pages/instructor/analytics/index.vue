<script setup>
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import {
  getInstructorAnalyticsAPIError,
  useInstructorAnalyticsApi,
} from "@/composables/instructor_analytics";
import InstructorOverviewCards from "@/components/instructor-analytics/InstructorOverviewCards.vue";
import InstructorQuizTable from "@/components/instructor-analytics/InstructorQuizTable.vue";
import LearnerPerformanceTable from "@/components/instructor-analytics/LearnerPerformanceTable.vue";
import ReleaseMonitoringPanel from "@/components/instructor-analytics/ReleaseMonitoringPanel.vue";
import InstructorActivityTimeline from "@/components/instructor-analytics/InstructorActivityTimeline.vue";

definePageMeta({ layout: "empty" });
useSeoMeta({
  title: "Instructor Analytics - GK Circle",
  description:
    "Instructor portfolio performance, learner analytics, and release monitoring.",
});

const router = useRouter();
const api = useInstructorAnalyticsApi();

const activeTab = ref("overview");
const loading = ref(true);
const error = ref("");

// Data states
const overview = ref(null);

const quizzes = ref([]);
const quizCursor = ref("");
const quizHasMore = ref(false);
const loadingQuizzes = ref(false);

const learners = ref([]);
const learnerCursor = ref("");
const learnerHasMore = ref(false);
const loadingLearners = ref(false);
const learnerSearch = ref("");

const releaseMonitoring = ref({ summary: {}, quizzes: [] });
const releaseCursor = ref("");
const releaseHasMore = ref(false);
const loadingReleases = ref(false);
const releaseFilter = ref("");

const timelineEvents = ref([]);
const timelineCursor = ref("");
const timelineHasMore = ref(false);
const loadingTimeline = ref(false);

const loadOverview = async () => {
  try {
    overview.value = await api.getOverview();
  } catch (err) {
    error.value = getInstructorAnalyticsAPIError(
      err,
      "Failed to load overview"
    );
  }
};

const loadQuizzes = async (isMore = false) => {
  loadingQuizzes.value = true;
  try {
    const res = await api.getQuizzes({
      cursor: isMore ? quizCursor.value : "",
    });
    if (isMore) {
      quizzes.value = [...quizzes.value, ...(res.quizzes || [])];
    } else {
      quizzes.value = res.quizzes || [];
    }
    quizCursor.value = res.next_cursor || "";
    quizHasMore.value = !!res.has_more;
  } catch (err) {
    error.value = getInstructorAnalyticsAPIError(err, "Failed to load quizzes");
  } finally {
    loadingQuizzes.value = false;
  }
};

const loadLearners = async (isMore = false) => {
  loadingLearners.value = true;
  try {
    const res = await api.getLearners({
      cursor: isMore ? learnerCursor.value : "",
      search: learnerSearch.value,
    });
    if (isMore) {
      learners.value = [...learners.value, ...(res.learners || [])];
    } else {
      learners.value = res.learners || [];
    }
    learnerCursor.value = res.next_cursor || "";
    learnerHasMore.value = !!res.has_more;
  } catch (err) {
    error.value = getInstructorAnalyticsAPIError(
      err,
      "Failed to load learners"
    );
  } finally {
    loadingLearners.value = false;
  }
};

const handleLearnerSearch = (q) => {
  learnerSearch.value = q;
  learnerCursor.value = "";
  loadLearners(false);
};

const loadReleases = async (isMore = false) => {
  loadingReleases.value = true;
  try {
    const res = await api.getReleases({
      cursor: isMore ? releaseCursor.value : "",
      classification: releaseFilter.value,
    });
    if (isMore) {
      releaseMonitoring.value = {
        summary: res.summary || releaseMonitoring.value.summary,
        quizzes: [
          ...(releaseMonitoring.value.quizzes || []),
          ...(res.quizzes || []),
        ],
      };
    } else {
      releaseMonitoring.value = res || { summary: {}, quizzes: [] };
    }
    releaseCursor.value = res.next_cursor || "";
    releaseHasMore.value = !!res.has_more;
  } catch (err) {
    error.value = getInstructorAnalyticsAPIError(
      err,
      "Failed to load releases"
    );
  } finally {
    loadingReleases.value = false;
  }
};

const handleReleaseFilter = (classification) => {
  releaseFilter.value = classification;
  releaseCursor.value = "";
  loadReleases(false);
};

const loadTimeline = async (isMore = false) => {
  loadingTimeline.value = true;
  try {
    const res = await api.getTimeline({
      cursor: isMore ? timelineCursor.value : "",
    });
    if (isMore) {
      timelineEvents.value = [...timelineEvents.value, ...(res.events || [])];
    } else {
      timelineEvents.value = res.events || [];
    }
    timelineCursor.value = res.next_cursor || "";
    timelineHasMore.value = !!res.has_more;
  } catch (err) {
    error.value = getInstructorAnalyticsAPIError(
      err,
      "Failed to load timeline"
    );
  } finally {
    loadingTimeline.value = false;
  }
};

const navigateToQuizAnalytics = (quizId) => {
  router.push(`/instructor/analytics/${quizId}`);
};

onMounted(async () => {
  loading.value = true;
  await Promise.all([
    loadOverview(),
    loadQuizzes(false),
    loadLearners(false),
    loadReleases(false),
    loadTimeline(false),
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
            Instructor Analytics
          </h1>
          <p class="text-sm text-text-secondary mt-1">
            Portfolio performance, learner activity, result releases, and event
            timelines.
          </p>
        </div>
        <NuxtLink
          to="/admin/quiz"
          class="px-4 py-2 text-sm font-medium border border-border rounded-md hover:bg-surface-hover transition-colors inline-flex items-center gap-1.5 self-start sm:self-auto"
        >
          ← Back to Quizzes
        </NuxtLink>
      </div>

      <!-- Error alert -->
      <div
        v-if="error"
        class="p-4 bg-danger/10 text-danger rounded-lg border border-danger/20 text-sm"
      >
        {{ error }}
      </div>

      <!-- Main Overview Cards -->
      <InstructorOverviewCards :overview="overview" />

      <!-- Navigation Tabs -->
      <div
        class="flex items-center gap-2 border-b border-border overflow-x-auto min-w-0 max-w-full pb-1"
      >
        <button
          v-for="tab in [
            { key: 'overview', label: 'Owned Quizzes' },
            { key: 'learners', label: 'Learner Performance' },
            { key: 'releases', label: 'Result Release Monitoring' },
            { key: 'timeline', label: 'Activity Timeline' },
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
      <div v-if="activeTab === 'overview'" class="min-w-0 max-w-full">
        <InstructorQuizTable
          :quizzes="quizzes"
          :loading="loadingQuizzes"
          :has-more="quizHasMore"
          @select-quiz="navigateToQuizAnalytics"
          @load-more="loadQuizzes(true)"
        />
      </div>

      <div v-else-if="activeTab === 'learners'" class="min-w-0 max-w-full">
        <LearnerPerformanceTable
          :learners="learners"
          :loading="loadingLearners"
          :has-more="learnerHasMore"
          @search="handleLearnerSearch"
          @load-more="loadLearners(true)"
        />
      </div>

      <div v-else-if="activeTab === 'releases'" class="min-w-0 max-w-full">
        <ReleaseMonitoringPanel
          :monitoring="releaseMonitoring"
          :loading="loadingReleases"
          :has-more="releaseHasMore"
          @filter-classification="handleReleaseFilter"
          @load-more="loadReleases(true)"
        />
      </div>

      <div v-else-if="activeTab === 'timeline'" class="min-w-0 max-w-full">
        <InstructorActivityTimeline
          :events="timelineEvents"
          :loading="loadingTimeline"
          :has-more="timelineHasMore"
          @load-more="loadTimeline(true)"
        />
      </div>
    </div>
  </div>
</template>
