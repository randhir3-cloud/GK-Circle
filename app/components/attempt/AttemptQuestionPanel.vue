<script setup>
import { computed } from "vue";
import {
  QUESTION_TYPE_SURVEY,
  SAVE_STATUS,
} from "@/utils/attempt_player_constants";

const props = defineProps({
  item: {
    type: Object,
    required: true,
  },
  index: {
    type: Number,
    required: true,
  },
  total: {
    type: Number,
    required: true,
  },
  draftSelection: {
    type: Array,
    required: true,
  },
  saveStatus: {
    type: String,
    required: true,
  },
  saveError: {
    type: String,
    default: "",
  },
  disabled: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["toggle-option", "clear-answer", "retry-save"]);

const optionEntries = computed(() =>
  Object.entries(props.item?.options || {}).sort(
    ([leftKey], [rightKey]) => Number(leftKey) - Number(rightKey)
  )
);

const isSurvey = computed(() => props.item?.type === QUESTION_TYPE_SURVEY);
const inputType = computed(() =>
  props.item?.type === QUESTION_TYPE_SURVEY ? "checkbox" : "radio"
);

const isSelected = (key) => props.draftSelection.includes(Number(key));

const saveMessage = computed(() => {
  if (props.saveStatus === SAVE_STATUS.SAVING) return "Saving your answer…";
  if (props.saveStatus === SAVE_STATUS.FAILED) {
    return props.saveError || "Could not save your answer.";
  }
  if (props.saveStatus === SAVE_STATUS.SAVED) return "Answer saved.";
  return "";
});

const isFetchableUrl = (value) =>
  typeof value === "string" && /^(https?:|data:|blob:)/i.test(value.trim());
</script>

<template>
  <section
    class="attempt-question"
    :aria-labelledby="`attempt-question-${item.question_id}`"
  >
    <header class="attempt-question__header">
      <p class="attempt-question__progress">
        Question {{ index + 1 }} of {{ total }}
      </p>
      <h2
        :id="`attempt-question-${item.question_id}`"
        class="attempt-question__stem"
      >
        {{ item.question }}
      </h2>
    </header>

    <fieldset class="attempt-question__options" :disabled="disabled">
      <legend class="attempt-question__legend">
        {{ isSurvey ? "Select all that apply" : "Select one answer" }}
      </legend>
      <label
        v-for="[key, label] in optionEntries"
        :key="key"
        class="attempt-question__option"
        :class="{ 'attempt-question__option--selected': isSelected(key) }"
      >
        <input
          :type="inputType"
          :name="`question-${item.question_id}`"
          :value="key"
          :checked="isSelected(key)"
          @change="emit('toggle-option', Number(key))"
        />
        <span class="attempt-question__option-key" aria-hidden="true">
          {{ String.fromCharCode(64 + Number(key)) }}
        </span>
        <span class="attempt-question__option-body">
          <img
            v-if="item.options_media === 'image' && isFetchableUrl(label)"
            :src="label"
            :alt="`Option ${key}`"
            class="attempt-question__option-image"
          />
          <span v-else>{{ label }}</span>
        </span>
      </label>
    </fieldset>

    <div class="attempt-question__actions">
      <button
        type="button"
        class="attempt-question__secondary"
        :disabled="disabled || draftSelection.length === 0"
        @click="emit('clear-answer')"
      >
        Clear answer
      </button>
    </div>

    <p
      v-if="saveMessage"
      class="attempt-question__save-status"
      role="status"
      aria-live="polite"
      :class="{
        'attempt-question__save-status--error':
          saveStatus === SAVE_STATUS.FAILED,
      }"
    >
      {{ saveMessage }}
      <button
        v-if="saveStatus === SAVE_STATUS.FAILED"
        type="button"
        class="attempt-question__retry"
        @click="emit('retry-save')"
      >
        Retry save
      </button>
    </p>
  </section>
</template>

<style scoped>
.attempt-question {
  display: grid;
  gap: 1rem;
}

.attempt-question__progress {
  margin: 0;
  font-size: 0.85rem;
  color: #4a5d73;
}

.attempt-question__stem {
  margin: 0.35rem 0 0;
  font-family: "Literata", "Georgia", serif;
  font-size: clamp(1.15rem, 3.5vw, 1.45rem);
  line-height: 1.45;
}

.attempt-question__options {
  border: 0;
  margin: 0;
  padding: 0;
  min-width: 0;
}

.attempt-question__legend {
  font-weight: 600;
  margin-bottom: 0.75rem;
}

.attempt-question__option {
  display: grid;
  grid-template-columns: auto auto 1fr;
  gap: 0.65rem;
  align-items: start;
  padding: 0.75rem;
  margin-bottom: 0.55rem;
  border: 1px solid #d7e0ea;
  border-radius: 0.4rem;
  cursor: pointer;
}

.attempt-question__option--selected {
  border-color: #0f6a5a;
  background: #f3fbf8;
}

.attempt-question__option input {
  margin-top: 0.2rem;
}

.attempt-question__option-key {
  display: inline-grid;
  place-items: center;
  width: 1.6rem;
  height: 1.6rem;
  border-radius: 0.3rem;
  border: 1px solid #12263a;
  font-size: 0.85rem;
  font-weight: 700;
}

.attempt-question__option-body {
  line-height: 1.45;
  word-break: break-word;
}

.attempt-question__option-image {
  max-height: 7rem;
  width: auto;
}

.attempt-question__actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.attempt-question__secondary,
.attempt-question__retry {
  appearance: none;
  border: 1px solid #b8c6d4;
  border-radius: 0.35rem;
  background: #fff;
  color: #12263a;
  font: inherit;
  padding: 0.55rem 0.85rem;
  cursor: pointer;
}

.attempt-question__retry {
  margin-left: 0.5rem;
  border-color: #b42318;
  color: #b42318;
}

.attempt-question__save-status {
  margin: 0;
  font-size: 0.92rem;
  color: #0f6a5a;
}

.attempt-question__save-status--error {
  color: #8a2f1d;
}

.attempt-question__secondary:focus-visible,
.attempt-question__retry:focus-visible,
.attempt-question__option:focus-within {
  outline: 2px solid #12263a;
  outline-offset: 2px;
}
</style>
