<script setup>
import { computed } from "vue";

const props = defineProps({
  engagement: {
    type: Object,
    default: () => ({}),
  },
});

const cards = computed(() => [
  {
    label: "Question Views",
    value: props.engagement?.total_question_views ?? 0,
  },
  {
    label: "Answer Selections",
    value: props.engagement?.total_answer_selections ?? 0,
  },
  {
    label: "Answer Changes",
    value: props.engagement?.total_answer_changes ?? 0,
  },
  {
    label: "Hints Opened",
    value: props.engagement?.total_hints_opened ?? 0,
  },
  {
    label: "Reviews Opened",
    value: props.engagement?.total_reviews_opened ?? 0,
  },
  {
    label: "Result Views",
    value: props.engagement?.total_result_views ?? 0,
  },
]);
</script>

<template>
  <div class="space-y-6 min-w-0 max-w-full">
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
      </div>
    </div>

    <div
      class="bg-surface p-4 rounded-lg border border-border flex flex-col sm:flex-row sm:items-center justify-between gap-4 min-w-0 max-w-full"
    >
      <div>
        <div class="text-sm font-medium text-text">
          Engaged Audience Summary
        </div>
        <div class="text-xs text-text-secondary mt-0.5">
          {{ props.engagement?.unique_engaged_learners ?? 0 }} unique learners
          across {{ props.engagement?.unique_engaged_attempts ?? 0 }} attempts
        </div>
      </div>
      <div class="text-right sm:text-right">
        <div class="text-2xl font-bold text-primary">
          {{ props.engagement?.average_events_per_attempt ?? 0 }}
        </div>
        <div class="text-xs text-text-tertiary">
          Average telemetry events per engaged attempt
        </div>
      </div>
    </div>
  </div>
</template>
