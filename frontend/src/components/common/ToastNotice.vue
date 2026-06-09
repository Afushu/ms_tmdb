<script setup lang="ts">
import type { ToastNoticeTone } from "@/composables/useToastNotice";

withDefaults(
  defineProps<{
    visible: boolean;
    message: string;
    tone?: ToastNoticeTone;
  }>(),
  {
    tone: "success",
  },
);

const emit = defineEmits<{
  close: [];
}>();
</script>

<template>
  <div v-if="visible" class="toast-notice" :class="`toast-notice-${tone}`" role="status" aria-live="polite">
    <span class="toast-notice-icon" aria-hidden="true"></span>
    <span class="toast-notice-text">{{ message }}</span>
    <button class="toast-notice-close" type="button" aria-label="关闭提示" @click="emit('close')"></button>
  </div>
</template>
