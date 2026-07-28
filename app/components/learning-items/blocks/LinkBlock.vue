<script setup>
import { computed } from "vue";
import { isExternalContentUrl, isSafeContentUrl } from "@/utils/content_url";

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
    isSafeContentUrl(props.data.url) &&
    typeof props.data.label === "string" &&
    props.data.label.length > 0
);
const external = computed(
  () => valid.value && isExternalContentUrl(props.data.url)
);
</script>

<template>
  <a
    v-if="valid"
    :href="data.url"
    :target="external ? '_blank' : undefined"
    :rel="external ? 'noopener noreferrer' : undefined"
    class="jv-card flex min-h-12 items-center justify-between gap-3 border-2 border-jv-ink bg-jv-mint px-4 py-3 font-black shadow-brutal-sm transition-transform hover:-translate-y-0.5 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-jv-ink"
    data-testid="link-block"
  >
    <span class="break-words">{{ data.label }}</span>
    <span aria-hidden="true">→</span>
  </a>
  <p
    v-else
    class="jv-card border-2 border-jv-ink bg-jv-yellow-soft p-4 font-bold"
    data-testid="malformed-block"
  >
    This content block is unavailable.
  </p>
</template>
