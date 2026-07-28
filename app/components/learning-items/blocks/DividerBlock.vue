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
    !Array.isArray(props.data)
);
</script>

<template>
  <hr
    v-if="valid"
    class="my-2 border-0 border-t-4 border-jv-ink"
    data-testid="divider-block"
  />
  <p
    v-else
    class="jv-card border-2 border-jv-ink bg-jv-yellow-soft p-4 font-bold"
    data-testid="malformed-block"
  >
    This content block is unavailable.
  </p>
</template>
