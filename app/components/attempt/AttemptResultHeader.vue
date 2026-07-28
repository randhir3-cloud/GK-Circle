<script setup>
import { computed } from "vue";

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
  instructionsPath: {
    type: String,
    required: true,
  },
});

const statusLabel = computed(() =>
  props.status === "AUTO_SUBMITTED" ? "Auto-Submitted" : "Submitted"
);

const formattedDate = computed(() => {
  if (!props.submittedAt) return "";
  try {
    const d = new Date(props.submittedAt);
    return d.toLocaleString("en-US", {
      dateStyle: "medium",
      timeStyle: "short",
    });
  } catch {
    return props.submittedAt;
  }
});
</script>

<template>
  <header class="result-header">
    <div class="result-header__content">
      <div class="result-header__title-row">
        <h1 class="result-header__title">Assessment Results</h1>
        <span
          class="result-header__badge"
          :class="{
            'result-header__badge--auto': status === 'AUTO_SUBMITTED',
          }"
        >
          {{ statusLabel }}
        </span>
      </div>
      <p class="result-header__meta">
        <span>Attempt ID: {{ attemptId.slice(0, 8) }}…</span>
        <span v-if="formattedDate" class="result-header__dot">•</span>
        <span v-if="formattedDate">Submitted on {{ formattedDate }}</span>
      </p>
    </div>
    <div class="result-header__action">
      <NuxtLink :to="instructionsPath" class="result-header__back-btn">
        Back to Instructions
      </NuxtLink>
    </div>
  </header>
</template>

<style scoped>
.result-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.5rem 2rem;
  background: #ffffff;
  border-bottom: 1px solid #dcdfe4;
  border-radius: 0.75rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.result-header__title-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.result-header__title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: #111827;
  font-family: "Outfit", "Source Sans 3", sans-serif;
}

.result-header__badge {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: #0f6a5a;
  background: #e6f4f1;
  border-radius: 9999px;
}

.result-header__badge--auto {
  color: #92400e;
  background: #fef3c7;
}

.result-header__meta {
  margin: 0.375rem 0 0;
  font-size: 0.875rem;
  color: #6b7280;
}

.result-header__dot {
  margin: 0 0.375rem;
}

.result-header__back-btn {
  display: inline-flex;
  align-items: center;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: #0f6a5a;
  background: #f0fdf9;
  border: 1px solid #0f6a5a;
  border-radius: 0.5rem;
  text-decoration: none;
  transition: all 0.15s ease;
}

.result-header__back-btn:hover {
  background: #0f6a5a;
  color: #ffffff;
}
</style>
