<template>
  <div
    v-if="job"
    class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4 space-y-3"
  >
    <div class="flex items-center justify-between">
      <div>
        <h4 class="text-sm font-semibold text-slate-900 dark:text-white">
          {{ job.title || "Export Task" }}
        </h4>
        <p class="text-xs text-slate-500 dark:text-slate-400">
          Format: {{ job.export_format }} | Type: {{ job.export_type }}
        </p>
      </div>

      <div class="flex items-center space-x-2">
        <span
          :class="[
            'px-2.5 py-1 text-xs font-medium rounded-full',
            job.status === 'COMPLETED'
              ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300'
              : job.status === 'RUNNING'
              ? 'bg-amber-100 text-amber-800 dark:bg-amber-950/60 dark:text-amber-300 animate-pulse'
              : job.status === 'FAILED'
              ? 'bg-rose-100 text-rose-800 dark:bg-rose-950/60 dark:text-rose-300'
              : 'bg-slate-100 text-slate-800 dark:bg-slate-700 dark:text-slate-300',
          ]"
        >
          {{ job.status }}
        </span>

        <a
          v-if="job.status === 'COMPLETED'"
          :href="downloadUrl"
          target="_blank"
          download
          class="px-3 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg shadow-sm"
        >
          Download File
        </a>
      </div>
    </div>

    <!-- Progress Indicator -->
    <div
      v-if="job.status === 'QUEUED' || job.status === 'RUNNING'"
      class="w-full bg-slate-100 dark:bg-slate-700 rounded-full h-1.5 overflow-hidden"
    >
      <div
        class="bg-indigo-600 h-1.5 rounded-full transition-all duration-500"
        :style="{ width: job.status === 'RUNNING' ? '65%' : '20%' }"
      ></div>
    </div>

    <div
      v-if="job.status === 'FAILED'"
      class="text-xs text-rose-600 dark:text-rose-400"
    >
      Error:
      {{ job.error_message || "Export failed during background generation." }}
    </div>
  </div>
</template>

<script setup>
defineProps({
  job: { type: Object, required: true },
  downloadUrl: { type: String, default: "" },
});
</script>
