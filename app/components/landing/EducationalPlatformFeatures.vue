<script setup>
import {
  BarChart3,
  BookOpenText,
  Bot,
  BrainCircuit,
  Radio,
  Users,
} from "lucide-vue-next";

const features = [
  {
    id: "courses",
    title: "Structured Courses",
    description:
      "Follow clear learning paths built around subjects and topics.",
    icon: BookOpenText,
    accent: "bg-jv-yellow/60",
    items: [
      "Topic-wise learning paths",
      "Revision notes",
      "Videos",
      "Resources",
    ],
  },
  {
    id: "practice",
    title: "Smart MCQ Practice",
    description: "Turn every attempt into a focused improvement plan.",
    icon: BrainCircuit,
    accent: "bg-jv-green/35",
    items: [
      "Timed tests",
      "Instant explanations",
      "Confidence tracking",
      "Bookmarks",
      "Weak topic detection",
    ],
  },
  {
    id: "live-quiz",
    title: "Live Quiz & Gamification",
    description: "Compete live, host sessions, and learn through momentum.",
    icon: Radio,
    accent: "bg-jv-coral/20",
    items: [
      "Join with Code",
      "Real-time leaderboard",
      "Socket.IO powered",
      "Host quiz",
      "Gamification",
    ],
  },
  {
    id: "analytics",
    title: "Analytics",
    description: "See what is improving and where your next effort belongs.",
    icon: BarChart3,
    accent: "bg-jv-slate/45",
    items: [
      "Accuracy",
      "Speed",
      "Weak topics",
      "Progress",
      "Daily streak",
      "Heat maps",
    ],
  },
  {
    id: "ai-assistant",
    title: "AI Assistant",
    description: "Use grounded assistance to understand, revise, and practise.",
    icon: Bot,
    accent: "bg-jv-yellow/35",
    items: [
      "RAG powered",
      "Explain answers",
      "Generate quizzes",
      "Summaries",
      "Revision",
    ],
  },
  {
    id: "community",
    title: "Community",
    description: "Learn with peers through shared preparation and resources.",
    icon: Users,
    accent: "bg-jv-coral/15",
    items: [
      "Discussion",
      "Groups",
      "Study circles",
      "Resources",
      "Followers",
      "Posts",
    ],
  },
];
</script>

<template>
  <section id="platform" class="scroll-mt-6">
    <div class="max-w-3xl">
      <p class="text-xs font-black uppercase tracking-wide text-jv-coral">
        One connected learning system
      </p>
      <h2 class="mt-2 font-headings text-3xl sm:text-5xl">
        Everything your preparation needs
      </h2>
      <p class="mt-3 font-bold leading-relaxed text-jv-muted">
        Move naturally from learning to practice, live competition, analytics,
        AI-guided revision, and collaborative study.
      </p>
    </div>

    <div class="mt-7 grid gap-5 md:grid-cols-2 xl:grid-cols-3">
      <article
        v-for="feature in features"
        :id="feature.id"
        :key="feature.title"
        class="jv-border-uneven group scroll-mt-6 bg-jv-white p-5 shadow-brutal-sm transition-transform duration-200 hover:-translate-y-1"
      >
        <div
          class="grid size-12 place-items-center rounded-[12px] border-[2px] border-jv-ink"
          :class="feature.accent"
        >
          <component :is="feature.icon" class="size-6" aria-hidden="true" />
        </div>
        <h3 class="mt-4 font-headings text-2xl">{{ feature.title }}</h3>
        <p class="mt-2 text-sm font-bold leading-relaxed text-jv-muted">
          {{ feature.description }}
        </p>
        <ul class="mt-4 flex flex-wrap gap-2">
          <li
            v-for="item in feature.items"
            :key="item"
            class="rounded-full border border-jv-ink/30 bg-jv-cream px-3 py-1 text-xs font-black"
          >
            {{ item }}
          </li>
        </ul>
      </article>
    </div>
  </section>
</template>
