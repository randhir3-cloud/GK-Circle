<script setup>
defineProps({
  summary: {
    type: Object,
    required: true,
  },
});
</script>

<template>
  <div class="attempt-summary">
    <h2 class="attempt-summary__heading">Performance Summary</h2>
    <div class="attempt-summary__grid">
      <div class="summary-metric summary-metric--answered">
        <span class="summary-metric__value">{{ summary.answered }}</span>
        <span class="summary-metric__label">Answered</span>
      </div>

      <div class="summary-metric summary-metric--correct">
        <span class="summary-metric__value">{{ summary.correct }}</span>
        <span class="summary-metric__label">Correct</span>
      </div>

      <div class="summary-metric summary-metric--incorrect">
        <span class="summary-metric__value">{{ summary.incorrect }}</span>
        <span class="summary-metric__label">Incorrect</span>
      </div>

      <div class="summary-metric summary-metric--unanswered">
        <span class="summary-metric__value">{{ summary.unanswered }}</span>
        <span class="summary-metric__label">Unanswered</span>
      </div>

      <div
        v-if="summary.unscored > 0"
        class="summary-metric summary-metric--unscored"
      >
        <span class="summary-metric__value">{{ summary.unscored }}</span>
        <span class="summary-metric__label">Unscored / Survey</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.attempt-summary {
  padding: 1.5rem;
  background: #ffffff;
  border-radius: 0.75rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  border: 1px solid #e5e7eb;
}

.attempt-summary__heading {
  margin: 0 0 1rem;
  font-size: 1.125rem;
  font-weight: 700;
  color: #1f2937;
  font-family: "Outfit", sans-serif;
}

.attempt-summary__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
  gap: 1rem;
}

.summary-metric {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1rem 0.75rem;
  border-radius: 0.5rem;
  text-align: center;
  border: 1px solid #e5e7eb;
}

.summary-metric__value {
  font-size: 1.75rem;
  font-weight: 800;
  font-family: "Outfit", sans-serif;
  line-height: 1.2;
}

.summary-metric__label {
  font-size: 0.8125rem;
  font-weight: 600;
  margin-top: 0.25rem;
}

.summary-metric--answered {
  background: #f0fdf4;
  border-color: #bbf7d0;
  color: #166534;
}

.summary-metric--correct {
  background: #ecfdf5;
  border-color: #a7f3d0;
  color: #047857;
}

.summary-metric--incorrect {
  background: #fef2f2;
  border-color: #fecaca;
  color: #b91c1c;
}

.summary-metric--unanswered {
  background: #f3f4f6;
  border-color: #e5e7eb;
  color: #4b5563;
}

.summary-metric--unscored {
  background: #fffbe6;
  border-color: #ffe58f;
  color: #d46b08;
}
</style>
