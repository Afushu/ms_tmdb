import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";

declare module "vue-router" {
  interface RouteMeta {
    activeMenu?: string;
    description?: string;
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
    meta: { title: "首页", section: "工作台", description: "检索热门内容，查看搜索结果和数据状态。", order: 10 },
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
      description: "查看、编辑电影资料，并处理 TMDB 同步差异。",
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
      description: "查看剧集资料、季信息和远端同步差异。",
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
      description: "查看人物资料与关联作品信息。",
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
      description: "管理本地缓存媒体、手动新建记录并进入详情。",
      order: 30,
    },
  },
  {
    path: "/logs",
    component: () => import("@/pages/LogsPage.vue"),
    meta: {
      title: "日志",
      section: "系统管理",
      description: "查看代理访问与 TMDB 回源请求日志。",
      order: 40,
    },
  },
  {
    path: "/system-settings",
    component: () => import("@/pages/SystemSettingsPage.vue"),
    meta: {
      title: "系统设置",
      section: "系统管理",
      description: "配置网络代理、自动同步任务和执行日志。",
      order: 50,
    },
  },
  { path: "/proxy-settings", redirect: "/system-settings", meta: { hideMenu: true, hideTab: true } },
  { path: "/auto-sync-settings", redirect: "/system-settings", meta: { hideMenu: true, hideTab: true } },
  {
    path: "/:pathMatch(.*)*",
    component: () => import("@/pages/NotFoundPage.vue"),
    meta: { title: "页面不存在", section: "系统", description: "请返回已有菜单或检查访问地址。", hideMenu: true },
  },
];

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(_to, _from, savedPosition) {
    return savedPosition ?? { top: 0 };
  },
  routes,
});

export default router;
