<template>
  <section
    class="w-full min-w-0 max-w-full rounded-xl border border-jv-ink/10 bg-white/80 p-4 shadow-sm"
    data-testid="recent-activity-table"
  >
    <h2 class="font-headings text-xl text-jv-ink">Recent activity</h2>
    <div class="mt-4 min-w-0 max-w-full overflow-x-auto">
      <table class="w-full min-w-[28rem] text-left text-sm">
        <thead
          class="border-b border-jv-ink/10 text-xs uppercase text-jv-muted"
        >
          <tr>
            <th class="py-2 pr-4">Quiz</th>
            <th class="py-2 pr-4">Status</th>
            <th class="py-2 pr-4">Result</th>
            <th class="py-2 pr-4">Score</th>
            <th class="py-2">When</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in items"
            :key="item.attempt_id"
            class="cursor-pointer border-b border-jv-ink/5 hover:bg-jv-cream/60"
            @click="$emit('select-attempt', item.attempt_id)"
          >
            <td class="py-2 pr-4 font-bold">{{ item.quiz_title }}</td>
            <td class="py-2 pr-4">{{ item.status }}</td>
            <td class="py-2 pr-4">{{ item.result_status }}</td>
            <td class="py-2 pr-4">
              <span v-if="item.result_status === 'Result Pending'"
                >Result Pending</span
              >
              <span v-else-if="item.percentage == null">—</span>
              <span v-else>{{ item.percentage }}%</span>
            </td>
            <td class="py-2">{{ formatWhen(item.activity_at) }}</td>
          </tr>
          <tr v-if="!items.length">
            <td colspan="5" class="py-4 text-jv-muted">No activity yet.</td>
          </tr>
        </tbody>
      </table>
    </div>
    <button
      v-if="hasMore"
      type="button"
      class="mt-4 rounded-md bg-jv-ink px-4 py-2 text-sm font-bold text-jv-cream"
      data-testid="activity-load-more"
      :disabled="loadingMore"
      @click="$emit('load-more')"
    >
      {{ loadingMore ? "Loading…" : "Load more" }}
    </button>
  </section>
</template>

<script setup>
defineProps({
  items: { type: Array, default: () => [] },
  hasMore: { type: Boolean, default: false },
  loadingMore: { type: Boolean, default: false },
});

defineEmits(["select-attempt", "load-more"]);

const formatWhen = (value) => {
  if (!value) return "—";
  try {
    return new Date(value).toLocaleString();
  } catch {
    return value;
  }
};
</script>
