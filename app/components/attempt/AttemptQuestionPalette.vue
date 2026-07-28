<script setup>
import { PALETTE_STATUS } from "@/utils/attempt_player_constants";

defineProps({
  items: {
    type: Array,
    required: true,
  },
  currentIndex: {
    type: Number,
    required: true,
  },
  paletteStatusFor: {
    type: Function,
    required: true,
  },
});

const emit = defineEmits(["select-index"]);

const statusLabel = (status) => {
  switch (status) {
    case PALETTE_STATUS.ANSWERED:
      return "Answered";
    case PALETTE_STATUS.SAVING:
      return "Saving";
    case PALETTE_STATUS.SAVE_FAILED:
      return "Save failed";
    case PALETTE_STATUS.VISITED_UNANSWERED:
      return "Visited, unanswered";
    default:
      return "Not visited";
  }
};
</script>

<template>
  <nav class="attempt-palette" aria-label="Question palette">
    <ol class="attempt-palette__list">
      <li v-for="(item, index) in items" :key="item.question_id">
        <button
          type="button"
          class="attempt-palette__button"
          :class="[
            `attempt-palette__button--${paletteStatusFor(item.question_id)}`,
            { 'attempt-palette__button--current': index === currentIndex },
          ]"
          :aria-current="index === currentIndex ? 'step' : undefined"
          :aria-label="`Question ${index + 1}, ${statusLabel(
            paletteStatusFor(item.question_id)
          )}`"
          @click="emit('select-index', index)"
        >
          <span aria-hidden="true">{{ index + 1 }}</span>
          <span class="attempt-palette__status-dot" aria-hidden="true"></span>
        </button>
      </li>
    </ol>
  </nav>
</template>

<style scoped>
.attempt-palette__list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  list-style: none;
  margin: 0;
  padding: 0;
}

.attempt-palette__button {
  position: relative;
  min-width: 2.4rem;
  min-height: 2.4rem;
  border: 1px solid #b8c6d4;
  border-radius: 0.35rem;
  background: #fff;
  color: #12263a;
  font: inherit;
  font-weight: 600;
  cursor: pointer;
}

.attempt-palette__button--current {
  outline: 2px solid #12263a;
  outline-offset: 2px;
}

.attempt-palette__button--answered {
  background: #e7f6ef;
  border-color: #0f6a5a;
}

.attempt-palette__button--visited_unanswered {
  background: #f4f7fa;
}

.attempt-palette__button--saving {
  background: #fff8e6;
  border-color: #c58a00;
}

.attempt-palette__button--save_failed {
  background: #fdecea;
  border-color: #b42318;
}

.attempt-palette__status-dot {
  position: absolute;
  right: 0.2rem;
  bottom: 0.2rem;
  width: 0.35rem;
  height: 0.35rem;
  border-radius: 999px;
  background: currentColor;
}

.attempt-palette__button--answered .attempt-palette__status-dot {
  color: #0f6a5a;
}

.attempt-palette__button--save_failed .attempt-palette__status-dot {
  color: #b42318;
}

.attempt-palette__button--saving .attempt-palette__status-dot {
  color: #c58a00;
}

.attempt-palette__button:focus-visible {
  outline: 2px solid #12263a;
  outline-offset: 2px;
}
</style>
