<template>
  <div class="bg-slate-900 border border-slate-800 rounded-xl p-6 shadow-sm">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-semibold text-slate-100">Release Status</h3>
      <span
        :class="[
          'px-3 py-1 text-xs font-semibold rounded-full border',
          isCurrentlyReleased
            ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
            : 'bg-amber-500/10 text-amber-400 border-amber-500/20',
        ]"
      >
        {{ isCurrentlyReleased ? "Results Released" : "Results Withheld" }}
      </span>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
      <div class="bg-slate-950/60 p-4 rounded-lg border border-slate-800/80">
        <span class="text-slate-400 block mb-1">Current Policy</span>
        <span class="font-medium text-slate-200 uppercase tracking-wide">
          {{ policy }}
        </span>
      </div>

      <div class="bg-slate-950/60 p-4 rounded-lg border border-slate-800/80">
        <span class="text-slate-400 block mb-1">Scheduled Release</span>
        <span class="font-medium text-slate-200">
          {{ scheduledAt ? formatDate(scheduledAt) : "Not Scheduled" }}
        </span>
      </div>

      <div class="bg-slate-950/60 p-4 rounded-lg border border-slate-800/80">
        <span class="text-slate-400 block mb-1">Submitted Attempts</span>
        <span class="font-medium text-slate-200">
          {{ totalSubmittedAttempts }}
        </span>
      </div>
    </div>

    <div v-if="releasedAt" class="mt-4 text-xs text-slate-400">
      Manually released at: {{ formatDate(releasedAt) }}
    </div>
  </div>
</template>

<script setup>
defineProps({
  policy: {
    type: String,
    default: "IMMEDIATE",
  },
  isCurrentlyReleased: {
    type: Boolean,
    default: true,
  },
  scheduledAt: {
    type: String,
    default: null,
  },
  releasedAt: {
    type: String,
    default: null,
  },
  totalSubmittedAttempts: {
    type: Number,
    default: 0,
  },
});

function formatDate(isoStr) {
  if (!isoStr) return "";
  try {
    return new Date(isoStr).toLocaleString();
  } catch (e) {
    return isoStr;
  }
}
</script>
