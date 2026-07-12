<script setup lang="ts">
type SelectOption = {
  label: string;
  value: string;
};

const props = withDefaults(
  defineProps<{
    modelValue: string;
    options: ReadonlyArray<SelectOption>;
    disabled?: boolean;
  }>(),
  {
    disabled: false,
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
  change: [value: string];
}>();

function onChange(event: Event) {
  const value = (event.target as HTMLSelectElement).value;
  if (props.modelValue === value) return;
  emit("update:modelValue", value);
  emit("change", value);
}
</script>

<template>
  <select
    class="field-control glass-select w-full text-sm"
    :class="disabled ? 'cursor-not-allowed opacity-80' : 'cursor-pointer'"
    :value="modelValue"
    :disabled="disabled"
    @change="onChange"
  >
    <option v-for="option in options" :key="option.value" :value="option.value">
      {{ option.label }}
    </option>
  </select>
</template>

<style scoped>
.glass-select {
  background: var(--field-bg) !important;
  border-color: var(--field-border) !important;
  color: var(--text-main) !important;
}

.glass-select:disabled {
  cursor: not-allowed;
  opacity: 0.8;
}
</style>
