import { computed, onMounted, reactive, watch } from "vue";
import {
  defaultPreferences,
  sidebarOptions,
  themeOptions,
  type AdminPreferences,
  type AdminSidebarColor,
  type AdminThemeColor,
} from "./adminLayoutConfig";

const preferenceStorageKey = "ms_tmdb_admin_preferences";

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

export function useAdminPreferences() {
  const preferences = reactive<AdminPreferences>({ ...defaultPreferences });
  const currentThemeOption = computed(
    () => themeOptions.find((item) => item.value === preferences.themeColor) ?? themeOptions[0],
  );
  const currentSidebarOption = computed(
    () => sidebarOptions.find((item) => item.value === preferences.sidebarColor) ?? sidebarOptions[0],
  );
  const adminThemeStyle = computed<Record<string, string>>(() => ({
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

  function setPreference(key: keyof AdminPreferences, value: AdminPreferences[keyof AdminPreferences]) {
    Object.assign(preferences, { [key]: value });
  }

  watch(preferences, savePreferences);
  onMounted(loadPreferences);

  return {
    adminThemeStyle,
    currentSidebarOption,
    currentThemeOption,
    preferences,
    resetPreferences,
    setPreference,
  };
}
