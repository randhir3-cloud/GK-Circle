<script setup>
import { computed } from "vue";
import { isSafeContentUrl } from "@/utils/content_url";

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
    typeof props.data.alt === "string" &&
    (props.data.caption === undefined || typeof props.data.caption === "string")
);
</script>

<template>
  <figure
    v-if="valid"
    class="jv-card overflow-hidden border-2 border-jv-ink bg-jv-white"
    data-testid="image-block"
  >
    <img
      :src="data.url"
      :alt="data.alt"
      class="max-h-[36rem] w-full object-contain"
      loading="lazy"
    />
    <figcaption
      v-if="data.caption !== undefined"
      class="border-t-2 border-jv-ink px-4 py-3 text-sm font-bold text-jv-muted"
    >
      {{ data.caption }}
    </figcaption>
  </figure>
  <p
    v-else
    class="jv-card border-2 border-jv-ink bg-jv-yellow-soft p-4 font-bold"
    data-testid="malformed-block"
  >
    This content block is unavailable.
  </p>
</template>
