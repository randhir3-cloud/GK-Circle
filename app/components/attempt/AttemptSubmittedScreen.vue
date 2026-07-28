<script setup>
import { computed } from "vue";
import { ATTEMPT_STATUS_AUTO_SUBMITTED } from "@/utils/attempt_player_constants";

const props = defineProps({
  attemptId: {
    type: String,
    required: true,
  },
  status: {
    type: String,
    default: "SUBMITTED",
  },
  submittedAt: {
    type: String,
    default: "",
  },
  summary: {
    type: Object,
    default: null,
  },
  instructionsPath: {
    type: String,
    required: true,
  },
});

const isAutoSubmitted = computed(
  () => props.status === ATTEMPT_STATUS_AUTO_SUBMITTED
);

const formattedSubmittedAt = computed(() => {
  if (!props.submittedAt) return new Date().toISOString();
  try {
    return (
      new Date(props.submittedAt).toLocaleString("en-US", {
        dateStyle: "medium",
        timeStyle: "medium",
        timeZone: "UTC",
      }) + " (UTC)"
    );
  } catch {
    return props.submittedAt;
  }
});
</script>

<template>
  <div class="submitted-screen">
    <div class="submitted-screen__card">
      <div class="submitted-screen__icon-wrapper">
        <span class="submitted-screen__icon" aria-hidden="true">
          {{ isAutoSubmitted ? "⏱️" : "✅" }}
        </span>
      </div>

      <h1 class="submitted-screen__title">
        {{ isAutoSubmitted ? "Attempt Auto-Submitted" : "Attempt Submitted" }}
      </h1>

      <p class="submitted-screen__subtitle">
        {{
          isAutoSubmitted
            ? "Your attempt time expired and your recorded answers have been automatically saved and submitted."
            : "Your assessment attempt has been successfully submitted."
        }}
      </p>

      <div class="submitted-screen__details">
        <div class="submitted-screen__detail-row">
          <span class="submitted-screen__label">Submission Type:</span>
          <span
            :class="[
              'submitted-screen__badge',
              isAutoSubmitted
                ? 'submitted-screen__badge--auto'
                : 'submitted-screen__badge--manual',
            ]"
          >
            {{
              isAutoSubmitted
                ? "Auto-submitted (time expired)"
                : "Submitted manually"
            }}
          </span>
        </div>

        <div class="submitted-screen__detail-row">
          <span class="submitted-screen__label">Attempt ID:</span>
          <code class="submitted-screen__code">{{ attemptId }}</code>
        </div>

        <div v-if="submittedAt" class="submitted-screen__detail-row">
          <span class="submitted-screen__label">Submitted At:</span>
          <span class="submitted-screen__value">{{
            formattedSubmittedAt
          }}</span>
        </div>

        <div v-if="summary" class="submitted-screen__detail-row">
          <span class="submitted-screen__label">Questions Answered:</span>
          <span class="submitted-screen__value">
            {{ summary.answered_count ?? summary.answeredCount ?? 0 }} of
            {{ summary.total_questions ?? summary.totalQuestions ?? 0 }}
          </span>
        </div>
      </div>

      <div class="submitted-screen__actions">
        <NuxtLink
          :to="`/attempt/quizzes/${encodeURIComponent(
            instructionsPath.split('/quizzes/')[1]?.split('?')[0] || ''
          )}/attempts/${encodeURIComponent(attemptId)}/result`"
          class="submitted-screen__button submitted-screen__button--primary"
        >
          View Assessment Results
        </NuxtLink>
        <NuxtLink
          :to="instructionsPath"
          class="submitted-screen__button submitted-screen__button--secondary"
        >
          Back to instructions
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<style scoped>
.submitted-screen {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 75vh;
  padding: 2rem 1rem;
  font-family: "Source Sans 3", "Segoe UI", sans-serif;
  color: #12263a;
}

.submitted-screen__card {
  width: 100%;
  max-width: 32rem;
  background: #ffffff;
  border-radius: 0.65rem;
  border: 1px solid #d0dbe5;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05);
  padding: 2rem 1.5rem;
  text-align: center;
}

.submitted-screen__icon-wrapper {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 4rem;
  height: 4rem;
  border-radius: 50%;
  background: #eef7f5;
  margin-bottom: 1rem;
}

.submitted-screen__icon {
  font-size: 2rem;
}

.submitted-screen__title {
  margin: 0;
  font-family: "Literata", "Georgia", serif;
  font-size: clamp(1.4rem, 3.5vw, 1.8rem);
  color: #12263a;
}

.submitted-screen__subtitle {
  margin: 0.5rem 0 1.5rem;
  color: #4a5d73;
  line-height: 1.5;
  font-size: 0.95rem;
}

.submitted-screen__details {
  background: #f8fafc;
  border-radius: 0.5rem;
  border: 1px solid #e2e8f0;
  padding: 1rem;
  display: grid;
  gap: 0.75rem;
  text-align: left;
  margin-bottom: 1.75rem;
}

.submitted-screen__detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  font-size: 0.9rem;
}

.submitted-screen__label {
  color: #64748b;
  font-weight: 600;
}

.submitted-screen__value {
  font-weight: 600;
  color: #0f172a;
}

.submitted-screen__code {
  font-family: monospace;
  font-size: 0.825rem;
  background: #e2e8f0;
  padding: 0.15rem 0.4rem;
  border-radius: 0.25rem;
  color: #334155;
  user-select: all;
}

.submitted-screen__badge {
  font-size: 0.8rem;
  font-weight: 700;
  padding: 0.2rem 0.6rem;
  border-radius: 9999px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.submitted-screen__badge--manual {
  background: #dcfce7;
  color: #166534;
}

.submitted-screen__badge--auto {
  background: #fef3c7;
  color: #92400e;
}

.submitted-screen__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.75rem;
}

.submitted-screen__button {
  display: inline-block;
  padding: 0.75rem 1.25rem;
  border-radius: 0.375rem;
  font-weight: 600;
  text-decoration: none;
  transition: all 0.15s ease-in-out;
  font-size: 0.9375rem;
}

.submitted-screen__button--primary {
  background: #0f6a5a;
  color: #ffffff;
}

.submitted-screen__button--primary:hover {
  background: #0b5044;
}

.submitted-screen__button--secondary {
  background: #f1f5f9;
  color: #334155;
  border: 1px solid #cbd5e1;
}

.submitted-screen__button--secondary:hover {
  background: #e2e8f0;
}
</style>
