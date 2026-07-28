<template>
  <form
    class="jv-border-rough bg-jv-white p-4 shadow-brutal-sm sm:p-5 lg:p-6"
    @submit.prevent
  >
    <div
      class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
    >
      <div>
        <p
          class="text-[12px] font-bold uppercase tracking-[0.14em] text-jv-coral"
        >
          {{ eyebrow }}
        </p>

        <h2
          class="mt-1 font-body text-[22px] font-bold leading-tight text-jv-ink sm:text-[26px]"
        >
          {{ title }}
        </h2>
      </div>

      <NuxtLink
        v-if="mode === 'edit' && quizId && questionId"
        :to="`/admin/quiz/list-quiz/${quizId}/${questionId}`"
        class="text-[13px] font-semibold text-jv-coral underline-offset-2 hover:underline"
      >
        Open full editor
      </NuxtLink>
    </div>

    <div class="mt-6">
      <McqQuestionEditor
        ref="editorRef"
        :question="question"
        :mode="mode"
        :quiz-id="quizId"
        :question-id="questionId"
        :saving="saving"
        :show-cancel="showCancel"
        :submit-label="submitLabel"
        @save="$emit('save', $event)"
        @cancel="$emit('cancel')"
      />
    </div>
  </form>
</template>

<script setup>
import { computed, ref } from "vue";

import McqQuestionEditor from "./McqQuestionEditor.vue";

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
});

defineEmits(["save", "cancel"]);

const editorRef = ref(null);

const title = computed(() =>
  props.mode === "edit" ? "Edit Question" : "New Question"
);

const eyebrow = computed(() => (props.mode === "edit" ? "Question" : "Create"));

const submitLabel = computed(() =>
  props.mode === "edit" ? "Save Changes" : "Add Question"
);

const showCancel = computed(
  () => props.mode === "edit" || props.mode === "create"
);

defineExpose({
  reloadRevisionHistory: () => editorRef.value?.reloadRevisionHistory?.(),
});
</script>
