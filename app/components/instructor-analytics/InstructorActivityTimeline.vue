<script setup>
defineProps({
  events: {
    type: Array,
    default: () => [],
  },
  loading: {
    type: Boolean,
    default: false,
  },
  hasMore: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["load-more"]);

const eventBadgeClass = (type) => {
  switch (type) {
    case "ATTEMPT_SUBMITTED":
    case "ATTEMPT_AUTO_SUBMITTED":
      return "bg-success/10 text-success border-success/20";
    case "ATTEMPT_STARTED":
      return "bg-info/10 text-info border-info/20";
    case "RESULT_VIEWED":
      return "bg-primary/10 text-primary border-primary/20";
    case "RESULT_RELEASE_OVERRIDE_APPLIED":
    case "RESULT_RELEASE_SCHEDULED_EFFECTIVE":
      return "bg-warning/10 text-warning border-warning/20";
    default:
      return "bg-border/30 text-text-secondary border-border/50";
  }
};
</script>

<template>
  <div
    class="bg-surface rounded-lg border border-border p-4 min-w-0 max-w-full"
  >
    <h3 class="text-base font-semibold text-text mb-4">
      Instructor Activity Timeline
    </h3>

    <div
      class="relative pl-6 space-y-6 before:absolute before:left-2.5 before:top-2 before:bottom-2 before:w-0.5 before:bg-border min-w-0 max-w-full"
    >
      <div v-for="evt in events" :key="evt.id" class="relative min-w-0">
        <!-- Timeline marker dot -->
        <div
          class="absolute -left-6 top-1 w-3 h-3 rounded-full bg-primary ring-4 ring-surface"
        />

        <div class="bg-background rounded-md border border-border p-3 min-w-0">
          <div class="flex items-center justify-between gap-2 flex-wrap mb-1">
            <span
              class="px-2 py-0.5 text-xs font-medium rounded border"
              :class="eventBadgeClass(evt.event_type)"
            >
              {{ evt.event_type }}
            </span>
            <span class="text-xs font-mono text-text-tertiary">
              {{ new Date(evt.occurred_at).toLocaleString() }}
            </span>
          </div>
          <p class="text-sm text-text font-medium mt-1">
            {{ evt.summary }}
          </p>
          <div class="text-xs text-text-secondary mt-0.5">
            Quiz:
            <span class="font-medium text-text">{{ evt.quiz_title }}</span>
          </div>
        </div>
      </div>

      <div
        v-if="!loading && events.length === 0"
        class="py-6 text-center text-text-tertiary"
      >
        No recent activity events recorded.
      </div>
    </div>

    <div v-if="hasMore" class="mt-6 text-center">
      <button
        class="px-4 py-2 text-sm font-medium border border-border rounded-md text-text hover:bg-surface-hover transition-colors disabled:opacity-50"
        :disabled="loading"
        @click="emit('load-more')"
      >
        {{ loading ? "Loading..." : "Load More Timeline Events" }}
      </button>
    </div>
  </div>
</template>
