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
  reloadPage: [];
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
    <div class="admin-topbar-left">
      <button
        class="admin-icon-btn"
        type="button"
        :aria-label="sidebarOpen ? '关闭导航' : '打开导航'"
        @click="emit('toggleSidebar')"
      >
        <span class="admin-icon-bars"></span>
      </button>

      <button class="admin-icon-btn" type="button" aria-label="刷新页面" @click="emit('reloadPage')">
        <svg class="admin-icon-svg" viewBox="0 0 24 24" aria-hidden="true" focusable="false">
          <path d="M21 12a9 9 0 0 1-15 6.7L3 16" />
          <path d="M3 22v-6h6" />
          <path d="M3 12a9 9 0 0 1 15-6.7L21 8" />
          <path d="M21 2v6h-6" />
        </svg>
      </button>

      <nav class="admin-breadcrumb" aria-label="面包屑">
        <span class="admin-breadcrumb-section">{{ currentSection }}</span>
        <span class="admin-breadcrumb-separator">›</span>
        <span class="admin-breadcrumb-current">{{ currentTitle }}</span>
      </nav>
    </div>

    <div class="admin-top-actions">
      <form class="admin-top-search" role="search" @submit.prevent="emit('submitSearch')">
        <span class="admin-icon-search" aria-hidden="true"></span>
        <input
          v-model="searchText"
          class="admin-top-search-input"
          type="search"
          placeholder="搜索"
        />
      </form>
      <button class="admin-preference-btn" type="button" aria-label="偏好设置" @click="emit('openPreferences')">
        <svg class="admin-preference-icon" viewBox="0 0 24 24" aria-hidden="true" focusable="false">
          <path d="M4 7h5M15 7h5M4 12h10M18 12h2M4 17h3M13 17h7" />
          <circle cx="12" cy="7" r="2.25" />
          <circle cx="16" cy="12" r="2.25" />
          <circle cx="10" cy="17" r="2.25" />
        </svg>
      </button>
    </div>
  </header>
</template>
