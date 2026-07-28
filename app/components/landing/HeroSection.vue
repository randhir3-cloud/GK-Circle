<script setup>
import {
  ArrowRight,
  Bot,
  BookOpenText,
  BrainCircuit,
  ChartNoAxesCombined,
  Gamepad2,
  Sparkles,
  Trophy,
  Users,
} from "lucide-vue-next";

const primaryActions = [
  { label: "Explore Courses", to: "/courses", icon: BookOpenText },
  { label: "Start Practice", to: "/#practice", icon: BrainCircuit },
  { label: "Join Live Quiz", to: "/join", icon: Gamepad2 },
  {
    label: "Create Test",
    to: "/admin/quiz/list-quiz?create=1",
    icon: Trophy,
  },
];

const secondaryActions = [
  { label: "Community", to: "/#community", icon: Users },
  { label: "AI Assistant", to: "/#ai-assistant", icon: Bot },
];

const platformSignals = [
  { label: "Structured learning", icon: BookOpenText },
  { label: "Smart practice", icon: BrainCircuit },
  { label: "Live competition", icon: Trophy },
  { label: "Actionable analytics", icon: ChartNoAxesCombined },
];
</script>

<template>
  <section
    class="grid items-center gap-8 overflow-hidden rounded-[20px] border-[3px] border-jv-ink bg-jv-white p-5 shadow-brutal-lg sm:p-8 lg:grid-cols-[minmax(0,1.25fr)_minmax(320px,0.75fr)] lg:p-10"
  >
    <div>
      <p
        class="inline-flex items-center gap-2 rounded-full border-[2px] border-jv-ink bg-jv-yellow/45 px-3 py-1.5 text-xs font-black uppercase tracking-wide"
      >
        <Sparkles class="size-4 text-jv-coral" aria-hidden="true" />
        GK Circle Educational Operating System
      </p>

      <h1
        class="mt-5 font-headings text-[42px] leading-[0.98] tracking-tight text-jv-ink sm:text-[64px] lg:text-[76px]"
      >
        Learn.<br />
        Practice.<br />
        Compete.
        <span class="text-jv-coral">Grow.</span>
      </h1>

      <p
        class="mt-6 max-w-3xl text-base font-bold leading-relaxed text-jv-muted sm:text-lg"
      >
        GK Circle is an Educational Operating System that combines structured
        learning, AI assistance, MCQ practice, live tests, analytics and
        collaborative learning into one intelligent platform.
      </p>

      <div class="mt-7 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <NuxtLink
          v-for="action in primaryActions"
          :key="action.label"
          :to="action.to"
          class="inline-flex min-h-12 items-center justify-center gap-2 rounded-[10px] border-[2px] border-jv-ink bg-jv-yellow px-4 py-2 text-center text-sm font-black text-jv-ink no-underline shadow-brutal-sm transition-transform hover:-translate-y-1"
        >
          <component :is="action.icon" class="size-4" aria-hidden="true" />
          {{ action.label }}
        </NuxtLink>
      </div>

      <div class="mt-3 flex flex-wrap gap-3">
        <NuxtLink
          v-for="action in secondaryActions"
          :key="action.label"
          :to="action.to"
          class="inline-flex min-h-11 items-center gap-2 rounded-[10px] border-[2px] border-jv-ink bg-white px-4 py-2 text-sm font-black text-jv-ink no-underline transition-colors hover:bg-jv-coral hover:text-white"
        >
          <component :is="action.icon" class="size-4" aria-hidden="true" />
          {{ action.label }}
          <ArrowRight class="size-4" aria-hidden="true" />
        </NuxtLink>
      </div>
    </div>

    <div
      class="relative mx-auto grid w-full max-w-lg gap-3 rounded-[18px] border-[3px] border-jv-ink bg-jv-cream p-4 shadow-brutal sm:p-5"
      aria-label="GK Circle platform capabilities"
    >
      <div
        class="absolute -right-4 -top-4 grid size-14 rotate-6 place-items-center rounded-[12px] border-[3px] border-jv-ink bg-jv-coral text-white shadow-brutal-sm"
        aria-hidden="true"
      >
        <BrainCircuit class="size-7" />
      </div>
      <article
        v-for="(signal, index) in platformSignals"
        :key="signal.label"
        class="flex items-center gap-4 rounded-[12px] border-[2px] border-jv-ink bg-white p-4 transition-transform hover:translate-x-1"
        :class="index % 2 ? 'rotate-[0.6deg]' : 'rotate-[-0.6deg]'"
      >
        <span
          class="grid size-11 shrink-0 place-items-center rounded-[10px] border-[2px] border-jv-ink bg-jv-yellow/60"
        >
          <component :is="signal.icon" class="size-5" aria-hidden="true" />
        </span>
        <span class="font-headings text-lg">{{ signal.label }}</span>
      </article>
    </div>
  </section>
</template>
