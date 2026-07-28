<template>
  <section
    class="w-full min-w-0 max-w-full overflow-x-hidden rounded-xl border border-jv-ink/10 bg-white/80 p-4 shadow-sm"
    data-testid="study-time-card"
  >
    <h2 class="font-headings text-xl text-jv-ink">Study time</h2>
    <dl class="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
      <div>
        <dt class="text-xs font-bold uppercase text-jv-muted">
          Assessment duration
        </dt>
        <dd class="mt-1 text-lg font-bold text-jv-ink">
          {{ formatSeconds(assessmentDurationSeconds) }}
        </dd>
      </div>
      <div>
        <dt class="text-xs font-bold uppercase text-jv-muted">
          Engaged question time
          <span
            class="ml-1 font-normal normal-case"
            title="Approximate telemetry metric capped per question and by attempt duration"
            >≈</span
          >
        </dt>
        <dd class="mt-1 text-lg font-bold text-jv-ink">
          {{ formatSeconds(engagedQuestionTimeSeconds) }}
        </dd>
        <p class="mt-1 text-xs text-jv-muted">Approximate telemetry estimate</p>
      </div>
    </dl>
  </section>
</template>

<script setup>
defineProps({
  assessmentDurationSeconds: { type: Number, default: 0 },
  engagedQuestionTimeSeconds: { type: Number, default: 0 },
});

const formatSeconds = (value) => {
  const total = Math.max(0, Math.floor(Number(value) || 0));
  const hours = Math.floor(total / 3600);
  const minutes = Math.floor((total % 3600) / 60);
  const secs = total % 60;
  if (hours > 0) return `${hours}h ${minutes}m`;
  if (minutes > 0) return `${minutes}m ${secs}s`;
  return `${secs}s`;
};
</script>
