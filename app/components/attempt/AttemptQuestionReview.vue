<script setup>
import { computed, ref } from "vue";
import { formatDurationSeconds } from "@/composables/assessment_attempts";

const props = defineProps({
  questions: {
    type: Array,
    required: true,
  },
  canShowCorrectness: {
    type: Boolean,
    default: true,
  },
  canShowExplanations: {
    type: Boolean,
    default: true,
  },
});

const currentIndex = ref(0);
const filter = ref("ALL"); // ALL, CORRECT, INCORRECT, UNANSWERED, MARKED

const filteredQuestions = computed(() => {
  if (filter.value === "CORRECT") {
    return props.questions.filter((q) => q.is_correct === true);
  }
  if (filter.value === "INCORRECT") {
    return props.questions.filter((q) => q.is_correct === false);
  }
  if (filter.value === "UNANSWERED") {
    return props.questions.filter((q) => !q.options?.some((o) => o.selected));
  }
  if (filter.value === "MARKED") {
    return props.questions.filter((q) => q.is_marked_review);
  }
  return props.questions;
});

const activeQuestion = computed(() => {
  if (filteredQuestions.value.length === 0) return null;
  if (currentIndex.value >= filteredQuestions.value.length) {
    return filteredQuestions.value[0];
  }
  return filteredQuestions.value[currentIndex.value];
});

const activeOriginalIndex = computed(() => {
  if (!activeQuestion.value) return 0;
  return props.questions.findIndex((q) => q.id === activeQuestion.value.id);
});

const setFilter = (newFilter) => {
  filter.value = newFilter;
  currentIndex.value = 0;
};

const selectQuestionByOriginalIndex = (idx) => {
  const targetQ = props.questions[idx];
  if (!targetQ) return;
  filter.value = "ALL";
  const filteredIdx = props.questions.findIndex((q) => q.id === targetQ.id);
  currentIndex.value = Math.max(0, filteredIdx);
};

const goNext = () => {
  if (currentIndex.value < filteredQuestions.value.length - 1) {
    currentIndex.value++;
  }
};

const goPrev = () => {
  if (currentIndex.value > 0) {
    currentIndex.value--;
  }
};
</script>

<template>
  <div class="question-review">
    <div class="question-review__header">
      <h2 class="question-review__heading">Question Review</h2>

      <div class="question-review__filters">
        <button
          class="filter-btn"
          :class="{ 'filter-btn--active': filter === 'ALL' }"
          @click="setFilter('ALL')"
        >
          All ({{ questions.length }})
        </button>
        <button
          v-if="canShowCorrectness"
          class="filter-btn filter-btn--correct"
          :class="{ 'filter-btn--active': filter === 'CORRECT' }"
          @click="setFilter('CORRECT')"
        >
          Correct
        </button>
        <button
          v-if="canShowCorrectness"
          class="filter-btn filter-btn--incorrect"
          :class="{ 'filter-btn--active': filter === 'INCORRECT' }"
          @click="setFilter('INCORRECT')"
        >
          Incorrect
        </button>
        <button
          class="filter-btn filter-btn--unanswered"
          :class="{ 'filter-btn--active': filter === 'UNANSWERED' }"
          @click="setFilter('UNANSWERED')"
        >
          Unanswered
        </button>
        <button
          class="filter-btn filter-btn--marked"
          :class="{ 'filter-btn--active': filter === 'MARKED' }"
          @click="setFilter('MARKED')"
        >
          Marked
        </button>
      </div>
    </div>

    <!-- Question Palette Grid -->
    <div class="question-palette">
      <button
        v-for="(q, idx) in questions"
        :key="q.id"
        class="palette-item"
        :class="{
          'palette-item--active': activeOriginalIndex === idx,
          'palette-item--correct': canShowCorrectness && q.is_correct === true,
          'palette-item--incorrect':
            canShowCorrectness && q.is_correct === false,
          'palette-item--unanswered': !q.options?.some((o) => o.selected),
          'palette-item--marked': q.is_marked_review,
        }"
        :aria-label="`Question ${idx + 1}`"
        @click="selectQuestionByOriginalIndex(idx)"
      >
        {{ idx + 1 }}
      </button>
    </div>

    <!-- Active Question View -->
    <div v-if="activeQuestion" class="review-card">
      <div class="review-card__top">
        <div class="review-card__num-badge">
          Question {{ activeOriginalIndex + 1 }} of {{ questions.length }}
        </div>
        <div class="review-card__badges">
          <span
            v-if="activeQuestion.is_marked_review"
            class="badge badge--marked"
          >
            Marked for Review
          </span>
          <template v-if="canShowCorrectness">
            <span
              v-if="activeQuestion.is_correct === true"
              class="badge badge--correct"
            >
              Correct (+{{ activeQuestion.score ?? activeQuestion.points }})
            </span>
            <span
              v-else-if="activeQuestion.is_correct === false"
              class="badge badge--incorrect"
            >
              Incorrect ({{ activeQuestion.score ?? 0 }})
            </span>
            <span v-else class="badge badge--unanswered">
              Unanswered / Unscored
            </span>
          </template>
        </div>
      </div>

      <div class="review-card__stem">
        <p class="review-card__question-text">{{ activeQuestion.question }}</p>
      </div>

      <!-- Options Presentation List -->
      <div class="review-options">
        <div
          v-for="opt in activeQuestion.options"
          :key="opt.id"
          class="review-option"
          :class="{
            'review-option--selected-correct':
              canShowCorrectness && opt.selected && opt.correct === true,
            'review-option--selected-incorrect':
              canShowCorrectness && opt.selected && opt.correct === false,
            'review-option--correct-answer':
              canShowCorrectness && !opt.selected && opt.correct === true,
            'review-option--selected-neutral':
              !canShowCorrectness && opt.selected,
          }"
        >
          <div class="review-option__indicator">
            <span
              v-if="canShowCorrectness && opt.selected && opt.correct === true"
              class="icon icon--check"
              >✓</span
            >
            <span
              v-else-if="
                canShowCorrectness && opt.selected && opt.correct === false
              "
              class="icon icon--cross"
              >✕</span
            >
            <span v-else-if="opt.correct === true" class="icon icon--key"
              >✓ (Key)</span
            >
            <span v-else class="icon icon--num">{{ opt.id }}</span>
          </div>

          <div class="review-option__text">{{ opt.text }}</div>

          <div class="review-option__meta">
            <span v-if="opt.selected" class="tag tag--selected"
              >Your Answer</span
            >
            <span v-if="opt.correct === true" class="tag tag--correct-key"
              >Correct Answer</span
            >
          </div>
        </div>
      </div>

      <!-- Time Taken & Explanation -->
      <div class="review-card__footer">
        <p v-if="activeQuestion.time_taken_seconds" class="review-card__time">
          Time spent:
          {{ formatDurationSeconds(activeQuestion.time_taken_seconds) }}
        </p>

        <div
          v-if="canShowExplanations && activeQuestion.explanation"
          class="explanation-box"
        >
          <h3 class="explanation-box__title">Explanation</h3>
          <p class="explanation-box__text">{{ activeQuestion.explanation }}</p>
        </div>
      </div>

      <!-- Nav Controls -->
      <div class="review-card__nav">
        <button class="nav-btn" :disabled="currentIndex <= 0" @click="goPrev">
          ← Previous
        </button>
        <span class="nav-count">
          {{ currentIndex + 1 }} / {{ filteredQuestions.length }}
        </span>
        <button
          class="nav-btn nav-btn--primary"
          :disabled="currentIndex >= filteredQuestions.length - 1"
          @click="goNext"
        >
          Next →
        </button>
      </div>
    </div>

    <div v-else class="review-card__empty">
      No questions match the selected filter.
    </div>
  </div>
</template>

<style scoped>
.question-review {
  margin-top: 1.5rem;
  padding: 1.5rem;
  background: #ffffff;
  border-radius: 0.75rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  border: 1px solid #e5e7eb;
}

.question-review__header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.25rem;
}

.question-review__heading {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: #111827;
  font-family: "Outfit", sans-serif;
}

.question-review__filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.filter-btn {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
  font-weight: 600;
  border-radius: 0.375rem;
  border: 1px solid #d1d5db;
  background: #ffffff;
  color: #374151;
  cursor: pointer;
  transition: all 0.15s ease;
}

.filter-btn:hover {
  background: #f3f4f6;
}

.filter-btn--active {
  background: #0f6a5a;
  color: #ffffff;
  border-color: #0f6a5a;
}

.question-palette {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
  padding: 0.75rem;
  background: #f9fafb;
  border-radius: 0.5rem;
  border: 1px solid #f3f4f6;
  margin-bottom: 1.5rem;
}

.palette-item {
  width: 2.25rem;
  height: 2.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8125rem;
  font-weight: 700;
  border-radius: 0.375rem;
  border: 1px solid #d1d5db;
  background: #ffffff;
  color: #374151;
  cursor: pointer;
}

.palette-item--active {
  ring: 2px solid #0f6a5a;
  font-weight: 900;
}

.palette-item--correct {
  background: #d1fae5;
  border-color: #6ee7b7;
  color: #065f46;
}

.palette-item--incorrect {
  background: #fee2e2;
  border-color: #fca5a5;
  color: #991b1b;
}

.palette-item--unanswered {
  background: #f3f4f6;
  color: #6b7280;
}

.review-card {
  padding: 1.5rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.625rem;
  background: #ffffff;
}

.review-card__top {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.review-card__num-badge {
  font-size: 0.875rem;
  font-weight: 700;
  color: #0f6a5a;
}

.badge {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 700;
  border-radius: 9999px;
  margin-left: 0.375rem;
}

.badge--correct {
  background: #d1fae5;
  color: #065f46;
}

.badge--incorrect {
  background: #fee2e2;
  color: #991b1b;
}

.badge--unanswered {
  background: #f3f4f6;
  color: #4b5563;
}

.badge--marked {
  background: #fef3c7;
  color: #92400e;
}

.review-card__question-text {
  font-size: 1.0625rem;
  font-weight: 600;
  color: #111827;
  margin-bottom: 1.25rem;
  line-height: 1.5;
}

.review-options {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
  margin-bottom: 1.5rem;
}

.review-option {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  background: #ffffff;
}

.review-option--selected-correct {
  border-color: #10b981;
  background: #ecfdf5;
}

.review-option--selected-incorrect {
  border-color: #ef4444;
  background: #fef2f2;
}

.review-option--correct-answer {
  border-color: #10b981;
  background: #f0fdf4;
  border-style: dashed;
}

.review-option__indicator {
  font-weight: 700;
  font-size: 0.875rem;
  min-width: 1.75rem;
}

.icon--check {
  color: #059669;
}

.icon--cross {
  color: #dc2626;
}

.icon--key {
  color: #059669;
  font-size: 0.75rem;
}

.review-option__text {
  flex: 1;
  font-size: 0.9375rem;
  color: #1f2937;
}

.tag {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
}

.tag--selected {
  background: #e0e7ff;
  color: #3730a3;
}

.tag--correct-key {
  background: #d1fae5;
  color: #065f46;
  margin-left: 0.375rem;
}

.explanation-box {
  margin-top: 1.25rem;
  padding: 1rem;
  background: #f8fafc;
  border-left: 4px solid #0f6a5a;
  border-radius: 0.375rem;
}

.explanation-box__title {
  margin: 0 0 0.375rem;
  font-size: 0.875rem;
  font-weight: 700;
  color: #0f6a5a;
}

.explanation-box__text {
  margin: 0;
  font-size: 0.875rem;
  color: #334155;
  line-height: 1.5;
}

.review-card__nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid #e5e7eb;
}

.nav-btn {
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  border-radius: 0.375rem;
  border: 1px solid #d1d5db;
  background: #ffffff;
  color: #374151;
  cursor: pointer;
}

.nav-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.nav-btn--primary {
  background: #0f6a5a;
  color: #ffffff;
  border-color: #0f6a5a;
}

.nav-btn--primary:hover:not(:disabled) {
  background: #0d594b;
}

.review-card__empty {
  padding: 2rem;
  text-align: center;
  color: #6b7280;
}
</style>
