<script setup>
import { computed } from "vue";

const props = defineProps({
  data: {
    type: [Object, Array, String, Number, Boolean],
    default: undefined,
  },
  blockType: {
    type: String,
    default: "QUOTE",
  },
});

const valid = computed(
  () =>
    props.data !== null &&
    typeof props.data === "object" &&
    !Array.isArray(props.data) &&
    typeof props.data.text === "string" &&
    (props.data.attribution === undefined ||
      typeof props.data.attribution === "string")
);
</script>

<template>
  <blockquote
    v-if="valid"
    class="jv-card border-2 border-jv-ink bg-jv-lavender p-5 shadow-brutal-sm sm:p-6"
    :data-testid="blockType === 'CALLOUT' ? 'callout-block' : 'quote-block'"
  >
    <p class="whitespace-pre-line text-lg font-bold leading-7">
      “{{ data.text }}”
    </p>
    <cite
      v-if="data.attribution !== undefined"
      class="mt-3 block text-sm font-black not-italic text-jv-muted"
    >
      — {{ data.attribution }}
    </cite>
  </blockquote>
  <p
    v-else
    class="jv-card border-2 border-jv-ink bg-jv-yellow-soft p-4 font-bold"
    data-testid="malformed-block"
  >
    This content block is unavailable.
  </p>
</template>
