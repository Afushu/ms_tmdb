export interface AdminMenuItem {
  order: number;
  path: string;
  section: string;
  title: string;
}

export interface AdminMenuGroup {
  items: AdminMenuItem[];
  section: string;
}

export interface AdminTab {
  fullPath: string;
  path: string;
  section: string;
  title: string;
}

export type AdminThemeColor = "teal" | "blue" | "green" | "amber" | "rose" | "dark";
export type AdminSidebarColor = "navy" | "light" | "dark" | "teal" | "blue" | "green" | "purple";

export interface AdminPreferences {
  compact: boolean;
  showTabs: boolean;
  sidebarCollapsed: boolean;
  sidebarColor: AdminSidebarColor;
  themeColor: AdminThemeColor;
}

export interface AdminThemeOption {
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

export interface AdminSidebarOption {
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

export const defaultPreferences: AdminPreferences = {
  compact: false,
  showTabs: true,
  sidebarCollapsed: false,
  sidebarColor: "navy",
  themeColor: "teal",
};

export const sidebarOptions: AdminSidebarOption[] = [
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

export const themeOptions: AdminThemeOption[] = [
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

const fallbackMenuIconPaths = ["M4 5h16", "M4 12h16", "M4 19h16"];
const menuIconPaths: Record<string, string[]> = {
  "/": ["M3 10.5 12 3l9 7.5", "M5 10v10h14V10", "M9 20v-6h6v6"],
  "/library": [
    "M4 6c0-1.1 3.6-2 8-2s8 .9 8 2-3.6 2-8 2-8-.9-8-2Z",
    "M4 6v6c0 1.1 3.6 2 8 2s8-.9 8-2V6",
    "M4 12v6c0 1.1 3.6 2 8 2s8-.9 8-2v-6",
  ],
  "/logs": ["M7 4h10", "M7 9h10", "M7 14h6", "M5 20h14a2 2 0 0 0 2-2V3H3v15a2 2 0 0 0 2 2Z"],
  "/system-settings": [
    "M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z",
    "M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.07a2 2 0 0 1-2.83 2.83l-.07-.06A1.7 1.7 0 0 0 15 19.4a1.7 1.7 0 0 0-1 1.55V21a2 2 0 0 1-4 0v-.09a1.7 1.7 0 0 0-1-1.55 1.7 1.7 0 0 0-1.88.34l-.07.06a2 2 0 1 1-2.83-2.83l.06-.07A1.7 1.7 0 0 0 4.6 15a1.7 1.7 0 0 0-1.55-1H3a2 2 0 0 1 0-4h.09a1.7 1.7 0 0 0 1.55-1 1.7 1.7 0 0 0-.34-1.88l-.06-.07a2 2 0 1 1 2.83-2.83l.07.06A1.7 1.7 0 0 0 9 4.6a1.7 1.7 0 0 0 1-1.55V3a2 2 0 0 1 4 0v.09a1.7 1.7 0 0 0 1 1.55 1.7 1.7 0 0 0 1.88-.34l.07-.06a2 2 0 1 1 2.83 2.83l-.06.07A1.7 1.7 0 0 0 19.4 9c.36.63.99 1 1.55 1H21a2 2 0 0 1 0 4h-.09a1.7 1.7 0 0 0-1.55 1Z",
  ],
};

export function getMenuIconPaths(path: string) {
  return menuIconPaths[path] ?? fallbackMenuIconPaths;
}

export function themeSwatchStyle(option: AdminThemeOption) {
  return {
    background: `linear-gradient(135deg, ${option.bgMain} 0 52%, ${option.accent} 52% 100%)`,
  };
}

export function sidebarControlStyle(option: AdminSidebarOption) {
  return {
    "--sidebar-preview": option.bg,
  };
}
