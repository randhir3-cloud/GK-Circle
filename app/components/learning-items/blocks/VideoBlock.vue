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
    typeof props.data.title === "string" &&
    props.data.title.length > 0 &&
    (props.data.caption === undefined || typeof props.data.caption === "string")
);
</script>

<template>
  <figure
    v-if="valid"
    class="jv-card overflow-hidden border-2 border-jv-ink bg-jv-white"
    data-testid="video-block"
  >
    <div class="aspect-video w-full bg-jv-ink">
      <iframe
        :src="data.url"
        :title="data.title"
        class="h-full w-full"
        loading="lazy"
        allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
        allowfullscreen
        referrerpolicy="strict-origin-when-cross-origin"
      />
    </div>
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
