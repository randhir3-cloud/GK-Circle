<script setup>
import { onBeforeUnmount, ref, watch } from "vue";
import {
  getLearnerLearningItemAPIError,
  isCourseEnrollmentRequiredError,
  useLearnerLearningItemsApi,
} from "@/composables/learner_learning_items";

definePageMeta({ layout: "empty" });
useSeoMeta({
  title: "Learning Items - GK Circle",
  description: "Course Learning Items.",
  robots: "noindex, nofollow",
});

const route = useRoute();
const api = useLearnerLearningItemsApi();

const items = ref([]);
const loading = ref(true);
const loaded = ref(false);
const error = ref("");
const enrollmentRequired = ref(false);
const enrolling = ref(false);
let requestGeneration = 0;

const routeIDs = () => ({
  courseId: String(route.params.course_id || ""),
  nodeId: String(route.params.node_id || ""),
});

const itemPath = (itemId) => {
  const { courseId, nodeId } = routeIDs();
  return `/courses/${encodeURIComponent(courseId)}/nodes/${encodeURIComponent(
    nodeId
  )}/learning-items/${encodeURIComponent(itemId)}`;
};

const loadItems = async () => {
  const generation = ++requestGeneration;
  const { courseId, nodeId } = routeIDs();
  items.value = [];
  loaded.value = false;
  error.value = "";
  enrollmentRequired.value = false;
  loading.value = true;

  try {
    const result = await api.listItems(courseId, nodeId);
    if (generation !== requestGeneration) return;
    items.value = Array.isArray(result) ? result : [];
    loaded.value = true;
  } catch (requestError) {
    if (generation !== requestGeneration) return;
    error.value = getLearnerLearningItemAPIError(
      requestError,
      "Unable to load Learning Items."
    );
    enrollmentRequired.value = isCourseEnrollmentRequiredError(requestError);
  } finally {
    if (generation === requestGeneration) loading.value = false;
  }
};

const enrollAndReload = async () => {
  const { courseId } = routeIDs();
  enrolling.value = true;
  error.value = "";
  try {
    await api.enroll(courseId);
    await loadItems();
  } catch (requestError) {
    error.value = getLearnerLearningItemAPIError(
      requestError,
      "Unable to enroll in this course."
    );
    enrollmentRequired.value = isCourseEnrollmentRequiredError(requestError);
  } finally {
    enrolling.value = false;
  }
};

watch(() => [route.params.course_id, route.params.node_id], loadItems, {
  immediate: true,
});

onBeforeUnmount(() => {
  requestGeneration += 1;
});
</script>

<template>
  <main
    class="min-h-screen w-full flex-1 bg-jv-canvas px-4 py-5 sm:px-6 md:px-8 md:py-7"
  >
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
      <header>
        <p class="font-black uppercase tracking-[0.16em] text-jv-muted">
          Course content
        </p>
        <h1
          class="mt-2 font-headings text-[38px] leading-none text-jv-ink sm:text-[52px]"
        >
          Learning Items
        </h1>
      </header>

      <section
        v-if="loading"
        class="jv-card border-2 border-jv-ink bg-jv-white p-5 font-bold"
        aria-live="polite"
        data-testid="items-loading"
      >
        Loading Learning Items…
      </section>

      <section
        v-else-if="error"
        class="jv-card border-2 border-jv-ink bg-jv-salmon p-5 font-bold"
        role="alert"
        data-testid="items-error"
      >
        <p>{{ error }}</p>
        <button
          v-if="enrollmentRequired"
          type="button"
          class="mt-4 border-2 border-jv-ink bg-jv-white px-4 py-2 font-black"
          data-testid="enroll-course"
          :disabled="enrolling"
          @click="enrollAndReload"
        >
          {{ enrolling ? "Enrolling…" : "Enroll in course" }}
        </button>
      </section>

      <section
        v-else-if="loaded && items.length === 0"
        class="jv-card border-2 border-jv-ink bg-jv-white p-6 text-center"
        data-testid="items-empty"
      >
        <h2 class="text-2xl font-bold">No Learning Items available.</h2>
      </section>

      <ol
        v-else
        class="grid grid-cols-1 gap-4 md:grid-cols-2"
        data-testid="learning-item-list"
      >
        <li
          v-for="item in items"
          :key="item.id"
          :data-testid="`learning-item-${item.id}`"
        >
          <NuxtLink
            :to="itemPath(item.id)"
            class="jv-card flex h-full min-h-32 flex-col justify-between border-2 border-jv-ink bg-jv-white p-5 shadow-brutal transition-transform hover:-translate-y-0.5 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-jv-ink"
          >
            <div>
              <p
                class="text-xs font-black uppercase tracking-[0.14em] text-jv-muted"
              >
                {{ item.item_type }}
              </p>
              <h2 class="mt-2 text-2xl font-bold">{{ item.title }}</h2>
              <p
                v-if="item.description"
                class="mt-2 line-clamp-3 text-sm font-semibold text-jv-muted"
              >
                {{ item.description }}
              </p>
            </div>
            <span class="mt-4 font-black">Open Learning Item →</span>
          </NuxtLink>
        </li>
      </ol>
    </div>
  </main>
</template>
