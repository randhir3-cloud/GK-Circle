<template>
  <div
    v-if="revisions.length"
    class="jv-border-rough border border-jv-ink/20 bg-jv-canvas p-4"
  >
    <h3
      class="font-body text-[13px] font-bold uppercase tracking-wide text-jv-ink"
    >
      Revision history
    </h3>
    <ul class="mt-3 flex flex-col gap-2">
      <li
        v-for="revision in revisions"
        :key="revision.id"
        class="border-b border-jv-ink/10 pb-2 font-body text-[14px] text-jv-ink last:border-b-0"
      >
        <span class="font-bold">#{{ revision.revision_number }}</span>
        · {{ revision.answer_review_status }}
        <span v-if="revision.created_at">
          · {{ formatRevisionDate(revision.created_at) }}
        </span>
      </li>
    </ul>
  </div>
  <p v-else-if="loaded && !pending" class="font-body text-[13px] text-jv-muted">
    No revision history yet.
  </p>
</template>

<script setup>
import { onMounted, ref, watch } from "vue";
import { useQuizQuestionsApi } from "@/composables/quiz_questions";

const props = defineProps({
  quizId: {
    type: String,
    default: "",
  },
  questionId: {
    type: String,
    default: "",
  },
});

const { listRevisions } = useQuizQuestionsApi();
const revisions = ref([]);
const pending = ref(false);
const loaded = ref(false);

const formatRevisionDate = (value) => {
  if (!value) return "";
  return new Date(value).toLocaleString();
};

const loadRevisions = async () => {
  if (!props.quizId || !props.questionId) {
    revisions.value = [];
    loaded.value = false;
    return;
  }

  pending.value = true;
  try {
    revisions.value =
      (await listRevisions(props.quizId, props.questionId)) || [];
    loaded.value = true;
  } catch (error) {
    console.error("Failed to load revision history", error);
    revisions.value = [];
    loaded.value = true;
  } finally {
    pending.value = false;
  }
};

onMounted(loadRevisions);

watch(
  () => [props.quizId, props.questionId],
  () => loadRevisions()
);

defineExpose({ reload: loadRevisions });
</script>
