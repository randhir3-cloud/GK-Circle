<template>
  <section
    class="w-full min-w-0 max-w-full overflow-x-hidden rounded-xl border border-jv-ink/10 bg-white/80 p-4 shadow-sm"
    :data-testid="testId"
  >
    <p class="text-xs font-bold uppercase tracking-wide text-jv-muted">
      {{ label }}
    </p>
    <p class="mt-2 font-headings text-2xl text-jv-ink">{{ value }}</p>
    <p v-if="hint" class="mt-1 text-xs text-jv-muted">{{ hint }}</p>
  </section>
</template>

<script setup>
defineProps({
  label: { type: String, required: true },
  value: { type: [String, Number], required: true },
  hint: { type: String, default: "" },
  testId: { type: String, default: "analytics-summary-card" },
});
</script>
