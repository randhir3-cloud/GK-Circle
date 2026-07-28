<script setup>
import {
  Home,
  BookOpen,
  Gamepad2,
  RotateCcw,
  AlertTriangle,
  FileQuestion,
  ShieldAlert,
  ServerCrash,
} from "lucide-vue-next";

const props = defineProps({
  error: {
    type: Object,
    default: null,
  },
});

const route = useRoute();

// Normalize error status & message from Nuxt error prop or route query params
const statusCode = computed(() => {
  if (props.error?.statusCode) return Number(props.error.statusCode);
  if (props.error?.status) return Number(props.error.status);
  if (route.query?.statusCode) return Number(route.query.statusCode);
  return 404;
});

const rawMessage = computed(() => {
  return (
    props.error?.message ||
    props.error?.statusMessage ||
    route.query?.error ||
    route.query?.message ||
    ""
  );
});

const is404 = computed(() => statusCode.value === 404 || !statusCode.value);
const is403 = computed(
  () => statusCode.value === 403 || statusCode.value === 401
);
const is500 = computed(() => statusCode.value >= 500);

const errorConfig = computed(() => {
  if (is404.value) {
    return {
      code: "404",
      badge: "PAGE NOT FOUND",
      title: "Looks like this page took a wrong turn in the syllabus",
      description:
        "The page or resource you requested doesn't exist or has been relocated.",
      icon: FileQuestion,
      accentBg: "bg-jv-yellow",
      badgeColor: "bg-jv-coral/20 text-jv-ink",
    };
  }

  if (is403.value) {
    return {
      code: statusCode.value.toString(),
      badge: "ACCESS RESTRICTED",
      title: "You need permission to enter this study module",
      description:
        rawMessage.value ||
        "Please sign in or request access to view this assessment area.",
      icon: ShieldAlert,
      accentBg: "bg-jv-salmon",
      badgeColor: "bg-jv-salmon/30 text-jv-ink",
    };
  }

  if (is500.value) {
    return {
      code: statusCode.value.toString(),
      badge: "SERVER RECALCULATING",
      title: "Our servers hit an unexpected error",
      description:
        rawMessage.value ||
        "We're already fixing this. Please try again in a few moments.",
      icon: ServerCrash,
      accentBg: "bg-jv-coral",
      badgeColor: "bg-jv-coral/30 text-jv-ink",
    };
  }

  return {
    code: statusCode.value ? statusCode.value.toString() : "ERROR",
    badge: "UNEXPECTED ISSUE",
    title: "Something went wrong while processing your request",
    description:
      rawMessage.value ||
      "Please return to the dashboard or try your action again.",
    icon: AlertTriangle,
    accentBg: "bg-jv-yellow",
    badgeColor: "bg-jv-yellow/40 text-jv-ink",
  };
});

useSeoMeta({
  title: `${errorConfig.value.code} - GK Circle`,
  description: errorConfig.value.description,
  robots: "noindex, nofollow",
});

const handleClearError = (targetPath = "/") => {
  clearError({ redirect: targetPath });
};
</script>

<template>
  <div
    class="flex min-h-screen flex-col items-center justify-center bg-jv-canvas px-4 py-12 text-jv-ink"
  >
    <div class="w-full max-w-2xl text-center">
      <!-- ROBOTIC MASCOT ILLUSTRATION (Hand-drawn Brutalist SVG) -->
      <div class="relative mx-auto mb-6 size-44 sm:size-52">
        <svg
          viewBox="0 0 200 200"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          class="size-full filter drop-shadow-[4px_4px_0px_#1E293B]"
          aria-hidden="true"
        >
          <!-- Antenna -->
          <line
            x1="100"
            y1="40"
            x2="100"
            y2="15"
            stroke="#1E293B"
            stroke-width="5"
            stroke-linecap="round"
          />
          <circle
            cx="100"
            cy="12"
            r="8"
            fill="#FF6B6B"
            stroke="#1E293B"
            stroke-width="4"
          />

          <!-- Robot Head Box -->
          <rect
            x="35"
            y="40"
            width="130"
            height="100"
            rx="16"
            fill="#FFFBEB"
            stroke="#1E293B"
            stroke-width="5"
          />

          <!-- Ears / Bolts -->
          <rect
            x="15"
            y="75"
            width="20"
            height="30"
            rx="6"
            fill="#FFE066"
            stroke="#1E293B"
            stroke-width="4"
          />
          <rect
            x="165"
            y="75"
            width="20"
            height="30"
            rx="6"
            fill="#FFE066"
            stroke="#1E293B"
            stroke-width="4"
          />

          <!-- Visor / Eye Screen -->
          <rect x="50" y="58" width="100" height="42" rx="10" fill="#1E293B" />

          <!-- Confused Eyes (X and O for 404 style) -->
          <g v-if="is404">
            <!-- Left Eye (X) -->
            <path
              d="M66 71L78 83M78 71L66 83"
              stroke="#FFE066"
              stroke-width="4"
              stroke-linecap="round"
            />
            <!-- Right Eye (X) -->
            <path
              d="M122 71L134 83M134 71L122 83"
              stroke="#FFE066"
              stroke-width="4"
              stroke-linecap="round"
            />
          </g>
          <g v-else>
            <!-- Glowing circles for general error -->
            <circle cx="72" cy="79" r="7" fill="#FF6B6B" />
            <circle cx="128" cy="79" r="7" fill="#FF6B6B" />
          </g>

          <!-- Mouth (Zig-Zag / Wavy line for confusion) -->
          <path
            d="M70 118 Q85 110 100 118 T130 118"
            stroke="#1E293B"
            stroke-width="5"
            stroke-linecap="round"
            fill="none"
          />

          <!-- Body Stand / Neck -->
          <rect
            x="80"
            y="140"
            width="40"
            height="25"
            fill="#E2E8F0"
            stroke="#1E293B"
            stroke-width="4"
          />

          <!-- Floating Question Mark / Warning Icon -->
          <g transform="translate(145, 25) rotate(12)">
            <circle
              cx="16"
              cy="16"
              r="16"
              fill="#FF6B6B"
              stroke="#1E293B"
              stroke-width="3"
            />
            <text
              x="16"
              y="22"
              font-family="sans-serif"
              font-weight="900"
              font-size="18"
              fill="#FFFFFF"
              text-anchor="middle"
            >
              ?
            </text>
          </g>
        </svg>
      </div>

      <!-- ERROR CONTENT CARD -->
      <main
        id="main-content"
        class="jv-border-rough relative -rotate-[0.5deg] bg-jv-white p-6 shadow-brutal sm:p-10"
        role="main"
      >
        <span
          class="absolute -top-3 left-1/2 h-4 w-28 -translate-x-1/2 rotate-1 bg-jv-yellow"
          aria-hidden="true"
        ></span>

        <div class="flex items-center justify-center gap-2">
          <span
            :class="[
              'rounded-full border-[2px] border-jv-ink px-3 py-1 font-headings text-xs font-black tracking-widest',
              errorConfig.badgeColor,
            ]"
          >
            {{ errorConfig.badge }} ({{ errorConfig.code }})
          </span>
        </div>

        <h1 class="mt-4 font-headings text-2xl sm:text-4xl">
          {{ errorConfig.title }}
        </h1>

        <p class="mt-3 font-body text-base text-jv-muted sm:text-lg">
          {{ errorConfig.description }}
        </p>

        <!-- ACTION BUTTONS -->
        <div class="mt-8 flex flex-wrap items-center justify-center gap-3">
          <button
            type="button"
            class="inline-flex h-11 items-center gap-2 rounded-[8px] border-[2px] border-jv-ink bg-jv-yellow px-5 font-headings text-sm font-bold text-jv-ink shadow-brutal-sm transition-transform hover:-translate-y-0.5"
            @click="handleClearError('/')"
          >
            <Home class="size-4" />
            Go to Homepage
          </button>

          <button
            type="button"
            class="inline-flex h-11 items-center gap-2 rounded-[8px] border-[2px] border-jv-ink bg-jv-white px-5 font-headings text-sm font-bold text-jv-ink shadow-brutal-sm transition-transform hover:-translate-y-0.5"
            @click="handleClearError('/courses')"
          >
            <BookOpen class="size-4" />
            Courses
          </button>

          <button
            type="button"
            class="inline-flex h-11 items-center gap-2 rounded-[8px] border-[2px] border-jv-ink bg-jv-coral/20 px-5 font-headings text-sm font-bold text-jv-ink shadow-brutal-sm transition-transform hover:-translate-y-0.5"
            @click="handleClearError('/join')"
          >
            <Gamepad2 class="size-4" />
            Play Quiz
          </button>

          <button
            type="button"
            class="inline-flex h-11 items-center gap-2 rounded-[8px] border-[2px] border-jv-ink/40 bg-jv-cream px-4 font-headings text-sm font-bold text-jv-muted shadow-brutal-sm transition-transform hover:-translate-y-0.5"
            @click="handleClearError(route.fullPath)"
          >
            <RotateCcw class="size-4" />
            Try Again
          </button>
        </div>
      </main>
    </div>
  </div>
</template>
