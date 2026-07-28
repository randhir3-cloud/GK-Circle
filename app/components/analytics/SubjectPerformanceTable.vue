<template>
  <section
    class="w-full min-w-0 max-w-full rounded-xl border border-jv-ink/10 bg-white/80 p-4 shadow-sm"
    data-testid="subject-performance-table"
  >
    <h2 class="font-headings text-xl text-jv-ink">Subject performance</h2>
    <div class="mt-4 min-w-0 max-w-full overflow-x-auto">
      <table class="w-full min-w-[28rem] text-left text-sm">
        <thead
          class="border-b border-jv-ink/10 text-xs uppercase text-jv-muted"
        >
          <tr>
            <th class="py-2 pr-4">Subject</th>
            <th class="py-2 pr-4">Attempts</th>
            <th class="py-2 pr-4">Scored</th>
            <th class="py-2 pr-4">Avg %</th>
            <th class="py-2">Duration</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="subject in subjects"
            :key="subject.subject_id || subject.subject_name"
            class="border-b border-jv-ink/5"
          >
            <td class="py-2 pr-4 font-bold">{{ subject.subject_name }}</td>
            <td class="py-2 pr-4">{{ subject.attempt_count }}</td>
            <td class="py-2 pr-4">{{ subject.scored_attempt_count }}</td>
            <td class="py-2 pr-4">
              {{
                subject.average_percentage == null
                  ? "—"
                  : `${subject.average_percentage}%`
              }}
            </td>
            <td class="py-2">
              {{ formatSeconds(subject.assessment_duration_seconds) }}
            </td>
          </tr>
          <tr v-if="!subjects.length">
            <td colspan="5" class="py-4 text-jv-muted">No subject data yet.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup>
defineProps({
  subjects: { type: Array, default: () => [] },
});

const formatSeconds = (value) => {
  const total = Math.max(0, Math.floor(Number(value) || 0));
  const minutes = Math.floor(total / 60);
  const secs = total % 60;
  return minutes > 0 ? `${minutes}m ${secs}s` : `${secs}s`;
};
</script>
