<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from "vue";
import type { AdminTab } from "./adminLayoutConfig";

defineProps<{
  currentFullPath: string;
  tabs: AdminTab[];
}>();

const emit = defineEmits<{
  close: [tab: AdminTab, event: MouseEvent];
  closeAll: [];
  closeLeft: [tab: AdminTab];
  closeOther: [tab: AdminTab];
  closeRight: [tab: AdminTab];
}>();

const menuVisible = ref(false);
const menuX = ref(0);
const menuY = ref(0);
const menuTab = ref<AdminTab | null>(null);
const menuRef = ref<HTMLElement | null>(null);

function openMenu(tab: AdminTab, event: MouseEvent) {
  event.preventDefault();
  event.stopPropagation();
  menuTab.value = tab;
  menuVisible.value = true;
  nextTick(() => {
    const menu = menuRef.value;
    if (!menu) return;
    const { innerWidth, innerHeight } = window;
    const rect = menu.getBoundingClientRect();
    menuX.value = Math.min(event.clientX, innerWidth - rect.width - 4);
    menuY.value = Math.min(event.clientY, innerHeight - rect.height - 4);
  });
}

function closeMenu() {
  menuVisible.value = false;
  menuTab.value = null;
}

function handleMenuAction(action: "close" | "closeAll" | "closeLeft" | "closeOther" | "closeRight") {
  const tab = menuTab.value;
  closeMenu();
  if (!tab) return;

  switch (action) {
    case "close":
      emit("close", tab, new MouseEvent("click"));
      break;
    case "closeOther":
      emit("closeOther", tab);
      break;
    case "closeLeft":
      emit("closeLeft", tab);
      break;
    case "closeRight":
      emit("closeRight", tab);
      break;
    case "closeAll":
      emit("closeAll");
      break;
  }
}

function onDocumentClick() {
  if (menuVisible.value) closeMenu();
}

function onDocumentContext(e: MouseEvent) {
  if (menuVisible.value && !menuRef.value?.contains(e.target as Node)) {
    closeMenu();
  }
}

onMounted(() => {
  document.addEventListener("click", onDocumentClick);
  document.addEventListener("contextmenu", onDocumentContext);
});

onBeforeUnmount(() => {
  document.removeEventListener("click", onDocumentClick);
  document.removeEventListener("contextmenu", onDocumentContext);
});
</script>

<template>
  <div class="admin-tabs" aria-label="已打开页面">
    <div
      v-for="tab in tabs"
      :key="tab.fullPath"
      class="admin-tab"
      :class="{ 'admin-tab-active': tab.fullPath === currentFullPath }"
      @contextmenu="openMenu(tab, $event)"
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

  <div
    v-if="menuVisible"
    ref="menuRef"
    class="tab-context-menu"
    :style="{ left: menuX + 'px', top: menuY + 'px' }"
  >
    <button v-if="menuTab?.path !== '/'" class="tab-context-item" @click="handleMenuAction('close')">
      关闭当前
    </button>
    <button class="tab-context-item" @click="handleMenuAction('closeOther')">关闭其他</button>
    <button class="tab-context-item" @click="handleMenuAction('closeLeft')">关闭左侧</button>
    <button class="tab-context-item" @click="handleMenuAction('closeRight')">关闭右侧</button>
    <div class="tab-context-divider"></div>
    <button class="tab-context-item" @click="handleMenuAction('closeAll')">关闭全部</button>
  </div>
</template>
