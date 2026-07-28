<script setup>
import { computed } from "vue";
import { useAssessmentAttemptsApi } from "@/composables/assessment_attempts";
import { createAttemptTimer } from "@/composables/attempt_timer";

const props = defineProps({
  expiresAt: {
    type: String,
    default: null,
  },
  quizId: {
    type: String,
    required: true,
  },
  attemptId: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(["expired", "terminalStatus"]);

const api = useAssessmentAttemptsApi();

const timer = createAttemptTimer({
  expiresAt: props.expiresAt,
  quizId: props.quizId,
  attemptId: props.attemptId,
  api,
  onExpired: () => emit("expired"),
  onTerminalStatus: (status) => emit("terminalStatus", status),
});

const timerClasses = computed(() => ({
  "attempt-timer": true,
  "attempt-timer--warning": timer.isWarning.value,
  "attempt-timer--critical": timer.isCritical.value,
}));
</script>

<template>
  <div v-if="timer.hasDeadline" :class="timerClasses" aria-live="polite">
    <span class="attempt-timer__icon" aria-hidden="true">⏱️</span>
    <span class="attempt-timer__label sr-only">Time remaining:</span>
    <span class="attempt-timer__value">{{ timer.formattedTime.value }}</span>
  </div>
  <div v-else class="attempt-timer attempt-timer--untimed">
    <span class="attempt-timer__icon" aria-hidden="true">♾️</span>
    <span class="attempt-timer__value">Untimed</span>
  </div>
</template>

<style scoped>
.attempt-timer {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.75rem;
  border-radius: 0.375rem;
  border: 1px solid #c2d1e0;
  background: #f4f8fb;
  color: #12263a;
  font-family: "Source Sans 3", "Segoe UI", sans-serif;
  font-weight: 700;
  font-size: 0.95rem;
  transition: all 0.2s ease-in-out;
}

.attempt-timer__icon {
  font-size: 1.05rem;
}

.attempt-timer--warning {
  border-color: #f39c12;
  background: #fffdf5;
  color: #935600;
}

.attempt-timer--critical {
  border-color: #d9534f;
  background: #fdf2f2;
  color: #a94442;
  animation: pulse-border 1.5s infinite;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@keyframes pulse-border {
  0% {
    box-shadow: 0 0 0 0 rgba(217, 83, 79, 0.4);
  }
  70% {
    box-shadow: 0 0 0 6px rgba(217, 83, 79, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(217, 83, 79, 0);
  }
}
</style>
