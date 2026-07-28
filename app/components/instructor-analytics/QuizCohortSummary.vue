<script setup>
import { computed } from "vue";

const props = defineProps({
  summary: {
    type: Object,
    default: () => ({}),
  },
});

const formatPct = (val) =>
  val == null || Number.isNaN(Number(val)) ? "—" : `${val}%`;

const formatDuration = (secs) => {
  if (!secs || secs <= 0) return "0s";
  const m = Math.floor(secs / 60);
  const s = secs % 60;
  return m > 0 ? `${m}m ${s}s` : `${s}s`;
};

const cards = computed(() => [
  {
    label: "Total Attempts",
    value: props.summary?.total_attempts ?? 0,
    hint: `${props.summary?.completed_attempts ?? 0} completed`,
  },
  {
    label: "Completion Rate",
    value: formatPct(props.summary?.completion_rate),
    hint: `${props.summary?.in_progress_attempts ?? 0} in progress, ${
      props.summary?.abandoned_attempts ?? 0
    } abandoned`,
  },
  {
    label: "Attempt-Weighted Avg Score",
    value: formatPct(props.summary?.average_score_percentage),
    hint: `High: ${formatPct(
      props.summary?.highest_score_percentage
    )} | Low: ${formatPct(props.summary?.lowest_score_percentage)}`,
  },
  {
    label: "Avg Attempt Duration",
    value: formatDuration(props.summary?.average_duration_seconds),
    hint: "over completed attempts",
  },
  {
    label: "Unique Learners",
    value: props.summary?.unique_learners ?? 0,
    hint: "distinct students",
  },
  {
    label: "Total Questions",
    value: props.summary?.total_questions ?? 0,
    hint: `Release: ${props.summary?.result_release_policy ?? "IMMEDIATE"}`,
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
