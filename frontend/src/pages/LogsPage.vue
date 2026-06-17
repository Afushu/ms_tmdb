<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import GlassSelect from "@/components/GlassSelect.vue";
import JsonFoldViewer from "@/components/common/JsonFoldViewer.vue";
import ToastNotice from "@/components/common/ToastNotice.vue";
import {
  clearProxyAccessLogs,
  clearTmdbRequestLogs,
  getProxyAccessLogDetail,
  getProxyAccessLogs,
  getTmdbRequestLogDetail,
  getTmdbRequestLogs,
  type AdminProxyAccessLogDetailResp,
  type AdminProxyAccessLogItem,
  type AdminTmdbRequestLogDetailResp,
  type AdminTmdbRequestLogItem,
} from "@/api/admin";
import { useToastNotice } from "@/composables/useToastNotice";

type LogTab = "access" | "tmdb";
type LogDetail = AdminProxyAccessLogDetailResp | AdminTmdbRequestLogDetailResp;

const activeTab = ref<LogTab>("access");

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

const detailVisible = ref(false);
const detailLoading = ref(false);
const detailType = ref<LogTab>("access");
const accessDetail = ref<AdminProxyAccessLogDetailResp | null>(null);
const tmdbDetail = ref<AdminTmdbRequestLogDetailResp | null>(null);

const clearConfirmVisible = ref(false);
const { toastVisible, toastText, toastTone, showToastNotice, closeToastNotice } = useToastNotice();

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "成功", value: "success" },
  { label: "失败", value: "error" },
];

const currentBusy = computed(() =>
  activeTab.value === "access" ? accessLoading.value || accessClearing.value : tmdbLoading.value || tmdbClearing.value,
);
const currentTotal = computed(() => (activeTab.value === "access" ? accessTotal.value : tmdbTotal.value));
const currentPage = computed(() => (activeTab.value === "access" ? accessPage.value : tmdbPage.value));
const currentPageSize = computed(() => (activeTab.value === "access" ? accessPageSize.value : tmdbPageSize.value));
const currentTotalPages = computed(() => totalPages(currentTotal.value, currentPageSize.value));
const accessTotalText = computed(() => formatRequestLogTotal(accessTotal.value));
const tmdbTotalText = computed(() => formatRequestLogTotal(tmdbTotal.value));
const currentTotalText = computed(() => (activeTab.value === "access" ? accessTotalText.value : tmdbTotalText.value));
const currentTotalLabel = computed(() => `共 ${currentTotalText.value}`);
const activeDetail = computed(() => (detailType.value === "access" ? accessDetail.value : tmdbDetail.value));
const currentKeyword = computed(() =>
  activeTab.value === "access" ? accessKeyword.value.trim() : tmdbKeyword.value.trim(),
);
const currentStatusLabel = computed(() => {
  const value = activeTab.value === "access" ? accessStatus.value : tmdbStatus.value;
  return statusOptions.find((item) => item.value === value)?.label ?? "全部状态";
});

function totalPages(total: number, pageSize: number) {
  const safeTotal = Math.max(0, Number(total) || 0);
  const safePageSize = normalizeNumber(Number(pageSize) || 20, 1, 100);
  return Math.max(1, Math.ceil(safeTotal / safePageSize));
}

function formatRequestLogTotal(total: number) {
  return `${Math.max(0, Number(total) || 0)}`;
}

function normalizeNumber(value: number, min: number, max: number) {
  const next = Number.isFinite(value) ? Math.trunc(value) : min;
  if (next < min) return min;
  if (next > max) return max;
  return next;
}

function setActiveTab(tab: LogTab) {
  activeTab.value = tab;
  if (tab === "access" && !accessLoaded.value) {
    void loadAccessLogs(1);
  }
  if (tab === "tmdb" && !tmdbLoaded.value) {
    void loadTmdbLogs(1);
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
    accessPage.value = normalizeNumber(Number(data.page) || 1, 1, totalPages(accessTotal.value, accessPageSize.value));
    accessPageSize.value = normalizeNumber(Number(data.page_size) || accessPageSize.value, 1, 100);
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
    tmdbPage.value = normalizeNumber(Number(data.page) || 1, 1, totalPages(tmdbTotal.value, tmdbPageSize.value));
    tmdbPageSize.value = normalizeNumber(Number(data.page_size) || tmdbPageSize.value, 1, 100);
    tmdbLoaded.value = true;
  } catch {
    // 错误已由全局请求拦截器提示，这里保留当前日志列表。
  } finally {
    tmdbLoading.value = false;
  }
}

async function refreshCurrentLogs() {
  if (activeTab.value === "access") {
    await loadAccessLogs(accessPage.value);
    return;
  }
  await loadTmdbLogs(tmdbPage.value);
}

async function applyStatusFilter() {
  if (activeTab.value === "access") {
    accessPage.value = 1;
    await loadAccessLogs(1);
    return;
  }
  tmdbPage.value = 1;
  await loadTmdbLogs(1);
}

async function applyKeywordSearch() {
  if (activeTab.value === "access") {
    accessPage.value = 1;
    await loadAccessLogs(1);
    return;
  }
  tmdbPage.value = 1;
  await loadTmdbLogs(1);
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
  if (!tmdbKeyword.value.trim()) {
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
  await loadTmdbLogs(target);
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
}

function formatDateTime(value: string) {
  const text = (value ?? "").trim();
  if (!text) {
    return "-";
  }
  const date = new Date(text);
  if (Number.isNaN(date.getTime())) {
    return text;
  }
  return date.toLocaleString("zh-CN", { hour12: false });
}

function formatDuration(durationMs: number) {
  const ms = Number.isFinite(durationMs) ? Math.max(0, Math.trunc(durationMs)) : 0;
  if (ms < 1000) {
    return `${ms}ms`;
  }
  const seconds = ms / 1000;
  return seconds < 60
    ? `${seconds.toFixed(seconds < 10 ? 1 : 0)}s`
    : `${Math.floor(seconds / 60)}m ${Math.round(seconds % 60)}s`;
}

function formatBytes(bytes: number) {
  const size = Number.isFinite(bytes) ? Math.max(0, bytes) : 0;
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(size < 10 * 1024 ? 1 : 0)} KB`;
  return `${(size / 1024 / 1024).toFixed(1)} MB`;
}

function formatStatusCode(code: number) {
  return code > 0 ? String(code) : "未返回";
}

function statusClass(code: number) {
  if (code >= 200 && code < 400) {
    return "bg-green-50 text-green-700 border border-green-200";
  }
  if (code === 0 || code >= 400) {
    return "bg-red-50 text-red-700 border border-red-200";
  }
  return "bg-amber-50 text-amber-700 border border-amber-200";
}

function trimMiddle(value: string, max = 72) {
  const text = (value ?? "").trim();
  if (!text) return "-";
  if (text.length <= max) return text;
  const head = Math.ceil((max - 3) * 0.6);
  const tail = Math.floor((max - 3) * 0.4);
  return `${text.slice(0, head)}...${text.slice(text.length - tail)}`;
}

function splitPathAndQuery(value: string) {
  const text = (value ?? "").trim();
  if (!text) {
    return { path: "-", query: "" };
  }
  const index = text.indexOf("?");
  if (index < 0) {
    return { path: text, query: "" };
  }
  return {
    path: text.slice(0, index) || "/",
    query: text.slice(index + 1),
  };
}

function accessPath(value: string) {
  return splitPathAndQuery(value).path;
}

function formatQueryForDisplay(query: string, max: number) {
  const text = (query ?? "").trim().replace(/^\?/, "");
  if (!text) {
    return "无查询参数";
  }
  const params = new window.URLSearchParams(text);
  for (const key of [...params.keys()]) {
    if (key.toLowerCase() === "api_key") {
      params.delete(key);
    }
  }
  const safeQuery = params.toString();
  return safeQuery ? `?${trimMiddle(safeQuery, max)}` : "无查询参数";
}

function accessQuery(value: string) {
  const query = splitPathAndQuery(value).query;
  return formatQueryForDisplay(query, 96);
}

function upstreamHost(value: string) {
  const text = (value ?? "").trim();
  if (!text) {
    return "-";
  }
  try {
    return new URL(text).origin;
  } catch {
    return "-";
  }
}

function upstreamQuery(value: string) {
  const text = (value ?? "").trim();
  if (!text) {
    return "无查询参数";
  }
  try {
    const parsed = new URL(text);
    return formatQueryForDisplay(parsed.search, 110);
  } catch {
    return accessQuery(text);
  }
}

function bodyText(text: string | undefined) {
  const raw = text ?? "";
  if (!raw) {
    return "(空)";
  }
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

function bodyMeta(bytes: number, truncated: boolean) {
  return `${formatBytes(bytes)}${truncated ? " · 已截断" : ""}`;
}

function isAccessDetail(detail: LogDetail): detail is AdminProxyAccessLogDetailResp {
  return "request_uri" in detail;
}

function detailEndpointPath(detail: LogDetail) {
  if (isAccessDetail(detail)) {
    return accessPath(detail.request_uri);
  }
  return detail.path || "-";
}

function detailEndpointDisplay(detail: LogDetail) {
  const path = detailEndpointPath(detail);
  const query = detailEndpointQuery(detail);
  return query === "无查询参数" ? path : `${path}${query}`;
}

function detailEndpointQuery(detail: LogDetail) {
  if (isAccessDetail(detail)) {
    return accessQuery(detail.request_uri);
  }
  return upstreamQuery(detail.url);
}

function detailEndpointSource(detail: LogDetail) {
  if (isAccessDetail(detail)) {
    return detail.client_ip || "-";
  }
  return upstreamHost(detail.url);
}

function detailEndpointTitle(detail: LogDetail) {
  return detailEndpointDisplay(detail);
}

onMounted(() => {
  void Promise.all([loadAccessLogs(1), loadTmdbLogs(1)]);
});
</script>

<template>
  <section class="grid gap-4">
    <section class="settings-toolbar card">
      <div class="min-w-0">
        <p class="section-label">Logs</p>
        <h2 class="library-toolbar-title">请求日志</h2>
        <p class="mt-1 text-sm text-black/55">代理访问与 TMDB 回源请求记录。</p>
      </div>

      <div class="library-toolbar-actions">
        <div class="glass-pill">
          <button
            type="button"
            class="glass-pill-btn"
            :class="activeTab === 'access' ? 'glass-pill-btn-active' : ''"
            @click="setActiveTab('access')"
          >
            外部访问
          </button>
          <button
            type="button"
            class="glass-pill-btn"
            :class="activeTab === 'tmdb' ? 'glass-pill-btn-active' : ''"
            @click="setActiveTab('tmdb')"
          >
            TMDB 请求
          </button>
        </div>
      </div>
    </section>

    <section class="logs-overview-strip">
      <div>
        <span>外部访问</span>
        <strong>{{ accessTotalText }}</strong>
      </div>
      <div>
        <span>TMDB 请求</span>
        <strong>{{ tmdbTotalText }}</strong>
      </div>
      <div>
        <span>当前分页</span>
        <strong>{{ currentPage }} / {{ currentTotalPages }}</strong>
      </div>
      <div>
        <span>状态筛选</span>
        <strong>{{ currentStatusLabel }}</strong>
      </div>
    </section>

    <section class="card settings-card-wide settings-log-card">
      <div class="settings-log-header">
        <div>
          <p class="section-label">{{ activeTab === "access" ? "Access" : "Upstream" }}</p>
          <h3 class="settings-section-title">{{ activeTab === "access" ? "外部访问日志" : "TMDB 请求日志" }}</h3>
          <p class="settings-note">
            {{ activeTab === "access" ? "记录命中代理入口的外部请求。" : "记录服务端实际访问 TMDB 的请求。" }}
          </p>
        </div>

        <div class="settings-log-actions">
          <form class="settings-log-search" @submit.prevent="applyKeywordSearch">
            <label class="settings-log-filter">
              关键字
              <input
                v-if="activeTab === 'access'"
                v-model.trim="accessKeyword"
                class="field-control settings-log-search-input"
                type="search"
                placeholder="搜索路径、查询、IP 或 UA"
                :disabled="currentBusy"
              />
              <input
                v-else
                v-model.trim="tmdbKeyword"
                class="field-control settings-log-search-input"
                type="search"
                placeholder="搜索上游路径或 URL"
                :disabled="currentBusy"
              />
            </label>
            <button class="btn-soft disabled:opacity-60" type="submit" :disabled="currentBusy">搜索</button>
            <button
              v-if="currentKeyword"
              class="btn-soft disabled:opacity-60"
              type="button"
              :disabled="currentBusy"
              @click="clearKeywordSearch"
            >
              清空关键字
            </button>
          </form>

          <label class="settings-log-filter">
            状态
            <GlassSelect
              v-if="activeTab === 'access'"
              v-model="accessStatus"
              :options="statusOptions"
              :disabled="currentBusy"
              class="min-w-[136px]"
              @change="applyStatusFilter"
            />
            <GlassSelect
              v-else
              v-model="tmdbStatus"
              :options="statusOptions"
              :disabled="currentBusy"
              class="min-w-[136px]"
              @change="applyStatusFilter"
            />
          </label>

          <button class="btn-soft disabled:opacity-60" :disabled="currentBusy" @click="refreshCurrentLogs">
            {{ currentBusy ? "刷新中..." : "刷新日志" }}
          </button>
          <button class="btn-danger-soft disabled:opacity-60" :disabled="currentBusy" @click="openClearConfirm">
            {{
              activeTab === "access" && accessClearing
                ? "清空中..."
                : activeTab === "tmdb" && tmdbClearing
                  ? "清空中..."
                  : "清空日志"
            }}
          </button>
        </div>
      </div>

      <template v-if="activeTab === 'access'">

        <div class="logs-list-shell">
          <div class="logs-list-head logs-grid-access">
            <span>时间</span>
            <span>请求</span>
            <span>状态</span>
            <span>耗时</span>
            <span>正文</span>
            <span>来源</span>
            <span>操作</span>
          </div>

          <article v-for="item in accessItems" :key="item.id" class="logs-row logs-grid-access">
            <time class="logs-time">{{ formatDateTime(item.created_at) }}</time>

            <div class="logs-main">
              <div class="logs-path-line">
                <span class="logs-method">{{ item.method }}</span>
                <code :title="accessPath(item.request_uri)">{{ accessPath(item.request_uri) }}</code>
              </div>
              <p v-if="item.error_message" class="logs-error-line" :title="item.error_message">
                {{ trimMiddle(item.error_message, 120) }}
              </p>
            </div>

            <div>
              <span class="settings-status-pill" :class="statusClass(item.status_code)">
                {{ formatStatusCode(item.status_code) }}
              </span>
            </div>

            <strong class="logs-duration">{{ formatDuration(item.duration_ms) }}</strong>

            <div class="logs-body-cell">
              <span>响应 {{ bodyMeta(item.response_body_bytes, item.response_body_truncated) }}</span>
              <small>请求 {{ bodyMeta(item.request_body_bytes, item.request_body_truncated) }}</small>
            </div>

            <div class="logs-source">
              <strong>{{ item.client_ip || "-" }}</strong>
              <span :title="item.user_agent">{{ item.user_agent || "-" }}</span>
            </div>

            <button class="btn-soft-xs logs-action" @click="openAccessDetail(item)">详情</button>
          </article>

          <p v-if="!accessLoading && accessItems.length === 0" class="logs-empty">暂无外部访问日志</p>
          <p v-if="accessLoading" class="logs-empty">日志加载中...</p>
        </div>
      </template>

      <template v-else>

        <div class="logs-list-shell">
          <div class="logs-list-head logs-grid-tmdb">
            <span>时间</span>
            <span>上游路径</span>
            <span>状态</span>
            <span>耗时</span>
            <span>响应正文</span>
            <span>操作</span>
          </div>

          <article v-for="item in tmdbItems" :key="item.id" class="logs-row logs-grid-tmdb">
            <time class="logs-time">{{ formatDateTime(item.created_at) }}</time>

            <div class="logs-main">
              <div class="logs-path-line">
                <span class="logs-method">{{ item.method }}</span>
                <code :title="item.path || '-'">{{ item.path || "-" }}</code>
              </div>
              <p class="logs-host" :title="upstreamHost(item.url)">{{ upstreamHost(item.url) }}</p>
              <p v-if="item.error_message" class="logs-error-line" :title="item.error_message">
                {{ trimMiddle(item.error_message, 120) }}
              </p>
            </div>

            <div>
              <span class="settings-status-pill" :class="statusClass(item.status_code)">
                {{ formatStatusCode(item.status_code) }}
              </span>
            </div>

            <strong class="logs-duration">{{ formatDuration(item.duration_ms) }}</strong>

            <div class="logs-body-cell">
              <span>{{ bodyMeta(item.response_body_bytes, item.response_body_truncated) }}</span>
            </div>

            <button class="btn-soft-xs logs-action" @click="openTmdbDetail(item)">详情</button>
          </article>

          <p v-if="!tmdbLoading && tmdbItems.length === 0" class="logs-empty">暂无 TMDB 请求日志</p>
          <p v-if="tmdbLoading" class="logs-empty">日志加载中...</p>
        </div>
      </template>

      <div class="settings-pagination-row">
        <p>{{ currentTotalLabel }} 条，当前第 {{ currentPage }} / {{ currentTotalPages }} 页</p>
        <div class="flex items-center gap-2">
          <button
            class="btn-soft px-3 py-1.5 disabled:opacity-60"
            :disabled="currentBusy || currentPage <= 1"
            @click="goToPage(currentPage - 1)"
          >
            上一页
          </button>
          <button
            class="btn-soft px-3 py-1.5 disabled:opacity-60"
            :disabled="currentBusy || currentPage >= currentTotalPages"
            @click="goToPage(currentPage + 1)"
          >
            下一页
          </button>
        </div>
      </div>
    </section>

    <div
      v-if="detailVisible"
      class="fixed inset-0 z-[1300] flex items-center justify-center bg-black/55 p-3 sm:p-4"
      role="dialog"
      aria-modal="true"
      @click.self="closeDetail"
    >
      <div class="panel-glass logs-detail-modal max-h-[92vh] w-full max-w-6xl overflow-hidden rounded-lg">
        <div class="modal-header-dark">
          <div>
            <p class="section-label">{{ detailType === "access" ? "Access Detail" : "TMDB Detail" }}</p>
            <h4 class="text-base font-semibold">日志详情</h4>
          </div>
          <button class="btn-soft px-3 py-1.5" @click="closeDetail">关闭</button>
        </div>

        <div class="logs-detail-content">
          <p v-if="detailLoading" class="text-sm text-black/60">详情加载中...</p>

          <template v-if="activeDetail">
            <div class="logs-detail-overview">
              <div class="logs-detail-overview-main">
                <span class="logs-method">{{ activeDetail.method }}</span>
                <code :title="detailEndpointTitle(activeDetail)">{{ detailEndpointDisplay(activeDetail) }}</code>
              </div>

              <div class="logs-detail-overview-meta">
                <span><small>状态</small><strong :title="activeDetail.error_message || 'ok'">{{ formatStatusCode(activeDetail.status_code) }}</strong></span>
                <span><small>耗时</small><strong>{{ formatDuration(activeDetail.duration_ms) }}</strong></span>
                <span><small>时间</small><strong>{{ formatDateTime(activeDetail.created_at) }}</strong></span>
                <span>
                  <small>{{ detailType === "access" ? "来源" : "上游" }}</small>
                  <strong>{{ detailEndpointSource(activeDetail) }}</strong>
                </span>
                <span><small>请求正文</small><strong>{{ formatBytes(activeDetail.request_body_bytes) }}</strong></span>
              </div>
            </div>

            <pre v-if="activeDetail.request_body" class="settings-diff-pre logs-detail-request-body">{{
              bodyText(activeDetail.request_body)
            }}</pre>

            <div class="settings-detail-section logs-detail-response-section">
              <div class="settings-detail-section-header">
                <div>
                  <h5 class="text-sm font-semibold">响应正文</h5>
                  <p class="settings-note">{{ formatBytes(activeDetail.response_body_bytes) }}</p>
                </div>
              </div>
              <JsonFoldViewer :value="activeDetail.response_body" />
            </div>
          </template>
        </div>
      </div>
    </div>

    <div
      v-if="clearConfirmVisible"
      class="fixed inset-0 z-[1300] flex items-center justify-center bg-black/45 p-4"
      role="dialog"
      aria-modal="true"
      @click.self="closeClearConfirm"
    >
      <div class="panel-glass w-full max-w-md rounded-lg p-5">
        <h4 class="text-base font-semibold text-red-700">确认清空日志</h4>
        <p class="mt-2 text-sm text-black/70">
          将清空当前视图的{{ activeTab === "access" ? "外部访问日志" : "TMDB 请求日志" }}，清空后无法恢复。
        </p>

        <div class="mt-5 flex items-center justify-end gap-2">
          <button class="btn-soft disabled:opacity-60" :disabled="currentBusy" @click="closeClearConfirm">取消</button>
          <button class="btn-danger-soft disabled:opacity-60" :disabled="currentBusy" @click="clearCurrentLogs">
            {{ currentBusy ? "清空中..." : "确认清空" }}
          </button>
        </div>
      </div>
    </div>

    <ToastNotice :visible="toastVisible" :message="toastText" :tone="toastTone" @close="closeToastNotice" />
  </section>
</template>
