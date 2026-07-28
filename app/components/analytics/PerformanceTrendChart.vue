<template>
  <section
    class="w-full min-w-0 overflow-x-auto rounded-xl border border-jv-ink/10 bg-white/80 p-4 shadow-sm"
    data-testid="performance-trend-chart"
  >
    <div class="flex flex-wrap items-center justify-between gap-3">
      <h2 class="font-headings text-xl text-jv-ink">Performance trends</h2>
      <div class="flex gap-2" role="group" aria-label="Trend granularity">
        <button
          v-for="option in granularities"
          :key="option.value"
          type="button"
          class="rounded-md px-3 py-1 text-sm font-bold"
          :class="
            option.value === modelValue
              ? 'bg-jv-ink text-jv-cream'
              : 'bg-jv-cream text-jv-ink'
          "
          @click="$emit('update:modelValue', option.value)"
        >
          {{ option.label }}
        </button>
      </div>
    </div>
    <div
      class="relative mt-4 h-[220px] w-full min-w-0 max-w-full overflow-hidden"
    >
      <Line v-if="chartData" :data="chartData" :options="chartOptions" />
      <p v-else class="text-sm font-bold text-jv-muted">No trend data yet.</p>
    </div>
  </section>
</template>

<script setup>
import { Line } from "vue-chartjs";
import { computed } from "vue";

const props = defineProps({
  modelValue: { type: String, default: "daily" },
  buckets: { type: Array, default: () => [] },
});

defineEmits(["update:modelValue"]);

const granularities = [
  { value: "daily", label: "Daily" },
  { value: "weekly", label: "Weekly" },
  { value: "monthly", label: "Monthly" },
];

const chartData = computed(() => {
  if (!props.buckets?.length) return null;
  return {
    labels: props.buckets.map((b) => b.label),
    datasets: [
      {
        label: "Average %",
        data: props.buckets.map((b) =>
          b.average_percentage == null ? null : b.average_percentage
        ),
        borderColor: "#1f3a2e",
        backgroundColor: "rgba(31, 58, 46, 0.15)",
        spanGaps: false,
        tension: 0.25,
      },
      {
        label: "Attempts",
        data: props.buckets.map((b) => b.attempt_count || 0),
        borderColor: "#c45c26",
        backgroundColor: "rgba(196, 92, 38, 0.12)",
        tension: 0.25,
        yAxisID: "y1",
      },
    ],
  };
});

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: "bottom" },
  },
  scales: {
    y: {
      beginAtZero: true,
      max: 100,
      title: { display: true, text: "Average %" },
    },
    y1: {
      beginAtZero: true,
      position: "right",
      grid: { drawOnChartArea: false },
      title: { display: true, text: "Attempts" },
    },
  },
};
</script>
