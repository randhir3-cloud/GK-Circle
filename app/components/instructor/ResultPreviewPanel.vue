<template>
  <div class="bg-slate-900 border border-slate-800 rounded-xl p-6 shadow-sm">
    <div class="flex items-center justify-between mb-4">
      <div>
        <h3 class="text-lg font-semibold text-slate-100">
          Learner View Preview
        </h3>
        <p class="text-xs text-slate-400">
          Pure client-side simulation. Does not modify database state.
        </p>
      </div>

      <div
        class="flex items-center space-x-1 bg-slate-950 p-1 rounded-lg border border-slate-800"
      >
        <button
          v-for="mode in ['VISIBLE', 'WITHHELD']"
          :key="mode"
          type="button"
          :class="[
            'px-3 py-1 text-xs font-medium rounded-md transition-colors',
            previewState === mode
              ? 'bg-indigo-600 text-white'
              : 'text-slate-400 hover:text-slate-200',
          ]"
          @click="previewState = mode"
        >
          {{ mode === "VISIBLE" ? "Results Visible" : "Results Withheld" }}
        </button>
      </div>
    </div>

    <!-- Preview Container -->
    <div class="border border-slate-800/80 rounded-lg p-6 bg-slate-950">
      <div v-if="previewState === 'WITHHELD'" class="text-center py-8">
        <div
          class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-amber-500/10 text-amber-400 mb-3"
        >
          <svg
            class="w-6 h-6"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>
        <h4 class="text-base font-semibold text-slate-200">
          Results Pending Release
        </h4>
        <p class="text-xs text-slate-400 max-w-sm mx-auto mt-1">
          Learners will see a pending screen explaining that results are
          withheld until release requirements are met.
        </p>
      </div>

      <div v-else class="space-y-6">
        <!-- Mock Score Summary -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3 text-center">
          <div class="p-3 bg-slate-900 rounded border border-slate-800">
            <span class="text-xs text-slate-400 block">Total Score</span>
            <span class="text-base font-bold text-slate-100">
              {{ mockPermissions.showScore ? "85.0 / 100.0" : "—" }}
            </span>
          </div>

          <div class="p-3 bg-slate-900 rounded border border-slate-800">
            <span class="text-xs text-slate-400 block">Percentage</span>
            <span class="text-base font-bold text-slate-100">
              {{ mockPermissions.showScore ? "85%" : "—" }}
            </span>
          </div>

          <div class="p-3 bg-slate-900 rounded border border-slate-800">
            <span class="text-xs text-slate-400 block">Pass / Fail</span>
            <span class="text-base font-bold text-emerald-400">
              {{ mockPermissions.showPassFail ? "PASSED" : "—" }}
            </span>
          </div>

          <div class="p-3 bg-slate-900 rounded border border-slate-800">
            <span class="text-xs text-slate-400 block">Correctness</span>
            <span class="text-base font-bold text-slate-100">
              {{
                mockPermissions.showCorrectness
                  ? "17 Correct / 3 Incorrect"
                  : "—"
              }}
            </span>
          </div>
        </div>

        <!-- Mock Question Review Section -->
        <div
          v-if="mockPermissions.allowAnswerReview"
          class="p-4 bg-slate-900 rounded-lg border border-slate-800 text-xs"
        >
          <span class="font-semibold text-slate-200 block mb-1"
            >Question 1 Review:</span
          >
          <p class="text-slate-300 mb-2">What is the capital of Bihar?</p>
          <div class="space-y-1">
            <div
              class="p-2 rounded bg-indigo-500/10 text-indigo-300 font-medium"
            >
              Option B: Patna (Selected)
              <span
                v-if="mockPermissions.showCorrectness"
                class="text-emerald-400 font-bold ml-2"
                >✓ Correct</span
              >
            </div>
          </div>
          <div
            v-if="mockPermissions.showExplanations"
            class="mt-3 p-2 bg-slate-950 rounded text-slate-400"
          >
            <strong>Explanation:</strong> Patna has been the capital city of
            Bihar since ancient times.
          </div>
        </div>
        <div v-else class="text-xs text-slate-500 text-center py-2">
          Question review is disabled according to current settings.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";

defineProps({
  mockPermissions: {
    type: Object,
    default: () => ({
      showScore: true,
      showPassFail: true,
      allowAnswerReview: true,
      showCorrectness: true,
      showExplanations: true,
    }),
  },
});

const previewState = ref("VISIBLE");
</script>
