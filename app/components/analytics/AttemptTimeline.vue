<template>
  <section
    class="w-full min-w-0 max-w-full rounded-xl border border-jv-ink/10 bg-white/80 p-4 shadow-sm"
    data-testid="attempt-timeline"
  >
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h2 class="font-headings text-xl text-jv-ink">Attempt timeline</h2>
        <p class="mt-1 text-sm text-jv-muted">
          {{ timeline?.quiz_title || "Select an attempt from recent activity" }}
        </p>
      </div>
      <p
        v-if="timeline"
        class="rounded-md bg-jv-cream px-3 py-1 text-xs font-bold"
      >
        {{ timeline.result_status }}
        <span v-if="timeline.percentage != null">
          · {{ timeline.percentage }}%</span
        >
      </p>
    </div>
    <ol v-if="timeline?.events?.length" class="mt-4 space-y-3">
      <li
        v-for="event in timeline.events"
        :key="event.id"
        class="border-l-2 border-jv-ink/20 pl-3"
      >
        <p class="text-sm font-bold text-jv-ink">{{ event.event_type }}</p>
        <p class="text-xs text-jv-muted">
          {{ formatWhen(event.occurred_at) }} · {{ event.event_source }}
        </p>
      </li>
    </ol>
    <p v-else class="mt-4 text-sm text-jv-muted">
      {{ emptyMessage }}
    </p>
  </section>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  timeline: { type: Object, default: null },
});

const emptyMessage = computed(() =>
  props.timeline
    ? "No timeline events for this attempt."
    : "Select an attempt to inspect its lifecycle timeline."
);

const formatWhen = (value) => {
  if (!value) return "—";
  try {
    return new Date(value).toLocaleString();
  } catch {
    return value;
  }
};
</script>
