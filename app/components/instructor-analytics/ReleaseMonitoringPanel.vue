<script setup>
import { ref } from "vue";

defineProps({
  monitoring: {
    type: Object,
    default: () => ({ summary: {}, quizzes: [] }),
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

const emit = defineEmits(["filter-classification", "load-more"]);

const activeFilter = ref("");

const setFilter = (cls) => {
  activeFilter.value = cls;
  emit("filter-classification", cls);
};

const classifications = [
  { key: "", label: "All Releases" },
  { key: "IMMEDIATE_RELEASE", label: "Immediate" },
  { key: "PENDING_MANUAL", label: "Pending Manual" },
  { key: "COMPLETED_MANUAL", label: "Completed Manual" },
  { key: "SCHEDULED_FOR_TODAY", label: "Scheduled Today" },
  { key: "UPCOMING_SCHEDULED", label: "Upcoming Scheduled" },
  { key: "OVERDUE_SCHEDULED", label: "Overdue Scheduled" },
  { key: "MANUALLY_OVERRIDDEN_SCHEDULED", label: "Manually Overridden" },
];

const badgeClass = (cls) => {
  switch (cls) {
    case "IMMEDIATE_RELEASE":
    case "COMPLETED_MANUAL":
      return "bg-success/10 text-success border-success/20";
    case "SCHEDULED_FOR_TODAY":
    case "UPCOMING_SCHEDULED":
      return "bg-info/10 text-info border-info/20";
    case "PENDING_MANUAL":
    case "OVERDUE_SCHEDULED":
      return "bg-warning/10 text-warning border-warning/20";
    case "MANUALLY_OVERRIDDEN_SCHEDULED":
      return "bg-primary/10 text-primary border-primary/20";
    default:
      return "bg-border/30 text-text-secondary";
  }
};
</script>

<template>
  <div
    class="bg-surface rounded-lg border border-border p-4 min-w-0 max-w-full"
  >
    <h3 class="text-base font-semibold text-text mb-3">
      Result Release Monitoring
    </h3>

    <!-- Summary pills -->
    <div class="flex flex-wrap gap-2 mb-4">
      <button
        v-for="c in classifications"
        :key="c.key"
        class="px-3 py-1 text-xs font-medium rounded-full border transition-colors"
        :class="
          activeFilter === c.key
            ? 'bg-primary text-primary-contrast border-primary'
            : 'bg-background text-text border-border hover:bg-surface-hover'
        "
        @click="setFilter(c.key)"
      >
        {{ c.label }}
        <span
          v-if="c.key && monitoring.summary"
          class="ml-1 px-1.5 py-0.2 bg-surface/50 rounded-full text-[10px]"
        >
          {{ monitoring.summary[c.key] || 0 }}
        </span>
      </button>
    </div>

    <div class="overflow-x-auto min-w-0 max-w-full">
      <table class="w-full text-left text-sm border-collapse min-w-[650px]">
        <thead>
          <tr
            class="border-b border-border text-text-secondary text-xs uppercase tracking-wider"
          >
            <th class="py-2 px-3">Quiz Title</th>
            <th class="py-2 px-3">Category</th>
            <th class="py-2 px-3">Policy</th>
            <th class="py-2 px-3">Classification</th>
            <th class="py-2 px-3 text-right">Completed Attempts</th>
            <th class="py-2 px-3 text-right">Scheduled At</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border text-text">
          <tr
            v-for="q in monitoring.quizzes"
            :key="q.quiz_id"
            class="hover:bg-surface-hover transition-colors"
          >
            <td class="py-3 px-3 font-medium text-text">{{ q.title }}</td>
            <td class="py-3 px-3 text-text-secondary">
              {{ q.category || "Uncategorised" }}
            </td>
            <td class="py-3 px-3 font-mono text-xs">
              {{ q.result_release_policy }}
            </td>
            <td class="py-3 px-3">
              <span
                class="px-2 py-0.5 text-xs font-medium rounded border"
                :class="badgeClass(q.classification)"
              >
                {{ q.classification.replace(/_/g, " ") }}
              </span>
            </td>
            <td class="py-3 px-3 text-right font-mono">
              {{ q.completed_attempts_count }}
            </td>
            <td
              class="py-3 px-3 text-right text-xs font-mono text-text-tertiary"
            >
              {{
                q.results_scheduled_at
                  ? new Date(q.results_scheduled_at).toLocaleString()
                  : "—"
              }}
            </td>
          </tr>
          <tr
            v-if="
              !loading &&
              (!monitoring.quizzes || monitoring.quizzes.length === 0)
            "
          >
            <td colspan="6" class="py-6 text-center text-text-tertiary">
              No release records match the selected filter.
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="hasMore" class="mt-4 text-center">
      <button
        class="px-4 py-2 text-sm font-medium border border-border rounded-md text-text hover:bg-surface-hover transition-colors disabled:opacity-50"
        :disabled="loading"
        @click="emit('load-more')"
      >
        {{ loading ? "Loading..." : "Load More Releases" }}
      </button>
    </div>
  </div>
</template>
