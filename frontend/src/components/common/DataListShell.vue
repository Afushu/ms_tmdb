<script setup lang="ts">

withDefaults(
  defineProps<{
    columns: string[];
    gridClass: string;
    shellClass?: string;
    loading?: boolean;
    empty?: boolean;
    emptyText?: string;
    loadingText?: string;
  }>(),
  {
    shellClass: "",
    loading: false,
    empty: false,
    emptyText: "暂无数据",
    loadingText: "加载中...",
  },
);
</script>

<template>
  <div class="logs-list-shell" :class="shellClass">
    <div class="logs-list-head" :class="gridClass">
      <span v-for="column in columns" :key="column">{{ column }}</span>
    </div>

    <slot />

    <p v-if="loading" class="logs-empty">{{ loadingText }}</p>
    <p v-else-if="empty" class="logs-empty">{{ emptyText }}</p>
  </div>
</template>
