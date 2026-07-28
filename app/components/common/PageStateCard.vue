<script setup>
import { ArrowRight, Sparkles } from "lucide-vue-next";

defineProps({
  eyebrow: { type: String, default: "GK Circle" },
  title: { type: String, required: true },
  description: { type: String, required: true },
  actionLabel: { type: String, default: "" },
  actionTo: { type: String, default: "" },
});
</script>

<template>
  <section
    class="jv-border-uneven relative overflow-hidden bg-jv-white p-6 shadow-brutal-sm sm:p-8"
  >
    <div
      class="absolute -right-10 -top-10 size-36 rounded-full bg-jv-yellow/35"
      aria-hidden="true"
    ></div>
    <div class="relative max-w-2xl">
      <p
        class="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-widest text-jv-coral"
      >
        <Sparkles class="size-4" aria-hidden="true" />
        {{ eyebrow }}
      </p>
      <h2 class="mt-3 font-headings text-2xl sm:text-3xl">{{ title }}</h2>
      <p class="mt-2 font-bold leading-relaxed text-jv-muted">
        {{ description }}
      </p>
      <NuxtLink
        v-if="actionLabel && actionTo"
        :to="actionTo"
        class="mt-5 inline-flex h-11 items-center gap-2 rounded-[8px] border-[2px] border-jv-ink bg-jv-yellow px-4 font-black text-jv-ink no-underline transition-transform hover:-translate-y-0.5"
      >
        {{ actionLabel }}
        <ArrowRight class="size-4" aria-hidden="true" />
      </NuxtLink>
    </div>
  </section>
</template>
