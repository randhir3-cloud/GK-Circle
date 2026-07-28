<script setup>
import { onBeforeUnmount, onMounted, ref } from "vue";

const props = defineProps({
  answeredCount: {
    type: Number,
    required: true,
  },
  unansweredCount: {
    type: Number,
    required: true,
  },
  totalQuestions: {
    type: Number,
    required: true,
  },
  formattedTime: {
    type: String,
    default: "",
  },
  submitting: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["cancel", "confirm"]);

const cancelButtonRef = ref(null);
const confirmButtonRef = ref(null);

const handleKeydown = (e) => {
  if (e.key === "Escape" && !props.submitting) {
    emit("cancel");
  }
};

onMounted(() => {
  if (typeof window !== "undefined") {
    window.addEventListener("keydown", handleKeydown);
  }
  confirmButtonRef.value?.focus();
});

onBeforeUnmount(() => {
  if (typeof window !== "undefined") {
    window.removeEventListener("keydown", handleKeydown);
  }
});
</script>

<template>
  <div
    class="submit-dialog-overlay"
    @click.self="!submitting && emit('cancel')"
  >
    <div
      class="submit-dialog"
      role="dialog"
      aria-modal="true"
      aria-labelledby="submit-dialog-title"
      aria-describedby="submit-dialog-desc"
    >
      <header class="submit-dialog__header">
        <h2 id="submit-dialog-title" class="submit-dialog__title">
          Submit Assessment Attempt?
        </h2>
      </header>

      <div id="submit-dialog-desc" class="submit-dialog__body">
        <p class="submit-dialog__intro">
          Are you sure you want to finish and submit your attempt? Once
          submitted, answers cannot be modified.
        </p>

        <div class="submit-dialog__stats">
          <div class="submit-dialog__stat">
            <span class="submit-dialog__stat-label">Answered:</span>
            <span
              class="submit-dialog__stat-value submit-dialog__stat-value--answered"
            >
              {{ answeredCount }} / {{ totalQuestions }}
            </span>
          </div>

          <div class="submit-dialog__stat">
            <span class="submit-dialog__stat-label">Unanswered:</span>
            <span
              class="submit-dialog__stat-value submit-dialog__stat-value--unanswered"
            >
              {{ unansweredCount }}
            </span>
          </div>

          <div v-if="formattedTime" class="submit-dialog__stat">
            <span class="submit-dialog__stat-label">Time Remaining:</span>
            <span class="submit-dialog__stat-value">
              {{ formattedTime }}
            </span>
          </div>
        </div>

        <p v-if="unansweredCount > 0" class="submit-dialog__warning">
          ⚠️ You have {{ unansweredCount }} unanswered question{{
            unansweredCount > 1 ? "s" : ""
          }}.
        </p>
      </div>

      <footer class="submit-dialog__actions">
        <button
          ref="cancelButtonRef"
          type="button"
          class="submit-dialog__button submit-dialog__button--cancel"
          :disabled="submitting"
          @click="emit('cancel')"
        >
          Continue Attempt
        </button>
        <button
          ref="confirmButtonRef"
          type="button"
          class="submit-dialog__button submit-dialog__button--confirm"
          :disabled="submitting"
          @click="emit('confirm')"
        >
          {{ submitting ? "Submitting…" : "Confirm Submission" }}
        </button>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.submit-dialog-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: rgba(18, 38, 58, 0.6);
  backdrop-filter: blur(2px);
}

.submit-dialog {
  width: 100%;
  max-width: 28rem;
  border-radius: 0.5rem;
  background: #ffffff;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1),
    0 10px 10px -5px rgba(0, 0, 0, 0.04);
  padding: 1.5rem;
  color: #12263a;
  font-family: "Source Sans 3", "Segoe UI", sans-serif;
}

.submit-dialog__title {
  margin: 0;
  font-family: "Literata", "Georgia", serif;
  font-size: 1.4rem;
  color: #12263a;
}

.submit-dialog__body {
  margin-top: 1rem;
}

.submit-dialog__intro {
  margin: 0;
  color: #4a5d73;
  line-height: 1.45;
}

.submit-dialog__stats {
  margin-top: 1rem;
  padding: 0.85rem;
  background: #f4f8fb;
  border-radius: 0.375rem;
  display: grid;
  gap: 0.5rem;
}

.submit-dialog__stat {
  display: flex;
  justify-content: space-between;
  font-size: 0.95rem;
}

.submit-dialog__stat-label {
  color: #4a5d73;
}

.submit-dialog__stat-value {
  font-weight: 700;
}

.submit-dialog__stat-value--answered {
  color: #0f6a5a;
}

.submit-dialog__stat-value--unanswered {
  color: #c0392b;
}

.submit-dialog__warning {
  margin: 0.75rem 0 0;
  padding: 0.5rem 0.75rem;
  border-radius: 0.25rem;
  background: #fff8e1;
  color: #896b00;
  font-size: 0.9rem;
  font-weight: 600;
}

.submit-dialog__actions {
  margin-top: 1.5rem;
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.submit-dialog__button {
  appearance: none;
  border: 1px solid #b8c6d4;
  border-radius: 0.375rem;
  background: #fff;
  color: #12263a;
  font: inherit;
  font-weight: 600;
  padding: 0.65rem 1rem;
  cursor: pointer;
  transition: background 0.15s ease-in-out;
}

.submit-dialog__button--confirm {
  background: #0f6a5a;
  border-color: #0f6a5a;
  color: #ffffff;
}

.submit-dialog__button--confirm:hover:not(:disabled) {
  background: #0b5044;
}

.submit-dialog__button:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.submit-dialog__button:focus-visible {
  outline: 2px solid #12263a;
  outline-offset: 2px;
}

@media (max-width: 360px) {
  .submit-dialog {
    min-width: 320px;
    padding: 1.25rem 1rem;
  }

  .submit-dialog__actions {
    flex-direction: column-reverse;
  }

  .submit-dialog__button {
    width: 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  .submit-dialog-overlay,
  .submit-dialog__button {
    transition: none;
  }
}
</style>
