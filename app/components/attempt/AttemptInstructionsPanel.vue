<script setup>
import { computed } from "vue";
import { formatDurationSeconds } from "@/composables/assessment_attempts";

const props = defineProps({
  instructions: {
    type: Object,
    required: true,
  },
  starting: {
    type: Boolean,
    default: false,
  },
  resuming: {
    type: Boolean,
    default: false,
  },
  actionError: {
    type: String,
    default: "",
  },
});

const emit = defineEmits(["start", "resume"]);

const quiz = computed(() => props.instructions?.quiz || {});
const snapshot = computed(() => props.instructions?.snapshot || {});
const activeAttempt = computed(
  () => props.instructions?.active_attempt || null
);

const durationLabel = computed(() =>
  formatDurationSeconds(quiz.value.duration_seconds)
);

const attemptsLeft = computed(() => {
  const max = Number(quiz.value.max_attempts || 0);
  const used = Number(props.instructions?.attempts_consumed || 0);
  return Math.max(0, max - used);
});
</script>

<template>
  <section
    class="attempt-instructions"
    aria-labelledby="attempt-instructions-title"
  >
    <header class="attempt-instructions__header">
      <p class="attempt-instructions__eyebrow">Self-paced test</p>
      <h1 id="attempt-instructions-title" class="attempt-instructions__title">
        {{ quiz.title || "Practice test" }}
      </h1>
      <p v-if="quiz.description" class="attempt-instructions__description">
        {{ quiz.description }}
      </p>
    </header>

    <ul class="attempt-instructions__rules" aria-label="Test rules">
      <li>
        <span class="attempt-instructions__label">Questions</span>
        <span>{{ snapshot.question_count }}</span>
      </li>
      <li>
        <span class="attempt-instructions__label">Duration</span>
        <span>{{ durationLabel }}</span>
      </li>
      <li>
        <span class="attempt-instructions__label">Attempts remaining</span>
        <span>{{ attemptsLeft }} of {{ quiz.max_attempts }}</span>
      </li>
      <li>
        <span class="attempt-instructions__label">Negative marking</span>
        <span>
          {{
            Number(quiz.negative_marks_per_question) > 0
              ? `−${quiz.negative_marks_per_question} per incorrect answer`
              : "None"
          }}
        </span>
      </li>
    </ul>

    <div class="attempt-instructions__guidance">
      <p>
        Read each question carefully. Your answers are saved on the server. You
        can mark questions for review before final submission.
      </p>
      <p
        v-if="instructions.block_reason"
        class="attempt-instructions__block"
        role="status"
      >
        {{ instructions.block_reason }}
      </p>
      <p v-if="actionError" class="attempt-instructions__error" role="alert">
        {{ actionError }}
      </p>
    </div>

    <div class="attempt-instructions__actions">
      <button
        v-if="instructions.can_resume && activeAttempt"
        type="button"
        class="attempt-instructions__primary"
        :disabled="resuming || starting"
        @click="emit('resume', activeAttempt.id)"
      >
        {{ resuming ? "Opening attempt…" : "Resume attempt" }}
      </button>
      <button
        v-else-if="instructions.can_start"
        type="button"
        class="attempt-instructions__primary"
        :disabled="starting || resuming"
        @click="emit('start')"
      >
        {{ starting ? "Starting…" : "Start attempt" }}
      </button>
      <p v-else class="attempt-instructions__unavailable" role="status">
        This test cannot be started right now.
      </p>
    </div>
  </section>
</template>

<style scoped>
.attempt-instructions {
  --attempt-ink: #12263a;
  --attempt-muted: #4a5d73;
  --attempt-line: #d7e0ea;
  --attempt-accent: #0f6a5a;
  --attempt-surface: #f7fafc;
  max-width: 40rem;
  margin: 0 auto;
  padding: 2rem 1.25rem 3rem;
  color: var(--attempt-ink);
  font-family: "Source Sans 3", "Segoe UI", sans-serif;
}

.attempt-instructions__eyebrow {
  margin: 0 0 0.5rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 0.75rem;
  color: var(--attempt-muted);
}

.attempt-instructions__title {
  margin: 0;
  font-family: "Literata", "Georgia", serif;
  font-size: clamp(1.75rem, 4vw, 2.35rem);
  line-height: 1.2;
  font-weight: 600;
}

.attempt-instructions__description {
  margin: 0.85rem 0 0;
  color: var(--attempt-muted);
  line-height: 1.55;
}

.attempt-instructions__rules {
  list-style: none;
  margin: 1.75rem 0 0;
  padding: 0;
  border-top: 1px solid var(--attempt-line);
}

.attempt-instructions__rules li {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.85rem 0;
  border-bottom: 1px solid var(--attempt-line);
}

.attempt-instructions__label {
  color: var(--attempt-muted);
}

.attempt-instructions__guidance {
  margin-top: 1.5rem;
  padding: 1rem 1.1rem;
  background: var(--attempt-surface);
  line-height: 1.55;
}

.attempt-instructions__guidance p {
  margin: 0;
}

.attempt-instructions__guidance p + p {
  margin-top: 0.75rem;
}

.attempt-instructions__block,
.attempt-instructions__error,
.attempt-instructions__unavailable {
  color: #8a2f1d;
}

.attempt-instructions__actions {
  margin-top: 1.75rem;
}

.attempt-instructions__primary {
  appearance: none;
  border: 0;
  border-radius: 0.35rem;
  background: var(--attempt-accent);
  color: #fff;
  font: inherit;
  font-weight: 600;
  padding: 0.85rem 1.35rem;
  cursor: pointer;
}

.attempt-instructions__primary:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.attempt-instructions__primary:focus-visible {
  outline: 2px solid #12263a;
  outline-offset: 3px;
}
</style>
