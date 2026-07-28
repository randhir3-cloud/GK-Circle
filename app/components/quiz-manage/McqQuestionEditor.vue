<template>
  <div class="flex flex-col gap-6">
    <div
      v-if="showTypeSelector"
      class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-end"
    >
      <select
        v-model.number="form.type"
        class="h-10 rounded-md border border-jv-ink/30 bg-jv-white px-3 text-[14px] font-semibold text-jv-ink outline-none focus:border-jv-ink"
      >
        <option :value="1">Multiple Choice</option>
        <option :value="2">Survey</option>
      </select>
    </div>

    <div class="space-y-2">
      <label
        class="text-[13px] font-semibold uppercase tracking-wide text-jv-muted"
      >
        Question
      </label>
      <input
        v-model.trim="form.question"
        type="text"
        required
        placeholder="Enter question..."
        class="h-11 w-full rounded-md border border-jv-ink/30 bg-jv-white px-3 text-[15px] font-medium text-jv-ink outline-none placeholder:text-jv-ink/40 focus:border-jv-ink focus:shadow-brutal-sm"
      />

      <div class="flex flex-wrap items-center gap-1.5 pt-1">
        <NavigationLink
          v-for="choice in mediaChoices"
          :key="`question-${choice.value}`"
          variant="toggle"
          type="button"
          :class="mediaButtonClass(form.question_media === choice.value)"
          @click="form.question_media = choice.value"
        >
          <component :is="choice.icon" class="size-3.5" :stroke-width="2.3" />
          <span>{{ choice.label }}</span>
        </NavigationLink>
      </div>

      <textarea
        v-if="form.question_media === 'code'"
        v-model="form.resource"
        rows="5"
        placeholder="Enter code..."
        class="mt-1 w-full resize-none rounded-md border border-jv-ink/30 bg-jv-canvas px-3 py-2 font-mono text-[14px] text-jv-ink outline-none focus:border-jv-ink focus:shadow-brutal-sm"
      ></textarea>

      <label
        v-else-if="form.question_media === 'image'"
        class="mt-1 flex min-h-12 cursor-pointer items-center justify-center gap-2 rounded-md border border-dashed border-jv-ink/35 bg-jv-canvas px-3 py-3 text-[14px] font-medium text-jv-muted transition-colors hover:bg-jv-yellow/20"
      >
        <ImageIcon class="size-4" :stroke-width="2.3" />
        <span>{{ questionFileName || "Upload question image" }}</span>
        <input
          type="file"
          class="hidden"
          accept="image/*"
          @change="handleQuestionImage"
        />
      </label>
    </div>

    <div class="space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <label
          class="text-[13px] font-semibold uppercase tracking-wide text-jv-muted"
        >
          Options
        </label>
        <div class="flex flex-wrap items-center gap-1.5">
          <NavigationLink
            v-for="choice in mediaChoices"
            :key="`options-${choice.value}`"
            variant="toggle"
            type="button"
            :class="mediaButtonClass(form.options_media === choice.value)"
            @click="form.options_media = choice.value"
          >
            <component :is="choice.icon" class="size-3.5" :stroke-width="2.3" />
            <span>{{ choice.label }}</span>
          </NavigationLink>
        </div>
      </div>

      <div class="flex flex-col">
        <div
          v-for="key in optionKeys"
          :key="key"
          class="flex min-w-0 items-center gap-3 border-b border-jv-ink/10 py-2.5 pl-2 pr-1 last:border-b-0"
          :class="optionRowClass(key)"
        >
          <input
            v-if="form.type === 1"
            v-model.number="correctAnswer"
            type="radio"
            :value="Number(key)"
            class="size-5 shrink-0 accent-jv-accent-green"
            :aria-label="`Operational correct answer option ${key}`"
          />

          <span class="w-6 shrink-0 text-[14px] font-bold text-jv-coral">
            {{ optionLetter(key) }}.
          </span>

          <input
            v-if="form.options_media === 'text'"
            v-model.trim="form.options[key]"
            type="text"
            :required="Number(key) <= 2"
            :placeholder="`Option ${optionLetter(key)}`"
            class="h-10 w-full min-w-0 flex-1 rounded-md border border-jv-ink/25 bg-jv-white px-3 text-[15px] font-medium text-jv-ink outline-none placeholder:text-jv-ink/40 focus:border-jv-ink focus:shadow-brutal-sm"
          />

          <textarea
            v-else-if="form.options_media === 'code'"
            v-model="form.options[key]"
            rows="2"
            :placeholder="`Code for option ${optionLetter(key)}`"
            class="min-w-0 flex-1 resize-none rounded-md border border-jv-ink/25 bg-jv-white px-3 py-2 font-mono text-[13px] text-jv-ink outline-none focus:border-jv-ink focus:shadow-brutal-sm"
          ></textarea>

          <label
            v-else
            class="flex h-10 min-w-0 flex-1 cursor-pointer items-center gap-2 rounded-md border border-dashed border-jv-ink/30 bg-jv-white px-3 text-[13px] font-medium text-jv-muted focus-within:shadow-brutal-sm hover:bg-jv-yellow/10"
          >
            <ImageIcon class="size-4 shrink-0" :stroke-width="2.3" />
            <span class="truncate">{{
              optionFileNames[key] || existingImageLabel(key)
            }}</span>
            <input
              type="file"
              class="hidden"
              accept="image/*"
              @change="handleOptionImage($event, key)"
            />
          </label>
        </div>
      </div>
    </div>

    <div
      class="jv-border-rough space-y-4 border border-jv-ink/20 bg-jv-canvas p-4"
    >
      <div>
        <h3
          class="font-body text-[13px] font-bold uppercase tracking-wide text-jv-ink"
        >
          Answer authority
        </h3>
        <p class="mt-1 font-body text-[12px] text-jv-muted">
          Operational answer drives live quiz scoring. Official and
          authoritative keys record documentary and review truth per ADR-024.
        </p>
      </div>

      <div class="space-y-2">
        <label
          for="mcq-answer-review-status"
          class="text-[13px] font-semibold uppercase tracking-wide text-jv-muted"
        >
          Answer review status
        </label>
        <select
          id="mcq-answer-review-status"
          v-model="answerReviewStatus"
          class="h-10 w-full rounded-md border border-jv-ink/30 bg-jv-white px-3 text-[14px] font-semibold text-jv-ink outline-none focus:border-jv-ink"
        >
          <option
            v-for="status in answerReviewStatuses"
            :key="status.value"
            :value="status.value"
          >
            {{ status.label }}
          </option>
        </select>
      </div>

      <div v-if="form.type === 1" class="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <fieldset class="space-y-2">
          <legend
            class="text-[13px] font-semibold uppercase tracking-wide text-jv-muted"
          >
            Official answer key
          </legend>
          <div class="flex flex-wrap gap-3">
            <label
              v-for="key in optionKeys"
              :key="`official-${key}`"
              class="inline-flex items-center gap-2 text-[14px] font-medium text-jv-ink"
            >
              <input
                v-model.number="officialAnswer"
                type="radio"
                :value="Number(key)"
                class="size-4 accent-jv-coral"
              />
              Option {{ optionLetter(key) }}
            </label>
          </div>
        </fieldset>

        <fieldset class="space-y-2">
          <legend
            class="text-[13px] font-semibold uppercase tracking-wide text-jv-muted"
          >
            Authoritative answer
          </legend>
          <div class="flex flex-wrap gap-3">
            <label
              v-for="key in optionKeys"
              :key="`auth-${key}`"
              class="inline-flex items-center gap-2 text-[14px] font-medium text-jv-ink"
            >
              <input
                v-model.number="authoritativeAnswer"
                type="radio"
                :value="Number(key)"
                class="size-4 accent-jv-coral"
              />
              Option {{ optionLetter(key) }}
            </label>
          </div>
        </fieldset>
      </div>

      <div
        v-if="
          answerReviewStatus === 'REVISED' || answerReviewStatus === 'DISPUTED'
        "
        class="grid grid-cols-1 gap-3 sm:grid-cols-2"
      >
        <input
          id="mcq-answer-revision-reason"
          v-model.trim="answerRevisionReason"
          type="text"
          placeholder="Revision reason"
          class="h-10 rounded-md border border-jv-ink/30 px-3 text-[14px] outline-none focus:border-jv-ink"
        />
        <input
          id="mcq-answer-revision-source"
          v-model.trim="answerRevisionSource"
          type="text"
          placeholder="Revision source / evidence"
          class="h-10 rounded-md border border-jv-ink/30 px-3 text-[14px] outline-none focus:border-jv-ink"
        />
      </div>

      <p
        v-if="mode === 'edit' && question?.revision_number"
        class="font-body text-[12px] text-jv-muted"
      >
        Revision {{ question.revision_number }}
        <span v-if="question.lineage_id">
          · lineage {{ question.lineage_id }}
        </span>
      </p>
    </div>

    <QuestionRevisionHistory
      v-if="mode === 'edit' && quizId && questionId"
      ref="revisionHistoryRef"
      :quiz-id="quizId"
      :question-id="questionId"
    />

    <div class="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
      <NavigationLink
        v-if="showCancel"
        type="button"
        class="h-11 bg-jv-white text-jv-ink"
        @click="$emit('cancel')"
      >
        Cancel
      </NavigationLink>
      <NavigationLink
        type="button"
        class="h-11 bg-jv-accent-green text-white"
        :disabled="saving"
        @click="submitForm"
      >
        {{ saving ? "Saving..." : resolvedSubmitLabel }}
      </NavigationLink>
    </div>
  </div>
</template>

<script setup>
import { computed, reactive, ref, watch } from "vue";
import { Code2, ImageIcon, Type } from "lucide-vue-next";
import { usePush } from "notivue";
import NavigationLink from "../common/NavigationLink.vue";
import QuestionRevisionHistory from "./QuestionRevisionHistory.vue";
import {
  ANSWER_REVIEW_STATUSES,
  buildAuthorityPayload,
  parseAnswerKeys,
} from "@/utils/question_authority";

const app = useNuxtApp();
const toast = usePush();
const { maxImageFileSize } = useRuntimeConfig().public;

const props = defineProps({
  question: {
    type: Object,
    default: null,
  },
  mode: {
    type: String,
    default: "create",
  },
  quizId: {
    type: String,
    default: "",
  },
  questionId: {
    type: String,
    default: "",
  },
  saving: {
    type: Boolean,
    default: false,
  },
  showTypeSelector: {
    type: Boolean,
    default: true,
  },
  showCancel: {
    type: Boolean,
    default: true,
  },
  submitLabel: {
    type: String,
    default: "",
  },
});

const emit = defineEmits(["save", "cancel"]);

const mediaChoices = [
  { label: "Text", value: "text", icon: Type },
  { label: "Image", value: "image", icon: ImageIcon },
  { label: "Code", value: "code", icon: Code2 },
];

const answerReviewStatuses = ANSWER_REVIEW_STATUSES;
const revisionHistoryRef = ref(null);

const form = reactive({
  question: "",
  type: 1,
  question_media: "text",
  options_media: "text",
  resource: "",
  options: {
    1: "",
    2: "",
    3: "",
    4: "",
  },
});

const correctAnswer = ref(1);
const officialAnswer = ref(1);
const authoritativeAnswer = ref(1);
const answerReviewStatus = ref("UNREVIEWED");
const answerRevisionReason = ref("");
const answerRevisionSource = ref("");
const questionFileName = ref("");
const optionFileNames = ref({});

const resolvedSubmitLabel = computed(() => {
  if (props.submitLabel) return props.submitLabel;
  return props.mode === "edit" ? "Save Changes" : "Add Question";
});

const mediaButtonClass = (active) =>
  active
    ? "border-jv-ink bg-jv-ink text-jv-white"
    : "border-jv-ink/25 bg-jv-white text-jv-muted hover:border-jv-ink/60 hover:text-jv-ink";

const optionRowClass = (key) =>
  form.type === 1 && Number(key) === Number(correctAnswer.value)
    ? "border-l-4 border-l-jv-accent-green bg-jv-accent-green/25 pl-1"
    : "border-l-4 border-l-transparent";

const resetForm = () => {
  const question = props.question;
  form.question = question?.question || "";
  form.type = Number(question?.question_type_id || 1);
  form.question_media = question?.question_media || "text";
  form.options_media = question?.options_media || "text";
  form.resource = question?.resource || "";
  form.options = {
    1: question?.options?.["1"] || "",
    2: question?.options?.["2"] || "",
    3: question?.options?.["3"] || "",
    4: question?.options?.["4"] || "",
  };

  if (question?.options?.["5"]) {
    form.options[5] = question.options["5"];
  }

  const operational = parseAnswerKeys(question?.correct_answer);
  const operationalKey = Number(operational[0] || 1);
  correctAnswer.value = operationalKey;
  officialAnswer.value = Number(
    parseAnswerKeys(question?.official_answer, operationalKey)[0]
  );
  authoritativeAnswer.value = Number(
    parseAnswerKeys(question?.authoritative_answer, operationalKey)[0]
  );
  answerReviewStatus.value = question?.answer_review_status || "UNREVIEWED";
  answerRevisionReason.value = question?.answer_revision_reason || "";
  answerRevisionSource.value = question?.answer_revision_source || "";
  questionFileName.value = "";
  optionFileNames.value = {};
};

watch(
  () => props.question,
  () => resetForm(),
  { immediate: true, deep: true }
);

const optionKeys = computed(() =>
  Object.keys(form.options)
    .sort((a, b) => Number(a) - Number(b))
    .filter((key) => Number(key) <= 5)
);

const optionLetter = (key) => String.fromCharCode(64 + Number(key));

const validateImageFile = (file) => {
  if (!app.$validImageTypes.includes(file.type)) {
    toast.error(
      "Please upload a valid image file (JPEG, PNG, GIF, WEBP, HEIC, HEIF)."
    );
    return false;
  }
  if (file.size > maxImageFileSize) {
    const limitKb = Math.round(maxImageFileSize / 1024);
    toast.error(`Please upload an image less than ${limitKb} KB.`);
    return false;
  }
  return true;
};

const readAsBase64 = (file) =>
  new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (e) => resolve(e.target.result);
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(file);
  });

const handleQuestionImage = async (event) => {
  const file = event.target.files?.[0];
  if (!file) return;
  if (!validateImageFile(file)) {
    event.target.value = "";
    return;
  }
  form.resource = await readAsBase64(file);
  questionFileName.value = file.name;
};

const handleOptionImage = async (event, key) => {
  const file = event.target.files?.[0];
  if (!file) return;
  if (!validateImageFile(file)) {
    event.target.value = "";
    return;
  }
  form.options[key] = await readAsBase64(file);
  optionFileNames.value = {
    ...optionFileNames.value,
    [key]: file.name,
  };
};

const existingImageLabel = (key) => {
  if (form.options[key]) return `Option ${key} image`;
  return `Upload image for option ${key}`;
};

const submitForm = () => {
  const nonEmptyOptionKeys = optionKeys.value.filter(
    (key) => form.options[key] || form.options_media === "image"
  );
  const operationalAnswers =
    form.type === 1
      ? [Number(correctAnswer.value)]
      : nonEmptyOptionKeys.map((key) => Number(key));

  const authority = buildAuthorityPayload({
    answers: operationalAnswers,
    officialAnswerKeys: [Number(officialAnswer.value)],
    authoritativeAnswerKeys: [Number(authoritativeAnswer.value)],
    answerReviewStatus: answerReviewStatus.value,
    answerRevisionReason: answerRevisionReason.value,
    answerRevisionSource: answerRevisionSource.value,
  });

  emit("save", {
    payload: {
      question: form.question,
      type: Number(form.type),
      options: { ...form.options },
      answers: authority.answers,
      official_answer: authority.official_answer,
      authoritative_answer: authority.authoritative_answer,
      answer_review_status: authority.answer_review_status,
      answer_revision_reason: authority.answer_revision_reason,
      answer_revision_source: authority.answer_revision_source,
      question_media: form.question_media,
      options_media: form.options_media,
      resource: form.resource,
    },
  });
};

defineExpose({
  reloadRevisionHistory: () => revisionHistoryRef.value?.reload?.(),
});
</script>
