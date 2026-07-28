<script setup>
import { onBeforeUnmount, ref, watch } from "vue";
import LearningItemRenderer from "@/components/learning-items/LearningItemRenderer.vue";
import {
  getLearnerLearningItemAPIError,
  isCourseEnrollmentRequiredError,
  useLearnerLearningItemsApi,
} from "@/composables/learner_learning_items";

definePageMeta({ layout: "empty" });
useSeoMeta({
  title: "Learning Item - GK Circle",
  description: "Course Learning Item content.",
  robots: "noindex, nofollow",
});

const route = useRoute();
const api = useLearnerLearningItemsApi();

const item = ref(null);
const previous = ref(null);
const next = ref(null);
const loading = ref(true);
const error = ref("");
const enrollmentRequired = ref(false);
const enrolling = ref(false);
let requestGeneration = 0;

const routeIDs = () => ({
  courseId: String(route.params.course_id || ""),
  nodeId: String(route.params.node_id || ""),
  itemId: String(route.params.item_id || ""),
});

const listPath = () => {
  const { courseId, nodeId } = routeIDs();
  return `/courses/${encodeURIComponent(courseId)}/nodes/${encodeURIComponent(
    nodeId
  )}/learning-items`;
};

const itemPath = (itemId) => `${listPath()}/${encodeURIComponent(itemId)}`;

const loadItem = async () => {
  const generation = ++requestGeneration;
  const { courseId, nodeId, itemId } = routeIDs();
  item.value = null;
  previous.value = null;
  next.value = null;
  error.value = "";
  enrollmentRequired.value = false;
  loading.value = true;

  try {
    const result = await api.getItem(courseId, nodeId, itemId);
    if (generation !== requestGeneration) return;
    item.value = result?.learning_item || null;
    previous.value = result?.previous || null;
    next.value = result?.next || null;
    if (!item.value) error.value = "Content unavailable.";
  } catch (requestError) {
    if (generation !== requestGeneration) return;
    error.value = getLearnerLearningItemAPIError(
      requestError,
      "Unable to load the Learning Item."
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
    await loadItem();
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

watch(
  () => [route.params.course_id, route.params.node_id, route.params.item_id],
  loadItem,
  { immediate: true }
);

onBeforeUnmount(() => {
  requestGeneration += 1;
});
</script>

<template>
  <main
    class="min-h-screen w-full flex-1 bg-jv-canvas px-4 py-5 sm:px-6 md:px-8 md:py-7"
  >
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
      <NuxtLink
        :to="listPath()"
        class="w-fit font-black underline decoration-2 underline-offset-4"
      >
        ← All Learning Items
      </NuxtLink>

      <section
        v-if="loading"
        class="jv-card border-2 border-jv-ink bg-jv-white p-5 font-bold"
        aria-live="polite"
        data-testid="item-loading"
      >
        Loading Learning Item…
      </section>

      <section
        v-else-if="error"
        class="jv-card border-2 border-jv-ink bg-jv-salmon p-5 font-bold"
        role="alert"
        data-testid="item-error"
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

      <template v-else-if="item">
        <header
          class="jv-card border-2 border-jv-ink bg-jv-white p-5 shadow-brutal sm:p-7"
        >
          <p
            class="text-xs font-black uppercase tracking-[0.14em] text-jv-muted"
          >
            {{ item.item_type }}
          </p>
          <h1
            class="mt-2 break-words font-headings text-[36px] leading-tight sm:text-[50px]"
          >
            {{ item.title }}
          </h1>
          <p
            v-if="item.description"
            class="mt-3 whitespace-pre-line text-base font-semibold leading-7 text-jv-muted"
          >
            {{ item.description }}
          </p>
        </header>

        <LearningItemRenderer :metadata="item.metadata" />

        <nav
          v-if="previous || next"
          class="grid grid-cols-1 gap-4 border-t-2 border-jv-ink pt-6 sm:grid-cols-2"
          aria-label="Learning Item navigation"
        >
          <NuxtLink
            v-if="previous"
            :to="itemPath(previous.id)"
            class="jv-card border-2 border-jv-ink bg-jv-white p-4 font-black shadow-brutal-sm"
            data-testid="previous-item"
          >
            <span
              class="block text-xs uppercase tracking-[0.12em] text-jv-muted"
            >
              Previous
            </span>
            ← {{ previous.title }}
          </NuxtLink>
          <NuxtLink
            v-if="next"
            :to="itemPath(next.id)"
            class="jv-card border-2 border-jv-ink bg-jv-yellow p-4 text-right font-black shadow-brutal-sm sm:col-start-2"
            data-testid="next-item"
          >
            <span
              class="block text-xs uppercase tracking-[0.12em] text-jv-muted"
            >
              Next
            </span>
            {{ next.title }} →
          </NuxtLink>
        </nav>
      </template>
    </div>
  </main>
</template>
