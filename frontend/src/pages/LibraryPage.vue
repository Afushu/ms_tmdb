<script setup lang="ts">
import GlassSelect from "@/components/GlassSelect.vue";
import LoadState from "@/components/common/LoadState.vue";
import ModalShell from "@/components/common/ModalShell.vue";
import LibraryDeleteDialog from "@/components/library/LibraryDeleteDialog.vue";
import LocalMediaCreateForm from "@/components/library/LocalMediaCreateForm.vue";
import type { MediaTab } from "@/components/library/types";
import { tmdbImg } from "@/api/tmdb";
import { useLibraryList } from "@/composables/useLibraryList";
import { useLocalMediaCreate } from "@/composables/useLocalMediaCreate";

let handleExternalTabChange: ((tab: MediaTab) => void) | undefined;

const {
  activeTab,
  viewMode,
  searchMode,
  keywordInput,
  keyword,
  loading,
  loadError,
  refreshError,
  items,
  total,
  page,
  initialLoading,
  deletingId,
  deleteModalVisible,
  pendingDeleteItem,
  searchModeOptions,
  loadData,
  switchTab,
  applySearch,
  resetSearch,
  totalPages,
  gotoPage,
  routeByItem,
  scheduleItemDetail,
  cancelItemDetail,
  touchItemDetail,
  openItemDetail,
  canDeleteItem,
  requestDeleteItem,
  closeDeleteModal,
  confirmDeleteItem,
} = useLibraryList({
  onExternalTabChange: (tab) => handleExternalTabChange?.(tab),
});

const {
  createPanelVisible,
  creating,
  uploadingKey,
  movieCreateForm,
  tvCreateForm,
  movieGenreOptions,
  tvGenreOptions,
  languageOptions,
  movieStatusOptions,
  tvStatusOptions,
  tvTypeOptions,
  createTitle,
  openCreatePanel,
  closeCreatePanel,
  onExternalTabChange,
  uploadCreateImage,
  submitCreate,
} = useLocalMediaCreate({
  activeTab,
  loadData,
});

handleExternalTabChange = onExternalTabChange;
</script>

<template>
  <section class="library-toolbar card">
    <div class="library-toolbar-main">
      <div class="library-toolbar-copy">
        <p class="section-label">本地库</p>
        <h2 class="library-toolbar-title">{{ activeTab === "movie" ? "电影库" : "剧集库" }}</h2>
        <p class="mt-1 text-sm text-black/55">管理本地缓存、手动新建条目，并快速进入详情页处理字段覆盖。</p>
      </div>
    </div>

    <div class="library-filter-panel">
      <div v-if="keyword" class="library-filter-header">
        <span v-if="keyword" class="badge">当前关键词：{{ keyword }}</span>
      </div>
      <div class="library-filter-form">
        <div class="library-search-combo">
          <div class="library-switch library-media-switch library-search-media-switch" role="group" aria-label="媒体类型">
            <button
              v-for="tab in [
                { key: 'movie', label: '电影' },
                { key: 'tv', label: '剧集' },
              ] as const"
              :key="tab.key"
              type="button"
              class="library-switch-btn"
              :class="activeTab === tab.key ? 'library-switch-btn-active' : ''"
              @click="switchTab(tab.key as MediaTab)"
            >
              {{ tab.label }}
            </button>
          </div>
          <input
            v-model="keywordInput"
            class="library-search-input text-sm"
            placeholder="快速筛选，支持 TMDB ID、片名或剧名；可切换包含/前缀匹配。"
            @keyup.enter="applySearch"
          />
        </div>
        <GlassSelect v-model="searchMode" :options="searchModeOptions" />
        <div class="library-filter-actions">
          <button class="btn-primary library-action-btn" @click="applySearch">搜索</button>
          <button class="btn-soft library-action-btn" @click="resetSearch">重置</button>
        </div>
      </div>
    </div>
  </section>

  <ModalShell
    :visible="createPanelVisible"
    :title="createTitle"
    variant="vben"
    max-width-class="max-w-4xl"
    content-class="modal-scroll-content max-h-[calc(86vh-120px)] overflow-y-auto px-5 py-4"
    @close="closeCreatePanel"
  >
    <LocalMediaCreateForm
      v-model:movie-form="movieCreateForm"
      v-model:tv-form="tvCreateForm"
      :media-type="activeTab"
      :movie-genre-options="movieGenreOptions"
      :tv-genre-options="tvGenreOptions"
      :uploading-key="uploadingKey"
      :language-options="languageOptions"
      :movie-status-options="movieStatusOptions"
      :tv-status-options="tvStatusOptions"
      :tv-type-options="tvTypeOptions"
      @upload="uploadCreateImage"
    />

    <template #footer>
      <button class="btn-soft library-modal-action-btn" @click="closeCreatePanel">取消</button>
      <button class="btn-primary disabled:opacity-60" :disabled="creating || uploadingKey !== ''" @click="submitCreate">
        {{ creating ? "创建中..." : "创建并进入详情" }}
      </button>
    </template>
  </ModalShell>

  <LibraryDeleteDialog
    :visible="deleteModalVisible"
    :deleting-id="deletingId"
    :pending-delete-item="pendingDeleteItem"
    :on-close="closeDeleteModal"
    :on-confirm="confirmDeleteItem"
  />

  <section class="library-list-summary mt-4">
    <p class="text-sm text-black/60">
      共 <strong>{{ total }}</strong> 条记录 · 第 {{ page }}/{{ totalPages() }} 页
    </p>
    <div class="library-list-controls">
      <div class="library-switch library-view-switch" role="group" aria-label="视图模式">
        <button
          type="button"
          class="library-switch-btn"
          :class="viewMode === 'grid' ? 'library-switch-btn-active' : ''"
          @click="viewMode = 'grid'"
        >
          卡片
        </button>
        <button
          type="button"
          class="library-switch-btn"
          :class="viewMode === 'table' ? 'library-switch-btn-active' : ''"
          @click="viewMode = 'table'"
        >
          表格
        </button>
      </div>

      <button type="button" class="btn-primary library-list-create-btn" @click="openCreatePanel">
        {{ createTitle }}
      </button>
    </div>
  </section>

  <div
    v-if="refreshError && !loadError"
    class="logs-refresh-error mt-4"
    role="status"
    aria-live="polite"
  >
    <span>刷新失败：{{ refreshError }}</span>
    <button type="button" class="btn-soft-xs" :disabled="loading" @click="loadData">重试</button>
  </div>

  <LoadState
    class="mt-4"
    :loading="initialLoading"
    :error="loadError"
    loading-text="媒体库加载中..."
    @retry="loadData"
  >
    <section v-if="viewMode === 'grid' && items.length" class="poster-grid">
      <div v-for="item in items" :key="item.tmdb_id" class="poster-card group relative">
        <button
          v-if="canDeleteItem(item)"
          type="button"
          class="library-delete-btn pointer-events-none absolute right-2 top-2 z-20 opacity-0 group-hover:pointer-events-auto group-hover:opacity-100 group-focus-within:pointer-events-auto group-focus-within:opacity-100"
          :class="deletingId === item.tmdb_id ? 'opacity-100 pointer-events-none' : ''"
          :disabled="deletingId === item.tmdb_id"
          :title="deletingId === item.tmdb_id ? '删除中' : '删除本地数据'"
          @click.stop="requestDeleteItem(item)"
        >
          <span v-if="deletingId === item.tmdb_id" class="text-[11px]">...</span>
          <svg
            v-else
            viewBox="0 0 24 24"
            class="h-4 w-4 fill-none stroke-current"
            stroke-width="1.8"
            aria-hidden="true"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M4 7h16M10 11v6M14 11v6M6 7l1 12a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-12M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"
            />
          </svg>
        </button>
        <RouterLink
          :to="routeByItem(item)"
          class="block"
          @pointerenter="scheduleItemDetail(item)"
          @pointerleave="cancelItemDetail(item)"
          @focus="scheduleItemDetail(item)"
          @blur="cancelItemDetail(item)"
          @touchstart.passive="touchItemDetail(item)"
        >
          <img
            :src="tmdbImg(item.poster_path, 'w342')"
            :srcset="`${tmdbImg(item.poster_path, 'w342')} 1x, ${tmdbImg(item.poster_path, 'w500')} 2x`"
            :alt="item.title || item.name"
            class="poster-img"
            loading="lazy"
          />
          <div class="poster-info">
            <p class="truncate text-sm font-medium">{{ item.title || item.name }}</p>
            <p class="poster-meta">
              <span class="rating-badge">{{ (item.vote_average ?? 0).toFixed(1) }} 分</span>
              <span>{{ (item.release_date || item.first_air_date || "").slice(0, 4) }}</span>
            </p>
            <span v-if="item.tmdb_id < 0" class="chip-local-new mt-1 text-[10px]"> 本地新建 </span>
            <span v-else-if="item.is_modified" class="chip-modified mt-1 text-[10px]"> 已修改 </span>
          </div>
        </RouterLink>
      </div>
    </section>

    <section v-else-if="viewMode === 'grid'" class="empty-state mt-4">
      暂无本地数据，可以尝试切换分类、重置搜索，或新建一条本地记录。
    </section>

    <section v-else class="table-shell library-table-shell">
      <table class="min-w-full text-left text-sm library-table">
        <thead class="table-head text-xs uppercase tracking-wide text-black/60">
          <tr>
            <th class="px-4 py-3 library-col-secondary">TMDB ID</th>
            <th class="px-4 py-3 library-col-primary">名称</th>
            <th class="px-4 py-3 library-col-primary">评分</th>
            <th class="px-4 py-3 library-col-primary">日期</th>
            <th class="px-4 py-3 library-col-secondary">类型</th>
            <th class="px-4 py-3 library-col-secondary">状态</th>
            <th class="px-4 py-3 library-col-actions">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.tmdb_id" class="table-row-hover">
            <td class="px-4 py-3 library-col-secondary">{{ item.tmdb_id }}</td>
            <td class="px-4 py-3 library-col-primary">
              <p class="font-medium">{{ item.title || item.name }}</p>
              <p class="text-xs text-black/50">{{ item.original_title || item.original_name }}</p>
            </td>
            <td class="px-4 py-3 library-col-primary">
              <span class="rating-badge">{{ (item.vote_average ?? 0).toFixed(1) }} 分</span>
            </td>
            <td class="px-4 py-3 library-col-primary">{{ item.release_date || item.first_air_date || "-" }}</td>
            <td class="px-4 py-3 library-col-secondary">
              <span class="text-xs text-black/70">
                {{ Array.isArray(item.genre_names) && item.genre_names.length ? item.genre_names.join(" / ") : "-" }}
              </span>
            </td>
            <td class="px-4 py-3 library-col-secondary">
              <span v-if="item.tmdb_id < 0" class="chip-local-new"> 本地新建 </span>
              <span v-else-if="item.is_modified" class="chip-modified"> 已修改 </span>
              <span v-else class="text-xs text-black/45">未修改</span>
            </td>
            <td class="px-4 py-3 library-col-actions">
              <div class="library-table-actions flex items-center gap-2">
                <button
                  class="table-action-btn table-action-btn-soft"
                  type="button"
                  data-tooltip="查看详情"
                  aria-label="查看详情"
                  @pointerenter="scheduleItemDetail(item)"
                  @pointerleave="cancelItemDetail(item)"
                  @focus="scheduleItemDetail(item)"
                  @blur="cancelItemDetail(item)"
                  @touchstart.passive="touchItemDetail(item)"
                  @click="openItemDetail(item)"
                >
                  <svg
                    viewBox="0 0 24 24"
                    class="h-4 w-4 fill-none stroke-current"
                    stroke-width="1.8"
                    aria-hidden="true"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M2.5 12s3.5-6 9.5-6 9.5 6 9.5 6-3.5 6-9.5 6-9.5-6-9.5-6Z"
                    />
                    <circle cx="12" cy="12" r="2.6" />
                  </svg>
                </button>
                <button
                  v-if="canDeleteItem(item)"
                  class="table-action-btn table-action-btn-danger"
                  type="button"
                  :data-tooltip="deletingId === item.tmdb_id ? '删除中...' : '删除'"
                  :aria-label="deletingId === item.tmdb_id ? '删除中' : '删除'"
                  :disabled="deletingId === item.tmdb_id"
                  @click="requestDeleteItem(item)"
                >
                  <span v-if="deletingId === item.tmdb_id" class="text-[11px]">...</span>
                  <svg
                    v-else
                    viewBox="0 0 24 24"
                    class="h-4 w-4 fill-none stroke-current"
                    stroke-width="1.8"
                    aria-hidden="true"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M4 7h16M10 11v6M14 11v6M6 7l1 12a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-12M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"
                    />
                  </svg>
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="items.length === 0">
            <td colspan="7" class="px-4 py-8 text-center text-black/50 library-table-empty">无数据</td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="mt-6 flex items-center justify-center gap-2">
      <button class="btn-soft px-3 py-1.5 disabled:opacity-40" :disabled="page <= 1" @click="gotoPage(page - 1)">
        上一页
      </button>
      <span class="px-3 text-sm text-black/60">{{ page }} / {{ totalPages() }}</span>
      <button
        class="btn-soft px-3 py-1.5 disabled:opacity-40"
        :disabled="page >= totalPages()"
        @click="gotoPage(page + 1)"
      >
        下一页
      </button>
    </section>
  </LoadState>
</template>
