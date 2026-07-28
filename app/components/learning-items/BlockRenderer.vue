<script setup>
import { computed } from "vue";
import DividerBlock from "@/components/learning-items/blocks/DividerBlock.vue";
import HeadingBlock from "@/components/learning-items/blocks/HeadingBlock.vue";
import ImageBlock from "@/components/learning-items/blocks/ImageBlock.vue";
import LinkBlock from "@/components/learning-items/blocks/LinkBlock.vue";
import PdfBlock from "@/components/learning-items/blocks/PdfBlock.vue";
import QuoteBlock from "@/components/learning-items/blocks/QuoteBlock.vue";
import TextBlock from "@/components/learning-items/blocks/TextBlock.vue";
import VideoBlock from "@/components/learning-items/blocks/VideoBlock.vue";

const props = defineProps({
  block: {
    type: [Object, Array, String, Number, Boolean],
    default: undefined,
  },
  index: {
    type: Number,
    required: true,
  },
});

const renderers = {
  TEXT: TextBlock,
  HEADING: HeadingBlock,
  IMAGE: ImageBlock,
  VIDEO: VideoBlock,
  PDF: PdfBlock,
  LINK: LinkBlock,
  QUOTE: QuoteBlock,
  CALLOUT: QuoteBlock,
  DIVIDER: DividerBlock,
};

const isBlockObject = computed(
  () =>
    props.block !== null &&
    typeof props.block === "object" &&
    !Array.isArray(props.block)
);
const blockType = computed(() =>
  isBlockObject.value && typeof props.block.type === "string"
    ? props.block.type
    : ""
);
const renderer = computed(() => renderers[blockType.value]);
const hasMalformedEnvelope = computed(
  () => !isBlockObject.value || blockType.value === ""
);
</script>

<template>
  <div
    class="min-w-0"
    :data-block-index="index"
    :data-block-type="blockType || undefined"
  >
    <component
      :is="renderer"
      v-if="renderer"
      :data="block.data"
      :block-type="blockType"
    />

    <p
      v-else-if="hasMalformedEnvelope"
      class="jv-card border-2 border-jv-ink bg-jv-yellow-soft p-4 font-bold"
      data-testid="malformed-block"
    >
      This content block is unavailable.
    </p>

    <p
      v-else
      class="jv-card border-2 border-jv-ink bg-jv-slate p-4 font-bold"
      data-testid="unsupported-block"
    >
      Unsupported content block.
    </p>
  </div>
</template>
