<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-sm p-4"
  >
    <div
      class="w-full max-w-md bg-white dark:bg-slate-800 rounded-xl shadow-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4"
    >
      <div
        class="flex items-center justify-between border-b border-slate-200 dark:border-slate-700 pb-3"
      >
        <h3 class="text-lg font-semibold text-slate-900 dark:text-white">
          Export Analytics Data
        </h3>
        <button
          class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
          @click="$emit('close')"
        >
          <span class="sr-only">Close</span>
          &times;
        </button>
      </div>

      <form class="space-y-4" @submit.prevent="handleSubmit">
        <!-- Export Type Select -->
        <div>
          <label
            class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1"
            >Export Subject</label
          >
          <select
            v-model="form.export_type"
            class="w-full rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white p-2.5 text-sm"
          >
            <option value="PORTFOLIO_OVERVIEW">Portfolio Overview</option>
            <option value="QUIZ_LIST">Quiz List & Aggregate Performance</option>
            <option value="LEARNER_PERFORMANCE">
              Learner Performance Metrics
            </option>
            <option value="RELEASE_MONITORING">
              Result Release Monitoring
            </option>
            <option v-if="quizId" value="QUIZ_SUMMARY">Quiz Summary</option>
            <option v-if="quizId" value="QUIZ_ATTEMPTS">
              Quiz Attempt Breakdown
            </option>
            <option v-if="quizId" value="QUESTION_METRICS">
              Question Quality Metrics
            </option>
          </select>
        </div>

        <!-- Format Select -->
        <div>
          <label
            class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1"
            >File Format</label
          >
          <div class="grid grid-cols-3 gap-3">
            <button
              v-for="fmt in ['CSV', 'XLSX', 'PDF']"
              :key="fmt"
              type="button"
              :class="[
                'py-2 px-3 text-sm font-medium rounded-lg border text-center transition-colors',
                form.export_format === fmt
                  ? 'border-indigo-600 bg-indigo-50 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400 font-semibold'
                  : 'border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700/50 text-slate-700 dark:text-slate-300',
              ]"
              @click="form.export_format = fmt"
            >
              {{ fmt }}
            </button>
          </div>
        </div>

        <!-- Custom Title (Optional) -->
        <div>
          <label
            class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1"
            >Report Title (Optional)</label
          >
          <input
            v-model="form.title"
            type="text"
            placeholder="e.g. Q3 PCS Mock Test Analytics"
            class="w-full rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white p-2.5 text-sm"
          />
        </div>

        <!-- Submit Buttons -->
        <div class="flex items-center justify-end space-x-3 pt-2">
          <button
            type="button"
            class="px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg"
            @click="$emit('close')"
          >
            Cancel
          </button>
          <button
            type="submit"
            :disabled="isSubmitting"
            class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 rounded-lg flex items-center shadow-sm"
          >
            <span v-if="isSubmitting" class="mr-2">Generating...</span>
            <span v-else>Request Export</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from "vue";

const props = defineProps({
  isOpen: { type: Boolean, default: false },
  quizId: { type: String, default: null },
});

const emit = defineEmits(["close", "submitted"]);

const isSubmitting = ref(false);
const form = reactive({
  export_type: props.quizId ? "QUIZ_SUMMARY" : "PORTFOLIO_OVERVIEW",
  export_format: "CSV",
  title: "",
});

const handleSubmit = async () => {
  isSubmitting.value = true;
  try {
    const payload = {
      export_type: form.export_type,
      export_format: form.export_format,
      title: form.title || undefined,
      filters: {},
    };
    if (props.quizId) {
      payload.quiz_id = props.quizId;
    }
    emit("submitted", payload);
  } finally {
    isSubmitting.value = false;
  }
};
</script>
