<script setup>
defineProps({
  questions: {
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

const formatDiscrim = (val) =>
  val == null || Number.isNaN(Number(val))
    ? "— (sample < 10)"
    : `${val > 0 ? "+" : ""}${val}%`;
</script>

<template>
  <div
    class="bg-surface rounded-lg border border-border p-4 min-w-0 max-w-full"
  >
    <h3 class="text-base font-semibold text-text mb-3">
      Question Quality Metrics
    </h3>
    <div class="overflow-x-auto min-w-0 max-w-full">
      <table class="w-full text-left text-sm border-collapse min-w-[700px]">
        <thead>
          <tr
            class="border-b border-border text-text-secondary text-xs uppercase tracking-wider"
          >
            <th class="py-2 px-3 text-center w-12">#</th>
            <th class="py-2 px-3">Question Prompt</th>
            <th class="py-2 px-3 text-right">Answered</th>
            <th class="py-2 px-3 text-right">Correct</th>
            <th class="py-2 px-3 text-right">Incorrect</th>
            <th class="py-2 px-3 text-right">Unanswered</th>
            <th class="py-2 px-3 text-right">Difficulty Index</th>
            <th
              class="py-2 px-3 text-right"
              title="Basic estimate based on top-27% vs bottom-27% scoring group correct rates"
            >
              Discrimination Est.
            </th>
            <th class="py-2 px-3 text-right">Avg Time</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border text-text">
          <tr
            v-for="q in questions"
            :key="q.question_id"
            class="hover:bg-surface-hover transition-colors"
          >
            <td
              class="py-3 px-3 text-center font-mono font-medium text-text-secondary"
            >
              {{ q.order_number }}
            </td>
            <td class="py-3 px-3">
              <div class="font-medium text-text line-clamp-2 max-w-md">
                {{ q.question_text }}
              </div>
              <!-- Objective answer distribution bars -->
              <div
                v-if="
                  q.answer_distribution &&
                  Object.keys(q.answer_distribution).length > 0
                "
                class="flex gap-2 mt-1.5 flex-wrap"
              >
                <span
                  v-for="(count, optKey) in q.answer_distribution"
                  :key="optKey"
                  class="px-1.5 py-0.5 text-[10px] font-mono rounded bg-border/40 text-text-secondary"
                >
                  Opt {{ optKey }}: {{ count }}
                </span>
              </div>
            </td>
            <td class="py-3 px-3 text-right font-mono">
              {{ q.total_answered }}
            </td>
            <td class="py-3 px-3 text-right font-mono text-success">
              {{ q.correct_count }}
            </td>
            <td class="py-3 px-3 text-right font-mono text-danger">
              {{ q.incorrect_count }}
            </td>
            <td class="py-3 px-3 text-right font-mono text-text-tertiary">
              {{ q.unanswered_count }}
            </td>
            <td class="py-3 px-3 text-right font-mono font-medium">
              {{ formatPct(q.difficulty_index) }}
            </td>
            <td class="py-3 px-3 text-right font-mono font-medium">
              {{ formatDiscrim(q.discrimination_index) }}
            </td>
            <td class="py-3 px-3 text-right font-mono text-text-secondary">
              {{
                q.average_time_seconds != null
                  ? `${q.average_time_seconds}s`
                  : "—"
              }}
            </td>
          </tr>
          <tr v-if="!loading && questions.length === 0">
            <td colspan="9" class="py-6 text-center text-text-tertiary">
              No question metrics available.
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
        {{ loading ? "Loading..." : "Load More Questions" }}
      </button>
    </div>
  </div>
</template>
