<template>
  <main
    class="mx-auto flex w-full max-w-7xl flex-col gap-10 px-4 py-5 sm:gap-14 sm:px-6 md:px-8 md:py-8"
  >
    <HeroSection />
    <EducationalPlatformFeatures />
    <JoinLiveQuiz />
    <PublicQuizSection />
  </main>
</template>

<script setup>
import HeroSection from "@/components/landing/HeroSection.vue";
import EducationalPlatformFeatures from "@/components/landing/EducationalPlatformFeatures.vue";
import JoinLiveQuiz from "@/components/Quiz/JoinLiveQuiz.vue";
import PublicQuizSection from "@/components/landing/PublicQuizSection.vue";

definePageMeta({
  layout: "empty",
});

const route = useRoute();
const { baseUrl } = useRuntimeConfig().public;
const canonicalUrl = new URL(route.path, baseUrl).href;
const shareImageUrl = new URL("/og-image.png", baseUrl).href;

useHead({
  link: [{ rel: "canonical", href: canonicalUrl }],
});

useSeoMeta({
  title: "GK Circle - Educational Operating System",
  description:
    "Learn, practise, compete and grow with structured Courses, smart MCQs, live tests, analytics, AI assistance and community learning.",
  ogTitle: "GK Circle - Educational Operating System",
  ogDescription:
    "Structured learning, smart practice, live competition, analytics, AI assistance and community in one platform.",
  ogType: "website",
  ogUrl: canonicalUrl,
  ogImage: shareImageUrl,
});

onMounted(() => {
  if (route.hash) {
    document.querySelector(route.hash)?.scrollIntoView({ behavior: "smooth" });
  }
});
</script>
