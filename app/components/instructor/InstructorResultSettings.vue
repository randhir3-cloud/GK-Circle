<template>
  <form
    class="bg-slate-900 border border-slate-800 rounded-xl p-6 shadow-sm space-y-6"
    @submit.prevent="handleSubmit"
  >
    <div>
      <h3 class="text-lg font-semibold text-slate-100">
        Release & Review Settings
      </h3>
      <p class="text-xs text-slate-400">
        Configure when and how learners access their assessment results.
      </p>
    </div>

    <!-- Release Policy Select -->
    <div class="space-y-3">
      <label class="block text-sm font-medium text-slate-200"
        >Release Policy</label
      >
      <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
        <label
          v-for="policyOption in policies"
          :key="policyOption.value"
          :class="[
            'flex flex-col p-4 rounded-lg border cursor-pointer transition-colors',
            form.result_release_policy === policyOption.value
              ? 'bg-indigo-600/10 border-indigo-500/50 text-slate-100'
              : 'bg-slate-950 border-slate-800 text-slate-400 hover:border-slate-700',
          ]"
        >
          <div class="flex items-center space-x-2">
            <input
              v-model="form.result_release_policy"
              type="radio"
              name="policy"
              :value="policyOption.value"
              class="text-indigo-600 focus:ring-indigo-500 bg-slate-900 border-slate-700"
            />
            <span class="font-semibold text-sm">{{ policyOption.label }}</span>
          </div>
          <span class="text-xs text-slate-400 mt-2">{{
            policyOption.description
          }}</span>
        </label>
      </div>
    </div>

    <!-- Scheduled Date Picker (Conditional) -->
    <div
      v-if="form.result_release_policy === 'SCHEDULED'"
      class="space-y-2 bg-slate-950 p-4 rounded-lg border border-slate-800"
    >
      <label class="block text-sm font-medium text-slate-200"
        >Scheduled Release Date & Time (UTC)</label
      >
      <input
        v-model="form.results_scheduled_at"
        type="datetime-local"
        class="w-full md:w-72 bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
        required
      />
      <p class="text-xs text-slate-400">
        Results will be automatically made visible to learners after this
        timestamp.
      </p>
    </div>

    <!-- Review Permission Toggles -->
    <div class="space-y-4 pt-4 border-t border-slate-800">
      <h4
        class="text-sm font-semibold text-slate-200 uppercase tracking-wider text-xs"
      >
        Review Controls
      </h4>

      <div class="space-y-3">
        <label
          class="flex items-center justify-between p-3 bg-slate-950 rounded-lg border border-slate-800/80 cursor-pointer"
        >
          <div>
            <span class="text-sm font-medium text-slate-200 block"
              >Show Total Score & Percentage</span
            >
            <span class="text-xs text-slate-400"
              >Display numerical scores to learners upon release.</span
            >
          </div>
          <input
            v-model="form.show_score"
            type="checkbox"
            class="w-5 h-5 rounded text-indigo-600 focus:ring-indigo-500 bg-slate-900 border-slate-700"
          />
        </label>

        <label
          class="flex items-center justify-between p-3 bg-slate-950 rounded-lg border border-slate-800/80 cursor-pointer"
        >
          <div>
            <span class="text-sm font-medium text-slate-200 block"
              >Show Pass / Fail Badge</span
            >
            <span class="text-xs text-slate-400"
              >Display pass/fail outcome status.</span
            >
          </div>
          <input
            v-model="form.show_pass_fail"
            type="checkbox"
            class="w-5 h-5 rounded text-indigo-600 focus:ring-indigo-500 bg-slate-900 border-slate-700"
          />
        </label>

        <label
          class="flex items-center justify-between p-3 bg-slate-950 rounded-lg border border-slate-800/80 cursor-pointer"
        >
          <div>
            <span class="text-sm font-medium text-slate-200 block"
              >Allow Question & Answer Review</span
            >
            <span class="text-xs text-slate-400"
              >Allow learners to inspect individual question choices.</span
            >
          </div>
          <input
            v-model="form.allow_answer_review"
            type="checkbox"
            class="w-5 h-5 rounded text-indigo-600 focus:ring-indigo-500 bg-slate-900 border-slate-700"
          />
        </label>

        <label
          class="flex items-center justify-between p-3 bg-slate-950 rounded-lg border border-slate-800/80 cursor-pointer"
        >
          <div>
            <span class="text-sm font-medium text-slate-200 block"
              >Show Option Correctness</span
            >
            <span class="text-xs text-slate-400"
              >Indicate whether selected choices were correct or
              incorrect.</span
            >
          </div>
          <input
            v-model="form.show_correctness"
            type="checkbox"
            class="w-5 h-5 rounded text-indigo-600 focus:ring-indigo-500 bg-slate-900 border-slate-700"
          />
        </label>

        <label
          class="flex items-center justify-between p-3 bg-slate-950 rounded-lg border border-slate-800/80 cursor-pointer"
        >
          <div>
            <span class="text-sm font-medium text-slate-200 block"
              >Show Explanations</span
            >
            <span class="text-xs text-slate-400"
              >Display detailed answer explanations and solution notes.</span
            >
          </div>
          <input
            v-model="form.show_explanations"
            type="checkbox"
            class="w-5 h-5 rounded text-indigo-600 focus:ring-indigo-500 bg-slate-900 border-slate-700"
          />
        </label>
      </div>
    </div>

    <!-- Actions -->
    <div
      class="flex items-center justify-between pt-4 border-t border-slate-800"
    >
      <button
        v-if="form.result_release_policy !== 'IMMEDIATE'"
        type="button"
        class="px-4 py-2 text-sm font-medium text-amber-300 bg-amber-500/10 border border-amber-500/20 hover:bg-amber-500/20 rounded-lg transition-colors"
        @click="$emit('open-release-dialog')"
      >
        Manual Release Now
      </button>
      <div v-else></div>

      <button
        type="submit"
        class="px-5 py-2.5 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-lg transition-colors disabled:opacity-50"
        :disabled="isSubmitting"
      >
        {{ isSubmitting ? "Saving..." : "Save Settings" }}
      </button>
    </div>
  </form>
</template>

<script setup>
import { reactive, watch } from "vue";

const props = defineProps({
  initialSettings: {
    type: Object,
    default: () => ({
      result_release_policy: "IMMEDIATE",
      results_scheduled_at: "",
      show_score: true,
      show_pass_fail: true,
      allow_answer_review: true,
      show_correctness: true,
      show_explanations: true,
    }),
  },
  isSubmitting: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["save", "open-release-dialog"]);

const policies = [
  {
    value: "IMMEDIATE",
    label: "Immediate",
    description: "Results are published automatically upon attempt submission.",
  },
  {
    value: "MANUAL",
    label: "Manual Release",
    description:
      "Results remain withheld until manually released by instructor.",
  },
  {
    value: "SCHEDULED",
    label: "Scheduled",
    description: "Results automatically release after a specified date/time.",
  },
];

const form = reactive({ ...props.initialSettings });

watch(
  () => props.initialSettings,
  (newVal) => {
    Object.assign(form, newVal);
  },
  { deep: true }
);

function handleSubmit() {
  let scheduledAtIso = null;
  if (form.result_release_policy === "SCHEDULED" && form.results_scheduled_at) {
    scheduledAtIso = new Date(form.results_scheduled_at).toISOString();
  }

  emit("save", {
    result_release_policy: form.result_release_policy,
    results_scheduled_at: scheduledAtIso,
    show_score: form.show_score,
    show_pass_fail: form.show_pass_fail,
    allow_answer_review: form.allow_answer_review,
    show_correctness: form.show_correctness,
    show_explanations: form.show_explanations,
  });
}
</script>
