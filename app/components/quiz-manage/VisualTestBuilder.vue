<template>
  <section
    class="jv-border-rough bg-jv-white p-4 shadow-brutal-sm sm:p-5"
    aria-labelledby="visual-test-builder-title"
  >
    <header
      class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between"
    >
      <div class="min-w-0">
        <h2
          id="visual-test-builder-title"
          class="font-headings text-[28px] leading-none text-jv-ink sm:text-[32px]"
        >
          Visual Test Builder
        </h2>
        <p class="mt-2 max-w-2xl text-[15px] leading-[1.5] text-jv-muted">
          Compose STATIC (ordered membership) and DYNAMIC (filter-based)
          collections from this quiz’s Question Bank. Collection logic stays on
          the server.
        </p>
      </div>
      <div v-if="canEdit" class="flex flex-wrap gap-2">
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-[999px] border-[3px] border-jv-ink bg-jv-yellow px-4 py-2 text-[14px] font-black text-jv-ink shadow-brutal-sm transition-transform hover:-rotate-[1deg] disabled:opacity-60"
          :disabled="pending || savingInlineQuestion"
          data-testid="inline-add-question"
          @click="openInlineQuestionForm"
        >
          <Plus class="size-4" :stroke-width="2.4" />
          Add New Question
        </button>
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-[999px] border-[3px] border-jv-ink bg-jv-white px-4 py-2 text-[14px] font-black text-jv-ink shadow-brutal-sm transition-transform hover:-rotate-[1deg] disabled:opacity-60"
          :disabled="pending"
          @click="startCreate('STATIC')"
        >
          <Plus class="size-4" :stroke-width="2.4" />
          Static collection
        </button>
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-[999px] border-[3px] border-jv-ink bg-jv-accent-green px-4 py-2 text-[14px] font-black text-white shadow-brutal-sm transition-transform hover:-rotate-[1deg] disabled:opacity-60"
          :disabled="pending"
          @click="startCreate('DYNAMIC')"
        >
          <Plus class="size-4" :stroke-width="2.4" />
          Dynamic collection
        </button>
      </div>
    </header>

    <section
      v-if="showInlineQuestionForm && canEdit"
      class="mt-5 grid gap-3"
      data-testid="inline-question-form"
      aria-label="Inline add new question"
    >
      <p class="text-[14px] text-jv-muted">
        Creates the question in this quiz’s Question Bank (lineage, revision,
        and answer authority). No parallel store.
      </p>
      <label
        v-if="selected?.kind === 'STATIC'"
        class="inline-flex items-center gap-2 text-[14px] font-semibold text-jv-ink"
      >
        <input
          v-model="linkToStaticCollection"
          type="checkbox"
          class="size-4 border-[2px] border-jv-ink"
          data-testid="link-to-static-collection"
        />
        Also link to selected STATIC collection
        <span class="font-normal text-jv-muted">({{ selected.title }})</span>
      </label>
      <p
        v-else-if="selected?.kind === 'DYNAMIC'"
        class="text-[14px] text-jv-muted"
      >
        DYNAMIC collections resolve from filters — the new question is linked to
        the quiz bank only (not forced into membership).
      </p>
      <QuestionFormCard
        mode="create"
        :quiz-id="quizId"
        :saving="savingInlineQuestion"
        @save="saveInlineQuestion"
        @cancel="closeInlineQuestionForm"
      />
      <p
        v-if="inlineQuestionError"
        class="text-[14px] font-semibold text-jv-coral"
        role="alert"
      >
        {{ inlineQuestionError }}
      </p>
    </section>

    <p
      v-if="loadError"
      class="mt-4 text-[15px] font-semibold text-jv-coral"
      role="alert"
    >
      {{ loadError }}
    </p>
    <p v-else-if="loading" class="mt-4 text-[15px] font-semibold text-jv-muted">
      Loading collections…
    </p>

    <div
      v-else
      class="mt-5 grid gap-5 lg:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]"
    >
      <div class="grid content-start gap-3">
        <h3
          class="text-[13px] font-black uppercase tracking-[0.16em] text-jv-ink"
        >
          Collections
        </h3>
        <p
          v-if="collections.length === 0"
          class="border-[3px] border-dashed border-jv-ink/30 bg-jv-canvas p-4 text-[15px] text-jv-muted"
        >
          No collections yet. Create a STATIC or DYNAMIC collection to begin.
        </p>
        <ul v-else class="grid gap-2">
          <li v-for="collection in collections" :key="collection.id">
            <button
              type="button"
              class="flex w-full items-start justify-between gap-3 border-[3px] border-jv-ink px-3 py-3 text-left transition-colors"
              :class="
                selectedId === collection.id
                  ? 'bg-jv-yellow/40 shadow-brutal-sm'
                  : 'bg-jv-canvas hover:bg-jv-white'
              "
              @click="selectCollection(collection.id)"
            >
              <span class="min-w-0">
                <span class="block truncate text-[16px] font-black text-jv-ink">
                  {{ collection.title }}
                </span>
                <span
                  class="mt-1 inline-flex items-center gap-2 text-[12px] font-bold uppercase tracking-[0.12em] text-jv-muted"
                >
                  <span
                    class="rounded-full border border-jv-ink/25 bg-jv-white px-2 py-0.5"
                  >
                    {{ collection.kind }}
                  </span>
                  <span>pos {{ collection.position }}</span>
                </span>
              </span>
            </button>
          </li>
        </ul>
      </div>

      <div class="min-w-0 border-[3px] border-jv-ink bg-jv-canvas p-4 sm:p-5">
        <template v-if="creating">
          <h3 class="text-[18px] font-black text-jv-ink">
            New {{ createKind }} collection
          </h3>
          <form class="mt-4 grid gap-4" @submit.prevent="submitCreate">
            <label class="grid gap-2">
              <span
                class="text-[13px] font-black uppercase tracking-[0.16em] text-jv-ink"
              >
                Title <span class="text-jv-coral">*</span>
              </span>
              <input
                v-model="createTitle"
                type="text"
                required
                maxlength="200"
                class="h-12 border-[3px] border-jv-ink bg-jv-white px-3 text-[15px] font-semibold text-jv-ink outline-none focus:shadow-brutal-sm"
              />
            </label>
            <label class="grid gap-2">
              <span
                class="text-[13px] font-black uppercase tracking-[0.16em] text-jv-ink"
              >
                Position
              </span>
              <input
                v-model.number="createPosition"
                type="number"
                min="0"
                class="h-12 w-28 border-[3px] border-jv-ink bg-jv-white px-3 text-[15px] font-semibold text-jv-ink outline-none"
              />
            </label>
            <div
              v-if="createKind === 'DYNAMIC'"
              class="grid gap-3 border-[2px] border-dashed border-jv-ink/35 bg-jv-white p-3"
            >
              <p
                class="text-[13px] font-black uppercase tracking-[0.14em] text-jv-ink"
              >
                Dynamic filters
              </p>
              <p class="text-[14px] text-jv-muted">
                Leave empty to resolve to all quiz-linked questions. Metadata
                filters are stored now; full taxonomy resolution ships later.
              </p>
              <DynamicFilterFields v-model="createFilter" />
            </div>
            <p
              v-if="actionError"
              class="text-[14px] font-semibold text-jv-coral"
            >
              {{ actionError }}
            </p>
            <div class="flex flex-wrap gap-2">
              <button
                type="submit"
                class="rounded-[999px] border-[3px] border-jv-ink bg-jv-accent-green px-4 py-2 text-[14px] font-black text-white disabled:opacity-60"
                :disabled="pending"
              >
                Create
              </button>
              <button
                type="button"
                class="rounded-[999px] border-[3px] border-jv-ink bg-jv-white px-4 py-2 text-[14px] font-black text-jv-ink"
                :disabled="pending"
                @click="cancelCreate"
              >
                Cancel
              </button>
            </div>
          </form>
        </template>

        <template v-else-if="selected">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="min-w-0">
              <p
                class="text-[12px] font-bold uppercase tracking-[0.14em] text-jv-coral"
              >
                {{ selected.kind }} collection
              </p>
              <h3 class="mt-1 break-words text-[22px] font-black text-jv-ink">
                {{ selected.title }}
              </h3>
            </div>
            <button
              v-if="canEdit"
              type="button"
              class="rounded-[999px] border-[3px] border-jv-ink bg-jv-coral px-3 py-1.5 text-[13px] font-black text-white disabled:opacity-60"
              :disabled="pending"
              @click="removeSelected"
            >
              Delete
            </button>
          </div>

          <form
            v-if="canEdit"
            class="mt-4 grid gap-4"
            @submit.prevent="saveDetails"
          >
            <label class="grid gap-2">
              <span
                class="text-[13px] font-black uppercase tracking-[0.16em] text-jv-ink"
              >
                Title
              </span>
              <input
                v-model="editTitle"
                type="text"
                required
                maxlength="200"
                class="h-12 border-[3px] border-jv-ink bg-jv-white px-3 text-[15px] font-semibold text-jv-ink outline-none"
              />
            </label>
            <label class="grid gap-2">
              <span
                class="text-[13px] font-black uppercase tracking-[0.16em] text-jv-ink"
              >
                Position
              </span>
              <input
                v-model.number="editPosition"
                type="number"
                min="0"
                class="h-12 w-28 border-[3px] border-jv-ink bg-jv-white px-3 text-[15px] font-semibold text-jv-ink outline-none"
              />
            </label>

            <div
              v-if="selected.kind === 'DYNAMIC'"
              class="grid gap-3 border-[2px] border-dashed border-jv-ink/35 bg-jv-white p-3"
            >
              <p
                class="text-[13px] font-black uppercase tracking-[0.14em] text-jv-ink"
              >
                Dynamic filters
              </p>
              <DynamicFilterFields v-model="editFilter" />
            </div>

            <button
              type="submit"
              class="w-fit rounded-[999px] border-[3px] border-jv-ink bg-jv-white px-4 py-2 text-[14px] font-black text-jv-ink disabled:opacity-60"
              :disabled="pending"
            >
              Save details
            </button>
          </form>

          <section
            v-if="selected.kind === 'STATIC'"
            class="mt-5 grid gap-3"
            aria-label="Static ordered membership"
          >
            <h4
              class="text-[13px] font-black uppercase tracking-[0.16em] text-jv-ink"
            >
              Ordered STATIC membership
            </h4>
            <p class="text-[14px] text-jv-muted">
              Membership is a fixed ordered list of questions already in this
              quiz’s bank. Reorder with the controls, then save.
            </p>

            <ul
              v-if="memberIds.length > 0"
              class="grid gap-2"
              data-testid="static-member-list"
            >
              <li
                v-for="(questionId, index) in memberIds"
                :key="questionId"
                class="flex items-center gap-2 border-[2px] border-jv-ink bg-jv-white px-3 py-2"
              >
                <span class="w-7 shrink-0 text-[13px] font-black text-jv-coral">
                  {{ index + 1 }}
                </span>
                <span
                  class="min-w-0 flex-1 truncate text-[14px] font-semibold text-jv-ink"
                >
                  {{ titleFor(questionId) }}
                </span>
                <div v-if="canEdit" class="flex shrink-0 gap-1">
                  <button
                    type="button"
                    class="grid size-8 place-items-center border border-jv-ink/30 text-jv-ink disabled:opacity-40"
                    aria-label="Move up"
                    :disabled="index === 0 || pending"
                    @click="moveMember(index, -1)"
                  >
                    <ChevronUp class="size-4" :stroke-width="2.4" />
                  </button>
                  <button
                    type="button"
                    class="grid size-8 place-items-center border border-jv-ink/30 text-jv-ink disabled:opacity-40"
                    aria-label="Move down"
                    :disabled="index === memberIds.length - 1 || pending"
                    @click="moveMember(index, 1)"
                  >
                    <ChevronDown class="size-4" :stroke-width="2.4" />
                  </button>
                  <button
                    type="button"
                    class="grid size-8 place-items-center border border-jv-ink/30 text-jv-coral"
                    aria-label="Remove from collection"
                    :disabled="pending"
                    @click="removeMember(questionId)"
                  >
                    <Trash2 class="size-4" :stroke-width="2.4" />
                  </button>
                </div>
              </li>
            </ul>
            <p
              v-else
              class="border-[2px] border-dashed border-jv-ink/30 bg-jv-white p-3 text-[14px] text-jv-muted"
            >
              No members yet. Add questions from the bank below.
            </p>

            <div v-if="canEdit" class="grid gap-2">
              <label class="grid gap-2">
                <span
                  class="text-[13px] font-black uppercase tracking-[0.16em] text-jv-ink"
                >
                  Add from Question Bank
                </span>
                <select
                  v-model="addQuestionId"
                  class="h-12 border-[3px] border-jv-ink bg-jv-white px-3 text-[15px] font-semibold text-jv-ink"
                >
                  <option value="">Select a question…</option>
                  <option
                    v-for="question in availableQuestions"
                    :key="question.question_id"
                    :value="question.question_id"
                  >
                    {{ question.question }}
                  </option>
                </select>
              </label>
              <div class="flex flex-wrap gap-2">
                <button
                  type="button"
                  class="rounded-[999px] border-[3px] border-jv-ink bg-jv-white px-4 py-2 text-[14px] font-black text-jv-ink disabled:opacity-60"
                  :disabled="!addQuestionId || pending"
                  @click="addMember"
                >
                  Add to membership
                </button>
                <button
                  type="button"
                  class="rounded-[999px] border-[3px] border-jv-ink bg-jv-accent-green px-4 py-2 text-[14px] font-black text-white disabled:opacity-60"
                  :disabled="pending || !membershipDirty"
                  @click="saveMembership"
                >
                  Save membership
                </button>
              </div>
            </div>
          </section>

          <section
            class="mt-5 grid gap-3"
            aria-label="Collection resolve preview"
          >
            <div class="flex flex-wrap items-center justify-between gap-2">
              <h4
                class="text-[13px] font-black uppercase tracking-[0.16em] text-jv-ink"
              >
                Resolve preview
              </h4>
              <button
                type="button"
                class="rounded-[999px] border-[3px] border-jv-ink bg-jv-white px-3 py-1.5 text-[13px] font-black text-jv-ink disabled:opacity-60"
                :disabled="pending"
                @click="runResolve"
              >
                Refresh resolve
              </button>
            </div>
            <p
              v-if="resolveError"
              class="text-[14px] font-semibold text-jv-coral"
            >
              {{ resolveError }}
            </p>
            <div
              v-else-if="resolution"
              class="grid gap-2 border-[2px] border-jv-ink bg-jv-white p-3"
              data-testid="resolve-preview"
            >
              <p class="text-[14px] font-black text-jv-ink">
                {{ describeResolutionStatus(resolution.resolution_status) }}
              </p>
              <p v-if="resolution.message" class="text-[14px] text-jv-muted">
                {{ resolution.message }}
              </p>
              <p
                class="text-[13px] font-bold uppercase tracking-[0.12em] text-jv-muted"
              >
                {{
                  selected.kind === "STATIC"
                    ? "Ordered member question IDs"
                    : "Dynamically resolved question IDs"
                }}
                · {{ (resolution.question_ids || []).length }}
              </p>
              <ol
                v-if="(resolution.question_ids || []).length > 0"
                class="grid gap-1"
              >
                <li
                  v-for="(questionId, index) in resolution.question_ids"
                  :key="`${questionId}-${index}`"
                  class="truncate text-[14px] text-jv-ink"
                >
                  {{ index + 1 }}. {{ titleFor(questionId) }}
                  <span class="text-[12px] text-jv-muted"
                    >({{ questionId }})</span
                  >
                </li>
              </ol>
              <p v-else class="text-[14px] text-jv-muted">
                No question IDs resolved yet.
              </p>
            </div>
            <p v-else class="text-[14px] text-jv-muted">
              Run resolve to preview the question set for this collection.
            </p>
          </section>

          <p
            v-if="actionError"
            class="mt-3 text-[14px] font-semibold text-jv-coral"
          >
            {{ actionError }}
          </p>
          <p
            v-if="actionSuccess"
            class="mt-3 text-[14px] font-semibold text-jv-accent-green"
          >
            {{ actionSuccess }}
          </p>
        </template>

        <p v-else class="text-[15px] text-jv-muted">
          Select a collection or create a new one.
        </p>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { ChevronDown, ChevronUp, Plus, Trash2 } from "lucide-vue-next";
import { useQuizCollectionsApi } from "@/composables/quiz_collections";
import {
  getQuizQuestionAPIError,
  useQuizQuestionsApi,
} from "@/composables/quiz_questions";
import {
  COLLECTION_KIND_DYNAMIC,
  COLLECTION_KIND_STATIC,
  describeResolutionStatus,
  emptyDynamicFilter,
  filterFromApi,
  filterToApiPayload,
  memberQuestionIds,
  questionTitleById,
} from "@/utils/question_collection";
import DynamicFilterFields from "@/components/quiz-manage/DynamicFilterFields.vue";
import QuestionFormCard from "@/components/quiz-manage/QuestionFormCard.vue";

const props = defineProps({
  quizId: {
    type: String,
    required: true,
  },
  questions: {
    type: Array,
    default: () => [],
  },
  canEdit: {
    type: Boolean,
    default: false,
  },
  defaultPoints: {
    type: Number,
    default: 10,
  },
  defaultDuration: {
    type: Number,
    default: 10,
  },
});

const emit = defineEmits(["question-created"]);

const {
  listCollections,
  getCollection,
  createCollection,
  updateCollection,
  deleteCollection,
  replaceMembers,
  resolveCollection,
  getQuizCollectionAPIError,
} = useQuizCollectionsApi();

const { createQuestion } = useQuizQuestionsApi();

const collections = ref([]);
const selectedId = ref("");
const selected = ref(null);
const loading = ref(false);
const pending = ref(false);
const loadError = ref("");
const actionError = ref("");
const actionSuccess = ref("");
const resolveError = ref("");
const resolution = ref(null);

const creating = ref(false);
const createKind = ref(COLLECTION_KIND_STATIC);
const createTitle = ref("");
const createPosition = ref(0);
const createFilter = ref(emptyDynamicFilter());

const editTitle = ref("");
const editPosition = ref(0);
const editFilter = ref(emptyDynamicFilter());
const memberIds = ref([]);
const savedMemberIds = ref([]);
const addQuestionId = ref("");

const showInlineQuestionForm = ref(false);
const savingInlineQuestion = ref(false);
const inlineQuestionError = ref("");
const linkToStaticCollection = ref(true);

const membershipDirty = computed(
  () => JSON.stringify(memberIds.value) !== JSON.stringify(savedMemberIds.value)
);

const availableQuestions = computed(() =>
  (props.questions || []).filter(
    (question) => !memberIds.value.includes(question.question_id)
  )
);

const titleFor = (questionId) => questionTitleById(props.questions, questionId);

const clearMessages = () => {
  actionError.value = "";
  actionSuccess.value = "";
  resolveError.value = "";
};

const applySelected = (collection) => {
  selected.value = collection;
  selectedId.value = collection?.id || "";
  editTitle.value = collection?.title || "";
  editPosition.value = Number(collection?.position || 0);
  editFilter.value = filterFromApi(collection?.filter);
  const ids = memberQuestionIds(collection);
  memberIds.value = [...ids];
  savedMemberIds.value = [...ids];
  addQuestionId.value = "";
  resolution.value = null;
};

const refreshList = async ({ preferId } = {}) => {
  loading.value = true;
  loadError.value = "";
  try {
    const rows = await listCollections(props.quizId);
    collections.value = Array.isArray(rows) ? rows : [];
    const nextId =
      preferId ||
      selectedId.value ||
      (collections.value[0] ? collections.value[0].id : "");
    if (nextId) {
      await selectCollection(nextId);
    } else {
      applySelected(null);
    }
  } catch (error) {
    loadError.value = getQuizCollectionAPIError(
      error,
      "Failed to load collections"
    );
  } finally {
    loading.value = false;
  }
};

const selectCollection = async (collectionId) => {
  clearMessages();
  creating.value = false;
  pending.value = true;
  try {
    const collection = await getCollection(props.quizId, collectionId);
    applySelected(collection);
    await runResolve({ silent: true });
  } catch (error) {
    actionError.value = getQuizCollectionAPIError(
      error,
      "Failed to load collection"
    );
  } finally {
    pending.value = false;
  }
};

const startCreate = (kind) => {
  clearMessages();
  creating.value = true;
  createKind.value = kind;
  createTitle.value = "";
  createPosition.value = collections.value.length;
  createFilter.value = emptyDynamicFilter();
  applySelected(null);
};

const cancelCreate = () => {
  creating.value = false;
  clearMessages();
};

const submitCreate = async () => {
  clearMessages();
  pending.value = true;
  try {
    const payload = {
      title: createTitle.value.trim(),
      kind: createKind.value,
      position: Number(createPosition.value) || 0,
    };
    if (createKind.value === COLLECTION_KIND_DYNAMIC) {
      payload.filter = filterToApiPayload(createFilter.value);
    }
    const created = await createCollection(props.quizId, payload);
    creating.value = false;
    await refreshList({ preferId: created.id });
    actionSuccess.value = "Collection created.";
  } catch (error) {
    actionError.value = getQuizCollectionAPIError(
      error,
      "Failed to create collection"
    );
  } finally {
    pending.value = false;
  }
};

const saveDetails = async () => {
  if (!selected.value) return;
  clearMessages();
  pending.value = true;
  try {
    const payload = {
      title: editTitle.value.trim(),
      position: Number(editPosition.value) || 0,
    };
    if (selected.value.kind === COLLECTION_KIND_DYNAMIC) {
      payload.filter = filterToApiPayload(editFilter.value);
    }
    const updated = await updateCollection(
      props.quizId,
      selected.value.id,
      payload
    );
    await refreshList({ preferId: updated.id });
    actionSuccess.value = "Collection details saved.";
  } catch (error) {
    actionError.value = getQuizCollectionAPIError(
      error,
      "Failed to update collection"
    );
  } finally {
    pending.value = false;
  }
};

const removeSelected = async () => {
  if (!selected.value) return;
  clearMessages();
  pending.value = true;
  try {
    await deleteCollection(props.quizId, selected.value.id);
    selectedId.value = "";
    await refreshList();
    actionSuccess.value = "Collection deleted.";
  } catch (error) {
    actionError.value = getQuizCollectionAPIError(
      error,
      "Failed to delete collection"
    );
  } finally {
    pending.value = false;
  }
};

const addMember = () => {
  if (!addQuestionId.value) return;
  memberIds.value = [...memberIds.value, addQuestionId.value];
  addQuestionId.value = "";
};

const removeMember = (questionId) => {
  memberIds.value = memberIds.value.filter((id) => id !== questionId);
};

const moveMember = (index, delta) => {
  const next = index + delta;
  if (next < 0 || next >= memberIds.value.length) return;
  const copy = [...memberIds.value];
  const [item] = copy.splice(index, 1);
  copy.splice(next, 0, item);
  memberIds.value = copy;
};

const saveMembership = async () => {
  if (!selected.value) return;
  clearMessages();
  pending.value = true;
  try {
    const updated = await replaceMembers(
      props.quizId,
      selected.value.id,
      memberIds.value
    );
    applySelected(updated);
    await runResolve({ silent: true });
    actionSuccess.value = "STATIC membership saved.";
  } catch (error) {
    actionError.value = getQuizCollectionAPIError(
      error,
      "Failed to save membership"
    );
  } finally {
    pending.value = false;
  }
};

const openInlineQuestionForm = () => {
  clearMessages();
  inlineQuestionError.value = "";
  showInlineQuestionForm.value = true;
  linkToStaticCollection.value =
    selected.value?.kind === COLLECTION_KIND_STATIC;
  creating.value = false;
};

const closeInlineQuestionForm = () => {
  showInlineQuestionForm.value = false;
  inlineQuestionError.value = "";
  savingInlineQuestion.value = false;
};

const normalizeCreatedQuestionId = (created) => {
  if (typeof created === "string" && created.trim()) {
    return created.trim();
  }
  if (created && typeof created === "object") {
    return String(
      created.question_id || created.id || created.QuestionId || ""
    ).trim();
  }
  return "";
};

const saveInlineQuestion = async ({ payload }) => {
  clearMessages();
  inlineQuestionError.value = "";
  savingInlineQuestion.value = true;
  try {
    const created = await createQuestion(props.quizId, {
      ...payload,
      points: Number(props.defaultPoints),
      duration_in_seconds: Number(props.defaultDuration),
    });
    const questionId = normalizeCreatedQuestionId(created);
    if (!questionId) {
      throw new Error("Question created but no question id was returned");
    }

    let linkedToCollection = false;
    if (
      selected.value?.kind === COLLECTION_KIND_STATIC &&
      linkToStaticCollection.value
    ) {
      const nextMembers = [...memberIds.value, questionId];
      const updated = await replaceMembers(
        props.quizId,
        selected.value.id,
        nextMembers
      );
      applySelected(updated);
      await runResolve({ silent: true });
      linkedToCollection = true;
    }

    showInlineQuestionForm.value = false;
    actionSuccess.value = linkedToCollection
      ? "Question added to Question Bank and linked to STATIC collection."
      : "Question added to Question Bank and linked to this quiz.";
    emit("question-created", {
      questionId,
      linkedToCollection,
      collectionId: linkedToCollection ? selected.value?.id : null,
    });
  } catch (error) {
    inlineQuestionError.value =
      getQuizQuestionAPIError(error, null) ||
      getQuizCollectionAPIError(error, "Failed to add question");
  } finally {
    savingInlineQuestion.value = false;
  }
};

const runResolve = async ({ silent = false } = {}) => {
  if (!selected.value) return;
  if (!silent) clearMessages();
  resolveError.value = "";
  try {
    resolution.value = await resolveCollection(props.quizId, selected.value.id);
  } catch (error) {
    resolution.value = null;
    resolveError.value = getQuizCollectionAPIError(
      error,
      "Failed to resolve collection"
    );
  }
};

watch(
  () => props.quizId,
  (next) => {
    if (next) refreshList();
  }
);

onMounted(() => {
  if (props.quizId) refreshList();
});

defineExpose({
  refreshList,
  selectCollection,
  collections,
  selected,
  memberIds,
  resolution,
  creating,
  createKind,
  showInlineQuestionForm,
  openInlineQuestionForm,
  saveInlineQuestion,
  linkToStaticCollection,
});
</script>
