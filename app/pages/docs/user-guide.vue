<script setup>
import {
  BookOpen,
  BrainCircuit,
  Trophy,
  BarChart3,
  Users,
  Compass,
  ArrowRight,
  HelpCircle,
  CheckCircle2,
  Sparkles,
} from "lucide-vue-next";

useSeoMeta({
  title: "User Guide - GK Circle PCS Prep Platform",
  description:
    "Learn how to use GK Circle for State PCS examination preparation: practice sets, mock exams, structured courses, live multiplayer quizzes, and analytics.",
});

const sections = [
  {
    id: "getting-started",
    title: "1. Getting Started",
    icon: Compass,
    accent: "bg-jv-yellow/40",
    content:
      "GK Circle is a state-PCS examination preparation platform. Whether you are preparing for BPSC, UPPCS, MPPSC, or RAS, the platform adapts to your syllabus and target stage.",
    steps: [
      "Create an account or sign in to save your attempt history and progress.",
      "Explore structured courses organized by subject, topic, and difficulty.",
      "Join live quiz competitions using a 6-digit session code provided by your instructor or host.",
      "Review your performance analytics to identify weak topics and revision priorities.",
    ],
  },
  {
    id: "courses-learning",
    title: "2. Structured Courses & Learning Items",
    icon: BookOpen,
    accent: "bg-jv-mint/40",
    content:
      "Courses group syllabus items into guided learning paths so you can study systematically.",
    steps: [
      "Navigate to the Courses tab in the sidebar.",
      "Select a course to view topic-wise modules and learning materials.",
      "Study revision notes, watch embedded video lessons, and complete attached practice quizzes.",
      "Mark items complete to track your overall course progress percentage.",
    ],
  },
  {
    id: "mcq-practice",
    title: "3. Practice Sets & Mock Examinations",
    icon: BrainCircuit,
    accent: "bg-jv-salmon/40",
    content:
      "Simulate actual exam conditions with timed practice sets and full-length State PCS mock papers.",
    steps: [
      "Choose between Subject-Wise Practice, Topic Tests, or Full-Length Mock Exams.",
      "Answer questions within timed boundaries with instant option selection.",
      "Submit your test to generate a snapshot of your score, accuracy rate, and time per question.",
      "Review detailed explanations and bookmark tricky questions for revision.",
    ],
  },
  {
    id: "live-competitions",
    title: "4. Timed Live Competitions",
    icon: Trophy,
    accent: "bg-jv-coral/25",
    content:
      "Test your speed and accuracy in real-time multiplayer quiz sessions hosted by educators.",
    steps: [
      "Click 'Play a Quiz' or go to /join.",
      "Enter your Display Name and the 6-digit Game Code.",
      "Answer questions simultaneously with other aspirants as the timer counts down.",
      "Watch live leaderboard rankings update after every question.",
    ],
  },
  {
    id: "analytics-reports",
    title: "5. Performance Analytics & Reports",
    icon: BarChart3,
    accent: "bg-jv-lavender/40",
    content:
      "Data-driven insights help you focus your study time on high-yield revision topics.",
    steps: [
      "Visit the Analytics dashboard to see overall accuracy, speed, and daily streak.",
      "Inspect topic-wise heatmaps to identify recurring weak areas.",
      "Download PDF attempt reports for offline review.",
    ],
  },
  {
    id: "community-sharing",
    title: "6. Shared Quizzes & Community",
    icon: Users,
    accent: "bg-jv-yellow-2/40",
    content:
      "Collaborate with peers and study circles by sharing custom-created quizzes.",
    steps: [
      "Access quizzes shared directly with your profile under 'Shared With Me'.",
      "Copy quiz links to invite fellow aspirants to practice together.",
    ],
  },
];
</script>

<template>
  <div class="min-h-screen bg-jv-canvas text-jv-ink">
    <div class="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
      <!-- HEADER BANNER -->
      <header
        class="relative -rotate-[0.5deg] jv-border-rough bg-jv-white p-6 shadow-brutal sm:p-10"
      >
        <span
          class="absolute -top-3 left-1/2 h-4 w-24 -translate-x-1/2 rotate-1 bg-jv-yellow"
          aria-hidden="true"
        ></span>
        <div
          class="inline-flex items-center gap-2 rounded-full border-[2px] border-jv-ink bg-jv-yellow-soft px-3 py-1 text-xs font-black uppercase tracking-widest text-jv-ink"
        >
          <Sparkles class="size-4" aria-hidden="true" />
          Documentation & Help
        </div>
        <h1 class="mt-2 font-headings text-3xl sm:text-5xl">
          GK Circle User Guide
        </h1>
        <p class="mt-3 max-w-3xl font-body text-base text-jv-muted sm:text-lg">
          Everything you need to know about preparing for State PCS
          examinations, taking practice tests, joining live quizzes, and
          tracking your progress.
        </p>

        <div class="mt-6 flex flex-wrap items-center gap-3">
          <NuxtLink
            to="/"
            class="inline-flex items-center gap-2 rounded-[8px] border-[2px] border-jv-ink bg-jv-yellow px-4 py-2 font-headings text-sm font-bold text-jv-ink no-underline shadow-brutal-sm transition-transform hover:-translate-y-0.5"
          >
            Go to Home
            <ArrowRight class="size-4" aria-hidden="true" />
          </NuxtLink>
          <NuxtLink
            to="/courses"
            class="inline-flex items-center gap-2 rounded-[8px] border-[2px] border-jv-ink bg-jv-white px-4 py-2 font-headings text-sm font-bold text-jv-ink no-underline shadow-brutal-sm transition-transform hover:-translate-y-0.5"
          >
            Explore Courses
          </NuxtLink>
          <NuxtLink
            to="/join"
            class="inline-flex items-center gap-2 rounded-[8px] border-[2px] border-jv-ink bg-jv-coral/20 px-4 py-2 font-headings text-sm font-bold text-jv-ink no-underline shadow-brutal-sm transition-transform hover:-translate-y-0.5"
          >
            Join Live Quiz
          </NuxtLink>
        </div>
      </header>

      <!-- SECTIONS -->
      <main class="mt-10 space-y-8">
        <section
          v-for="sec in sections"
          :id="sec.id"
          :key="sec.id"
          class="jv-border-uneven bg-jv-white p-6 shadow-brutal-sm sm:p-8"
        >
          <div class="flex items-center gap-3">
            <div
              :class="[
                'grid size-12 place-items-center rounded-[12px] border-[2px] border-jv-ink shadow-brutal-sm',
                sec.accent,
              ]"
            >
              <component :is="sec.icon" class="size-6" aria-hidden="true" />
            </div>
            <h2 class="font-headings text-2xl sm:text-3xl">{{ sec.title }}</h2>
          </div>

          <p class="mt-4 font-body leading-relaxed text-jv-muted sm:text-base">
            {{ sec.content }}
          </p>

          <div
            class="mt-5 rounded-[10px] border-[2px] border-jv-ink/20 bg-jv-cream p-4 sm:p-5"
          >
            <h3
              class="flex items-center gap-2 font-headings text-lg text-jv-ink"
            >
              <CheckCircle2 class="size-5 text-jv-coral" />
              Key Workflow Steps
            </h3>
            <ul
              class="mt-3 space-y-2 font-body text-sm text-jv-ink sm:text-base"
            >
              <li
                v-for="(step, idx) in sec.steps"
                :key="idx"
                class="flex items-start gap-2.5"
              >
                <span
                  class="mt-0.5 grid size-5 shrink-0 place-items-center rounded-full border border-jv-ink bg-jv-yellow text-xs font-black text-jv-ink"
                >
                  {{ idx + 1 }}
                </span>
                <span>{{ step }}</span>
              </li>
            </ul>
          </div>
        </section>

        <!-- NEED HELP BANNER -->
        <section
          class="jv-border-rough flex flex-col items-start justify-between gap-4 bg-jv-yellow/30 p-6 sm:flex-row sm:items-center sm:p-8"
        >
          <div>
            <div
              class="flex items-center gap-2 font-headings text-xl sm:text-2xl"
            >
              <HelpCircle class="size-6 text-jv-ink" />
              Have Questions or Feedback?
            </div>
            <p class="mt-1 font-body text-sm text-jv-muted sm:text-base">
              Reach out to our academic team or join the learner discussion
              community.
            </p>
          </div>
          <NuxtLink
            to="/join"
            class="inline-flex h-11 items-center gap-2 rounded-[8px] border-[2px] border-jv-ink bg-jv-yellow px-5 font-headings text-base font-bold text-jv-ink no-underline shadow-brutal-sm transition-transform hover:-translate-y-0.5"
          >
            Start Preparing Now
            <ArrowRight class="size-4" />
          </NuxtLink>
        </section>
      </main>
    </div>
  </div>
</template>
