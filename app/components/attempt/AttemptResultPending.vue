<script setup>
import { computed } from "vue";

const props = defineProps({
  attemptId: {
    type: String,
    required: true,
  },
  submittedAt: {
    type: String,
    default: "",
  },
  message: {
    type: String,
    default: "Results for this assessment have not been released yet.",
  },
  instructionsPath: {
    type: String,
    required: true,
  },
});

const formattedSubmittedAt = computed(() => {
  if (!props.submittedAt) return "";
  try {
    return new Date(props.submittedAt).toLocaleString("en-US", {
      dateStyle: "medium",
      timeStyle: "short",
    });
  } catch {
    return props.submittedAt;
  }
});
</script>

<template>
  <div class="result-pending">
    <div class="result-pending__card">
      <div class="result-pending__icon-wrapper">
        <span class="result-pending__icon" aria-hidden="true">⏳</span>
      </div>

      <h1 class="result-pending__title">Results Pending Release</h1>

      <p class="result-pending__subtitle">
        {{ message }}
      </p>

      <div class="result-pending__meta-box">
        <div class="result-pending__row">
          <span class="result-pending__label">Attempt ID:</span>
          <code class="result-pending__code">{{ attemptId.slice(0, 8) }}…</code>
        </div>
        <div v-if="formattedSubmittedAt" class="result-pending__row">
          <span class="result-pending__label">Submitted On:</span>
          <span class="result-pending__value">{{ formattedSubmittedAt }}</span>
        </div>
      </div>

      <div class="result-pending__actions">
        <NuxtLink :to="instructionsPath" class="result-pending__button">
          Back to Instructions
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<style scoped>
.result-pending {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 3rem 1rem;
  font-family: "Source Sans 3", "Segoe UI", sans-serif;
}

.result-pending__card {
  width: 100%;
  max-width: 32rem;
  background: #ffffff;
  border-radius: 0.75rem;
  border: 1px solid #d0dbe5;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05);
  padding: 2.5rem 2rem;
  text-align: center;
}

.result-pending__icon-wrapper {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 4.5rem;
  height: 4.5rem;
  border-radius: 50%;
  background: #fef3c7;
  margin-bottom: 1.25rem;
}

.result-pending__icon {
  font-size: 2.25rem;
}

.result-pending__title {
  margin: 0;
  font-family: "Outfit", "Literata", serif;
  font-size: 1.625rem;
  font-weight: 700;
  color: #12263a;
}

.result-pending__subtitle {
  margin: 0.75rem 0 1.75rem;
  color: #4b5563;
  line-height: 1.5;
  font-size: 1rem;
}

.result-pending__meta-box {
  background: #f8fafc;
  border-radius: 0.5rem;
  border: 1px solid #e2e8f0;
  padding: 1rem;
  display: grid;
  gap: 0.625rem;
  text-align: left;
  margin-bottom: 2rem;
}

.result-pending__row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.875rem;
}

.result-pending__label {
  color: #64748b;
  font-weight: 600;
}

.result-pending__value {
  font-weight: 600;
  color: #0f172a;
}

.result-pending__code {
  font-family: monospace;
  font-size: 0.825rem;
  background: #e2e8f0;
  padding: 0.15rem 0.4rem;
  border-radius: 0.25rem;
  color: #334155;
}

.result-pending__actions {
  display: flex;
  justify-content: center;
}

.result-pending__button {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  background: #0f6a5a;
  color: #ffffff;
  font-weight: 600;
  text-decoration: none;
  transition: background 0.15s ease-in-out;
}

.result-pending__button:hover {
  background: #0b5044;
}
</style>
