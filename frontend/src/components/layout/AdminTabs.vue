<script setup lang="ts">
import type { AdminTab } from "./adminLayoutConfig";

defineProps<{
  currentFullPath: string;
  tabs: AdminTab[];
}>();

const emit = defineEmits<{
  close: [tab: AdminTab, event: MouseEvent];
}>();
</script>

<template>
  <div class="admin-tabs" aria-label="已打开页面">
    <div
      v-for="tab in tabs"
      :key="tab.fullPath"
      class="admin-tab"
      :class="{ 'admin-tab-active': tab.fullPath === currentFullPath }"
    >
      <RouterLink :to="tab.fullPath" class="admin-tab-link">{{ tab.title }}</RouterLink>
      <button
        v-if="tab.path !== '/'"
        type="button"
        class="admin-tab-close"
        aria-label="关闭标签"
        @click="emit('close', tab, $event)"
      ></button>
    </div>
  </div>
</template>
