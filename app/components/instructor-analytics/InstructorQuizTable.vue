<script setup>
defineProps({
  quizzes: {
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

const emit = defineEmits(["select-quiz", "load-more"]);

const formatPct = (val) =>
  val == null || Number.isNaN(Number(val)) ? "—" : `${val}%`;
</script>

<template>
  <div
    class="bg-surface rounded-lg border border-border p-4 min-w-0 max-w-full"
  >
    <h3 class="text-base font-semibold text-text mb-3">
      Owned Quizzes Overview
    </h3>
    <div class="overflow-x-auto min-w-0 max-w-full">
      <table class="w-full text-left text-sm border-collapse min-w-[600px]">
        <thead>
          <tr
            class="border-b border-border text-text-secondary text-xs uppercase tracking-wider"
          >
            <th class="py-2 px-3">Quiz Title</th>
            <th class="py-2 px-3">Category</th>
            <th class="py-2 px-3 text-right">Attempts</th>
            <th class="py-2 px-3 text-right">Completed</th>
            <th class="py-2 px-3 text-right">Avg Score</th>
            <th class="py-2 px-3 text-right">Learners</th>
            <th class="py-2 px-3">Release Policy</th>
            <th class="py-2 px-3">Action</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border text-text">
          <tr
            v-for="q in quizzes"
            :key="q.quiz_id"
            class="hover:bg-surface-hover transition-colors cursor-pointer"
            @click="emit('select-quiz', q.quiz_id)"
          >
            <td class="py-3 px-3 font-medium text-primary">
              {{ q.title }}
            </td>
            <td class="py-3 px-3 text-text-secondary">
              {{ q.category_name || "Uncategorised" }}
            </td>
            <td class="py-3 px-3 text-right font-mono">
              {{ q.total_attempts }}
            </td>
            <td class="py-3 px-3 text-right font-mono">
              {{ q.completed_attempts }}
            </td>
            <td class="py-3 px-3 text-right font-mono font-medium">
              {{ formatPct(q.average_score_percentage) }}
            </td>
            <td class="py-3 px-3 text-right font-mono">
              {{ q.unique_learners }}
            </td>
            <td class="py-3 px-3">
              <span
                class="px-2 py-0.5 text-xs rounded font-medium"
                :class="
                  q.results_released
                    ? 'bg-success/10 text-success'
                    : 'bg-warning/10 text-warning'
                "
              >
                {{ q.result_release_policy }} ({{
                  q.results_released ? "Released" : "Pending"
                }})
              </span>
            </td>
            <td class="py-3 px-3">
              <button
                class="px-2.5 py-1 text-xs font-medium bg-primary text-primary-contrast rounded hover:bg-primary-hover"
                @click.stop="emit('select-quiz', q.quiz_id)"
              >
                Analytics
              </button>
            </td>
          </tr>
          <tr v-if="!loading && quizzes.length === 0">
            <td colspan="8" class="py-6 text-center text-text-tertiary">
              No owned quizzes found.
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
        {{ loading ? "Loading..." : "Load More Quizzes" }}
      </button>
    </div>
  </div>
</template>
