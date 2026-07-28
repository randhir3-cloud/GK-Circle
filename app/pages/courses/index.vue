<script setup>
import { onMounted, ref } from "vue";
import { useLearnerLearningItemsApi } from "@/composables/learner_learning_items";
import { setUserDataStore } from "@/composables/auth";
import { getSafeAPIErrorMessage } from "@/utils/api_error";
import { useUsersStore } from "~~/store/users";
import PageStateCard from "@/components/common/PageStateCard.vue";

definePageMeta({ layout: "empty" });
useSeoMeta({
  title: "Courses - GK Circle",
  description:
    "Discover structured courses created by educators and the GK Circle community.",
});

const api = useLearnerLearningItemsApi();
const usersStore = useUsersStore();
const courses = ref([]);
const loading = ref(true);
const error = ref("");
const authenticated = ref(false);

onMounted(async () => {
  try {
    const user = await setUserDataStore(usersStore);
    if (user) authenticated.value = true;
  } catch {
    authenticated.value = false;
  }

  try {
    courses.value = (await api.listPublishedCourses()) || [];
  } catch (requestError) {
    error.value = getSafeAPIErrorMessage(
      requestError,
      "Courses could not be loaded. Please try again."
    );
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div class="min-h-screen bg-jv-cream px-4 py-6 text-jv-ink sm:px-6 lg:px-8">
    <header class="mx-auto w-full max-w-7xl">
      <p
        class="inline-flex rounded-full border-[2px] border-jv-ink bg-jv-yellow-soft px-3 py-1 text-xs font-black uppercase tracking-widest text-jv-ink"
      >
        Structured learning
      </p>
      <h1 class="mt-1 font-headings text-3xl sm:text-5xl">Courses</h1>
      <p class="mt-2 max-w-3xl text-sm font-bold text-jv-muted sm:text-base">
        Discover structured courses created by educators and the GK Circle
        community.
      </p>
    </header>

    <div
      v-if="loading"
      class="mx-auto mt-6 grid w-full max-w-7xl gap-4 md:grid-cols-2 xl:grid-cols-3"
      aria-label="Loading Courses"
    >
      <div
        v-for="index in 6"
        :key="index"
        class="h-44 animate-pulse rounded-[12px] border-[2px] border-jv-ink/15 bg-jv-white"
      ></div>
    </div>

    <div v-else-if="error" class="mx-auto mt-6 w-full max-w-7xl" role="alert">
      <PageStateCard
        eyebrow="Courses unavailable"
        title="We could not load Courses right now."
        :description="error"
      />
    </div>

    <div
      v-else-if="!courses.length"
      class="mx-auto mt-6 w-full max-w-7xl"
      data-testid="published-course-empty"
    >
      <PageStateCard
        eyebrow="Course catalogue"
        title="No published Courses yet."
        description="Educators are preparing structured learning paths. Please check back soon."
      />
    </div>

    <ul
      v-else
      data-testid="published-course-list"
      class="mx-auto mt-6 grid w-full max-w-7xl gap-4 md:grid-cols-2 xl:grid-cols-3"
    >
      <li
        v-for="course in courses"
        :key="course.id"
        class="jv-border-uneven group min-h-44 bg-jv-white p-5 shadow-brutal-sm transition-transform hover:-translate-y-1"
      >
        <p
          class="inline-flex rounded-full border border-jv-ink bg-jv-yellow-soft px-2.5 py-1 text-xs font-black uppercase tracking-wide text-jv-ink"
        >
          Structured Course
        </p>
        <NuxtLink
          :to="`/courses/${encodeURIComponent(course.id)}`"
          class="mt-3 block font-headings text-xl text-jv-ink underline decoration-jv-yellow decoration-4 underline-offset-4"
        >
          {{ course.title }}
        </NuxtLink>
        <p
          v-if="course.short_description"
          class="mt-3 text-sm font-bold leading-relaxed text-jv-muted"
        >
          {{ course.short_description }}
        </p>
      </li>
    </ul>
  </div>
</template>
