<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import VbenPage from "@/components/layout/VbenPage.vue";
import { buildSearchQuery, readQueryString } from "@/utils/routeSearch";

interface AdminMenuItem {
  order: number;
  path: string;
  section: string;
  title: string;
}

interface AdminMenuGroup {
  items: AdminMenuItem[];
  section: string;
}

interface AdminTab {
  fullPath: string;
  path: string;
  section: string;
  title: string;
}

type AdminThemeColor = "teal" | "blue" | "green" | "amber" | "rose" | "dark";
type AdminSidebarColor = "navy" | "light" | "dark" | "teal" | "blue" | "green" | "purple";

interface AdminPreferences {
  compact: boolean;
  showTabs: boolean;
  sidebarCollapsed: boolean;
  sidebarColor: AdminSidebarColor;
  themeColor: AdminThemeColor;
}

interface AdminThemeOption {
  accent: string;
  accentSoft: string;
  accentStrong: string;
  bgMain: string;
  borderMuted: string;
  colorScheme: "dark" | "light";
  dataTheme: "msdark" | "mslight";
  fieldBorderFocus: string;
  fieldBg: string;
  fieldBorder: string;
  glassBg: string;
  glassBgStrong: string;
  glassBorder: string;
  glassShadow: string;
  glassShadowSoft: string;
  label: string;
  sidebarBg: string;
  surface: string;
  surfaceMuted: string;
  surfaceStrong: string;
  textMain: string;
  textMuted: string;
  topbarBg: string;
  value: AdminThemeColor;
}

interface AdminSidebarOption {
  activeBg: string;
  activeIconBg: string;
  activeIconText: string;
  activeText: string;
  bg: string;
  border: string;
  hoverBg: string;
  hoverIconBg: string;
  hoverIconText: string;
  hoverText: string;
  iconBg: string;
  label: string;
  muted: string;
  text: string;
  value: AdminSidebarColor;
}

const preferenceStorageKey = "ms_tmdb_admin_preferences";
const defaultPreferences: AdminPreferences = {
  compact: false,
  showTabs: true,
  sidebarCollapsed: false,
  sidebarColor: "navy",
  themeColor: "teal",
};

const sidebarOptions: AdminSidebarOption[] = [
  {
    activeBg: "var(--accent)",
    activeIconBg: "rgba(255, 255, 255, 0.18)",
    activeIconText: "#ffffff",
    activeText: "#ffffff",
    bg: "#001529",
    border: "rgba(255, 255, 255, 0.08)",
    hoverBg: "rgba(255, 255, 255, 0.07)",
    hoverIconBg: "rgba(255, 255, 255, 0.14)",
    hoverIconText: "#ffffff",
    hoverText: "#ffffff",
    iconBg: "rgba(255, 255, 255, 0.1)",
    label: "深蓝",
    muted: "rgba(255, 255, 255, 0.42)",
    text: "rgba(255, 255, 255, 0.72)",
    value: "navy",
  },
  {
    activeBg: "var(--accent)",
    activeIconBg: "rgba(255, 255, 255, 0.22)",
    activeIconText: "#ffffff",
    activeText: "#ffffff",
    bg: "#ffffff",
    border: "#e5e7eb",
    hoverBg: "#f3f6fb",
    hoverIconBg: "#e5e7eb",
    hoverIconText: "#111827",
    hoverText: "#111827",
    iconBg: "#f2f4f7",
    label: "浅色",
    muted: "#98a2b3",
    text: "#344054",
    value: "light",
  },
  {
    activeBg: "var(--accent)",
    activeIconBg: "rgba(255, 255, 255, 0.18)",
    activeIconText: "#ffffff",
    activeText: "#ffffff",
    bg: "#111827",
    border: "rgba(255, 255, 255, 0.08)",
    hoverBg: "rgba(255, 255, 255, 0.07)",
    hoverIconBg: "rgba(255, 255, 255, 0.14)",
    hoverIconText: "#ffffff",
    hoverText: "#ffffff",
    iconBg: "rgba(255, 255, 255, 0.1)",
    label: "深灰",
    muted: "rgba(255, 255, 255, 0.42)",
    text: "rgba(255, 255, 255, 0.72)",
    value: "dark",
  },
  {
    activeBg: "var(--accent)",
    activeIconBg: "rgba(255, 255, 255, 0.18)",
    activeIconText: "#ffffff",
    activeText: "#ffffff",
    bg: "#0f3f3f",
    border: "rgba(255, 255, 255, 0.08)",
    hoverBg: "rgba(255, 255, 255, 0.08)",
    hoverIconBg: "rgba(255, 255, 255, 0.14)",
    hoverIconText: "#ffffff",
    hoverText: "#ffffff",
    iconBg: "rgba(255, 255, 255, 0.1)",
    label: "青绿",
    muted: "rgba(255, 255, 255, 0.46)",
    text: "rgba(255, 255, 255, 0.76)",
    value: "teal",
  },
  {
    activeBg: "var(--accent)",
    activeIconBg: "rgba(255, 255, 255, 0.18)",
    activeIconText: "#ffffff",
    activeText: "#ffffff",
    bg: "#0f2f66",
    border: "rgba(255, 255, 255, 0.08)",
    hoverBg: "rgba(255, 255, 255, 0.08)",
    hoverIconBg: "rgba(255, 255, 255, 0.14)",
    hoverIconText: "#ffffff",
    hoverText: "#ffffff",
    iconBg: "rgba(255, 255, 255, 0.1)",
    label: "蓝色",
    muted: "rgba(255, 255, 255, 0.46)",
    text: "rgba(255, 255, 255, 0.76)",
    value: "blue",
  },
  {
    activeBg: "var(--accent)",
    activeIconBg: "rgba(255, 255, 255, 0.18)",
    activeIconText: "#ffffff",
    activeText: "#ffffff",
    bg: "#12361f",
    border: "rgba(255, 255, 255, 0.08)",
    hoverBg: "rgba(255, 255, 255, 0.08)",
    hoverIconBg: "rgba(255, 255, 255, 0.14)",
    hoverIconText: "#ffffff",
    hoverText: "#ffffff",
    iconBg: "rgba(255, 255, 255, 0.1)",
    label: "绿色",
    muted: "rgba(255, 255, 255, 0.46)",
    text: "rgba(255, 255, 255, 0.76)",
    value: "green",
  },
  {
    activeBg: "var(--accent)",
    activeIconBg: "rgba(255, 255, 255, 0.18)",
    activeIconText: "#ffffff",
    activeText: "#ffffff",
    bg: "#2f1f56",
    border: "rgba(255, 255, 255, 0.08)",
    hoverBg: "rgba(255, 255, 255, 0.08)",
    hoverIconBg: "rgba(255, 255, 255, 0.14)",
    hoverIconText: "#ffffff",
    hoverText: "#ffffff",
    iconBg: "rgba(255, 255, 255, 0.1)",
    label: "紫色",
    muted: "rgba(255, 255, 255, 0.46)",
    text: "rgba(255, 255, 255, 0.76)",
    value: "purple",
  },
];

const themeOptions: AdminThemeOption[] = [
  {
    accent: "#0ea5a4",
    accentSoft: "#90cea1",
    accentStrong: "#0d7c8a",
    bgMain: "#f5f7fb",
    borderMuted: "#e5e7eb",
    colorScheme: "light",
    dataTheme: "mslight",
    fieldBorderFocus: "rgba(14, 165, 164, 0.68)",
    fieldBg: "#ffffff",
    fieldBorder: "rgba(15, 23, 42, 0.14)",
    glassBg: "rgba(255, 255, 255, 0.86)",
    glassBgStrong: "rgba(255, 255, 255, 0.98)",
    glassBorder: "#e5e7eb",
    glassShadow: "0 1px 3px rgba(15, 23, 42, 0.06)",
    glassShadowSoft: "0 1px 2px rgba(15, 23, 42, 0.04)",
    label: "青绿",
    sidebarBg: "#001529",
    surface: "#ffffff",
    surfaceMuted: "#f5f7fb",
    surfaceStrong: "#ffffff",
    textMain: "#1f2937",
    textMuted: "#667085",
    topbarBg: "#ffffff",
    value: "teal",
  },
  {
    accent: "#2563eb",
    accentSoft: "#93c5fd",
    accentStrong: "#1d4ed8",
    bgMain: "#f5f7fb",
    borderMuted: "#e5e7eb",
    colorScheme: "light",
    dataTheme: "mslight",
    fieldBorderFocus: "rgba(37, 99, 235, 0.68)",
    fieldBg: "#ffffff",
    fieldBorder: "rgba(15, 23, 42, 0.14)",
    glassBg: "rgba(255, 255, 255, 0.86)",
    glassBgStrong: "rgba(255, 255, 255, 0.98)",
    glassBorder: "#e5e7eb",
    glassShadow: "0 1px 3px rgba(15, 23, 42, 0.06)",
    glassShadowSoft: "0 1px 2px rgba(15, 23, 42, 0.04)",
    label: "蓝色",
    sidebarBg: "#001529",
    surface: "#ffffff",
    surfaceMuted: "#f5f7fb",
    surfaceStrong: "#ffffff",
    textMain: "#1f2937",
    textMuted: "#667085",
    topbarBg: "#ffffff",
    value: "blue",
  },
  {
    accent: "#16a34a",
    accentSoft: "#86efac",
    accentStrong: "#15803d",
    bgMain: "#f5f7fb",
    borderMuted: "#e5e7eb",
    colorScheme: "light",
    dataTheme: "mslight",
    fieldBorderFocus: "rgba(22, 163, 74, 0.68)",
    fieldBg: "#ffffff",
    fieldBorder: "rgba(15, 23, 42, 0.14)",
    glassBg: "rgba(255, 255, 255, 0.86)",
    glassBgStrong: "rgba(255, 255, 255, 0.98)",
    glassBorder: "#e5e7eb",
    glassShadow: "0 1px 3px rgba(15, 23, 42, 0.06)",
    glassShadowSoft: "0 1px 2px rgba(15, 23, 42, 0.04)",
    label: "绿色",
    sidebarBg: "#001529",
    surface: "#ffffff",
    surfaceMuted: "#f5f7fb",
    surfaceStrong: "#ffffff",
    textMain: "#1f2937",
    textMuted: "#667085",
    topbarBg: "#ffffff",
    value: "green",
  },
  {
    accent: "#d97706",
    accentSoft: "#fcd34d",
    accentStrong: "#b45309",
    bgMain: "#f5f7fb",
    borderMuted: "#e5e7eb",
    colorScheme: "light",
    dataTheme: "mslight",
    fieldBorderFocus: "rgba(217, 119, 6, 0.68)",
    fieldBg: "#ffffff",
    fieldBorder: "rgba(15, 23, 42, 0.14)",
    glassBg: "rgba(255, 255, 255, 0.86)",
    glassBgStrong: "rgba(255, 255, 255, 0.98)",
    glassBorder: "#e5e7eb",
    glassShadow: "0 1px 3px rgba(15, 23, 42, 0.06)",
    glassShadowSoft: "0 1px 2px rgba(15, 23, 42, 0.04)",
    label: "琥珀",
    sidebarBg: "#001529",
    surface: "#ffffff",
    surfaceMuted: "#f5f7fb",
    surfaceStrong: "#ffffff",
    textMain: "#1f2937",
    textMuted: "#667085",
    topbarBg: "#ffffff",
    value: "amber",
  },
  {
    accent: "#e11d48",
    accentSoft: "#fda4af",
    accentStrong: "#be123c",
    bgMain: "#f5f7fb",
    borderMuted: "#e5e7eb",
    colorScheme: "light",
    dataTheme: "mslight",
    fieldBorderFocus: "rgba(225, 29, 72, 0.68)",
    fieldBg: "#ffffff",
    fieldBorder: "rgba(15, 23, 42, 0.14)",
    glassBg: "rgba(255, 255, 255, 0.86)",
    glassBgStrong: "rgba(255, 255, 255, 0.98)",
    glassBorder: "#e5e7eb",
    glassShadow: "0 1px 3px rgba(15, 23, 42, 0.06)",
    glassShadowSoft: "0 1px 2px rgba(15, 23, 42, 0.04)",
    label: "玫红",
    sidebarBg: "#001529",
    surface: "#ffffff",
    surfaceMuted: "#f5f7fb",
    surfaceStrong: "#ffffff",
    textMain: "#1f2937",
    textMuted: "#667085",
    topbarBg: "#ffffff",
    value: "rose",
  },
  {
    accent: "#0ea5a4",
    accentSoft: "#90cea1",
    accentStrong: "#0d7c8a",
    bgMain: "#0f1115",
    borderMuted: "rgba(208, 216, 228, 0.12)",
    colorScheme: "dark",
    dataTheme: "msdark",
    fieldBorderFocus: "rgba(14, 165, 164, 0.68)",
    fieldBg: "rgba(18, 22, 28, 0.96)",
    fieldBorder: "rgba(208, 216, 228, 0.16)",
    glassBg: "rgba(25, 30, 38, 0.72)",
    glassBgStrong: "rgba(29, 35, 43, 0.92)",
    glassBorder: "rgba(208, 216, 228, 0.12)",
    glassShadow: "0 18px 38px rgba(0, 0, 0, 0.28)",
    glassShadowSoft: "0 10px 22px rgba(0, 0, 0, 0.2)",
    label: "深色",
    sidebarBg: "#001529",
    surface: "#161a20",
    surfaceMuted: "#1d232b",
    surfaceStrong: "#101317",
    textMain: "#edf2f7",
    textMuted: "#aeb9c7",
    topbarBg: "rgba(18, 22, 28, 0.94)",
    value: "dark",
  },
];

const route = useRoute();
const router = useRouter();
const sidebarOpen = ref(false);
const preferencesOpen = ref(false);
const showBackToTop = ref(false);
const openedTabs = ref<AdminTab[]>([]);
const preferences = reactive<AdminPreferences>({ ...defaultPreferences });
const topbarSearchQuery = ref("");

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
const currentThemeOption = computed(
  () => themeOptions.find((item) => item.value === preferences.themeColor) ?? themeOptions[0],
);
const currentSidebarOption = computed(
  () => sidebarOptions.find((item) => item.value === preferences.sidebarColor) ?? sidebarOptions[0],
);
const adminThemeStyle = computed(() => ({
  "--bg-main": currentThemeOption.value.bgMain,
  "--border-muted": currentThemeOption.value.borderMuted,
  "--accent": currentThemeOption.value.accent,
  "--accent-active-bg": `${currentThemeOption.value.accent}24`,
  "--accent-active-border": `${currentThemeOption.value.accent}52`,
  "--accent-hover-border": `${currentThemeOption.value.accent}57`,
  "--accent-ring": `${currentThemeOption.value.accent}1f`,
  "--accent-soft": currentThemeOption.value.accentSoft,
  "--accent-strong": currentThemeOption.value.accentStrong,
  "--field-border-focus": currentThemeOption.value.fieldBorderFocus,
  "--field-bg": currentThemeOption.value.fieldBg,
  "--field-border": currentThemeOption.value.fieldBorder,
  "--glass-bg": currentThemeOption.value.glassBg,
  "--glass-bg-strong": currentThemeOption.value.glassBgStrong,
  "--glass-border": currentThemeOption.value.glassBorder,
  "--glass-shadow": currentThemeOption.value.glassShadow,
  "--glass-shadow-soft": currentThemeOption.value.glassShadowSoft,
  "--sidebar-active-bg": currentSidebarOption.value.activeBg,
  "--sidebar-active-icon-bg": currentSidebarOption.value.activeIconBg,
  "--sidebar-active-icon-text": currentSidebarOption.value.activeIconText,
  "--sidebar-active-text": currentSidebarOption.value.activeText,
  "--sidebar-bg": currentSidebarOption.value.bg,
  "--sidebar-border": currentSidebarOption.value.border,
  "--sidebar-hover-bg": currentSidebarOption.value.hoverBg,
  "--sidebar-hover-icon-bg": currentSidebarOption.value.hoverIconBg,
  "--sidebar-hover-icon-text": currentSidebarOption.value.hoverIconText,
  "--sidebar-hover-text": currentSidebarOption.value.hoverText,
  "--sidebar-icon-bg": currentSidebarOption.value.iconBg,
  "--sidebar-muted": currentSidebarOption.value.muted,
  "--sidebar-text": currentSidebarOption.value.text,
  "--surface": currentThemeOption.value.surface,
  "--surface-muted": currentThemeOption.value.surfaceMuted,
  "--surface-strong": currentThemeOption.value.surfaceStrong,
  "--text-main": currentThemeOption.value.textMain,
  "--text-muted": currentThemeOption.value.textMuted,
  "--topbar-bg": currentThemeOption.value.topbarBg,
  "color-scheme": currentThemeOption.value.colorScheme,
}));

function isActiveMenu(path: string) {
  return path === activeMenuPath.value;
}

function buildCurrentTab(): AdminTab | null {
  if (route.meta.hideTab) return null;
  if (!hasRouteTitle.value && route.path !== "/") return null;
  return {
    fullPath: route.fullPath,
    path: route.path,
    section: currentSection.value,
    title: currentTitle.value,
  };
}

function ensureHomeTab() {
  if (openedTabs.value.some((tab) => tab.path === "/")) return;
  openedTabs.value.unshift({
    fullPath: "/",
    path: "/",
    section: "工作台",
    title: "首页",
  });
}

function trimTabs() {
  while (openedTabs.value.length > 8) {
    const removableIndex = openedTabs.value.findIndex((tab) => tab.path !== "/" && tab.fullPath !== route.fullPath);
    if (removableIndex < 0) return;
    openedTabs.value.splice(removableIndex, 1);
  }
}

function removeDuplicateTabs(currentTab: AdminTab) {
  const seenPaths = new Set<string>();
  openedTabs.value = openedTabs.value.filter((tab) => {
    if (tab.path === currentTab.path) {
      if (tab.fullPath === currentTab.fullPath) {
        seenPaths.add(tab.path);
        return true;
      }
      return false;
    }

    if (seenPaths.has(tab.path)) return false;
    seenPaths.add(tab.path);
    return true;
  });
}

function syncOpenedTab() {
  ensureHomeTab();
  const currentTab = buildCurrentTab();
  if (!currentTab) return;

  const existedIndex = openedTabs.value.findIndex((tab) => tab.path === currentTab.path);
  if (existedIndex >= 0) {
    openedTabs.value.splice(existedIndex, 1, currentTab);
  } else {
    openedTabs.value.push(currentTab);
  }
  removeDuplicateTabs(currentTab);
  trimTabs();
}

function closeTab(tab: AdminTab, event: MouseEvent) {
  event.preventDefault();
  event.stopPropagation();
  if (tab.path === "/") return;

  const tabIndex = openedTabs.value.findIndex((item) => item.fullPath === tab.fullPath);
  if (tabIndex < 0) return;
  openedTabs.value.splice(tabIndex, 1);

  if (tab.fullPath !== route.fullPath) return;
  const nextTab = openedTabs.value[tabIndex - 1] ?? openedTabs.value[tabIndex] ?? openedTabs.value[0];
  void router.push(nextTab?.fullPath ?? "/");
}

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

function normalizePreferences(raw: unknown): AdminPreferences {
  const payload = raw && typeof raw === "object" ? (raw as Partial<Record<keyof AdminPreferences, unknown>>) : {};
  const themeColor = themeOptions.some((item) => item.value === payload.themeColor)
    ? (payload.themeColor as AdminThemeColor)
    : defaultPreferences.themeColor;
  const sidebarColor = sidebarOptions.some((item) => item.value === payload.sidebarColor)
    ? (payload.sidebarColor as AdminSidebarColor)
    : defaultPreferences.sidebarColor;
  return {
    compact: typeof payload.compact === "boolean" ? payload.compact : defaultPreferences.compact,
    showTabs: typeof payload.showTabs === "boolean" ? payload.showTabs : defaultPreferences.showTabs,
    sidebarCollapsed:
      typeof payload.sidebarCollapsed === "boolean" ? payload.sidebarCollapsed : defaultPreferences.sidebarCollapsed,
    sidebarColor,
    themeColor,
  };
}

function loadPreferences() {
  try {
    const raw = window.localStorage.getItem(preferenceStorageKey);
    if (!raw) return;
    Object.assign(preferences, normalizePreferences(JSON.parse(raw) as unknown));
  } catch {
    Object.assign(preferences, defaultPreferences);
  }
}

function savePreferences() {
  window.localStorage.setItem(preferenceStorageKey, JSON.stringify(preferences));
}

function resetPreferences() {
  Object.assign(preferences, defaultPreferences);
}

function themeSwatchStyle(option: AdminThemeOption) {
  return {
    background: `linear-gradient(135deg, ${option.bgMain} 0 52%, ${option.accent} 52% 100%)`,
  };
}

function sidebarControlStyle(option: AdminSidebarOption) {
  return {
    "--sidebar-preview": option.bg,
  };
}

const menuIconPaths: Record<string, string[]> = {
  "/": ["M3 10.5 12 3l9 7.5", "M5 10v10h14V10", "M9 20v-6h6v6"],
  "/library": ["M4 6c0-1.1 3.6-2 8-2s8 .9 8 2-3.6 2-8 2-8-.9-8-2Z", "M4 6v6c0 1.1 3.6 2 8 2s8-.9 8-2V6", "M4 12v6c0 1.1 3.6 2 8 2s8-.9 8-2v-6"],
  "/logs": ["M7 4h10", "M7 9h10", "M7 14h6", "M5 20h14a2 2 0 0 0 2-2V3H3v15a2 2 0 0 0 2 2Z"],
  "/system-settings": ["M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z", "M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.07a2 2 0 0 1-2.83 2.83l-.07-.06A1.7 1.7 0 0 0 15 19.4a1.7 1.7 0 0 0-1 1.55V21a2 2 0 0 1-4 0v-.09a1.7 1.7 0 0 0-1-1.55 1.7 1.7 0 0 0-1.88.34l-.07.06a2 2 0 1 1-2.83-2.83l.06-.07A1.7 1.7 0 0 0 4.6 15a1.7 1.7 0 0 0-1.55-1H3a2 2 0 0 1 0-4h.09a1.7 1.7 0 0 0 1.55-1 1.7 1.7 0 0 0-.34-1.88l-.06-.07a2 2 0 1 1 2.83-2.83l.07.06A1.7 1.7 0 0 0 9 4.6a1.7 1.7 0 0 0 1-1.55V3a2 2 0 0 1 4 0v.09a1.7 1.7 0 0 0 1 1.55 1.7 1.7 0 0 0 1.88-.34l.07-.06a2 2 0 1 1 2.83 2.83l-.06.07A1.7 1.7 0 0 0 19.4 9c.36.63.99 1 1.55 1H21a2 2 0 0 1 0 4h-.09a1.7 1.7 0 0 0-1.55 1Z"],
};

function getMenuIconPaths(path: string) {
  return menuIconPaths[path] ?? ["M4 5h16", "M4 12h16", "M4 19h16"];
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
  () => route.fullPath,
  () => {
    syncOpenedTab();
    sidebarOpen.value = false;
  },
  { immediate: true },
);

watch(
  () => route.query.q,
  (value) => {
    topbarSearchQuery.value = readQueryString(value);
  },
  { immediate: true },
);

watch(preferences, savePreferences);

onMounted(() => {
  loadPreferences();
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
    <aside class="admin-sidebar" :class="{ 'admin-sidebar-open': sidebarOpen }" aria-label="主导航">
      <RouterLink to="/" class="admin-brand">
        <span class="admin-brand-mark">MS</span>
        <span class="admin-brand-copy">
          <strong>MS TMDB</strong>
          <small>Media Service</small>
        </span>
      </RouterLink>

      <nav class="admin-menu">
        <section v-for="group in menuGroups" :key="group.section" class="admin-menu-section">
          <p class="admin-menu-title">{{ group.section }}</p>
          <RouterLink
            v-for="item in group.items"
            :key="item.path"
            :to="item.path"
            class="admin-menu-link"
            :class="{ 'admin-menu-link-active': isActiveMenu(item.path) }"
          >
            <span class="admin-menu-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" focusable="false">
                <path
                  v-for="pathData in getMenuIconPaths(item.path)"
                  :key="pathData"
                  :d="pathData"
                  pathLength="24"
                />
              </svg>
            </span>
            <span>{{ item.title }}</span>
          </RouterLink>
        </section>
      </nav>
    </aside>

    <button
      v-if="sidebarOpen"
      class="admin-sidebar-mask"
      type="button"
      aria-label="关闭导航"
      @click="sidebarOpen = false"
    ></button>

    <section class="admin-workspace">
      <header class="admin-topbar">
        <button
          class="admin-icon-btn"
          type="button"
          :aria-label="sidebarOpen ? '关闭导航' : '打开导航'"
          @click="toggleSidebar"
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
          <form class="admin-top-search" role="search" @submit.prevent="submitTopbarSearch">
            <input
              v-model="topbarSearchQuery"
              class="admin-top-search-input"
              type="search"
              placeholder="搜索电影、剧集、人物"
            />
            <button class="admin-top-search-btn" type="submit" aria-label="搜索">
              <span class="admin-icon-search"></span>
            </button>
          </form>
          <button class="admin-preference-btn" type="button" aria-label="偏好设置" @click="preferencesOpen = true">
            <span class="admin-icon-gear"></span>
          </button>
        </div>
      </header>

      <div v-if="preferences.showTabs" class="admin-tabs" aria-label="已打开页面">
        <div
          v-for="tab in openedTabs"
          :key="tab.fullPath"
          class="admin-tab"
          :class="{ 'admin-tab-active': tab.fullPath === route.fullPath }"
        >
          <RouterLink :to="tab.fullPath" class="admin-tab-link">{{ tab.title }}</RouterLink>
          <button
            v-if="tab.path !== '/'"
            type="button"
            class="admin-tab-close"
            aria-label="关闭标签"
            @click="closeTab(tab, $event)"
          ></button>
        </div>
      </div>

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

    <button
      v-if="preferencesOpen"
      class="admin-preference-mask"
      type="button"
      aria-label="关闭偏好设置"
      @click="preferencesOpen = false"
    ></button>
    <aside v-if="preferencesOpen" class="admin-preference-drawer" aria-label="偏好设置">
      <header class="admin-preference-header">
        <h2>偏好设置</h2>
        <button class="admin-drawer-close" type="button" aria-label="关闭偏好设置" @click="preferencesOpen = false">
        </button>
      </header>

      <div class="admin-preference-options">
        <section class="admin-preference-group" aria-label="布局">
          <p class="admin-preference-section-title">布局</p>
          <label class="admin-preference-option">
            <span>折叠菜单</span>
            <input v-model="preferences.sidebarCollapsed" type="checkbox" class="admin-switch-input" />
            <span class="admin-switch-track" aria-hidden="true">
              <span class="admin-switch-thumb"></span>
            </span>
          </label>
          <label class="admin-preference-option">
            <span>显示标签栏</span>
            <input v-model="preferences.showTabs" type="checkbox" class="admin-switch-input" />
            <span class="admin-switch-track" aria-hidden="true">
              <span class="admin-switch-thumb"></span>
            </span>
          </label>
        </section>

        <section class="admin-preference-group" aria-label="外观">
          <p class="admin-preference-section-title">外观</p>
          <p class="admin-preference-label">主题色 · {{ currentThemeOption.label }}</p>
          <div class="admin-theme-grid">
            <button
              v-for="option in themeOptions"
              :key="option.value"
              type="button"
              class="admin-theme-swatch"
              :class="{ 'admin-theme-swatch-active': preferences.themeColor === option.value }"
              :style="themeSwatchStyle(option)"
              :aria-label="`主题色：${option.label}`"
              :title="option.label"
              @click="preferences.themeColor = option.value"
            >
              <span v-if="preferences.themeColor === option.value" class="admin-theme-check"></span>
            </button>
          </div>
          <p class="admin-preference-label">侧栏色 · {{ currentSidebarOption.label }}</p>
          <div class="admin-sidebar-control" role="group" aria-label="侧栏色">
            <button
              v-for="option in sidebarOptions"
              :key="option.value"
              type="button"
              class="admin-sidebar-control-btn"
              :class="{ 'admin-sidebar-control-btn-active': preferences.sidebarColor === option.value }"
              :style="sidebarControlStyle(option)"
              :aria-label="`侧栏色：${option.label}`"
              :title="option.label"
              @click="preferences.sidebarColor = option.value"
            >
              <span class="admin-sidebar-control-preview"></span>
              <span>{{ option.label }}</span>
            </button>
          </div>
        </section>

        <label class="admin-preference-option">
          <span>紧凑间距</span>
          <input v-model="preferences.compact" type="checkbox" class="admin-switch-input" />
          <span class="admin-switch-track" aria-hidden="true">
            <span class="admin-switch-thumb"></span>
          </span>
        </label>
      </div>

      <button class="btn-soft admin-preference-reset" type="button" @click="resetPreferences">恢复默认</button>
    </aside>

    <button v-if="showBackToTop" class="back-top-btn" type="button" aria-label="返回顶部" @click="scrollToTop"></button>
  </div>
</template>
