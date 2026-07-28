<script setup>
import LearningItemRow from "./LearningItemRow.vue";

defineProps({
  items: { type: Array, default: () => [] },
});

const emit = defineEmits(["edit", "delete"]);
</script>

<template>
  <div
    class="overflow-hidden rounded-[10px] border-[3px] border-jv-ink bg-jv-white"
    role="table"
    aria-label="Learning Items"
    data-testid="learning-item-table"
  >
    <div
      class="hidden grid-cols-[72px_minmax(0,1fr)_150px_120px_150px] bg-jv-ink px-3 py-3 text-xs font-black uppercase tracking-wide text-white sm:grid"
      role="row"
    >
      <span>Position</span><span>Title</span><span>Type</span>
      <span>Publish state</span><span class="text-right">Actions</span>
    </div>
    <LearningItemRow
      v-for="item in items"
      :key="item.id"
      :item="item"
      @edit="emit('edit', $event)"
      @delete="emit('delete', $event)"
    />
  </div>
</template>
