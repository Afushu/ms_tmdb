import { ref, watch, type ComputedRef } from "vue";
import type { RouteLocationNormalizedLoaded, Router } from "vue-router";
import type { AdminTab } from "./adminLayoutConfig";

interface UseAdminTabsOptions {
  currentSection: ComputedRef<string>;
  currentTitle: ComputedRef<string>;
  hasRouteTitle: ComputedRef<boolean>;
  onRouteChange?: () => void;
  route: RouteLocationNormalizedLoaded;
  router: Router;
}

export function useAdminTabs(options: UseAdminTabsOptions) {
  const { currentSection, currentTitle, hasRouteTitle, onRouteChange, route, router } = options;
  const openedTabs = ref<AdminTab[]>([]);

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

  function closeOtherTabs(target: AdminTab) {
    openedTabs.value = openedTabs.value.filter((tab) => tab.path === "/" || tab.fullPath === target.fullPath);
    if (route.fullPath !== target.fullPath) {
      void router.push(target.fullPath);
    }
  }

  function closeLeftTabs(target: AdminTab) {
    const index = openedTabs.value.findIndex((tab) => tab.fullPath === target.fullPath);
    if (index < 0) return;
    openedTabs.value = openedTabs.value.filter((tab, i) => i >= index || tab.path === "/");
    if (!openedTabs.value.some((tab) => tab.fullPath === route.fullPath)) {
      void router.push(target.fullPath);
    }
  }

  function closeRightTabs(target: AdminTab) {
    const index = openedTabs.value.findIndex((tab) => tab.fullPath === target.fullPath);
    if (index < 0) return;
    openedTabs.value = openedTabs.value.filter((tab, i) => i <= index || tab.path === "/");
    if (!openedTabs.value.some((tab) => tab.fullPath === route.fullPath)) {
      void router.push(target.fullPath);
    }
  }

  function closeAllTabs() {
    openedTabs.value = openedTabs.value.filter((tab) => tab.path === "/");
    void router.push("/");
  }

  watch(
    () => route.fullPath,
    () => {
      syncOpenedTab();
      onRouteChange?.();
    },
    { immediate: true },
  );

  return {
    closeAllTabs,
    closeLeftTabs,
    closeOtherTabs,
    closeRightTabs,
    closeTab,
    openedTabs,
  };
}
