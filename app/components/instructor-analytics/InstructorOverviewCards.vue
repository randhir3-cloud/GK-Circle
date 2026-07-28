<script setup>
import { computed } from "vue";

const props = defineProps({
  overview: {
    type: Object,
    default: () => ({}),
  },
});

const formatPct = (val) =>
  val == null || Number.isNaN(Number(val)) ? "—" : `${val}%`;

const cards = computed(() => [
  {
    label: "Total Owned Quizzes",
    value: props.overview?.total_quizzes ?? 0,
    hint: `${props.overview?.released_quizzes ?? 0} released`,
  },
  {
    label: "Total Learner Attempts",
    value: props.overview?.total_attempts ?? 0,
    hint: `${props.overview?.completed_attempts ?? 0} completed`,
  },
  {
    label: "Overall Completion Rate",
    value: formatPct(props.overview?.completion_rate),
    hint: `${props.overview?.completed_attempts ?? 0} of ${
      props.overview?.total_attempts ?? 0
    } attempts`,
  },
  {
    label: "Attempt-Weighted Avg Score",
    value: formatPct(props.overview?.average_score_percentage),
    hint: `across ${
      props.overview?.completed_scored_attempts_count ?? 0
    } scored attempts`,
  },
  {
    label: "Unique Engaged Learners",
    value: props.overview?.unique_learners ?? 0,
    hint: "distinct students",
  },
  {
    label: "Pending Release Quizzes",
    value: props.overview?.pending_release_quizzes ?? 0,
    hint: "quizzes awaiting release",
  },
]);
</script>

<template>
  <div
    class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 min-w-0 max-w-full"
  >
    <div
      v-for="(card, idx) in cards"
      :key="idx"
      class="bg-surface p-4 rounded-lg border border-border shadow-sm min-w-0"
    >
      <div class="text-sm font-medium text-text-secondary truncate">
        {{ card.label }}
      </div>
      <div class="text-2xl font-bold text-text mt-1 truncate">
        {{ card.value }}
      </div>
      <div class="text-xs text-text-tertiary mt-1 truncate">
        {{ card.hint }}
      </div>
    </div>
  </div>
</template>
