<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  currentSection: string;
  currentTitle: string;
  searchQuery: string;
  sidebarOpen: boolean;
}>();

const emit = defineEmits<{
  openPreferences: [];
  submitSearch: [];
  toggleSidebar: [];
  "update:searchQuery": [value: string];
}>();

const searchText = computed({
  get: () => props.searchQuery,
  set: (value: string) => emit("update:searchQuery", value),
});
</script>

<template>
  <header class="admin-topbar">
    <button
      class="admin-icon-btn"
      type="button"
      :aria-label="sidebarOpen ? '关闭导航' : '打开导航'"
      @click="emit('toggleSidebar')"
    >
      <span class="admin-icon-bars"></span>
    </button>

    <div class="admin-title-block">
      <p class="admin-breadcrumb">
        <span>首页</span>
        <span>/</span>
        <span>{{ currentSection }}</span>
      </p>
      <h1 class="admin-page-title">{{ currentTitle }}</h1>
    </div>

    <div class="admin-top-actions">
      <form class="admin-top-search" role="search" @submit.prevent="emit('submitSearch')">
        <input
          v-model="searchText"
          class="admin-top-search-input"
          type="search"
          placeholder="搜索电影、剧集、人物"
        />
        <button class="admin-top-search-btn" type="submit" aria-label="搜索">
          <span class="admin-icon-search"></span>
        </button>
      </form>
      <button class="admin-preference-btn" type="button" aria-label="偏好设置" @click="emit('openPreferences')">
        <span class="admin-icon-gear"></span>
      </button>
    </div>
  </header>
</template>
