<script setup>
defineProps({
  courses: { type: Array, default: () => [] },
  modelValue: { type: String, default: "" },
  loading: { type: Boolean, default: false },
  error: { type: String, default: "" },
});

const emit = defineEmits(["update:modelValue"]);
</script>

<template>
  <section class="jv-border-uneven bg-jv-white p-4 shadow-brutal-sm sm:p-5">
    <label
      for="course-selector"
      class="font-headings text-lg text-jv-ink sm:text-xl"
    >
      Course
    </label>
    <p class="mt-1 text-sm font-bold text-jv-muted">
      Choose the Course that owns the content.
    </p>
    <select
      id="course-selector"
      data-testid="course-selector"
      class="mt-3 h-11 w-full rounded-[8px] border-[2px] border-jv-ink bg-jv-white px-3 font-bold text-jv-ink disabled:opacity-60"
      :value="modelValue"
      :disabled="loading"
      @change="emit('update:modelValue', $event.target.value)"
    >
      <option value="">
        {{ loading ? "Loading Courses…" : "Select a Course" }}
      </option>
      <option v-for="course in courses" :key="course.id" :value="course.id">
        {{ course.title }}
      </option>
    </select>
    <p v-if="error" class="mt-2 text-sm font-bold text-red-700" role="alert">
      {{ error }}
    </p>
  </section>
</template>
