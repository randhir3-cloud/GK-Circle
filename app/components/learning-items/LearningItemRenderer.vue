<script setup>
import { computed } from "vue";
import BlockRenderer from "@/components/learning-items/BlockRenderer.vue";

const props = defineProps({
  metadata: {
    type: [Object, Array, String, Number, Boolean],
    default: undefined,
  },
});

const hasUsableMetadata = computed(
  () =>
    props.metadata !== null &&
    typeof props.metadata === "object" &&
    !Array.isArray(props.metadata) &&
    Array.isArray(props.metadata.blocks)
);
</script>

<template>
  <section
    class="flex min-w-0 flex-col gap-5"
    aria-label="Learning Item content"
    data-testid="learning-item-renderer"
  >
    <p
      v-if="!hasUsableMetadata"
      class="jv-card border-2 border-jv-ink bg-jv-yellow-soft p-4 font-bold"
      data-testid="content-unavailable"
    >
      Content unavailable.
    </p>

    <p
      v-else-if="metadata.blocks.length === 0"
      class="jv-card border-2 border-jv-ink bg-jv-white p-4 font-bold"
      data-testid="content-empty"
    >
      No content available.
    </p>

    <BlockRenderer
      v-for="(block, index) in metadata.blocks"
      v-else
      :key="index"
      :block="block"
      :index="index"
    />
  </section>
</template>
