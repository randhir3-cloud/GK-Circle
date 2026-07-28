<template>
  <div class="flex flex-col gap-5">
    <McqQuestionEditor
      ref="editorRef"
      :question="question"
      mode="edit"
      :quiz-id="quizId"
      :question-id="questionId"
      :saving="saving"
      :show-type-selector="false"
      submit-label="Save Changes"
      @save="handleSave"
    />
  </div>
</template>

<script setup>
import { ref } from "vue";

import { usePush } from "notivue";

import McqQuestionEditor from "@/components/quiz-manage/McqQuestionEditor.vue";

import {
  getQuizQuestionAPIError,
  useQuizQuestionsApi,
} from "@/composables/quiz_questions";

const toast = usePush();

const { updateQuestion } = useQuizQuestionsApi();

const props = defineProps({
  question: {
    type: Object,

    required: true,

    default: () => ({}),
  },

  quizId: {
    type: String,

    required: false,

    default: "",
  },

  questionId: {
    type: String,

    required: false,

    default: "",
  },
});

const saving = ref(false);

const editorRef = ref(null);

const handleSave = async ({ payload }) => {
  if (!props.quizId || !props.questionId) {
    toast.error("Missing quiz or question identifier.");

    return;
  }

  saving.value = true;

  try {
    await updateQuestion(props.quizId, props.questionId, {
      ...payload,

      points: props.question.points,

      duration_in_seconds: props.question.duration_in_seconds,
    });

    toast.success("Question updated successfully!");

    editorRef.value?.reloadRevisionHistory?.();
  } catch (error) {
    console.error("Failed to update the question", error);

    toast.error(
      getQuizQuestionAPIError(error, "Failed to update the question.")
    );
  } finally {
    saving.value = false;
  }
};
</script>
