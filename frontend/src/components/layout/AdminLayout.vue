<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import AdminPreferencesDrawer from "@/components/layout/AdminPreferencesDrawer.vue";
import AdminSidebar from "@/components/layout/AdminSidebar.vue";
import AdminTabs from "@/components/layout/AdminTabs.vue";
import AdminTopbar from "@/components/layout/AdminTopbar.vue";
import VbenPage from "@/components/layout/VbenPage.vue";
import { buildSearchQuery, readQueryString } from "@/utils/routeSearch";
import { sidebarOptions, themeOptions, type AdminMenuGroup, type AdminMenuItem } from "./adminLayoutConfig";
import { useAdminPreferences } from "./useAdminPreferences";
import { useAdminTabs } from "./useAdminTabs";

const route = useRoute();
const router = useRouter();
const sidebarOpen = ref(false);
const preferencesOpen = ref(false);
const showBackToTop = ref(false);
const topbarSearchQuery = ref("");

const { adminThemeStyle, currentSidebarOption, currentThemeOption, preferences, resetPreferences, setPreference } =
  useAdminPreferences();

const menuItems = computed<AdminMenuItem[]>(() =>
  router
    .getRoutes()
    .filter((item) => item.meta.title && !item.meta.hideMenu && !item.redirect)
    .map((item) => ({
      order: Number(item.meta.order ?? 100),
      path: item.path,
      section: String(item.meta.section ?? "系统"),
      title: String(item.meta.menuTitle ?? item.meta.title),
    }))
    .sort((left, right) => left.order - right.order),
);

const menuGroups = computed<AdminMenuGroup[]>(() => {
  const groups = new Map<string, AdminMenuItem[]>();
  for (const item of menuItems.value) {
    const groupItems = groups.get(item.section) ?? [];
    groupItems.push(item);
    groups.set(item.section, groupItems);
  }
  return Array.from(groups.entries()).map(([section, items]) => ({ section, items }));
});

const activeMenuPath = computed(() => String(route.meta.activeMenu ?? route.path));
const currentMenu = computed(() => menuItems.value.find((item) => item.path === activeMenuPath.value));
const hasRouteTitle = computed(() => typeof route.meta.title === "string" && route.meta.title.length > 0);
const currentSection = computed(() =>
  currentMenu.value?.section ?? String(route.meta.section ?? (route.path === "/" ? "工作台" : "系统")),
);
const currentTitle = computed(() => String(route.meta.title ?? currentMenu.value?.title ?? "首页"));
const currentDescription = computed(() => String(route.meta.description ?? ""));

const { closeTab, openedTabs } = useAdminTabs({
  currentSection,
  currentTitle,
  hasRouteTitle,
  onRouteChange: () => {
    sidebarOpen.value = false;
  },
  route,
  router,
});

function handleScroll() {
  showBackToTop.value = window.scrollY > 360;
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function toggleSidebar() {
  if (window.innerWidth < 1024) {
    sidebarOpen.value = !sidebarOpen.value;
    return;
  }
  preferences.sidebarCollapsed = !preferences.sidebarCollapsed;
}

async function submitTopbarSearch() {
  const trimmedQuery = topbarSearchQuery.value.trim();
  if (!trimmedQuery) {
    await router.push("/");
    return;
  }

  await router.push({
    path: "/",
    query: buildSearchQuery("multi", trimmedQuery),
  });
}

watch(
  () => route.query.q,
  (value) => {
    topbarSearchQuery.value = readQueryString(value);
  },
  { immediate: true },
);

onMounted(() => {
  handleScroll();
  window.addEventListener("scroll", handleScroll, { passive: true });
});

onBeforeUnmount(() => {
  window.removeEventListener("scroll", handleScroll);
});
</script>

<template>
  <div
    class="admin-app"
    :class="{ 'admin-app-sidebar-collapsed': preferences.sidebarCollapsed }"
    :data-theme="currentThemeOption.dataTheme"
    :style="adminThemeStyle"
  >
    <AdminSidebar :active-path="activeMenuPath" :groups="menuGroups" :open="sidebarOpen" />

    <button
      v-if="sidebarOpen"
      class="admin-sidebar-mask"
      type="button"
      aria-label="关闭导航"
      @click="sidebarOpen = false"
    ></button>

    <section class="admin-workspace">
      <AdminTopbar
        v-model:search-query="topbarSearchQuery"
        :current-section="currentSection"
        :current-title="currentTitle"
        :sidebar-open="sidebarOpen"
        @open-preferences="preferencesOpen = true"
        @submit-search="submitTopbarSearch"
        @toggle-sidebar="toggleSidebar"
      />

      <AdminTabs
        v-if="preferences.showTabs"
        :current-full-path="route.fullPath"
        :tabs="openedTabs"
        @close="closeTab"
      />

      <main class="page-shell admin-content" :class="{ 'admin-content-compact': preferences.compact }">
        <VbenPage
          :description="currentDescription"
          :section="currentSection"
          :show-header="false"
          :title="currentTitle"
        >
          <RouterView />
        </VbenPage>
      </main>
    </section>

    <AdminPreferencesDrawer
      :current-sidebar-option="currentSidebarOption"
      :current-theme-option="currentThemeOption"
      :preferences="preferences"
      :sidebar-options="sidebarOptions"
      :theme-options="themeOptions"
      :visible="preferencesOpen"
      @close="preferencesOpen = false"
      @reset="resetPreferences"
      @update-preference="setPreference"
    />

    <button v-if="showBackToTop" class="back-top-btn" type="button" aria-label="返回顶部" @click="scrollToTop"></button>
  </div>
</template>
