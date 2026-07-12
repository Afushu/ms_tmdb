import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import { startGlobalPageLoading, stopGlobalPageLoading } from "@/composables/useGlobalPageLoading";

declare module "vue-router" {
  interface RouteMeta {
    activeMenu?: string;
    hideMenu?: boolean;
    hideTab?: boolean;
    menuTitle?: string;
    order?: number;
    section?: string;
    title?: string;
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    component: () => import("@/pages/HomePage.vue"),
    meta: { title: "首页", section: "工作台", order: 10 },
  },
  {
    path: "/search",
    redirect: (to) => ({ path: "/", query: to.query }),
    meta: { hideMenu: true, hideTab: true },
  },
  {
    path: "/movie/:id",
    component: () => import("@/pages/MoviePage.vue"),
    meta: {
      title: "电影详情",
      section: "媒体数据",
      activeMenu: "/library",
      hideMenu: true,
    },
  },
  {
    path: "/tv/:id",
    component: () => import("@/pages/TVPage.vue"),
    meta: {
      title: "剧集详情",
      section: "媒体数据",
      activeMenu: "/library",
      hideMenu: true,
    },
  },
  {
    path: "/person/:id",
    component: () => import("@/pages/PersonPage.vue"),
    meta: {
      title: "人物详情",
      section: "媒体数据",
      activeMenu: "/",
      hideMenu: true,
    },
  },
  {
    path: "/library",
    component: () => import("@/pages/LibraryPage.vue"),
    meta: {
      title: "本地库",
      section: "媒体数据",
      order: 30,
    },
  },
  {
    path: "/logs",
    component: () => import("@/pages/LogsPage.vue"),
    meta: {
      title: "日志",
      section: "系统管理",
      order: 50,
    },
  },
  {
    path: "/system-settings",
    component: () => import("@/pages/SystemSettingsPage.vue"),
    meta: {
      title: "系统设置",
      section: "系统管理",
      order: 40,
    },
  },
  { path: "/proxy-settings", redirect: "/system-settings", meta: { hideMenu: true, hideTab: true } },
  { path: "/auto-sync-settings", redirect: "/system-settings", meta: { hideMenu: true, hideTab: true } },
  {
    path: "/:pathMatch(.*)*",
    component: () => import("@/pages/NotFoundPage.vue"),
    meta: { title: "页面不存在", section: "系统", hideMenu: true },
  },
];

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(_to, _from, savedPosition) {
    return savedPosition ?? { top: 0 };
  },
  routes,
});

router.beforeEach(() => {
  startGlobalPageLoading();
  return true;
});

const APP_TITLE = "媒体数据管理";

router.afterEach((to, _from, failure) => {
  stopGlobalPageLoading();
  if (failure != null) {
    return;
  }
  const routeTitle = typeof to.meta.title === "string" ? to.meta.title.trim() : "";
  document.title = routeTitle ? `${routeTitle} - ${APP_TITLE}` : APP_TITLE;
});

router.onError(() => {
  stopGlobalPageLoading();
});

export default router;
