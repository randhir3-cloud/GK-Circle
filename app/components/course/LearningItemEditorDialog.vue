<script setup>
import { computed, reactive, ref, watch } from "vue";
import { Modal } from "@/components/ui/modal";
import { isSafeContentUrl } from "@/utils/content_url";

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  mode: { type: String, default: "create" },
  item: { type: Object, default: null },
  quizzes: { type: Array, default: () => [] },
  quizzesLoading: { type: Boolean, default: false },
  saving: { type: Boolean, default: false },
  error: { type: String, default: "" },
});

const emit = defineEmits(["update:modelValue", "save"]);

const itemTypes = [
  { value: "ARTICLE", label: "Article" },
  { value: "VIDEO", label: "Video" },
  { value: "PDF", label: "PDF document" },
  { value: "LINK", label: "External link" },
  { value: "QUIZ_REFERENCE", label: "Test / Quiz" },
];
const publishStates = ["DRAFT", "PUBLISHED"];
const blockTypes = [
  { value: "TEXT", label: "Paragraph" },
  { value: "HEADING", label: "Heading" },
  { value: "VIDEO", label: "Video" },
  { value: "PDF", label: "PDF" },
  { value: "LINK", label: "Link" },
  { value: "IMAGE", label: "Image" },
  { value: "DIVIDER", label: "Divider" },
];

let blockSequence = 0;
const blockId = () => `content-${Date.now()}-${++blockSequence}`;

const defaultDataFor = (type) => {
  switch (type) {
    case "TEXT":
      return { text: "" };
    case "HEADING":
      return { text: "", level: 2 };
    case "VIDEO":
      return { url: "", title: "", caption: "" };
    case "PDF":
      return { url: "", title: "" };
    case "LINK":
      return { url: "", label: "" };
    case "IMAGE":
      return { url: "", alt: "", caption: "" };
    default:
      return {};
  }
};

const createBlock = (type) => ({
  id: blockId(),
  type,
  data: defaultDataFor(type),
  visibility: { mode: "ALL" },
});
const cloneJSON = (value) => JSON.parse(JSON.stringify(value));

const defaultBlockTypeForItem = (itemType) => {
  if (itemType === "VIDEO") return "VIDEO";
  if (itemType === "PDF") return "PDF";
  if (itemType === "LINK") return "LINK";
  return "TEXT";
};

const cloneBlocks = (metadata) => {
  if (
    metadata === null ||
    typeof metadata !== "object" ||
    !Array.isArray(metadata.blocks)
  ) {
    return [];
  }
  return metadata.blocks.map((block) => ({
    id: block.id,
    type: block.type,
    data:
      block.data && typeof block.data === "object" ? cloneJSON(block.data) : {},
    visibility:
      block.visibility && typeof block.visibility === "object"
        ? cloneJSON(block.visibility)
        : { mode: "ALL" },
  }));
};

const quizIdFromMetadata = (metadata) => {
  if (
    metadata === null ||
    typeof metadata !== "object" ||
    !Array.isArray(metadata.blocks)
  ) {
    return "";
  }
  return (
    metadata.blocks.find(
      (block) =>
        block?.type === "LINK" && typeof block?.data?.quiz_id === "string"
    )?.data?.quiz_id || ""
  );
};

const form = reactive({
  title: "",
  itemType: "ARTICLE",
  description: "",
  publishState: "DRAFT",
  quizId: "",
  blocks: [],
});
const selectedBlockType = ref("TEXT");
const localError = ref("");

const usesContentBlocks = computed(() => form.itemType !== "QUIZ_REFERENCE");
const blockIsComplete = (block) => {
  if (block.type === "DIVIDER") return true;
  if (block.type === "TEXT") return Boolean(block.data.text?.trim());
  if (block.type === "HEADING") {
    return (
      Boolean(block.data.text?.trim()) &&
      [2, 3, 4, 5, 6].includes(block.data.level)
    );
  }
  if (block.type === "VIDEO" || block.type === "PDF") {
    return (
      isSafeContentUrl(block.data.url) && Boolean(block.data.title?.trim())
    );
  }
  if (block.type === "LINK") {
    return (
      isSafeContentUrl(block.data.url) && Boolean(block.data.label?.trim())
    );
  }
  if (block.type === "IMAGE") {
    return (
      isSafeContentUrl(block.data.url) && typeof block.data.alt === "string"
    );
  }
  return false;
};
const contentIsComplete = computed(
  () =>
    !usesContentBlocks.value ||
    (form.blocks.length > 0 && form.blocks.every(blockIsComplete))
);
const submitDisabled = computed(
  () =>
    props.saving ||
    (form.itemType === "QUIZ_REFERENCE" && !form.quizId) ||
    !contentIsComplete.value
);

watch(
  () => [props.modelValue, props.mode, props.item],
  ([open]) => {
    if (!open) return;
    localError.value = "";
    form.title = props.item?.title || "";
    form.itemType = props.item?.item_type || "ARTICLE";
    form.description = props.item?.description || "";
    form.publishState = props.item?.publish_state || "DRAFT";
    form.quizId =
      props.item?.quiz_id || quizIdFromMetadata(props.item?.metadata);
    form.blocks = cloneBlocks(props.item?.metadata);
    if (form.itemType !== "QUIZ_REFERENCE" && form.blocks.length === 0) {
      form.blocks.push(createBlock(defaultBlockTypeForItem(form.itemType)));
    }
    selectedBlockType.value = defaultBlockTypeForItem(form.itemType);
  },
  { immediate: true }
);

watch(
  () => form.itemType,
  (itemType, previousType) => {
    if (!props.modelValue || itemType === previousType) return;
    form.quizId = "";
    if (
      itemType !== "QUIZ_REFERENCE" &&
      props.mode === "create" &&
      form.blocks.length === 1 &&
      !blockIsComplete(form.blocks[0])
    ) {
      const type = defaultBlockTypeForItem(itemType);
      form.blocks.splice(0, 1, createBlock(type));
      selectedBlockType.value = type;
    } else if (itemType !== "QUIZ_REFERENCE" && form.blocks.length === 0) {
      const type = defaultBlockTypeForItem(itemType);
      form.blocks.push(createBlock(type));
      selectedBlockType.value = type;
    }
  }
);

const addBlock = () => {
  form.blocks.push(createBlock(selectedBlockType.value));
};

const removeBlock = (index) => {
  form.blocks.splice(index, 1);
};

const moveBlock = (index, offset) => {
  const destination = index + offset;
  if (destination < 0 || destination >= form.blocks.length) return;
  const [block] = form.blocks.splice(index, 1);
  form.blocks.splice(destination, 0, block);
};

const cleanData = (block) => {
  if (block.type === "VIDEO" || block.type === "IMAGE") {
    const data = { ...block.data };
    if (!data.caption) delete data.caption;
    return data;
  }
  return { ...block.data };
};

const submit = () => {
  localError.value = "";
  if (form.itemType === "QUIZ_REFERENCE" && !form.quizId) {
    localError.value = "Select a Test before saving.";
    return;
  }
  if (!contentIsComplete.value) {
    localError.value =
      "Complete every content block with valid text and HTTP(S) or site-relative URLs.";
    return;
  }
  const selectedQuiz = props.quizzes.find((quiz) => quiz.id === form.quizId);
  const metadataBlocks =
    form.itemType === "QUIZ_REFERENCE"
      ? [
          {
            id:
              form.blocks.find((block) => block?.data?.quiz_id)?.id ||
              blockId(),
            type: "LINK",
            data: {
              url: `/admin/quiz/list-quiz/${form.quizId}`,
              label: selectedQuiz?.title || form.title,
              quiz_id: form.quizId,
            },
            visibility: { mode: "ALL" },
          },
        ]
      : form.blocks.map((block) => ({
          id: block.id,
          type: block.type,
          data: cleanData(block),
          visibility: block.visibility || { mode: "ALL" },
        }));
  const payload = {
    title: form.title,
    item_type: form.itemType,
    publish_state: form.publishState,
    metadata: {
      version: 1,
      blocks: metadataBlocks,
    },
    quiz_id: form.itemType === "QUIZ_REFERENCE" ? form.quizId : null,
  };
  if (props.mode === "update") {
    payload.description = form.description || null;
  } else if (form.description) {
    payload.description = form.description;
  }
  emit("save", payload);
};
</script>

<template>
  <Modal
    :model-value="modelValue"
    :title="mode === 'create' ? 'New Learning Item' : 'Edit Learning Item'"
    description="Add the content learners will read, watch, download, or attempt."
    size="lg"
    :close-on-backdrop="!saving"
    :hide-close="saving"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <form
      class="grid max-h-[75vh] gap-4 overflow-y-auto pr-1"
      data-testid="learning-item-editor"
      @submit.prevent="submit"
    >
      <label class="grid gap-1 text-sm font-black">
        Title
        <input
          v-model="form.title"
          name="title"
          required
          class="h-11 rounded-[8px] border-[2px] border-jv-ink px-3"
        />
      </label>

      <div class="grid gap-4 sm:grid-cols-2">
        <label class="grid gap-1 text-sm font-black">
          Learning Item type
          <select
            v-model="form.itemType"
            name="item_type"
            class="h-11 rounded-[8px] border-[2px] border-jv-ink bg-white px-3"
          >
            <option
              v-for="type in itemTypes"
              :key="type.value"
              :value="type.value"
            >
              {{ type.label }}
            </option>
          </select>
        </label>
        <label class="grid gap-1 text-sm font-black">
          Publish state
          <select
            v-model="form.publishState"
            name="publish_state"
            class="h-11 rounded-[8px] border-[2px] border-jv-ink bg-white px-3"
          >
            <option v-for="state in publishStates" :key="state" :value="state">
              {{ state }}
            </option>
          </select>
        </label>
      </div>

      <label class="grid gap-1 text-sm font-black">
        Description
        <textarea
          v-model="form.description"
          name="description"
          rows="3"
          class="rounded-[8px] border-[2px] border-jv-ink px-3 py-2"
        ></textarea>
      </label>

      <section
        v-if="form.itemType === 'QUIZ_REFERENCE'"
        class="grid gap-3 rounded-[10px] border-[2px] border-jv-ink bg-jv-yellow-soft p-4"
        data-testid="quiz-reference-editor"
      >
        <div>
          <h3 class="font-headings text-xl">Attach a Test</h3>
          <p class="text-sm font-bold text-jv-muted">
            Select one of your existing tests, or create one in the Test
            Builder.
          </p>
        </div>
        <label class="grid gap-1 text-sm font-black">
          Test / Quiz
          <select
            v-model="form.quizId"
            name="quiz_id"
            required
            class="h-11 rounded-[8px] border-[2px] border-jv-ink bg-white px-3"
            :disabled="quizzesLoading"
          >
            <option value="">
              {{ quizzesLoading ? "Loading tests…" : "Select a Test" }}
            </option>
            <option v-for="quiz in quizzes" :key="quiz.id" :value="quiz.id">
              {{ quiz.title }}
            </option>
          </select>
        </label>
        <NuxtLink
          to="/admin/quiz/list-quiz?create=1"
          target="_blank"
          class="w-fit rounded-full border-[2px] border-jv-ink bg-jv-white px-4 py-2 text-sm font-black shadow-brutal-sm"
        >
          Create a new Test
        </NuxtLink>
      </section>

      <section
        v-else
        class="grid gap-3 rounded-[10px] border-[2px] border-jv-ink bg-jv-canvas p-4"
        data-testid="content-block-editor"
      >
        <div>
          <h3 class="font-headings text-xl">Content</h3>
          <p class="text-sm font-bold text-jv-muted">
            Add and order the blocks learners will see.
          </p>
        </div>

        <article
          v-for="(block, index) in form.blocks"
          :key="block.id"
          class="grid gap-3 rounded-[8px] border-[2px] border-jv-ink bg-white p-3"
          :data-testid="`content-block-${index}`"
        >
          <div class="flex flex-wrap items-center justify-between gap-2">
            <strong>{{ index + 1 }}. {{ block.type }}</strong>
            <div class="flex gap-2">
              <button
                type="button"
                class="rounded border border-jv-ink px-2 py-1 text-xs font-black disabled:opacity-40"
                :disabled="index === 0"
                :aria-label="`Move block ${index + 1} up`"
                @click="moveBlock(index, -1)"
              >
                ↑
              </button>
              <button
                type="button"
                class="rounded border border-jv-ink px-2 py-1 text-xs font-black disabled:opacity-40"
                :disabled="index === form.blocks.length - 1"
                :aria-label="`Move block ${index + 1} down`"
                @click="moveBlock(index, 1)"
              >
                ↓
              </button>
              <button
                type="button"
                class="rounded border border-red-700 px-2 py-1 text-xs font-black text-red-700"
                :aria-label="`Remove block ${index + 1}`"
                @click="removeBlock(index)"
              >
                Remove
              </button>
            </div>
          </div>

          <label
            v-if="block.type === 'TEXT'"
            class="grid gap-1 text-sm font-black"
          >
            Article text
            <textarea
              v-model="block.data.text"
              :name="`block_${index}_text`"
              rows="7"
              required
              class="rounded-[8px] border-[2px] border-jv-ink px-3 py-2 font-normal"
            ></textarea>
          </label>

          <template v-else-if="block.type === 'HEADING'">
            <label class="grid gap-1 text-sm font-black">
              Heading
              <input
                v-model="block.data.text"
                :name="`block_${index}_heading`"
                required
                class="h-11 rounded-[8px] border-[2px] border-jv-ink px-3 font-normal"
              />
            </label>
            <label class="grid gap-1 text-sm font-black">
              Heading level
              <select
                v-model.number="block.data.level"
                :name="`block_${index}_level`"
                class="h-11 rounded-[8px] border-[2px] border-jv-ink bg-white px-3"
              >
                <option v-for="level in [2, 3, 4, 5, 6]" :key="level">
                  {{ level }}
                </option>
              </select>
            </label>
          </template>

          <template v-else-if="block.type === 'VIDEO'">
            <label class="grid gap-1 text-sm font-black">
              Video embed URL
              <input
                v-model="block.data.url"
                :name="`block_${index}_url`"
                required
                placeholder="https://www.youtube.com/embed/…"
                class="h-11 rounded-[8px] border-[2px] border-jv-ink px-3 font-normal"
              />
            </label>
            <label class="grid gap-1 text-sm font-black">
              Accessible video title
              <input
                v-model="block.data.title"
                :name="`block_${index}_title`"
                required
                class="h-11 rounded-[8px] border-[2px] border-jv-ink px-3 font-normal"
              />
            </label>
            <label class="grid gap-1 text-sm font-black">
              Caption (optional)
              <input
                v-model="block.data.caption"
                :name="`block_${index}_caption`"
                class="h-11 rounded-[8px] border-[2px] border-jv-ink px-3 font-normal"
              />
            </label>
          </template>

          <template v-else-if="block.type === 'PDF'">
            <label class="grid gap-1 text-sm font-black">
              PDF URL
              <input
                v-model="block.data.url"
                :name="`block_${index}_url`"
                required
                placeholder="https://…/document.pdf"
                class="h-11 rounded-[8px] border-[2px] border-jv-ink px-3 font-normal"
              />
            </label>
            <label class="grid gap-1 text-sm font-black">
              Document title
              <input
                v-model="block.data.title"
                :name="`block_${index}_title`"
                required
                class="h-11 rounded-[8px] border-[2px] border-jv-ink px-3 font-normal"
              />
            </label>
          </template>

          <template v-else-if="block.type === 'LINK'">
            <label class="grid gap-1 text-sm font-black">
              Link URL
              <input
                v-model="block.data.url"
                :name="`block_${index}_url`"
                required
                placeholder="https://…"
                class="h-11 rounded-[8px] border-[2px] border-jv-ink px-3 font-normal"
              />
            </label>
            <label class="grid gap-1 text-sm font-black">
              Link label
              <input
                v-model="block.data.label"
                :name="`block_${index}_label`"
                required
                class="h-11 rounded-[8px] border-[2px] border-jv-ink px-3 font-normal"
              />
            </label>
          </template>

          <template v-else-if="block.type === 'IMAGE'">
            <label class="grid gap-1 text-sm font-black">
              Image URL
              <input
                v-model="block.data.url"
                :name="`block_${index}_url`"
                required
                class="h-11 rounded-[8px] border-[2px] border-jv-ink px-3 font-normal"
              />
            </label>
            <label class="grid gap-1 text-sm font-black">
              Alternative text
              <input
                v-model="block.data.alt"
                :name="`block_${index}_alt`"
                required
                class="h-11 rounded-[8px] border-[2px] border-jv-ink px-3 font-normal"
              />
            </label>
            <label class="grid gap-1 text-sm font-black">
              Caption (optional)
              <input
                v-model="block.data.caption"
                :name="`block_${index}_caption`"
                class="h-11 rounded-[8px] border-[2px] border-jv-ink px-3 font-normal"
              />
            </label>
          </template>

          <p
            v-else-if="block.type === 'DIVIDER'"
            class="border-t-2 border-dashed border-jv-ink py-2 text-center text-sm font-bold text-jv-muted"
          >
            Section divider
          </p>
        </article>

        <div class="flex flex-col gap-2 sm:flex-row">
          <select
            v-model="selectedBlockType"
            data-testid="new-block-type"
            class="h-11 flex-1 rounded-[8px] border-[2px] border-jv-ink bg-white px-3 font-bold"
          >
            <option
              v-for="type in blockTypes"
              :key="type.value"
              :value="type.value"
            >
              {{ type.label }}
            </option>
          </select>
          <button
            type="button"
            data-testid="add-content-block"
            class="h-11 rounded-[8px] border-[2px] border-jv-ink bg-jv-yellow px-4 font-black"
            @click="addBlock"
          >
            Add content block
          </button>
        </div>
      </section>

      <p
        v-if="localError || error"
        class="text-sm font-bold text-red-700"
        role="alert"
      >
        {{ localError || error }}
      </p>
      <div class="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
        <button
          type="button"
          class="h-11 rounded-full border-[2px] border-jv-ink bg-white px-5 font-black"
          :disabled="saving"
          @click="emit('update:modelValue', false)"
        >
          Cancel
        </button>
        <button
          type="submit"
          class="h-11 rounded-full border-[2px] border-jv-ink bg-jv-coral px-5 font-black text-white shadow-brutal-sm disabled:opacity-60"
          :disabled="submitDisabled"
        >
          {{
            saving ? "Saving…" : mode === "create" ? "Create" : "Save changes"
          }}
        </button>
      </div>
    </form>
  </Modal>
</template>
