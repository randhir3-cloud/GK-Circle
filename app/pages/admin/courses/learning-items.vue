<script setup>
import { onMounted, ref } from "vue";
import { Plus } from "lucide-vue-next";
import { usePush } from "notivue";
import CourseSelector from "@/components/course/CourseSelector.vue";
import CourseNodeSelector from "@/components/course/CourseNodeSelector.vue";
import LearningItemDeleteDialog from "@/components/course/LearningItemDeleteDialog.vue";
import LearningItemEditorDialog from "@/components/course/LearningItemEditorDialog.vue";
import LearningItemEmptyState from "@/components/course/LearningItemEmptyState.vue";
import LearningItemTable from "@/components/course/LearningItemTable.vue";
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
  title: "Course Content - GK Circle",
  description: "Manage CourseNode LearningItems.",
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

const quizzes = ref([]);
const quizzesLoading = ref(true);

const nodeLevels = ref([]);
const nodesLoading = ref(false);
const nodesError = ref("");
const selectedNodeId = ref("");

const items = ref([]);
const itemsLoading = ref(false);
const itemsLoaded = ref(false);
const itemsError = ref("");

const editorOpen = ref(false);
const editorMode = ref("create");
const editingItem = ref(null);
const saving = ref(false);
const mutationError = ref("");

const deleteOpen = ref(false);
const deletingItem = ref(null);
const deleting = ref(false);
const deleteError = ref("");

let courseLoadToken = 0;
let nodeLoadToken = 0;
let itemLoadToken = 0;

const loadCourses = async () => {
  const token = ++courseLoadToken;
  coursesLoading.value = true;
  coursesError.value = "";
  try {
    const result = await api.listCourses();
    if (token === courseLoadToken) courses.value = result || [];
  } catch (error) {
    if (token === courseLoadToken) {
      coursesError.value = getCourseAdminAPIError(
        error,
        "Unable to load Courses."
      );
    }
  } finally {
    if (token === courseLoadToken) coursesLoading.value = false;
  }
};

const loadQuizzes = async () => {
  quizzesLoading.value = true;
  try {
    quizzes.value = (await api.listQuizzes()) || [];
  } catch {
    quizzes.value = [];
  } finally {
    quizzesLoading.value = false;
  }
};

const resetItems = () => {
  selectedNodeId.value = "";
  items.value = [];
  itemsLoaded.value = false;
  itemsLoading.value = false;
  itemsError.value = "";
  itemLoadToken += 1;
};

const selectCourse = async (courseId) => {
  selectedCourseId.value = courseId;
  nodeLevels.value = [];
  nodesError.value = "";
  nodeLoadToken += 1;
  resetItems();
  if (!courseId) return;

  const token = ++nodeLoadToken;
  nodesLoading.value = true;
  nodeLevels.value = [
    { parentId: null, selectedId: "", nodes: [], loading: true },
  ];
  try {
    const roots = await api.listRootNodes(courseId);
    if (token !== nodeLoadToken) return;
    nodeLevels.value = [
      { parentId: null, selectedId: "", nodes: roots || [], loading: false },
    ];
  } catch (error) {
    if (token === nodeLoadToken) {
      nodeLevels.value = [];
      nodesError.value = getCourseAdminAPIError(
        error,
        "Unable to load CourseNodes."
      );
    }
  } finally {
    if (token === nodeLoadToken) nodesLoading.value = false;
  }
};

const loadItems = async (courseId, nodeId) => {
  const token = ++itemLoadToken;
  itemsLoading.value = true;
  itemsLoaded.value = false;
  itemsError.value = "";
  try {
    const result = await api.listItems(courseId, nodeId);
    if (token !== itemLoadToken) return false;
    items.value = result || [];
    itemsLoaded.value = true;
    return true;
  } catch (error) {
    if (token === itemLoadToken) {
      items.value = [];
      itemsError.value = getCourseAdminAPIError(
        error,
        "Unable to load Learning Items."
      );
    }
    return false;
  } finally {
    if (token === itemLoadToken) itemsLoading.value = false;
  }
};

const selectNode = async ({ levelIndex, nodeId }) => {
  const current = nodeLevels.value[levelIndex];
  if (!current) return;
  current.selectedId = nodeId;
  nodeLevels.value = nodeLevels.value.slice(0, levelIndex + 1);
  nodesError.value = "";
  nodeLoadToken += 1;
  resetItems();
  if (!nodeId) return;

  selectedNodeId.value = nodeId;
  const courseId = selectedCourseId.value;
  const token = ++nodeLoadToken;
  nodesLoading.value = true;
  void loadItems(courseId, nodeId);

  try {
    const children = await api.listChildren(courseId, nodeId);
    if (token !== nodeLoadToken || !children?.length) return;
    nodeLevels.value.push({
      parentId: nodeId,
      selectedId: "",
      nodes: children,
      loading: false,
    });
  } catch (error) {
    if (token === nodeLoadToken) {
      nodesError.value = getCourseAdminAPIError(
        error,
        "Unable to load child CourseNodes."
      );
    }
  } finally {
    if (token === nodeLoadToken) nodesLoading.value = false;
  }
};

const openCreate = () => {
  editorMode.value = "create";
  editingItem.value = null;
  mutationError.value = "";
  editorOpen.value = true;
};

const openEdit = (item) => {
  editorMode.value = "update";
  editingItem.value = item;
  mutationError.value = "";
  editorOpen.value = true;
};

const saveItem = async (payload) => {
  if (!selectedCourseId.value || !selectedNodeId.value || saving.value) return;
  saving.value = true;
  mutationError.value = "";
  try {
    if (editorMode.value === "create") {
      await api.createItem(
        selectedCourseId.value,
        selectedNodeId.value,
        payload
      );
    } else {
      await api.updateItem(
        selectedCourseId.value,
        selectedNodeId.value,
        editingItem.value.id,
        payload
      );
    }
    editorOpen.value = false;
    toast.success(
      editorMode.value === "create"
        ? "Learning Item created."
        : "Learning Item updated."
    );
    await loadItems(selectedCourseId.value, selectedNodeId.value);
  } catch (error) {
    mutationError.value = getCourseAdminAPIError(
      error,
      "Unable to save the Learning Item."
    );
    toast.error(mutationError.value);
  } finally {
    saving.value = false;
  }
};

const openDelete = (item) => {
  deletingItem.value = item;
  deleteError.value = "";
  deleteOpen.value = true;
};

const deleteItem = async () => {
  if (!deletingItem.value || deleting.value) return;
  deleting.value = true;
  deleteError.value = "";
  try {
    await api.deleteItem(
      selectedCourseId.value,
      selectedNodeId.value,
      deletingItem.value.id
    );
    deleteOpen.value = false;
    toast.success("Learning Item deleted.");
    await loadItems(selectedCourseId.value, selectedNodeId.value);
  } catch (error) {
    deleteError.value = getCourseAdminAPIError(
      error,
      "Unable to delete the Learning Item."
    );
    toast.error(deleteError.value);
  } finally {
    deleting.value = false;
  }
};

onMounted(async () => {
  if (!usersStore.userData) await setUserDataStore();
  await Promise.all([loadCourses(), loadQuizzes()]);
  const requestedCourseId = String(route.query.course || "");
  if (
    requestedCourseId &&
    courses.value.some((course) => course.id === requestedCourseId)
  ) {
    await selectCourse(requestedCourseId);
  }
});
</script>

<template>
  <main class="min-h-screen bg-jv-canvas px-4 py-5 sm:px-6 md:px-8 md:py-7">
    <div class="mx-auto flex max-w-6xl flex-col gap-6">
      <header>
        <h1
          class="font-headings text-[38px] leading-none text-jv-ink sm:text-[52px]"
        >
          Course Content
        </h1>
        <p class="mt-2 max-w-2xl font-bold text-jv-muted">
          Manage ordered LearningItems on one selected CourseNode.
        </p>
      </header>

      <CourseSelector
        :courses="courses"
        :model-value="selectedCourseId"
        :loading="coursesLoading"
        :error="coursesError"
        @update:model-value="selectCourse"
      />
      <CourseNodeSelector
        :levels="nodeLevels"
        :loading="nodesLoading"
        :error="nodesError"
        @select="selectNode"
      />

      <section class="jv-border-uneven bg-jv-white p-4 shadow-brutal-sm sm:p-5">
        <div
          class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
        >
          <div>
            <h2 class="font-headings text-2xl text-jv-ink">Learning Items</h2>
            <p class="text-sm font-bold text-jv-muted">
              Position and ordering come directly from the server.
            </p>
          </div>
          <button
            type="button"
            class="inline-flex h-11 items-center justify-center gap-2 rounded-full border-[2px] border-jv-ink bg-jv-coral px-5 font-black text-white shadow-brutal-sm disabled:opacity-50"
            :disabled="!selectedNodeId || itemsLoading"
            @click="openCreate"
          >
            <Plus class="size-5" /> New Learning Item
          </button>
        </div>

        <p
          v-if="!selectedNodeId"
          class="mt-5 rounded-[8px] bg-jv-canvas p-4 font-bold text-jv-muted"
          data-testid="item-prompt"
        >
          Select a CourseNode to load its LearningItems.
        </p>
        <p
          v-else-if="itemsLoading"
          class="mt-5 rounded-[8px] bg-jv-canvas p-4 font-bold text-jv-muted"
          data-testid="items-loading"
        >
          Loading Learning Items…
        </p>
        <p
          v-else-if="itemsError"
          class="mt-5 rounded-[8px] bg-red-100 p-4 font-bold text-red-800"
          role="alert"
        >
          {{ itemsError }}
        </p>
        <LearningItemEmptyState
          v-else-if="itemsLoaded && items.length === 0"
          class="mt-5"
        />
        <LearningItemTable
          v-else-if="itemsLoaded"
          class="mt-5"
          :items="items"
          @edit="openEdit"
          @delete="openDelete"
        />
      </section>
    </div>

    <LearningItemEditorDialog
      v-model="editorOpen"
      :mode="editorMode"
      :item="editingItem"
      :quizzes="quizzes"
      :quizzes-loading="quizzesLoading"
      :saving="saving"
      :error="mutationError"
      @save="saveItem"
    />
    <LearningItemDeleteDialog
      v-model="deleteOpen"
      :item="deletingItem"
      :deleting="deleting"
      :error="deleteError"
      @confirm="deleteItem"
    />
  </main>
</template>
