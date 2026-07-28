<script setup>
import { Pencil, Trash2 } from "lucide-vue-next";

defineProps({
  item: { type: Object, required: true },
});

const emit = defineEmits(["edit", "delete"]);
</script>

<template>
  <div
    class="grid gap-3 border-t-[2px] border-dashed border-jv-ink/20 px-3 py-4 first:border-t-0 sm:grid-cols-[72px_minmax(0,1fr)_150px_120px_150px] sm:items-center"
    role="row"
    :data-testid="`learning-item-${item.id}`"
  >
    <div class="font-black text-jv-muted">
      <span class="sm:hidden">Position: </span>{{ item.position }}
    </div>
    <div class="min-w-0">
      <p class="truncate font-headings text-lg text-jv-ink">{{ item.title }}</p>
      <p
        v-if="item.description"
        class="truncate text-sm font-bold text-jv-muted"
      >
        {{ item.description }}
      </p>
    </div>
    <div class="text-sm font-black text-jv-ink">
      <span class="sm:hidden">Type: </span>{{ item.item_type }}
    </div>
    <div>
      <span
        class="inline-flex rounded-full border-[2px] border-jv-ink px-3 py-1 text-xs font-black"
        :class="
          item.publish_state === 'PUBLISHED' ? 'bg-emerald-200' : 'bg-jv-yellow'
        "
      >
        {{ item.publish_state }}
      </span>
    </div>
    <div class="flex gap-2 sm:justify-end">
      <button
        type="button"
        class="inline-flex h-9 items-center gap-1 rounded-full border-[2px] border-jv-ink bg-jv-white px-3 text-sm font-black shadow-brutal-sm"
        :aria-label="`Edit ${item.title}`"
        @click="emit('edit', item)"
      >
        <Pencil class="size-4" /> Edit
      </button>
      <button
        type="button"
        class="inline-flex h-9 items-center gap-1 rounded-full border-[2px] border-jv-ink bg-jv-coral px-3 text-sm font-black text-white shadow-brutal-sm"
        :aria-label="`Delete ${item.title}`"
        @click="emit('delete', item)"
      >
        <Trash2 class="size-4" /> Delete
      </button>
    </div>
  </div>
</template>
