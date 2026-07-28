<script setup>
import { onMounted, ref } from "vue";
import {
  getLearnerLearningItemAPIError,
  useLearnerLearningItemsApi,
} from "@/composables/learner_learning_items";
import { setUserDataStore } from "@/composables/auth";
import { useUsersStore } from "~~/store/users";

definePageMeta({ layout: "empty" });
useSeoMeta({
  title: "Courses - GK Circle",
  description: "Browse published PCS courses.",
});

const api = useLearnerLearningItemsApi();
const usersStore = useUsersStore();
const courses = ref([]);
const loading = ref(true);
const error = ref("");

onMounted(async () => {
  try {
    await setUserDataStore(usersStore);
  } catch {
    /* ignore */
  }
  try {
    courses.value = (await api.listPublishedCourses()) || [];
  } catch (err) {
    error.value = getLearnerLearningItemAPIError(
      err,
      "Unable to load published courses."
    );
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div class="min-h-screen bg-jv-cream px-4 py-6 text-jv-ink sm:px-6">
    <header class="mx-auto max-w-3xl">
      <h1 class="font-headings text-3xl sm:text-4xl">Published courses</h1>
      <p class="mt-2 text-sm font-bold text-jv-muted">
        Open a course outline to enrol and study.
      </p>
    </header>

    <p v-if="loading" class="mx-auto mt-6 max-w-3xl font-bold">Loading…</p>
    <p
      v-else-if="error"
      class="mx-auto mt-6 max-w-3xl font-bold text-red-700"
      role="alert"
    >
      {{ error }}
    </p>
    <ul
      v-else
      data-testid="published-course-list"
      class="mx-auto mt-6 max-w-3xl space-y-3"
    >
      <li
        v-for="course in courses"
        :key="course.id"
        class="jv-border-uneven bg-jv-white p-4 shadow-brutal-sm"
      >
        <NuxtLink
          :to="`/courses/${encodeURIComponent(course.id)}`"
          class="font-headings text-xl underline"
        >
          {{ course.title }}
        </NuxtLink>
        <p v-if="course.short_description" class="mt-1 text-sm font-bold">
          {{ course.short_description }}
        </p>
      </li>
      <li
        v-if="!courses.length"
        class="font-bold text-jv-muted"
        data-testid="published-course-empty"
      >
        No published courses yet.
      </li>
    </ul>
  </div>
</template>
