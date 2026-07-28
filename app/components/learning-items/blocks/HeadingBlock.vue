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
    typeof props.data.text === "string" &&
    [2, 3, 4, 5, 6].includes(props.data.level)
);
const headingTag = computed(() => (valid.value ? `h${props.data.level}` : "p"));
</script>

<template>
  <component
    :is="headingTag"
    v-if="valid"
    class="break-words font-headings text-2xl font-bold leading-tight sm:text-3xl"
    data-testid="heading-block"
  >
    {{ data.text }}
  </component>
  <p
    v-else
    class="jv-card border-2 border-jv-ink bg-jv-yellow-soft p-4 font-bold"
    data-testid="malformed-block"
  >
    This content block is unavailable.
  </p>
</template>
