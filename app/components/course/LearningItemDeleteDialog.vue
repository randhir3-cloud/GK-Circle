<script setup>
import { Modal } from "@/components/ui/modal";

defineProps({
  modelValue: { type: Boolean, default: false },
  item: { type: Object, default: null },
  deleting: { type: Boolean, default: false },
  error: { type: String, default: "" },
});

const emit = defineEmits(["update:modelValue", "confirm"]);
</script>

<template>
  <Modal
    :model-value="modelValue"
    title="Delete Learning Item?"
    :description="
      item
        ? `Delete “${item.title}”? This action uses the server delete policy.`
        : ''
    "
    size="sm"
    :close-on-backdrop="!deleting"
    :hide-close="deleting"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <p v-if="error" class="text-sm font-bold text-red-700" role="alert">
      {{ error }}
    </p>
    <div class="mt-5 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
      <button
        type="button"
        class="h-11 rounded-full border-[2px] border-jv-ink bg-white px-5 font-black"
        :disabled="deleting"
        @click="emit('update:modelValue', false)"
      >
        Cancel
      </button>
      <button
        type="button"
        class="h-11 rounded-full border-[2px] border-jv-ink bg-jv-coral px-5 font-black text-white shadow-brutal-sm disabled:opacity-60"
        :disabled="deleting"
        @click="emit('confirm')"
      >
        {{ deleting ? "Deleting…" : "Delete" }}
      </button>
    </div>
  </Modal>
</template>
