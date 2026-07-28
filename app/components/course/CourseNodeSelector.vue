<script setup>
defineProps({
  levels: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  error: { type: String, default: "" },
});

const emit = defineEmits(["select"]);
</script>

<template>
  <section class="jv-border-uneven bg-jv-white p-4 shadow-brutal-sm sm:p-5">
    <h2 class="font-headings text-lg text-jv-ink sm:text-xl">CourseNode</h2>
    <p class="mt-1 text-sm font-bold text-jv-muted">
      Drill down one server-provided child level at a time.
    </p>

    <div v-if="levels.length" class="mt-3 grid gap-3 md:grid-cols-2">
      <label
        v-for="(level, index) in levels"
        :key="`${level.parentId || 'root'}-${index}`"
        class="block"
      >
        <span class="text-xs font-black uppercase tracking-wide text-jv-muted">
          {{ index === 0 ? "Root node" : `Child level ${index}` }}
        </span>
        <select
          :data-testid="`node-selector-${index}`"
          class="mt-1 h-11 w-full rounded-[8px] border-[2px] border-jv-ink bg-jv-white px-3 font-bold text-jv-ink disabled:opacity-60"
          :value="level.selectedId"
          :disabled="loading || level.loading"
          @change="
            emit('select', {
              levelIndex: index,
              nodeId: $event.target.value,
            })
          "
        >
          <option value="">
            {{ level.loading ? "Loading nodes…" : "Select a CourseNode" }}
          </option>
          <option v-for="node in level.nodes" :key="node.id" :value="node.id">
            {{ node.title }} · {{ node.node_type }}
          </option>
        </select>
      </label>
    </div>

    <p
      v-else-if="!loading"
      class="mt-3 text-sm font-bold text-jv-muted"
      data-testid="node-prompt"
    >
      Select a Course to load its root CourseNodes.
    </p>
    <p v-if="error" class="mt-2 text-sm font-bold text-red-700" role="alert">
      {{ error }}
    </p>
  </section>
</template>
