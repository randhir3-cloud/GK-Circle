<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { Plus } from "lucide-vue-next";
import { usePush } from "notivue";
import {
  getCourseAdminAPIError,
  useCourseLearningItemsApi,
} from "@/composables/course_learning_items";
import { setUserDataStore } from "@/composables/auth";
import { useUsersStore } from "~~/store/users";

definePageMeta({ layout: "empty" });
useSeoMeta({
  title: "Course Builder - GK Circle",
  description: "Create courses, subjects, and topics for PCS preparation.",
  robots: "noindex, nofollow",
});

const api = useCourseLearningItemsApi();
const toast = usePush();
const usersStore = useUsersStore();
const route = useRoute();

const courses = ref([]);
const coursesLoading = ref(true);
const coursesError = ref("");
const selectedCourseId = ref("");
const selectedCourse = computed(
  () => courses.value.find((c) => c.id === selectedCourseId.value) || null
);

const tree = ref(null);
const treeLoading = ref(false);
const treeError = ref("");

const newCourseTitle = ref("");
const creatingCourse = ref(false);

const nodeTitle = ref("");
const nodeType = ref("SUBJECT");
const parentId = ref("");
const creatingNode = ref(false);
const publishing = ref(false);
const formError = ref("");

const flattenNodes = (roots) => {
  const out = [];
  const walk = (items, depth = 0) => {
    for (const entry of items || []) {
      out.push({
        id: entry.node.id,
        title: entry.node.title,
        node_type: entry.node.node_type,
        depth,
      });
      walk(entry.children, depth + 1);
    }
  };
  walk(roots);
  return out;
};

const flatNodes = computed(() => flattenNodes(tree.value?.roots || []));
const subjectNodes = computed(() =>
  flatNodes.value.filter((node) => node.node_type === "SUBJECT")
);
const availableParentNodes = computed(() =>
  nodeType.value === "TOPIC" ? subjectNodes.value : flatNodes.value
);
const nodeFormHeading = computed(() => {
  if (nodeType.value === "TOPIC") return "Add topic under a subject";
  if (nodeType.value === "SECTION") return "Add structural section";
  return "Add top-level subject";
});
const parentFieldLabel = computed(() =>
  nodeType.value === "TOPIC" ? "Subject (required)" : "Parent (optional)"
);
const nodeButtonLabel = computed(() => {
  if (nodeType.value === "TOPIC") return "Add topic";
  if (nodeType.value === "SECTION") return "Add section";
  return "Add subject";
});

const findHierarchyEntry = (entries, nodeId) => {
  for (const entry of entries || []) {
    if (entry.node.id === nodeId) return entry;
    const childMatch = findHierarchyEntry(entry.children, nodeId);
    if (childMatch) return childMatch;
  }
  return null;
};

const nextSiblingPosition = (selectedParentId) => {
  if (!selectedParentId) {
    return tree.value?.roots?.length || 0;
  }
  const parent = findHierarchyEntry(tree.value?.roots, selectedParentId);
  return parent?.children?.length || 0;
};

const loadCourses = async () => {
  coursesLoading.value = true;
  coursesError.value = "";
  try {
    courses.value = (await api.listCourses()) || [];
  } catch (error) {
    coursesError.value = getCourseAdminAPIError(
      error,
      "Unable to load Courses."
    );
  } finally {
    coursesLoading.value = false;
  }
};

const loadTree = async (courseId) => {
  if (!courseId) {
    tree.value = null;
    return;
  }
  treeLoading.value = true;
  treeError.value = "";
  try {
    tree.value = await api.getTree(courseId);
  } catch (error) {
    treeError.value = getCourseAdminAPIError(
      error,
      "Unable to load course outline."
    );
    tree.value = null;
  } finally {
    treeLoading.value = false;
  }
};

watch(selectedCourseId, (id) => {
  parentId.value = "";
  loadTree(id);
});

watch(nodeType, () => {
  parentId.value = "";
  formError.value = "";
});

const createCourse = async () => {
  formError.value = "";
  const title = newCourseTitle.value.trim();
  if (!title) {
    formError.value = "Course title is required.";
    return;
  }
  creatingCourse.value = true;
  try {
    const course = await api.createCourse({ title });
    newCourseTitle.value = "";
    await loadCourses();
    selectedCourseId.value = course.id;
    toast.success("Course created as DRAFT.");
  } catch (error) {
    formError.value = getCourseAdminAPIError(error, "Unable to create course.");
  } finally {
    creatingCourse.value = false;
  }
};

const createNode = async () => {
  formError.value = "";
  if (!selectedCourseId.value) {
    formError.value = "Select a course first.";
    return;
  }
  const title = nodeTitle.value.trim();
  if (!title) {
    formError.value = "Node title is required.";
    return;
  }
  if (nodeType.value === "TOPIC" && !parentId.value) {
    formError.value = "Select a subject for this topic.";
    return;
  }
  if (
    nodeType.value === "TOPIC" &&
    !subjectNodes.value.some((subject) => subject.id === parentId.value)
  ) {
    formError.value = "Topics must be created under a subject.";
    return;
  }
  const selectedParentId = nodeType.value === "SUBJECT" ? "" : parentId.value;
  creatingNode.value = true;
  try {
    const body = {
      title,
      node_type: nodeType.value,
      position: nextSiblingPosition(selectedParentId),
    };
    if (selectedParentId) body.parent_id = selectedParentId;
    await api.createNode(selectedCourseId.value, body);
    nodeTitle.value = "";
    await loadTree(selectedCourseId.value);
    toast.success("Node added.");
  } catch (error) {
    formError.value = getCourseAdminAPIError(error, "Unable to create node.");
  } finally {
    creatingNode.value = false;
  }
};

const setStatus = async (status) => {
  if (!selectedCourseId.value) return;
  publishing.value = true;
  formError.value = "";
  try {
    await api.updateCourse(selectedCourseId.value, { status });
    await loadCourses();
    toast.success(`Course status set to ${status}.`);
  } catch (error) {
    formError.value = getCourseAdminAPIError(
      error,
      "Unable to update course status."
    );
  } finally {
    publishing.value = false;
  }
};

onMounted(async () => {
  try {
    await setUserDataStore(usersStore);
  } catch {
    /* auth store optional for page shell */
  }
  await loadCourses();
  const requestedCourseId =
    typeof route.query.course === "string" ? route.query.course : "";
  if (courses.value.some((course) => course.id === requestedCourseId)) {
    selectedCourseId.value = requestedCourseId;
  }
});
</script>

<template>
  <div class="min-h-screen bg-jv-cream px-4 py-6 text-jv-ink sm:px-6">
    <header class="mx-auto max-w-5xl">
      <p class="text-xs font-black uppercase tracking-wide text-jv-muted">
        Exam Platform · P1
      </p>
      <h1 class="font-headings text-3xl sm:text-4xl">Course Builder</h1>
      <p class="mt-2 max-w-2xl text-sm font-bold text-jv-muted">
        Create a PCS course, add subjects and topics, then publish when ready.
      </p>
      <NuxtLink
        to="/admin/courses/learning-items"
        class="mt-3 inline-block text-sm font-black underline"
      >
        Open Learning Items editor
      </NuxtLink>
      <span class="mx-2 text-jv-muted" aria-hidden="true">·</span>
      <NuxtLink
        to="/admin/courses/list"
        class="mt-3 inline-block text-sm font-black underline"
      >
        Browse all Courses
      </NuxtLink>
    </header>

    <div class="mx-auto mt-6 grid max-w-5xl gap-4 lg:grid-cols-2">
      <section class="jv-border-uneven bg-jv-white p-4 shadow-brutal-sm sm:p-5">
        <h2 class="font-headings text-xl">Create course</h2>
        <label class="mt-3 block text-sm font-bold" for="new-course-title">
          Title
        </label>
        <input
          id="new-course-title"
          v-model="newCourseTitle"
          data-testid="new-course-title"
          class="mt-1 h-11 w-full rounded-[8px] border-[2px] border-jv-ink px-3 font-bold"
          type="text"
          maxlength="200"
          placeholder="e.g. PCS Prelims GS"
        />
        <button
          type="button"
          data-testid="create-course-button"
          class="mt-3 inline-flex h-11 items-center gap-2 rounded-[8px] border-[2px] border-jv-ink bg-jv-yellow px-4 font-black disabled:opacity-60"
          :disabled="creatingCourse"
          @click="createCourse"
        >
          <Plus class="h-4 w-4" />
          Create draft course
        </button>
      </section>

      <section class="jv-border-uneven bg-jv-white p-4 shadow-brutal-sm sm:p-5">
        <h2 class="font-headings text-xl">Select course</h2>
        <label class="mt-3 block text-sm font-bold" for="builder-course">
          Course
        </label>
        <select
          id="builder-course"
          v-model="selectedCourseId"
          data-testid="builder-course-selector"
          class="mt-1 h-11 w-full rounded-[8px] border-[2px] border-jv-ink px-3 font-bold"
          :disabled="coursesLoading"
        >
          <option value="">
            {{ coursesLoading ? "Loading…" : "Select a Course" }}
          </option>
          <option v-for="course in courses" :key="course.id" :value="course.id">
            {{ course.title }} · {{ course.status }}
          </option>
        </select>
        <p v-if="coursesError" class="mt-2 text-sm font-bold text-red-700">
          {{ coursesError }}
        </p>
        <div v-if="selectedCourse" class="mt-4 flex flex-wrap gap-2">
          <button
            type="button"
            data-testid="publish-course-button"
            class="h-10 rounded-[8px] border-[2px] border-jv-ink bg-jv-green px-3 text-sm font-black text-white disabled:opacity-60"
            :disabled="publishing || selectedCourse.status === 'PUBLISHED'"
            @click="setStatus('PUBLISHED')"
          >
            Publish
          </button>
          <button
            type="button"
            data-testid="draft-course-button"
            class="h-10 rounded-[8px] border-[2px] border-jv-ink bg-jv-white px-3 text-sm font-black disabled:opacity-60"
            :disabled="publishing || selectedCourse.status === 'DRAFT'"
            @click="setStatus('DRAFT')"
          >
            Unpublish to draft
          </button>
          <button
            type="button"
            data-testid="archive-course-button"
            class="h-10 rounded-[8px] border-[2px] border-jv-ink bg-jv-ink px-3 text-sm font-black text-white disabled:opacity-60"
            :disabled="publishing || selectedCourse.status === 'ARCHIVED'"
            @click="setStatus('ARCHIVED')"
          >
            Archive
          </button>
        </div>
      </section>

      <section
        id="course-node-editor"
        class="jv-border-uneven bg-jv-white p-4 shadow-brutal-sm sm:p-5 lg:col-span-2"
      >
        <h2 class="font-headings text-xl">{{ nodeFormHeading }}</h2>
        <p class="mt-1 text-sm font-bold text-jv-muted">
          {{
            nodeType === "TOPIC"
              ? "Choose the Subject that will contain this Topic."
              : nodeType === "SUBJECT"
              ? "Subjects are always created at the Course root."
              : "Sections may be placed at the root or below another node."
          }}
        </p>
        <div class="mt-3 grid gap-3 md:grid-cols-3">
          <div>
            <label class="block text-sm font-bold" for="node-title"
              >Title</label
            >
            <input
              id="node-title"
              v-model="nodeTitle"
              data-testid="node-title"
              class="mt-1 h-11 w-full rounded-[8px] border-[2px] border-jv-ink px-3 font-bold"
              type="text"
            />
          </div>
          <div>
            <label class="block text-sm font-bold" for="node-type">Type</label>
            <select
              id="node-type"
              v-model="nodeType"
              data-testid="node-type"
              class="mt-1 h-11 w-full rounded-[8px] border-[2px] border-jv-ink px-3 font-bold"
            >
              <option value="SECTION">SECTION</option>
              <option value="SUBJECT">SUBJECT</option>
              <option value="TOPIC">TOPIC</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-bold" for="node-parent">
              {{ parentFieldLabel }}
            </label>
            <select
              id="node-parent"
              v-model="parentId"
              data-testid="node-parent"
              class="mt-1 h-11 w-full rounded-[8px] border-[2px] border-jv-ink px-3 font-bold"
            >
              <option value="">
                {{
                  nodeType === "TOPIC" ? "Select a Subject" : "Top-level root"
                }}
              </option>
              <option
                v-for="node in availableParentNodes"
                :key="node.id"
                :value="node.id"
                :disabled="nodeType === 'SUBJECT'"
              >
                {{ "—".repeat(node.depth) }} {{ node.title }} ({{
                  node.node_type
                }})
              </option>
            </select>
          </div>
        </div>
        <button
          type="button"
          data-testid="create-node-button"
          class="mt-3 h-11 rounded-[8px] border-[2px] border-jv-ink bg-jv-yellow px-4 font-black disabled:opacity-60"
          :disabled="
            creatingNode ||
            !selectedCourseId ||
            (nodeType === 'TOPIC' && !parentId)
          "
          @click="createNode"
        >
          {{ nodeButtonLabel }}
        </button>
      </section>

      <section
        class="jv-border-uneven bg-jv-white p-4 shadow-brutal-sm sm:p-5 lg:col-span-2"
      >
        <h2 class="font-headings text-xl">Outline</h2>
        <p v-if="treeLoading" class="mt-2 text-sm font-bold text-jv-muted">
          Loading outline…
        </p>
        <p
          v-else-if="treeError"
          class="mt-2 text-sm font-bold text-red-700"
          role="alert"
        >
          {{ treeError }}
        </p>
        <ul
          v-else-if="flatNodes.length"
          data-testid="course-outline"
          class="mt-3 space-y-2"
        >
          <li
            v-for="node in flatNodes"
            :key="node.id"
            class="rounded-[8px] border-[2px] px-3 py-2 font-bold"
            :class="
              node.depth
                ? 'border-jv-green/60 bg-jv-green/10'
                : 'border-jv-ink/20'
            "
            :data-depth="node.depth"
            :style="{ marginLeft: `${node.depth * 32}px` }"
          >
            <span v-if="node.depth" aria-hidden="true">↳ </span>
            <span class="text-xs font-black uppercase text-jv-muted">{{
              node.node_type
            }}</span>
            — {{ node.title }}
          </li>
        </ul>
        <p v-else class="mt-2 text-sm font-bold text-jv-muted">
          {{
            selectedCourseId
              ? "No nodes yet. Add a SUBJECT or TOPIC above."
              : "Select a course to view its outline."
          }}
        </p>
      </section>
    </div>

    <p
      v-if="formError"
      class="mx-auto mt-4 max-w-5xl text-sm font-bold text-red-700"
      role="alert"
      data-testid="builder-form-error"
    >
      {{ formError }}
    </p>
  </div>
</template>
