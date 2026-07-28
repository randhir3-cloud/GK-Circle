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
    typeof props.data.title === "string" &&
    props.data.title.length > 0
);
const external = computed(
  () => valid.value && isExternalContentUrl(props.data.url)
);
</script>

<template>
  <section
    v-if="valid"
    class="jv-card overflow-hidden border-2 border-jv-ink bg-jv-white"
    data-testid="pdf-block"
  >
    <iframe
      :src="data.url"
      :title="data.title"
      class="h-[28rem] w-full sm:h-[38rem]"
      loading="lazy"
    />
    <div class="border-t-2 border-jv-ink p-4">
      <a
        :href="data.url"
        :target="external ? '_blank' : undefined"
        :rel="external ? 'noopener noreferrer' : undefined"
        class="inline-flex min-h-11 items-center rounded-lg border-2 border-jv-ink bg-jv-yellow px-4 py-2 font-black shadow-brutal-sm"
      >
        Open or download {{ data.title }}
      </a>
    </div>
  </section>
  <p
    v-else
    class="jv-card border-2 border-jv-ink bg-jv-yellow-soft p-4 font-bold"
    data-testid="malformed-block"
  >
    This content block is unavailable.
  </p>
</template>
