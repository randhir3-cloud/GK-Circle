<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4"
  >
    <div
      class="bg-slate-900 border border-slate-800 rounded-xl p-6 max-w-md w-full shadow-2xl"
    >
      <h3 class="text-xl font-bold text-slate-100 mb-2">
        Release Quiz Results
      </h3>

      <p class="text-sm text-slate-300 mb-4">
        You are about to release results for this assessment. Learners will
        immediately gain access to view their assessment results according to
        configured review permissions.
      </p>

      <div
        class="bg-amber-500/10 border border-amber-500/20 rounded-lg p-3 text-xs text-amber-300 mb-6"
      >
        <strong class="font-semibold">Irreversible Action:</strong> Once
        released, results cannot be withheld for attempts already submitted.
        Total submitted attempts affected:
        <strong>{{ totalSubmittedAttempts }}</strong
        >.
      </div>

      <div class="flex items-center justify-end space-x-3">
        <button
          type="button"
          class="px-4 py-2 text-sm font-medium text-slate-300 bg-slate-800 hover:bg-slate-700 rounded-lg transition-colors"
          :disabled="isSubmitting"
          @click="$emit('close')"
        >
          Cancel
        </button>
        <button
          type="button"
          class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-lg transition-colors disabled:opacity-50"
          :disabled="isSubmitting"
          @click="$emit('confirm')"
        >
          {{ isSubmitting ? "Releasing..." : "Confirm Release" }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  isOpen: {
    type: Boolean,
    default: false,
  },
  totalSubmittedAttempts: {
    type: Number,
    default: 0,
  },
  isSubmitting: {
    type: Boolean,
    default: false,
  },
});

defineEmits(["close", "confirm"]);
</script>
