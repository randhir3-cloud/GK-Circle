<template>
  <div
    :data-kratos-node-index="index"
    :data-kratos-node-group="node.group || undefined"
    :data-kratos-node-type="node.type || undefined"
    :data-kratos-node-input-type="attributes.type || undefined"
    :data-kratos-node-name="attributes.name || undefined"
  >
    <input
      v-if="isHiddenInput"
      :name="attributes.name"
      type="hidden"
      :value="modelValue"
      :disabled="attributes.disabled"
    />

    <button
      v-else-if="isSubmitInput"
      type="submit"
      :name="attributes.name"
      :value="attributes.value"
      :disabled="attributes.disabled || loading"
      class="jv-card mt-2 inline-flex h-12 w-full items-center justify-center gap-2 border-2 border-jv-ink bg-jv-coral font-headings text-base text-white shadow-brutal-sm transition-transform hover:rotate-[1deg] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none disabled:cursor-not-allowed disabled:opacity-70 sm:text-lg"
      @click="$emit('submit-node', node)"
    >
      {{ label }}
      <ArrowRight class="size-5" :stroke-width="2.4" />
    </button>

    <div v-else-if="isEditableInput" class="flex flex-col gap-1.5">
      <label
        :for="inputId"
        class="px-0.5 font-body text-xs font-bold uppercase tracking-wide text-jv-ink sm:text-[13px]"
      >
        {{ label }}
      </label>
      <div
        class="jv-card flex items-center gap-2.5 border-2 border-jv-ink bg-jv-white px-3 py-2.5 shadow-brutal-sm transition-all focus-within:translate-x-[1px] focus-within:translate-y-[1px] focus-within:shadow-none"
      >
        <Mail
          v-if="attributes.name === 'email'"
          class="size-[18px] shrink-0 text-jv-ink/70"
          :stroke-width="2.2"
        />
        <ShieldCheck
          v-else-if="attributes.name === 'code'"
          class="size-[18px] shrink-0 text-jv-ink/70"
          :stroke-width="2.2"
        />
        <input
          :id="inputId"
          :name="attributes.name"
          :type="attributes.type || 'text'"
          :value="modelValue"
          :required="attributes.required"
          :disabled="attributes.disabled"
          :autocomplete="attributes.autocomplete"
          :pattern="attributes.pattern"
          :placeholder="attributes.placeholder"
          :readonly="attributes.readonly"
          :inputmode="attributes.inputmode"
          :min="attributes.min"
          :max="attributes.max"
          :minlength="attributes.minlength"
          :maxlength="attributes.maxlength"
          class="min-w-0 flex-1 border-0 bg-transparent font-body text-sm text-jv-ink outline-none placeholder:text-jv-ink/40 sm:text-base"
          :class="{ 'tracking-[0.2em]': attributes.name === 'code' }"
          @input="$emit('update-value', attributes.name, $event.target.value)"
        />
      </div>
    </div>

    <p
      v-else-if="node.type === 'text'"
      class="m-0 font-body text-sm text-jv-ink/80"
    >
      {{ attributes.text?.text || attributes.text || label }}
    </p>

    <a
      v-else-if="node.type === 'a' && safeUrl"
      :href="safeUrl"
      :title="attributes.title"
      class="font-body text-sm text-jv-coral underline underline-offset-4"
      rel="noopener noreferrer"
    >
      {{ label }}
    </a>

    <img
      v-else-if="node.type === 'img' && safeUrl"
      :src="safeUrl"
      :alt="label"
      :title="attributes.title"
      :width="attributes.width"
      :height="attributes.height"
      class="max-w-full"
    />

    <div
      v-else
      class="font-body text-xs text-jv-ink/60"
      data-kratos-node-unsupported="true"
    >
      {{ label }}
    </div>

    <p
      v-for="(message, messageIndex) in node.messages || []"
      :key="message.id || messageIndex"
      class="m-0 flex items-center gap-1 px-0.5 font-body text-xs text-jv-accent-red"
    >
      <AlertCircle class="size-3.5 shrink-0" :stroke-width="2.2" />
      {{ message.text }}
    </p>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { AlertCircle, ArrowRight, Mail, ShieldCheck } from "lucide-vue-next";
import { safeNodeUrl } from "~/utils/verificationFlow";

const props = defineProps({
  node: { type: Object, required: true },
  index: { type: Number, required: true },
  modelValue: { type: [String, Number, Boolean], default: "" },
  loading: { type: Boolean, default: false },
  currentOrigin: { type: String, required: true },
});

defineEmits(["update-value", "submit-node"]);

const attributes = computed(() => props.node.attributes || {});
const inputId = computed(
  () => `verification-${attributes.value.name || "node"}-${props.index}`
);
const label = computed(
  () =>
    props.node.meta?.label?.text ||
    attributes.value.label?.text ||
    attributes.value.name ||
    "Verification step"
);
const isInput = computed(() => props.node.type === "input");
const isHiddenInput = computed(
  () => isInput.value && attributes.value.type === "hidden"
);
const isSubmitInput = computed(
  () => isInput.value && attributes.value.type === "submit"
);
const isEditableInput = computed(
  () => isInput.value && !isHiddenInput.value && !isSubmitInput.value
);
const safeUrl = computed(() =>
  safeNodeUrl(
    attributes.value.href || attributes.value.src,
    props.currentOrigin
  )
);
</script>
