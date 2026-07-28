<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 grid place-items-center bg-jv-ink/35 px-4 py-6 backdrop-blur-[2px]"
      @click.self="handleClose"
    >
      <div
        class="flex max-h-[90vh] w-full max-w-[760px] flex-col border-[4px] border-jv-ink bg-jv-white shadow-brutal-lg"
        role="dialog"
        aria-modal="true"
        aria-labelledby="quiz-import-wizard-title"
      >
        <div
          class="flex shrink-0 items-center justify-between gap-4 border-b-[3px] border-jv-ink bg-jv-ink px-5 py-4 text-jv-white sm:px-6"
        >
          <h2
            id="quiz-import-wizard-title"
            class="font-body text-[24px] font-black leading-none text-jv-white sm:text-[28px]"
          >
            Import Questions from CSV
          </h2>
          <button
            type="button"
            class="grid size-9 place-items-center text-jv-white transition-transform hover:rotate-[6deg]"
            aria-label="Close import wizard"
            :disabled="commitPending"
            @click="handleClose"
          >
            <X class="size-6" :stroke-width="2.4" />
          </button>
        </div>

        <div class="min-h-0 flex-1 overflow-y-auto px-5 py-6 sm:px-8">
          <ol
            class="mb-6 flex flex-wrap gap-2 text-[12px] font-black uppercase tracking-[0.14em]"
          >
            <li
              :class="stepClass('upload')"
              class="rounded-full border-[2px] border-jv-ink px-3 py-1"
            >
              1. Upload
            </li>
            <li
              :class="stepClass('preview')"
              class="rounded-full border-[2px] border-jv-ink px-3 py-1"
            >
              2. Preview
            </li>
            <li
              :class="stepClass('result')"
              class="rounded-full border-[2px] border-jv-ink px-3 py-1"
            >
              3. Result
            </li>
          </ol>

          <section v-if="step === 'upload'" class="grid gap-4">
            <label class="grid gap-2">
              <span
                class="text-[13px] font-black uppercase tracking-[0.16em] text-jv-ink"
              >
                Choose CSV File <span class="text-jv-coral">*</span>
              </span>
              <span
                class="flex h-14 cursor-pointer border-[3px] border-jv-ink bg-jv-canvas"
              >
                <span
                  class="inline-flex h-full items-center gap-2 bg-jv-ink px-4 text-[16px] font-black text-jv-white"
                >
                  <Upload class="size-4" :stroke-width="2.4" />
                  Choose File
                </span>
                <span
                  class="flex min-w-0 flex-1 items-center px-4 text-[15px] font-semibold text-jv-muted"
                >
                  {{ csvFileName || "No file chosen" }}
                </span>
                <input
                  ref="fileInputRef"
                  type="file"
                  class="hidden"
                  accept=".csv,text/csv"
                  @change="handleCsvFile"
                />
              </span>
            </label>
            <p class="text-[15px] leading-[1.6] text-jv-muted">
              Upload a CSV file to validate rows before adding questions to this
              quiz. Invalid rows are reported separately; only valid rows can be
              committed.
            </p>
            <NavigationLink
              url="/files/demo.csv"
              url-name="Download Sample"
              external
              download="demo.csv"
              class="w-fit rounded-[999px] bg-jv-white"
            >
              <Download class="size-4" :stroke-width="2.4" />
            </NavigationLink>
            <p
              v-if="uploadError"
              class="text-[15px] font-semibold text-jv-coral"
            >
              {{ uploadError }}
            </p>
          </section>

          <section
            v-else-if="step === 'preview' && previewJob"
            class="grid gap-4"
          >
            <div class="grid gap-2 border-[3px] border-jv-ink bg-jv-canvas p-4">
              <p class="text-[15px] font-black text-jv-ink">
                {{ previewJob.source_filename }}
              </p>
              <p class="text-[14px] text-jv-muted">
                {{ previewJob.valid_row_count }} valid ·
                {{ previewJob.error_row_count }} invalid ·
                {{ previewJob.total_rows }} total rows
              </p>
            </div>

            <div
              v-if="duplicateErrors.length"
              class="grid gap-2 border-[3px] border-jv-coral bg-jv-white p-4"
            >
              <h3
                class="text-[14px] font-black uppercase tracking-[0.12em] text-jv-coral"
              >
                Duplicate rows
              </h3>
              <ul class="grid max-h-40 gap-2 overflow-y-auto text-[14px]">
                <li
                  v-for="rowError in duplicateErrors"
                  :key="`dup-${rowError.row_number}`"
                  class="border-l-[3px] border-jv-coral pl-3"
                >
                  <span class="font-black">Row {{ rowError.row_number }}:</span>
                  {{ rowError.messages.join("; ") }}
                </li>
              </ul>
            </div>

            <div
              v-if="validationErrors.length"
              class="grid gap-2 border-[3px] border-jv-coral bg-jv-white p-4"
            >
              <h3
                class="text-[14px] font-black uppercase tracking-[0.12em] text-jv-coral"
              >
                Invalid rows
              </h3>
              <ul class="grid max-h-40 gap-2 overflow-y-auto text-[14px]">
                <li
                  v-for="rowError in validationErrors"
                  :key="`err-${rowError.row_number}`"
                  class="border-l-[3px] border-jv-coral pl-3"
                >
                  <span class="font-black">Row {{ rowError.row_number }}:</span>
                  {{ rowError.messages.join("; ") }}
                </li>
              </ul>
            </div>

            <div
              v-if="previewJob.preview?.valid_rows?.length"
              class="grid gap-2 border-[3px] border-jv-accent-green bg-jv-white p-4"
            >
              <h3
                class="text-[14px] font-black uppercase tracking-[0.12em] text-jv-ink"
              >
                Valid rows (preview)
              </h3>
              <ul class="grid max-h-48 gap-2 overflow-y-auto text-[14px]">
                <li
                  v-for="row in previewJob.preview.valid_rows"
                  :key="`valid-${row.row_number}`"
                  class="border-l-[3px] border-jv-accent-green pl-3"
                >
                  <span class="font-black">Row {{ row.row_number }}:</span>
                  {{ row.question }}
                </li>
              </ul>
            </div>

            <p
              v-if="previewJob.valid_row_count === 0"
              class="text-[15px] font-semibold text-jv-coral"
            >
              No valid rows to import. Fix the CSV and upload again.
            </p>
            <p
              v-if="commitError"
              class="text-[15px] font-semibold text-jv-coral"
            >
              {{ commitError }}
            </p>
          </section>

          <section
            v-else-if="step === 'result' && resultJob"
            class="grid gap-4"
          >
            <div
              class="grid gap-2 border-[3px] border-jv-accent-green bg-jv-canvas p-4"
            >
              <h3 class="text-[18px] font-black text-jv-ink">
                Import complete
              </h3>
              <p class="text-[15px] text-jv-muted">
                {{ resultJob.commit_result?.committed_count || 0 }} question(s)
                added to this quiz.
              </p>
            </div>
            <p
              v-if="previewJob?.error_row_count"
              class="text-[14px] text-jv-muted"
            >
              {{ previewJob.error_row_count }} row(s) were skipped due to
              validation errors.
            </p>
          </section>
        </div>

        <div
          class="flex shrink-0 flex-col-reverse gap-3 border-t-[3px] border-jv-ink bg-jv-canvas px-5 py-4 sm:flex-row sm:justify-end sm:px-8"
        >
          <button
            v-if="step !== 'result'"
            type="button"
            class="inline-flex h-12 items-center justify-center border-[3px] border-jv-ink bg-jv-white px-6 text-[17px] font-black text-jv-ink shadow-brutal-sm transition-transform hover:rotate-[-1deg] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="uploadPending || commitPending"
            @click="handleClose"
          >
            Cancel
          </button>

          <button
            v-if="step === 'upload'"
            type="button"
            class="inline-flex h-12 items-center justify-center border-[3px] border-jv-ink bg-jv-accent-green px-6 text-[17px] font-black text-jv-ink shadow-brutal-sm transition-transform hover:rotate-[1deg] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="!csvFile || uploadPending"
            @click="uploadPreview"
          >
            {{ uploadPending ? "Validating..." : "Validate CSV" }}
          </button>

          <button
            v-if="step === 'preview'"
            type="button"
            class="inline-flex h-12 items-center justify-center border-[3px] border-jv-ink bg-jv-white px-6 text-[17px] font-black text-jv-ink shadow-brutal-sm transition-transform hover:rotate-[-1deg] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="uploadPending || commitPending"
            @click="resetToUpload"
          >
            Upload another file
          </button>

          <button
            v-if="step === 'preview'"
            type="button"
            class="inline-flex h-12 items-center justify-center border-[3px] border-jv-ink bg-jv-accent-green px-6 text-[17px] font-black text-jv-ink shadow-brutal-sm transition-transform hover:rotate-[1deg] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="!canCommit || commitPending"
            @click="commitImport"
          >
            {{
              commitPending
                ? "Importing..."
                : `Import ${previewJob?.valid_row_count || 0} question(s)`
            }}
          </button>

          <button
            v-if="step === 'result'"
            type="button"
            class="inline-flex h-12 items-center justify-center border-[3px] border-jv-ink bg-jv-accent-green px-6 text-[17px] font-black text-jv-ink shadow-brutal-sm transition-transform hover:rotate-[1deg] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none"
            @click="handleDone"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { Download, Upload, X } from "lucide-vue-next";
import NavigationLink from "~/components/common/NavigationLink.vue";
import { useQuizImportJobsApi } from "@/composables/quiz_import_jobs";

const props = defineProps({
  open: { type: Boolean, default: false },
  quizId: { type: String, required: true },
});

const emit = defineEmits(["update:open", "imported"]);

const { createPreviewJob, commitPreviewJob, getQuizImportJobAPIError } =
  useQuizImportJobsApi();

const step = ref("upload");
const csvFile = ref(null);
const csvFileName = ref("");
const fileInputRef = ref(null);
const uploadPending = ref(false);
const commitPending = ref(false);
const uploadError = ref("");
const commitError = ref("");
const previewJob = ref(null);
const resultJob = ref(null);

const canCommit = computed(
  () =>
    Boolean(previewJob.value?.valid_row_count) &&
    previewJob.value?.status !== "COMMITTED"
);

const previewErrors = computed(() => previewJob.value?.preview?.errors || []);

const duplicateErrors = computed(() =>
  previewErrors.value.filter((rowError) => rowError.kind === "duplicate")
);

const validationErrors = computed(() =>
  previewErrors.value.filter((rowError) => rowError.kind !== "duplicate")
);

const stepClass = (target) => {
  const order = { upload: 1, preview: 2, result: 3 };
  const current = order[step.value] || 1;
  return order[target] <= current
    ? "bg-jv-ink text-jv-white"
    : "bg-jv-white text-jv-muted";
};

const resetState = () => {
  step.value = "upload";
  csvFile.value = null;
  csvFileName.value = "";
  uploadPending.value = false;
  commitPending.value = false;
  uploadError.value = "";
  commitError.value = "";
  previewJob.value = null;
  resultJob.value = null;
  if (fileInputRef.value) {
    fileInputRef.value.value = "";
  }
};

const handleClose = () => {
  if (commitPending.value) {
    return;
  }
  emit("update:open", false);
};

const handleDone = () => {
  emit("imported");
  emit("update:open", false);
};

const handleCsvFile = (event) => {
  const file = event.target.files?.[0];
  csvFile.value = file || null;
  csvFileName.value = file?.name || "";
  uploadError.value = "";
};

const resetToUpload = () => {
  step.value = "upload";
  csvFile.value = null;
  csvFileName.value = "";
  uploadError.value = "";
  commitError.value = "";
  previewJob.value = null;
  if (fileInputRef.value) {
    fileInputRef.value.value = "";
  }
};

const uploadPreview = async () => {
  if (!csvFile.value || !props.quizId) {
    uploadError.value = "Please select a CSV file.";
    return;
  }

  try {
    uploadPending.value = true;
    uploadError.value = "";
    previewJob.value = await createPreviewJob(props.quizId, csvFile.value);
    step.value = "preview";
  } catch (error) {
    uploadError.value = getQuizImportJobAPIError(
      error,
      "Failed to validate CSV."
    );
  } finally {
    uploadPending.value = false;
  }
};

const commitImport = async () => {
  if (!previewJob.value?.id || commitPending.value || !canCommit.value) {
    return;
  }

  try {
    commitPending.value = true;
    commitError.value = "";
    resultJob.value = await commitPreviewJob(props.quizId, previewJob.value.id);
    previewJob.value = resultJob.value;
    step.value = "result";
  } catch (error) {
    const status = error?.status || error?.statusCode;
    if (status === 409) {
      commitError.value =
        "Import is already in progress. Wait a moment and try again.";
    } else {
      commitError.value = getQuizImportJobAPIError(
        error,
        "Failed to import questions."
      );
    }
  } finally {
    commitPending.value = false;
  }
};

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) {
      resetState();
    }
  }
);
</script>
