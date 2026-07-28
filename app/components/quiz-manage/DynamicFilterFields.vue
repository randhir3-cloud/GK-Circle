<template>
  <div class="grid gap-3 sm:grid-cols-2">
    <label class="grid gap-1.5">
      <span
        class="text-[12px] font-bold uppercase tracking-[0.12em] text-jv-muted"
      >
        Subject
      </span>
      <input
        :value="modelValue.subject"
        type="text"
        class="h-11 border-[2px] border-jv-ink bg-jv-canvas px-3 text-[14px] font-semibold text-jv-ink outline-none"
        @input="update('subject', $event.target.value)"
      />
    </label>
    <label class="grid gap-1.5">
      <span
        class="text-[12px] font-bold uppercase tracking-[0.12em] text-jv-muted"
      >
        Topic
      </span>
      <input
        :value="modelValue.topic"
        type="text"
        class="h-11 border-[2px] border-jv-ink bg-jv-canvas px-3 text-[14px] font-semibold text-jv-ink outline-none"
        @input="update('topic', $event.target.value)"
      />
    </label>
    <label class="grid gap-1.5">
      <span
        class="text-[12px] font-bold uppercase tracking-[0.12em] text-jv-muted"
      >
        Year
      </span>
      <input
        :value="modelValue.year"
        type="number"
        min="1900"
        max="2100"
        class="h-11 border-[2px] border-jv-ink bg-jv-canvas px-3 text-[14px] font-semibold text-jv-ink outline-none"
        @input="update('year', $event.target.value)"
      />
    </label>
    <label class="grid gap-1.5">
      <span
        class="text-[12px] font-bold uppercase tracking-[0.12em] text-jv-muted"
      >
        Difficulty
      </span>
      <input
        :value="modelValue.difficulty"
        type="text"
        class="h-11 border-[2px] border-jv-ink bg-jv-canvas px-3 text-[14px] font-semibold text-jv-ink outline-none"
        @input="update('difficulty', $event.target.value)"
      />
    </label>
    <label class="grid gap-1.5 sm:col-span-2">
      <span
        class="text-[12px] font-bold uppercase tracking-[0.12em] text-jv-muted"
      >
        PYQ status
      </span>
      <select
        :value="modelValue.pyq_status"
        class="h-11 border-[2px] border-jv-ink bg-jv-canvas px-3 text-[14px] font-semibold text-jv-ink outline-none"
        @change="update('pyq_status', $event.target.value)"
      >
        <option value="">Any / unset</option>
        <option value="true">PYQ only</option>
        <option value="false">Non-PYQ only</option>
      </select>
    </label>
  </div>
</template>

<script setup>
const props = defineProps({
  modelValue: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(["update:modelValue"]);

const update = (key, value) => {
  emit("update:modelValue", {
    ...props.modelValue,
    [key]: value,
  });
};
</script>
