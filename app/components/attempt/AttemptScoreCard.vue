<script setup>
import { computed } from "vue";
import { formatDurationSeconds } from "@/composables/assessment_attempts";

const props = defineProps({
  summary: {
    type: Object,
    required: true,
  },
  canShowScore: {
    type: Boolean,
    default: true,
  },
  canShowPassFail: {
    type: Boolean,
    default: true,
  },
});

const formattedScore = computed(() => {
  const s = props.summary?.total_score ?? 0;
  return Number.isInteger(s) ? s.toString() : s.toFixed(2);
});

const formattedMaxScore = computed(() => {
  const m = props.summary?.max_score ?? 0;
  return Number.isInteger(m) ? m.toString() : m.toFixed(2);
});

const percentage = computed(() => {
  const p = props.summary?.percentage ?? 0;
  return Math.max(0, Math.floor(p));
});

const formattedDuration = computed(() =>
  formatDurationSeconds(props.summary?.duration_seconds ?? 0)
);

const isPassed = computed(() => props.summary?.passed);
</script>

<template>
  <div class="score-card">
    <div v-if="canShowScore" class="score-card__primary">
      <span class="score-card__label">Total Score</span>
      <div class="score-card__score-val">
        <span class="score-card__score">{{ formattedScore }}</span>
        <span class="score-card__max">/ {{ formattedMaxScore }}</span>
      </div>
    </div>

    <div v-if="canShowScore" class="score-card__divider"></div>

    <div v-if="canShowScore" class="score-card__stat">
      <span class="score-card__label">Percentage</span>
      <div class="score-card__badge-val">
        <span
          class="score-card__percentage"
          :class="{
            'score-card__percentage--high': percentage >= 60,
            'score-card__percentage--mid': percentage >= 40 && percentage < 60,
            'score-card__percentage--low': percentage < 40,
          }"
        >
          {{ percentage }}%
        </span>
        <span
          v-if="canShowPassFail && isPassed !== undefined && isPassed !== null"
          class="score-card__pass-badge"
          :class="{
            'score-card__pass-badge--pass': isPassed,
            'score-card__pass-badge--fail': !isPassed,
          }"
        >
          {{ isPassed ? "PASSED" : "FAILED" }}
        </span>
      </div>
    </div>

    <div v-if="canShowScore" class="score-card__divider"></div>

    <div class="score-card__stat">
      <span class="score-card__label">Time Taken</span>
      <span class="score-card__duration">{{ formattedDuration }}</span>
    </div>
  </div>
</template>

<style scoped>
.score-card {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-around;
  gap: 1.5rem;
  padding: 1.75rem 2rem;
  background: linear-gradient(135deg, #0f6a5a 0%, #134e48 100%);
  color: #ffffff;
  border-radius: 0.75rem;
  box-shadow: 0 4px 14px rgba(15, 106, 90, 0.2);
}

.score-card__label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: rgba(255, 255, 255, 0.8);
  margin-bottom: 0.375rem;
}

.score-card__score-val {
  display: flex;
  align-items: baseline;
  gap: 0.375rem;
}

.score-card__score {
  font-size: 2.25rem;
  font-weight: 800;
  font-family: "Outfit", sans-serif;
  line-height: 1;
}

.score-card__max {
  font-size: 1.125rem;
  color: rgba(255, 255, 255, 0.7);
  font-weight: 500;
}

.score-card__divider {
  width: 1px;
  height: 2.5rem;
  background: rgba(255, 255, 255, 0.2);
}

.score-card__badge-val {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.score-card__percentage {
  font-size: 1.875rem;
  font-weight: 800;
  font-family: "Outfit", sans-serif;
}

.score-card__percentage--high {
  color: #a7f3d0;
}

.score-card__percentage--mid {
  color: #fde68a;
}

.score-card__percentage--low {
  color: #fca5a5;
}

.score-card__pass-badge {
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 700;
  border-radius: 0.375rem;
}

.score-card__pass-badge--pass {
  background: #065f46;
  color: #a7f3d0;
}

.score-card__pass-badge--fail {
  background: #991b1b;
  color: #fecaca;
}

.score-card__duration {
  font-size: 1.5rem;
  font-weight: 700;
  font-family: "Outfit", sans-serif;
}

@media (max-width: 640px) {
  .score-card__divider {
    display: none;
  }
}
</style>
