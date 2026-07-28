<template>
  <div
    class="min-h-screen bg-slate-950 text-slate-100 py-8 px-4 sm:px-6 lg:px-8"
  >
    <div class="max-w-4xl mx-auto space-y-8">
      <!-- Header -->
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">
            Result Release Administration
          </h1>
          <p class="text-xs text-slate-400 mt-1">Quiz ID: {{ quizId }}</p>
        </div>
      </div>

      <!-- Loading / Error States -->
      <div v-if="isLoading" class="text-center py-12 text-slate-400">
        Loading result release status...
      </div>

      <div
        v-else-if="errorMessage"
        class="p-4 bg-red-500/10 border border-red-500/20 text-red-400 rounded-lg text-sm"
      >
        {{ errorMessage }}
      </div>

      <div v-else class="space-y-8">
        <!-- Status Card -->
        <ResultReleaseStatusCard
          :policy="status.result_release_policy"
          :is-currently-released="status.is_currently_released"
          :scheduled-at="status.results_scheduled_at"
          :released-at="status.results_released_at"
          :total-submitted-attempts="status.total_submitted_attempts"
        />

        <!-- Settings Form -->
        <InstructorResultSettings
          :initial-settings="settingsForm"
          :is-submitting="isSaving"
          @save="handleSaveSettings"
          @open-release-dialog="isReleaseDialogOpen = true"
        />

        <!-- Client-side Simulator Preview Panel -->
        <ResultPreviewPanel :mock-permissions="previewPermissions" />
      </div>

      <!-- Release Confirmation Dialog -->
      <ReleaseResultsDialog
        :is-open="isReleaseDialogOpen"
        :total-submitted-attempts="status.total_submitted_attempts"
        :is-submitting="isReleasing"
        @close="isReleaseDialogOpen = false"
        @confirm="handleConfirmRelease"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import { useQuizResultAdminApi } from "~/composables/quiz_result_admin";

import ResultReleaseStatusCard from "~/components/instructor/ResultReleaseStatusCard.vue";
import InstructorResultSettings from "~/components/instructor/InstructorResultSettings.vue";
import ResultPreviewPanel from "~/components/instructor/ResultPreviewPanel.vue";
import ReleaseResultsDialog from "~/components/instructor/ReleaseResultsDialog.vue";

const route = useRoute();
const quizId = computed(() => route.params.id);

const { getReleaseStatus, updateResultSettings, releaseResults } =
  useQuizResultAdminApi();

const isLoading = ref(true);
const isSaving = ref(false);
const isReleasing = ref(false);
const isReleaseDialogOpen = ref(false);
const errorMessage = ref("");

const status = reactive({
  result_release_policy: "IMMEDIATE",
  results_released: false,
  results_scheduled_at: null,
  results_released_at: null,
  is_currently_released: true,
  show_score: true,
  show_pass_fail: true,
  allow_answer_review: true,
  show_correctness: true,
  show_explanations: true,
  total_submitted_attempts: 0,
});

const settingsForm = computed(() => ({
  result_release_policy: status.result_release_policy,
  results_scheduled_at: status.results_scheduled_at
    ? formatDateForInput(status.results_scheduled_at)
    : "",
  show_score: status.show_score,
  show_pass_fail: status.show_pass_fail,
  allow_answer_review: status.allow_answer_review,
  show_correctness: status.show_correctness,
  show_explanations: status.show_explanations,
}));

const previewPermissions = computed(() => ({
  showScore: status.show_score,
  showPassFail: status.show_pass_fail,
  allowAnswerReview: status.allow_answer_review,
  showCorrectness: status.show_correctness,
  showExplanations: status.show_explanations,
}));

function formatDateForInput(isoStr) {
  if (!isoStr) return "";
  const d = new Date(isoStr);
  const pad = (n) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(
    d.getHours()
  )}:${pad(d.getMinutes())}`;
}

async function loadStatus() {
  isLoading.value = true;
  errorMessage.value = "";
  try {
    const data = await getReleaseStatus(quizId.value);
    Object.assign(status, data);
  } catch (err) {
    errorMessage.value =
      err.data?.message || err.message || "Failed to load release status";
  } finally {
    isLoading.value = false;
  }
}

async function handleSaveSettings(newSettings) {
  isSaving.value = true;
  errorMessage.value = "";
  try {
    const data = await updateResultSettings(quizId.value, newSettings);
    Object.assign(status, data);
  } catch (err) {
    errorMessage.value =
      err.data?.message || err.message || "Failed to update settings";
  } finally {
    isSaving.value = false;
  }
}

async function handleConfirmRelease() {
  isReleasing.value = true;
  errorMessage.value = "";
  try {
    const data = await releaseResults(quizId.value);
    Object.assign(status, data);
    isReleaseDialogOpen.value = false;
  } catch (err) {
    errorMessage.value =
      err.data?.message || err.message || "Failed to release results";
  } finally {
    isReleasing.value = false;
  }
}

onMounted(() => {
  if (quizId.value) {
    loadStatus();
  }
});
</script>
