<script setup>
defineProps({
  attempts: {
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
    <h3 class="text-base font-semibold text-text mb-3">Quiz Attempts</h3>
    <div class="overflow-x-auto min-w-0 max-w-full">
      <table class="w-full text-left text-sm border-collapse min-w-[650px]">
        <thead>
          <tr
            class="border-b border-border text-text-secondary text-xs uppercase tracking-wider"
          >
            <th class="py-2 px-3">Learner</th>
            <th class="py-2 px-3 text-right">Attempt #</th>
            <th class="py-2 px-3">Status</th>
            <th class="py-2 px-3 text-right">Score</th>
            <th class="py-2 px-3 text-right">Percentage</th>
            <th class="py-2 px-3 text-right">Time Taken</th>
            <th class="py-2 px-3 text-right">Started At</th>
            <th class="py-2 px-3 text-right">Submitted At</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border text-text">
          <tr
            v-for="a in attempts"
            :key="a.attempt_id"
            class="hover:bg-surface-hover transition-colors"
          >
            <td class="py-3 px-3">
              <div class="flex items-center gap-2">
                <div
                  class="w-6 h-6 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-xs"
                >
                  {{ a.display_name.charAt(0) }}
                </div>
                <span class="font-medium text-text">{{ a.display_name }}</span>
              </div>
            </td>
            <td class="py-3 px-3 text-right font-mono">
              {{ a.attempt_number }}
            </td>
            <td class="py-3 px-3">
              <span
                class="px-2 py-0.5 text-xs rounded font-medium"
                :class="
                  a.status === 'SUBMITTED' || a.status === 'AUTO_SUBMITTED'
                    ? 'bg-success/10 text-success'
                    : 'bg-warning/10 text-warning'
                "
              >
                {{ a.status }}
              </span>
            </td>
            <td class="py-3 px-3 text-right font-mono">
              {{
                a.total_score != null && a.max_score != null
                  ? `${a.total_score} / ${a.max_score}`
                  : "—"
              }}
            </td>
            <td class="py-3 px-3 text-right font-mono font-medium">
              {{ formatPct(a.percentage) }}
            </td>
            <td class="py-3 px-3 text-right font-mono text-text-secondary">
              {{ formatDuration(a.time_taken_seconds) }}
            </td>
            <td class="py-3 px-3 text-right text-xs text-text-tertiary">
              {{ new Date(a.started_at).toLocaleString() }}
            </td>
            <td class="py-3 px-3 text-right text-xs text-text-tertiary">
              {{
                a.submitted_at ? new Date(a.submitted_at).toLocaleString() : "—"
              }}
            </td>
          </tr>
          <tr v-if="!loading && attempts.length === 0">
            <td colspan="8" class="py-6 text-center text-text-tertiary">
              No attempts found for this quiz.
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
        {{ loading ? "Loading..." : "Load More Attempts" }}
      </button>
    </div>
  </div>
</template>
