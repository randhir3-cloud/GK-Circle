<script setup>
import { computed, onMounted, ref } from "vue";
import {
  getLearnerLearningItemAPIError,
  useLearnerLearningItemsApi,
} from "@/composables/learner_learning_items";
import { setUserDataStore } from "@/composables/auth";
import { useUsersStore } from "~~/store/users";

definePageMeta({ layout: "empty" });

const route = useRoute();
const courseId = computed(() => String(route.params.course_id || ""));
const api = useLearnerLearningItemsApi();
const usersStore = useUsersStore();

const course = ref(null);
const outline = ref(null);
const enrollment = ref(null);
const loading = ref(true);
const error = ref("");
const enrolling = ref(false);

useSeoMeta({
  title: computed(() =>
    course.value?.title
      ? `${course.value.title} - GK Circle`
      : "Course outline - GK Circle"
  ),
});

const flatten = (roots) => {
  const out = [];
  const walk = (items, depth = 0) => {
    for (const entry of items || []) {
      out.push({ ...entry.node, depth });
      walk(entry.children, depth + 1);
    }
  };
  walk(roots);
  return out;
};

const nodes = computed(() => flatten(outline.value?.roots || []));

const router = useRouter();

const load = async () => {
  loading.value = true;
  error.value = "";
  try {
    course.value = await api.getPublishedCourse(courseId.value);
    outline.value = await api.getPublishedOutline(courseId.value);
    try {
      enrollment.value = await api.getEnrollment(courseId.value);
    } catch {
      enrollment.value = { enrolled: false };
    }
  } catch (err) {
    error.value = getLearnerLearningItemAPIError(
      err,
      "Unable to load course outline."
    );
  } finally {
    loading.value = false;
  }
};

const enroll = async () => {
  if (!usersStore.getUserData()) {
    router.push(
      `/account/login?redirect=/courses/${encodeURIComponent(courseId.value)}`
    );
    return;
  }
  enrolling.value = true;
  error.value = "";
  try {
    enrollment.value = await api.enroll(courseId.value);
  } catch (err) {
    error.value = getLearnerLearningItemAPIError(
      err,
      "Unable to enrol in this course."
    );
  } finally {
    enrolling.value = false;
  }
};

onMounted(async () => {
  try {
    await setUserDataStore(usersStore);
  } catch {
    /* ignore */
  }
  await load();
});
</script>

<template>
  <div class="min-h-screen bg-jv-cream px-4 py-6 text-jv-ink sm:px-6">
    <div class="mx-auto max-w-3xl">
      <NuxtLink to="/courses" class="text-sm font-black underline"
        >All courses</NuxtLink
      >
      <p v-if="loading" class="mt-4 font-bold">Loading outline…</p>
      <p v-else-if="error" class="mt-4 font-bold text-red-700" role="alert">
        {{ error }}
      </p>
      <template v-else-if="course">
        <h1 class="mt-3 font-headings text-3xl">{{ course.title }}</h1>
        <p v-if="course.short_description" class="mt-2 font-bold text-jv-muted">
          {{ course.short_description }}
        </p>

        <div class="mt-4 flex flex-wrap gap-2">
          <button
            v-if="!enrollment?.enrolled"
            type="button"
            data-testid="enroll-course-button"
            class="h-11 rounded-[8px] border-[2px] border-jv-ink bg-jv-yellow px-4 font-black disabled:opacity-60"
            :disabled="enrolling"
            @click="enroll"
          >
            Enrol
          </button>
          <p
            v-else
            class="font-black text-jv-green"
            data-testid="enrolled-badge"
          >
            Enrolled
          </p>
        </div>

        <h2 class="mt-8 font-headings text-xl">Outline</h2>
        <ul data-testid="learner-outline" class="mt-3 space-y-2">
          <li
            v-for="node in nodes"
            :key="node.id"
            class="rounded-[8px] border-[2px] border-jv-ink/20 bg-jv-white px-3 py-2 font-bold"
            :style="{ marginLeft: `${node.depth * 16}px` }"
          >
            <span class="text-xs font-black uppercase text-jv-muted">{{
              node.node_type
            }}</span>
            —
            <NuxtLink
              v-if="enrollment?.enrolled"
              :to="`/courses/${encodeURIComponent(
                courseId
              )}/nodes/${encodeURIComponent(node.id)}/learning-items`"
              class="underline"
            >
              {{ node.title }}
            </NuxtLink>
            <span v-else>{{ node.title }}</span>
          </li>
          <li v-if="!nodes.length" class="font-bold text-jv-muted">
            This course has no subjects or topics yet.
          </li>
        </ul>
        <p
          v-if="!enrollment?.enrolled && nodes.length"
          class="mt-3 text-sm font-bold text-jv-muted"
        >
          Enrol to open learning items under each topic.
        </p>
      </template>
    </div>
  </div>
</template>
