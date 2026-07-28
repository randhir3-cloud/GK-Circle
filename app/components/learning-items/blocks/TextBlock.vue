<script setup>
import { computed } from "vue";

const props = defineProps({
  data: {
    type: [Object, Array, String, Number, Boolean],
    default: undefined,
  },
});

const valid = computed(
  () =>
    props.data !== null &&
    typeof props.data === "object" &&
    !Array.isArray(props.data) &&
    typeof props.data.text === "string"
);
</script>

<template>
  <p
    v-if="valid"
    class="whitespace-pre-line break-words text-base leading-7 sm:text-lg"
    data-testid="text-block"
  >
    {{ data.text }}
  </p>
  <p
    v-else
    class="jv-card border-2 border-jv-ink bg-jv-yellow-soft p-4 font-bold"
    data-testid="malformed-block"
  >
    This content block is unavailable.
  </p>
</template>
