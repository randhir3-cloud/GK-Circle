<script setup>
import { ref } from "vue";

defineProps({
  learners: {
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

const emit = defineEmits(["search", "load-more"]);

const searchQuery = ref("");

const handleSearch = () => {
  emit("search", searchQuery.value);
};

const formatPct = (val) =>
  val == null || Number.isNaN(Number(val)) ? "—" : `${val}%`;

const formatDuration = (secs) => {
  if (!secs || secs <= 0) return "0s";
  const m = Math.floor(secs / 60);
  const s = secs % 60;
  return m > 0 ? `${m}m ${s}s` : `${s}s`;
};
</script>

<template>
  <div
    class="bg-surface rounded-lg border border-border p-4 min-w-0 max-w-full"
  >
    <div
      class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4"
    >
      <h3 class="text-base font-semibold text-text">
        Learner Aggregated Performance
      </h3>
      <div class="flex items-center gap-2">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search by learner name..."
          class="px-3 py-1.5 text-sm border border-border rounded-md bg-background text-text focus:outline-none focus:border-primary"
          @keyup.enter="handleSearch"
        />
        <button
          class="px-3 py-1.5 text-sm font-medium bg-primary text-primary-contrast rounded-md hover:bg-primary-hover"
          @click="handleSearch"
        >
          Search
        </button>
      </div>
    </div>

    <div class="overflow-x-auto min-w-0 max-w-full">
      <table class="w-full text-left text-sm border-collapse min-w-[700px]">
        <thead>
          <tr
            class="border-b border-border text-text-secondary text-xs uppercase tracking-wider"
          >
            <th class="py-2 px-3">Learner</th>
            <th class="py-2 px-3 text-right">Quizzes</th>
            <th class="py-2 px-3 text-right">Attempts</th>
            <th class="py-2 px-3 text-right">Completed</th>
            <th class="py-2 px-3 text-right">Completion Rate</th>
            <th class="py-2 px-3 text-right">Avg Score</th>
            <th class="py-2 px-3 text-right">Engaged Time</th>
            <th class="py-2 px-3 text-right">Last Activity</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border text-text">
          <tr
            v-for="l in learners"
            :key="l.learner_id"
            class="hover:bg-surface-hover transition-colors"
          >
            <td class="py-3 px-3">
              <div class="flex items-center gap-2">
                <div
                  class="w-7 h-7 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-xs"
                >
                  {{ l.display_name.charAt(0) }}
                </div>
                <span class="font-medium text-text">{{ l.display_name }}</span>
              </div>
            </td>
            <td class="py-3 px-3 text-right font-mono">
              {{ l.unique_quizzes_attempted }}
            </td>
            <td class="py-3 px-3 text-right font-mono">
              {{ l.total_attempts }}
            </td>
            <td class="py-3 px-3 text-right font-mono">
              {{ l.completed_attempts }}
            </td>
            <td class="py-3 px-3 text-right font-mono font-medium">
              {{ formatPct(l.completion_rate) }}
            </td>
            <td class="py-3 px-3 text-right font-mono font-medium">
              {{ formatPct(l.average_percentage) }}
            </td>
            <td
              class="py-3 px-3 text-right font-mono text-text-secondary"
              title="Approximate active question time"
            >
              {{ formatDuration(l.engaged_question_time_seconds) }}
            </td>
            <td class="py-3 px-3 text-right text-xs text-text-tertiary">
              {{ new Date(l.last_activity_at).toLocaleDateString() }}
            </td>
          </tr>
          <tr v-if="!loading && learners.length === 0">
            <td colspan="8" class="py-6 text-center text-text-tertiary">
              No learner performance records found.
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
        {{ loading ? "Loading..." : "Load More Learners" }}
      </button>
    </div>
  </div>
</template>
