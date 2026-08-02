<script setup>
import { computed, onMounted, ref } from "vue";
import { Plus, Search } from "lucide-vue-next";
import { usePush } from "notivue";
import {
  getCourseAdminAPIError,
  useCourseLearningItemsApi,
} from "@/composables/course_learning_items";
import { setUserDataStore } from "@/composables/auth";
import { useUsersStore } from "~~/store/users";

definePageMeta({
  layout: "empty",
  middleware: ["authorization"],
  requiredRoles: ["super_admin", "admin"],
});
useSeoMeta({
  title: "Courses - GK Circle",
  description: "Search and manage created GK Circle courses.",
  robots: "noindex, nofollow",
});

const api = useCourseLearningItemsApi();
const toast = usePush();
const usersStore = useUsersStore();

const courses = ref([]);
const loading = ref(true);
const errorMessage = ref("");
const searchQuery = ref("");
const statusFilter = ref("ALL");
const editingCourseId = ref("");
const editingCourseTitle = ref("");
const savingCourseId = ref("");
const updatingStatusId = ref("");

const filteredCourses = computed(() => {
  const query = searchQuery.value.trim().toLocaleLowerCase();
  return courses.value.filter((course) => {
    const matchesQuery =
      !query || course.title.toLocaleLowerCase().includes(query);
    const matchesStatus =
      statusFilter.value === "ALL" || course.status === statusFilter.value;
    return matchesQuery && matchesStatus;
  });
});

const resultSummary = computed(() => {
  const total = courses.value.length;
  const visible = filteredCourses.value.length;
  return visible === total
    ? `${total} ${total === 1 ? "Course" : "Courses"}`
    : `${visible} of ${total} Courses`;
});

const loadCourses = async () => {
  loading.value = true;
  errorMessage.value = "";
  try {
    courses.value = (await api.listCourses()) || [];
  } catch (error) {
    errorMessage.value = getCourseAdminAPIError(
      error,
      "Unable to load Courses."
    );
  } finally {
    loading.value = false;
  }
};

const startCourseEdit = (course) => {
  editingCourseId.value = course.id;
  editingCourseTitle.value = course.title;
  errorMessage.value = "";
};

const cancelCourseEdit = () => {
  editingCourseId.value = "";
  editingCourseTitle.value = "";
};

const saveCourseTitle = async (courseId) => {
  const title = editingCourseTitle.value.trim();
  if (!title) {
    errorMessage.value = "Course title is required.";
    return;
  }
  savingCourseId.value = courseId;
  errorMessage.value = "";
  try {
    await api.updateCourse(courseId, { title });
    await loadCourses();
    cancelCourseEdit();
    toast.success("Course title updated.");
  } catch (error) {
    errorMessage.value = getCourseAdminAPIError(
      error,
      "Unable to update the Course."
    );
  } finally {
    savingCourseId.value = "";
  }
};

const changeCourseStatus = async (course, status) => {
  if (status === course.status) return;
  updatingStatusId.value = course.id;
  errorMessage.value = "";
  try {
    await api.updateCourse(course.id, { status });
    await loadCourses();
    toast.success(`Course status set to ${status}.`);
  } catch (error) {
    errorMessage.value = getCourseAdminAPIError(
      error,
      "Unable to update course status."
    );
  } finally {
    updatingStatusId.value = "";
  }
};

onMounted(async () => {
  try {
    await setUserDataStore(usersStore);
  } catch {
    /* auth store optional for page shell */
  }
  await loadCourses();
});
</script>

<template>
  <div class="min-h-screen bg-jv-cream px-4 py-6 text-jv-ink sm:px-6">
    <header
      class="mx-auto flex max-w-6xl flex-wrap items-end justify-between gap-4"
    >
      <div>
        <p
          class="inline-flex rounded-full border-[2px] border-jv-ink bg-jv-yellow-soft px-3 py-1 text-xs font-black uppercase tracking-widest text-jv-ink"
        >
          Exam Platform · Course administration
        </p>
        <h1 class="font-headings text-3xl sm:text-4xl">Courses</h1>
        <p class="mt-2 max-w-2xl text-sm font-bold text-jv-muted">
          Search every created Course, then edit its structure, content, title,
          or publication state.
        </p>
      </div>
      <NuxtLink
        to="/admin/courses"
        class="inline-flex h-11 items-center gap-2 rounded-[8px] border-[2px] border-jv-ink bg-jv-yellow px-4 font-black text-jv-ink no-underline"
      >
        <Plus class="h-4 w-4" />
        Create or build Course
      </NuxtLink>
    </header>

    <main class="mx-auto mt-6 max-w-6xl">
      <section class="jv-border-uneven bg-jv-white p-4 shadow-brutal-sm sm:p-5">
        <div
          class="grid gap-3 md:grid-cols-[minmax(0,1fr)_220px_auto] md:items-end"
        >
          <label class="grid gap-1 text-sm font-black" for="course-search">
            Search Courses
            <span class="relative">
              <Search
                class="pointer-events-none absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-jv-muted"
                aria-hidden="true"
              />
              <input
                id="course-search"
                v-model="searchQuery"
                data-testid="course-search"
                type="search"
                class="h-11 w-full rounded-[8px] border-[2px] border-jv-ink bg-white pl-10 pr-3 font-bold"
                placeholder="Search by Course title"
              />
            </span>
          </label>
          <label
            class="grid gap-1 text-sm font-black"
            for="course-status-filter"
          >
            Publication state
            <select
              id="course-status-filter"
              v-model="statusFilter"
              data-testid="course-status-filter"
              class="h-11 rounded-[8px] border-[2px] border-jv-ink bg-white px-3 font-bold"
            >
              <option value="ALL">All states</option>
              <option value="DRAFT">Draft</option>
              <option value="PUBLISHED">Published</option>
              <option value="ARCHIVED">Archived</option>
            </select>
          </label>
          <p
            class="pb-2 text-sm font-black text-jv-muted"
            data-testid="course-result-summary"
            aria-live="polite"
          >
            {{ resultSummary }}
          </p>
        </div>
      </section>

      <p
        v-if="errorMessage"
        class="mt-4 rounded-[8px] border-[2px] border-red-700 bg-red-50 p-3 text-sm font-bold text-red-700"
        role="alert"
      >
        {{ errorMessage }}
      </p>

      <p v-if="loading" class="mt-5 font-bold text-jv-muted">
        Loading Courses…
      </p>
      <section
        v-else-if="courses.length === 0"
        class="jv-border-uneven mt-5 bg-jv-white p-6 text-center shadow-brutal-sm"
      >
        <h2 class="font-headings text-xl">No Courses created yet</h2>
        <p class="mt-2 font-bold text-jv-muted">
          Start in Course Builder to create the first Course.
        </p>
      </section>
      <section
        v-else-if="filteredCourses.length === 0"
        class="jv-border-uneven mt-5 bg-jv-white p-6 text-center shadow-brutal-sm"
        data-testid="course-empty-search"
      >
        <h2 class="font-headings text-xl">No matching Courses</h2>
        <p class="mt-2 font-bold text-jv-muted">
          Try a different title or publication state.
        </p>
      </section>
      <section
        v-else
        class="mt-5 grid gap-4 lg:grid-cols-2"
        data-testid="course-list"
      >
        <article
          v-for="course in filteredCourses"
          :key="course.id"
          class="jv-border-uneven flex min-h-52 flex-col bg-jv-white p-5 shadow-brutal-sm"
          :data-testid="`course-card-${course.id}`"
        >
          <form
            v-if="editingCourseId === course.id"
            class="grid gap-3"
            @submit.prevent="saveCourseTitle(course.id)"
          >
            <label class="grid gap-1 text-sm font-black">
              Course title
              <input
                v-model="editingCourseTitle"
                name="course_title"
                required
                maxlength="200"
                class="h-11 rounded-[8px] border-[2px] border-jv-ink bg-white px-3"
              />
            </label>
            <div class="flex flex-wrap gap-2">
              <button
                type="submit"
                class="h-10 rounded-[8px] border-[2px] border-jv-ink bg-jv-yellow px-3 text-sm font-black"
                :disabled="savingCourseId === course.id"
              >
                {{ savingCourseId === course.id ? "Saving…" : "Save title" }}
              </button>
              <button
                type="button"
                class="h-10 rounded-[8px] border-[2px] border-jv-ink bg-white px-3 text-sm font-black"
                :disabled="savingCourseId === course.id"
                @click="cancelCourseEdit"
              >
                Cancel
              </button>
            </div>
          </form>
          <template v-else>
            <div class="flex items-start justify-between gap-3">
              <h2 class="font-headings text-xl leading-snug">
                {{ course.title }}
              </h2>
              <span
                class="shrink-0 rounded-full border border-jv-ink px-2 py-1 text-xs font-black"
              >
                {{ course.status }}
              </span>
            </div>

            <label
              class="mt-4 grid max-w-52 gap-1 text-xs font-black uppercase"
            >
              Publication state
              <select
                :value="course.status"
                class="h-10 rounded-[8px] border-[2px] border-jv-ink bg-white px-3 text-sm font-bold normal-case"
                :aria-label="`Publication state for ${course.title}`"
                :disabled="updatingStatusId === course.id"
                @change="changeCourseStatus(course, $event.target.value)"
              >
                <option value="DRAFT">Draft</option>
                <option value="PUBLISHED">Published</option>
                <option value="ARCHIVED">Archived</option>
              </select>
            </label>

            <div class="mt-auto flex flex-wrap gap-2 pt-5">
              <NuxtLink
                :to="`/admin/courses?course=${course.id}`"
                class="inline-flex h-10 items-center rounded-[8px] border-[2px] border-jv-ink bg-jv-yellow px-3 text-sm font-black text-jv-ink no-underline"
              >
                Edit structure
              </NuxtLink>
              <NuxtLink
                :to="`/admin/courses/learning-items?course=${course.id}`"
                class="inline-flex h-10 items-center rounded-[8px] border-[2px] border-jv-ink bg-jv-green px-3 text-sm font-black text-jv-ink no-underline"
              >
                Manage content
              </NuxtLink>
              <button
                type="button"
                class="h-10 rounded-[8px] border-[2px] border-jv-ink bg-white px-3 text-sm font-black"
                :aria-label="`Rename ${course.title}`"
                @click="startCourseEdit(course)"
              >
                Rename
              </button>
            </div>
          </template>
        </article>
      </section>
    </main>
  </div>
</template>
