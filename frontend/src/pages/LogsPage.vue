<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import AccessLogList from "@/components/logs/AccessLogList.vue";
import AutoSyncLogDetailModal from "@/components/logs/AutoSyncLogDetailModal.vue";
import AutoSyncLogList from "@/components/logs/AutoSyncLogList.vue";
import ClearLogsConfirmDialog from "@/components/logs/ClearLogsConfirmDialog.vue";
import LogsHero from "@/components/logs/LogsHero.vue";
import LogsPagination from "@/components/logs/LogsPagination.vue";
import LogsToolbar from "@/components/logs/LogsToolbar.vue";
import RequestLogDetailModal from "@/components/logs/RequestLogDetailModal.vue";
import TmdbLogList from "@/components/logs/TmdbLogList.vue";
import ToastNotice from "@/components/common/ToastNotice.vue";
import {
  clearAutoSyncLogs,
  clearProxyAccessLogs,
  clearTmdbRequestLogs,
  getAutoSyncLogDetail,
  getAutoSyncLogs,
  getProxyAccessLogDetail,
  getProxyAccessLogs,
  getTmdbRequestLogDetail,
  getTmdbRequestLogs,
  type AdminAutoSyncLogDetailParams,
  type AdminAutoSyncLogDetailResp,
  type AdminAutoSyncLogItem,
  type AdminProxyAccessLogDetailResp,
  type AdminProxyAccessLogItem,
  type AdminTmdbRequestLogDetailResp,
  type AdminTmdbRequestLogItem,
} from "@/api/admin";
import { useToastNotice } from "@/composables/useToastNotice";
import { formatRequestLogTotal } from "@/utils/logFormatters";
import { detailTotalPages, normalizeNumber, totalPages } from "@/utils/logPagination";

type LogTab = "access" | "tmdb" | "autoSync";
type RequestLogTab = "access" | "tmdb";
type RequestLogDetail = AdminProxyAccessLogDetailResp | AdminTmdbRequestLogDetailResp;

const route = useRoute();
const router = useRouter();
const activeTab = ref<LogTab>(readLogTab(route.query.tab));

const accessLoading = ref(false);
const accessClearing = ref(false);
const accessLoaded = ref(false);
const accessStatus = ref("");
const accessKeyword = ref("");
const accessPage = ref(1);
const accessPageSize = ref(20);
const accessTotal = ref(0);
const accessItems = ref<AdminProxyAccessLogItem[]>([]);

const tmdbLoading = ref(false);
const tmdbClearing = ref(false);
const tmdbLoaded = ref(false);
const tmdbStatus = ref("");
const tmdbKeyword = ref("");
const tmdbPage = ref(1);
const tmdbPageSize = ref(20);
const tmdbTotal = ref(0);
const tmdbItems = ref<AdminTmdbRequestLogItem[]>([]);

const autoSyncLoading = ref(false);
const autoSyncClearing = ref(false);
const autoSyncLoaded = ref(false);
const autoSyncStatus = ref("");
const autoSyncPage = ref(1);
const autoSyncPageSize = ref(10);
const autoSyncTotal = ref(0);
const autoSyncItems = ref<AdminAutoSyncLogItem[]>([]);

const detailVisible = ref(false);
const detailLoading = ref(false);
const detailType = ref<RequestLogTab>("access");
const accessDetail = ref<AdminProxyAccessLogDetailResp | null>(null);
const tmdbDetail = ref<AdminTmdbRequestLogDetailResp | null>(null);

const autoSyncDetailVisible = ref(false);
const autoSyncDetailLoading = ref(false);
const activeAutoSyncDetail = ref<AdminAutoSyncLogDetailResp | null>(null);
const activeAutoSyncLogId = ref<number | null>(null);
const detailSyncedPage = ref(1);
const detailSyncedPageSize = ref(10);
const detailFailedPage = ref(1);
const detailFailedPageSize = ref(10);

const clearConfirmVisible = ref(false);
const { toastVisible, toastText, toastTone, showToastNotice, closeToastNotice } = useToastNotice();

const requestStatusOptions = [
  { label: "全部状态", value: "" },
  { label: "成功", value: "success" },
  { label: "失败", value: "error" },
];

const autoSyncStatusOptions = [
  { label: "全部状态", value: "" },
  { label: "成功", value: "success" },
  { label: "部分失败", value: "partial_failed" },
  { label: "异常", value: "panic" },
  { label: "已取消", value: "canceled" },
];

const currentBusy = computed(() => {
  if (activeTab.value === "access") {
    return accessLoading.value || accessClearing.value;
  }
  if (activeTab.value === "tmdb") {
    return tmdbLoading.value || tmdbClearing.value;
  }
  return autoSyncLoading.value || autoSyncClearing.value;
});
const currentTotal = computed(() => {
  if (activeTab.value === "access") return accessTotal.value;
  if (activeTab.value === "tmdb") return tmdbTotal.value;
  return autoSyncTotal.value;
});
const currentPage = computed(() => {
  if (activeTab.value === "access") return accessPage.value;
  if (activeTab.value === "tmdb") return tmdbPage.value;
  return autoSyncPage.value;
});
const currentPageSize = computed(() => {
  if (activeTab.value === "access") return accessPageSize.value;
  if (activeTab.value === "tmdb") return tmdbPageSize.value;
  return autoSyncPageSize.value;
});
const currentTotalPages = computed(() => totalPages(currentTotal.value, currentPageSize.value));
const accessTotalText = computed(() => formatRequestLogTotal(accessTotal.value));
const tmdbTotalText = computed(() => formatRequestLogTotal(tmdbTotal.value));
const autoSyncTotalText = computed(() => formatRequestLogTotal(autoSyncTotal.value));
const activeDetail = computed<RequestLogDetail | null>(() => (detailType.value === "access" ? accessDetail.value : tmdbDetail.value));
const currentKeyword = computed(() => {
  if (activeTab.value === "access") return accessKeyword.value.trim();
  if (activeTab.value === "tmdb") return tmdbKeyword.value.trim();
  return "";
});
const currentPanelLabel = computed(() => {
  if (activeTab.value === "access") return "Access";
  if (activeTab.value === "tmdb") return "Upstream";
  return "Schedule";
});
const currentPanelTitle = computed(() => {
  if (activeTab.value === "access") return "外部访问日志";
  if (activeTab.value === "tmdb") return "TMDB 请求日志";
  return "定时任务执行日志";
});
const currentPanelNote = computed(() => {
  if (activeTab.value === "access") return "记录命中代理入口的外部请求。";
  if (activeTab.value === "tmdb") return "记录服务端实际访问 TMDB 的请求。";
  return "记录定时同步任务每次执行后的检查、同步和失败结果。";
});
const currentClearLabel = computed(() => {
  if (activeTab.value === "access") return "外部访问日志";
  if (activeTab.value === "tmdb") return "TMDB 请求日志";
  return "定时任务执行日志";
});

function readLogTab(value: unknown): LogTab {
  const text = Array.isArray(value) ? value[0] : value;
  if (text === "tmdb" || text === "autoSync") {
    return text;
  }
  return "access";
}

function syncActiveTabQuery(tab: LogTab) {
  const nextQuery = { ...route.query };
  if (tab === "access") {
    delete nextQuery.tab;
  } else {
    nextQuery.tab = tab;
  }
  void router.replace({ query: nextQuery });
}

function setActiveTab(tab: LogTab, syncQuery = true) {
  activeTab.value = tab;
  if (syncQuery) {
    syncActiveTabQuery(tab);
  }
  if (tab === "access" && !accessLoaded.value) {
    void loadAccessLogs(1);
  }
  if (tab === "tmdb" && !tmdbLoaded.value) {
    void loadTmdbLogs(1);
  }
  if (tab === "autoSync" && !autoSyncLoaded.value) {
    void loadAutoSyncLogs(1);
  }
}

async function loadAccessLogs(page = accessPage.value) {
  accessLoading.value = true;
  try {
    const safePage = Math.max(1, Math.trunc(page));
    const resp = await getProxyAccessLogs({
      page: safePage,
      page_size: accessPageSize.value,
      status: accessStatus.value || undefined,
      keyword: accessKeyword.value.trim() || undefined,
    });
    const data = resp.data;
    accessItems.value = Array.isArray(data.results) ? data.results : [];
    accessTotal.value = Math.max(0, Number(data.total) || 0);
    accessPageSize.value = normalizeNumber(Number(data.page_size) || accessPageSize.value, 1, 100);
    accessPage.value = normalizeNumber(Number(data.page) || 1, 1, totalPages(accessTotal.value, accessPageSize.value));
    accessLoaded.value = true;
  } catch {
    // 错误已由全局请求拦截器提示，这里保留当前日志列表。
  } finally {
    accessLoading.value = false;
  }
}

async function loadTmdbLogs(page = tmdbPage.value) {
  tmdbLoading.value = true;
  try {
    const safePage = Math.max(1, Math.trunc(page));
    const resp = await getTmdbRequestLogs({
      page: safePage,
      page_size: tmdbPageSize.value,
      status: tmdbStatus.value || undefined,
      keyword: tmdbKeyword.value.trim() || undefined,
    });
    const data = resp.data;
    tmdbItems.value = Array.isArray(data.results) ? data.results : [];
    tmdbTotal.value = Math.max(0, Number(data.total) || 0);
    tmdbPageSize.value = normalizeNumber(Number(data.page_size) || tmdbPageSize.value, 1, 100);
    tmdbPage.value = normalizeNumber(Number(data.page) || 1, 1, totalPages(tmdbTotal.value, tmdbPageSize.value));
    tmdbLoaded.value = true;
  } catch {
    // 错误已由全局请求拦截器提示，这里保留当前日志列表。
  } finally {
    tmdbLoading.value = false;
  }
}

async function loadAutoSyncLogs(page = autoSyncPage.value) {
  autoSyncLoading.value = true;
  try {
    const safePage = Math.max(1, Math.trunc(page));
    const resp = await getAutoSyncLogs({
      page: safePage,
      page_size: autoSyncPageSize.value,
      status: autoSyncStatus.value || undefined,
    });
    const data = resp.data;
    autoSyncItems.value = Array.isArray(data.results) ? data.results : [];
    autoSyncTotal.value = Math.max(0, Number(data.total) || 0);
    autoSyncPageSize.value = normalizeNumber(Number(data.page_size) || autoSyncPageSize.value, 1, 100);
    autoSyncPage.value = normalizeNumber(
      Number(data.page) || 1,
      1,
      totalPages(autoSyncTotal.value, autoSyncPageSize.value),
    );
    autoSyncLoaded.value = true;
  } catch {
    // 错误已由全局请求拦截器提示，这里保留当前日志列表。
  } finally {
    autoSyncLoading.value = false;
  }
}

async function refreshCurrentLogs() {
  if (activeTab.value === "access") {
    await loadAccessLogs(accessPage.value);
    return;
  }
  if (activeTab.value === "tmdb") {
    await loadTmdbLogs(tmdbPage.value);
    return;
  }
  await loadAutoSyncLogs(autoSyncPage.value);
}

async function applyStatusFilter() {
  if (activeTab.value === "access") {
    accessPage.value = 1;
    await loadAccessLogs(1);
    return;
  }
  if (activeTab.value === "tmdb") {
    tmdbPage.value = 1;
    await loadTmdbLogs(1);
    return;
  }
  autoSyncPage.value = 1;
  await loadAutoSyncLogs(1);
}

async function applyKeywordSearch() {
  if (activeTab.value === "access") {
    accessPage.value = 1;
    await loadAccessLogs(1);
    return;
  }
  if (activeTab.value === "tmdb") {
    tmdbPage.value = 1;
    await loadTmdbLogs(1);
  }
}

async function clearKeywordSearch() {
  if (activeTab.value === "access") {
    if (!accessKeyword.value.trim()) {
      return;
    }
    accessKeyword.value = "";
    accessPage.value = 1;
    await loadAccessLogs(1);
    return;
  }
  if (activeTab.value !== "tmdb" || !tmdbKeyword.value.trim()) {
    return;
  }
  tmdbKeyword.value = "";
  tmdbPage.value = 1;
  await loadTmdbLogs(1);
}

async function goToPage(page: number) {
  const target = normalizeNumber(page, 1, currentTotalPages.value);
  if (activeTab.value === "access") {
    await loadAccessLogs(target);
    return;
  }
  if (activeTab.value === "tmdb") {
    await loadTmdbLogs(target);
    return;
  }
  await loadAutoSyncLogs(target);
}

async function changePageSize(pageSize: number) {
  const nextPageSize = normalizeNumber(Number(pageSize), 1, 100);
  if (activeTab.value === "access") {
    accessPageSize.value = nextPageSize;
    accessPage.value = 1;
    await loadAccessLogs(1);
    return;
  }
  if (activeTab.value === "tmdb") {
    tmdbPageSize.value = nextPageSize;
    tmdbPage.value = 1;
    await loadTmdbLogs(1);
    return;
  }
  autoSyncPageSize.value = nextPageSize;
  autoSyncPage.value = 1;
  await loadAutoSyncLogs(1);
}

async function openAccessDetail(item: AdminProxyAccessLogItem) {
  detailVisible.value = true;
  detailLoading.value = true;
  detailType.value = "access";
  accessDetail.value = null;
  tmdbDetail.value = null;
  try {
    const resp = await getProxyAccessLogDetail(item.id);
    accessDetail.value = resp.data;
  } catch {
    // 错误已由全局请求拦截器提示，这里只关闭详情加载态。
  } finally {
    detailLoading.value = false;
  }
}

async function openTmdbDetail(item: AdminTmdbRequestLogItem) {
  detailVisible.value = true;
  detailLoading.value = true;
  detailType.value = "tmdb";
  accessDetail.value = null;
  tmdbDetail.value = null;
  try {
    const resp = await getTmdbRequestLogDetail(item.id);
    tmdbDetail.value = resp.data;
  } catch {
    // 错误已由全局请求拦截器提示，这里只关闭详情加载态。
  } finally {
    detailLoading.value = false;
  }
}

function closeDetail() {
  detailVisible.value = false;
  detailLoading.value = false;
  accessDetail.value = null;
  tmdbDetail.value = null;
}

async function loadAutoSyncLogDetail(id: number, params: AdminAutoSyncLogDetailParams = {}, reset = false) {
  autoSyncDetailLoading.value = true;
  activeAutoSyncLogId.value = id;
  if (reset) {
    activeAutoSyncDetail.value = null;
  }

  try {
    const resp = await getAutoSyncLogDetail(id, {
      synced_page: params.synced_page ?? detailSyncedPage.value,
      synced_page_size: params.synced_page_size ?? detailSyncedPageSize.value,
      failed_page: params.failed_page ?? detailFailedPage.value,
      failed_page_size: params.failed_page_size ?? detailFailedPageSize.value,
    });
    const data = resp.data;
    if (!autoSyncDetailVisible.value || activeAutoSyncLogId.value !== id) {
      return;
    }
    activeAutoSyncDetail.value = data;
    detailSyncedPageSize.value = normalizeNumber(Number(data.synced_page_size) || detailSyncedPageSize.value, 1, 100);
    detailFailedPageSize.value = normalizeNumber(Number(data.failed_page_size) || detailFailedPageSize.value, 1, 100);
    detailSyncedPage.value = normalizeNumber(
      Number(data.synced_page) || 1,
      1,
      detailTotalPages(data.synced, detailSyncedPageSize.value),
    );
    detailFailedPage.value = normalizeNumber(
      Number(data.failed_page) || 1,
      1,
      detailTotalPages(data.failed, detailFailedPageSize.value),
    );
  } catch {
    // 错误已由全局请求拦截器提示，这里只关闭详情加载态。
  } finally {
    if (activeAutoSyncLogId.value === id) {
      autoSyncDetailLoading.value = false;
    }
  }
}

async function openAutoSyncDetail(item: AdminAutoSyncLogItem) {
  autoSyncDetailVisible.value = true;
  detailSyncedPage.value = 1;
  detailFailedPage.value = 1;
  await loadAutoSyncLogDetail(
    item.id,
    {
      synced_page: 1,
      synced_page_size: detailSyncedPageSize.value,
      failed_page: 1,
      failed_page_size: detailFailedPageSize.value,
    },
    true,
  );
}

async function goToDetailSyncedPage(page: number) {
  if (!activeAutoSyncLogId.value) {
    return;
  }
  const target = normalizeNumber(page, 1, detailSyncedTotalPages());
  await loadAutoSyncLogDetail(activeAutoSyncLogId.value, {
    synced_page: target,
    failed_page: detailFailedPage.value,
  });
}

async function goToDetailFailedPage(page: number) {
  if (!activeAutoSyncLogId.value) {
    return;
  }
  const target = normalizeNumber(page, 1, detailFailedTotalPages());
  await loadAutoSyncLogDetail(activeAutoSyncLogId.value, {
    synced_page: detailSyncedPage.value,
    failed_page: target,
  });
}

function closeAutoSyncDetail() {
  autoSyncDetailVisible.value = false;
  autoSyncDetailLoading.value = false;
  activeAutoSyncDetail.value = null;
  activeAutoSyncLogId.value = null;
  detailSyncedPage.value = 1;
  detailFailedPage.value = 1;
}

function detailSyncedTotalPages() {
  return detailTotalPages(activeAutoSyncDetail.value?.synced ?? 0, detailSyncedPageSize.value);
}

function detailFailedTotalPages() {
  return detailTotalPages(activeAutoSyncDetail.value?.failed ?? 0, detailFailedPageSize.value);
}

function openClearConfirm() {
  clearConfirmVisible.value = true;
}

function closeClearConfirm() {
  if (currentBusy.value) {
    return;
  }
  clearConfirmVisible.value = false;
}

async function clearCurrentLogs() {
  if (activeTab.value === "access") {
    accessClearing.value = true;
    try {
      const resp = await clearProxyAccessLogs();
      showToastNotice(resp.data.message || "代理访问日志已清空");
      accessPage.value = 1;
      await loadAccessLogs(1);
      closeDetail();
      clearConfirmVisible.value = false;
    } catch {
      // 错误已由全局请求拦截器提示，这里只恢复清空状态。
    } finally {
      accessClearing.value = false;
    }
    return;
  }

  if (activeTab.value === "tmdb") {
    tmdbClearing.value = true;
    try {
      const resp = await clearTmdbRequestLogs();
      showToastNotice(resp.data.message || "TMDB 请求日志已清空");
      tmdbPage.value = 1;
      await loadTmdbLogs(1);
      closeDetail();
      clearConfirmVisible.value = false;
    } catch {
      // 错误已由全局请求拦截器提示，这里只恢复清空状态。
    } finally {
      tmdbClearing.value = false;
    }
    return;
  }

  autoSyncClearing.value = true;
  try {
    const resp = await clearAutoSyncLogs();
    showToastNotice(resp.data.message || "定时任务执行日志已清空");
    autoSyncPage.value = 1;
    await loadAutoSyncLogs(1);
    closeAutoSyncDetail();
    clearConfirmVisible.value = false;
  } catch {
    // 错误已由全局请求拦截器提示，这里只恢复清空状态。
  } finally {
    autoSyncClearing.value = false;
  }
}

watch(
  () => route.query.tab,
  (value) => {
    const nextTab = readLogTab(value);
    if (nextTab !== activeTab.value) {
      setActiveTab(nextTab, false);
    }
  },
);

onMounted(() => {
  void Promise.all([loadAccessLogs(1), loadTmdbLogs(1), loadAutoSyncLogs(1)]);
});
</script>

<template>
  <section class="logs-page-shell">
    <LogsHero
      :active-tab="activeTab"
      :access-total-text="accessTotalText"
      :tmdb-total-text="tmdbTotalText"
      :auto-sync-total-text="autoSyncTotalText"
      @select-tab="setActiveTab"
    />

    <section class="card settings-card-wide settings-log-card">
      <LogsToolbar
        v-model:access-keyword="accessKeyword"
        v-model:tmdb-keyword="tmdbKeyword"
        v-model:access-status="accessStatus"
        v-model:tmdb-status="tmdbStatus"
        v-model:auto-sync-status="autoSyncStatus"
        :active-tab="activeTab"
        :busy="currentBusy"
        :panel-label="currentPanelLabel"
        :panel-title="currentPanelTitle"
        :panel-note="currentPanelNote"
        :request-status-options="requestStatusOptions"
        :auto-sync-status-options="autoSyncStatusOptions"
        :current-keyword="currentKeyword"
        @search="applyKeywordSearch"
        @clear-keyword="clearKeywordSearch"
        @status-change="applyStatusFilter"
        @refresh="refreshCurrentLogs"
        @clear="openClearConfirm"
      />

      <AccessLogList
        v-if="activeTab === 'access'"
        :items="accessItems"
        :loading="accessLoading"
        @open-detail="openAccessDetail"
      />
      <TmdbLogList
        v-else-if="activeTab === 'tmdb'"
        :items="tmdbItems"
        :loading="tmdbLoading"
        @open-detail="openTmdbDetail"
      />
      <AutoSyncLogList v-else :items="autoSyncItems" :loading="autoSyncLoading" @open-detail="openAutoSyncDetail" />

      <LogsPagination
        :total="currentTotal"
        :page="currentPage"
        :page-size="currentPageSize"
        :total-pages="currentTotalPages"
        :busy="currentBusy"
        @change-page="goToPage"
        @change-page-size="changePageSize"
      />
    </section>

    <RequestLogDetailModal
      :visible="detailVisible"
      :loading="detailLoading"
      :detail-type="detailType"
      :detail="activeDetail"
      @close="closeDetail"
    />

    <AutoSyncLogDetailModal
      :visible="autoSyncDetailVisible"
      :loading="autoSyncDetailLoading"
      :detail="activeAutoSyncDetail"
      :synced-page="detailSyncedPage"
      :synced-total-pages="detailSyncedTotalPages()"
      :failed-page="detailFailedPage"
      :failed-total-pages="detailFailedTotalPages()"
      @close="closeAutoSyncDetail"
      @change-synced-page="goToDetailSyncedPage"
      @change-failed-page="goToDetailFailedPage"
    />

    <ClearLogsConfirmDialog
      :visible="clearConfirmVisible"
      :busy="currentBusy"
      :label="currentClearLabel"
      @close="closeClearConfirm"
      @confirm="clearCurrentLogs"
    />

    <ToastNotice :visible="toastVisible" :message="toastText" :tone="toastTone" @close="closeToastNotice" />
  </section>
</template>
