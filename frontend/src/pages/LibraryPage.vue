<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter, type LocationQueryRaw } from "vue-router";
import GlassSelect from "@/components/GlassSelect.vue";
import ModalShell from "@/components/common/ModalShell.vue";
import LocalMediaCreateForm from "@/components/library/LocalMediaCreateForm.vue";
import type { LocalMovieCreateForm, LocalTVCreateForm, MediaTab, UploadingKey } from "@/components/library/types";
import { prefetchMediaDetail } from "@/api/prefetch";
import {
  createMovie,
  createTV,
  deleteMovie,
  deleteTV,
  listMovies,
  listTV,
  uploadAdminImage,
  type AdminListResp,
  type AdminMovieListItem,
  type AdminTVListItem,
  type AdminCreateMoviePayload,
  type AdminCreateTVPayload,
} from "@/api/admin";
import { tmdbImg } from "@/api/tmdb";
import { getMovieGenreList } from "@/api/movie";
import { getTVGenreList } from "@/api/tv";
import { movieStatusOptions, tvStatusOptions, tvTypeOptions } from "@/constants/mediaStatus";
import { normalizeGenreOptions, type GenreOption } from "@/utils/mediaNormalizers";
import { isSameQuery, readQueryString } from "@/utils/routeQuery";

type ViewMode = "grid" | "table";
type SearchMode = "contains" | "prefix";

type LibraryListItem = AdminMovieListItem | AdminTVListItem;

const route = useRoute();
const router = useRouter();
const activeTab = ref<MediaTab>(normalizeTab(route.query.tab));
const viewMode = ref<ViewMode>(normalizeViewMode(route.query.view));
const searchMode = ref<SearchMode>(normalizeSearchMode(route.query.mode));
const keywordInput = ref(readQueryString(route.query.q));
const keyword = ref(readQueryString(route.query.q));
const loading = ref(false);
const items = ref<LibraryListItem[]>([]);
const total = ref(0);
const page = ref(normalizePage(route.query.page));
const pageSize = 20;

const createPanelVisible = ref(false);
const creating = ref(false);
const createError = ref("");
const uploadingKey = ref<UploadingKey>("");
const deletingId = ref<number | null>(null);
const deleteModalVisible = ref(false);
const pendingDeleteItem = ref<LibraryListItem | null>(null);
const movieCreateForm = ref<LocalMovieCreateForm>(emptyMovieForm());
const tvCreateForm = ref<LocalTVCreateForm>(emptyTVForm());
const movieGenreOptions = ref<GenreOption[]>([]);
const tvGenreOptions = ref<GenreOption[]>([]);

const languageOptions = [
  { label: "中文 (zh-CN)", value: "zh-CN" },
  { label: "英语 (en-US)", value: "en-US" },
  { label: "日语 (ja-JP)", value: "ja-JP" },
  { label: "韩语 (ko-KR)", value: "ko-KR" },
] as const;

const searchModeOptions = [
  { label: "模糊包含", value: "contains" },
  { label: "前缀匹配", value: "prefix" },
] as const;

const createTitle = computed(() => (activeTab.value === "movie" ? "新建本地电影" : "新建本地剧集"));
let previousBodyOverflow = "";
let loadReqSeq = 0;

function normalizeTab(value: unknown): MediaTab {
  return readQueryString(value) === "tv" ? "tv" : "movie";
}

function normalizeViewMode(value: unknown): ViewMode {
  return readQueryString(value) === "table" ? "table" : "grid";
}

function normalizeSearchMode(value: unknown): SearchMode {
  return readQueryString(value) === "prefix" ? "prefix" : "contains";
}

function normalizePage(value: unknown): number {
  const parsed = Number(readQueryString(value));
  return Number.isInteger(parsed) && parsed > 0 ? parsed : 1;
}

function buildLibraryQuery(): LocationQueryRaw {
  const nextQuery: LocationQueryRaw = {};
  if (activeTab.value !== "movie") nextQuery.tab = activeTab.value;
  if (page.value > 1) nextQuery.page = String(page.value);
  if (keyword.value) nextQuery.q = keyword.value;
  if (searchMode.value !== "contains") nextQuery.mode = searchMode.value;
  if (viewMode.value !== "grid") nextQuery.view = viewMode.value;
  return nextQuery;
}

function isSameLibraryQuery(nextQuery: LocationQueryRaw): boolean {
  return isSameQuery(route.query, nextQuery);
}

function syncLibraryQuery() {
  const nextQuery = buildLibraryQuery();
  if (isSameLibraryQuery(nextQuery)) return;
  void router.replace({ path: "/library", query: nextQuery });
}

function normalizeListItem(item: unknown): LibraryListItem | null {
  if (!item || typeof item !== "object") return null;
  const raw = item as Record<string, unknown>;
  const tmdbId = Number(raw.tmdb_id);
  if (!Number.isFinite(tmdbId)) return null;

  const genres = Array.isArray(raw.genre_names)
    ? raw.genre_names.map((v) => String(v ?? "").trim()).filter(Boolean)
    : [];

  if (typeof raw.name === "string") {
    return {
      tmdb_id: tmdbId,
      name: String(raw.name ?? ""),
      original_name: String(raw.original_name ?? ""),
      poster_path: String(raw.poster_path ?? ""),
      vote_average: Number(raw.vote_average ?? 0),
      first_air_date: String(raw.first_air_date ?? ""),
      number_of_seasons: Number(raw.number_of_seasons ?? 0),
      number_of_episodes: Number(raw.number_of_episodes ?? 0),
      popularity: Number(raw.popularity ?? 0),
      is_modified: Boolean(raw.is_modified),
      genre_names: genres,
      id: Number.isFinite(Number(raw.id)) ? Number(raw.id) : tmdbId,
      media_type: "tv",
    };
  }

  return {
    tmdb_id: tmdbId,
    title: String(raw.title ?? ""),
    original_title: String(raw.original_title ?? ""),
    poster_path: String(raw.poster_path ?? ""),
    vote_average: Number(raw.vote_average ?? 0),
    release_date: String(raw.release_date ?? ""),
    popularity: Number(raw.popularity ?? 0),
    is_modified: Boolean(raw.is_modified),
    genre_names: genres,
    id: Number.isFinite(Number(raw.id)) ? Number(raw.id) : tmdbId,
    media_type: "movie",
  };
}

function normalizeListResults(raw: unknown): LibraryListItem[] {
  if (!Array.isArray(raw)) return [];
  return raw.map((item) => normalizeListItem(item)).filter((item): item is LibraryListItem => item !== null);
}

function normalizeListResponse(raw: unknown): AdminListResp<LibraryListItem> {
  const payload = raw && typeof raw === "object" ? (raw as Record<string, unknown>) : {};
  return {
    total: Number(payload.total ?? 0),
    page: Number(payload.page ?? 1),
    page_size: Number(payload.page_size ?? pageSize),
    results: normalizeListResults(payload.results),
  };
}

function emptyMovieForm(): LocalMovieCreateForm {
  return {
    title: "",
    original_title: "",
    genre_names: [],
    release_date: "",
    status: "Released",
    runtime: "",
    original_language: "zh-CN",
    poster_path: "",
    backdrop_path: "",
    vote_average: "",
    popularity: "",
    overview: "",
  };
}

function emptyTVForm(): LocalTVCreateForm {
  return {
    name: "",
    original_name: "",
    genre_names: [],
    first_air_date: "",
    status: "Returning Series",
    type: "Scripted",
    number_of_seasons: "",
    number_of_episodes: "",
    original_language: "zh-CN",
    poster_path: "",
    backdrop_path: "",
    vote_average: "",
    popularity: "",
    overview: "",
  };
}

async function loadData() {
  const requestSeq = ++loadReqSeq;
  const targetTab = activeTab.value;
  const targetPage = page.value;
  const targetKeyword = keyword.value;
  const targetSearchMode = searchMode.value;
  loading.value = true;
  try {
    const resp =
      targetTab === "movie"
        ? await listMovies(targetPage, pageSize, targetKeyword, targetSearchMode)
        : await listTV(targetPage, pageSize, targetKeyword, targetSearchMode);
    if (requestSeq !== loadReqSeq) {
      return;
    }
    const normalized = normalizeListResponse(resp.data);
    items.value = normalized.results;
    total.value = normalized.total;
  } catch {
    /* handled by global toast */
  } finally {
    if (requestSeq === loadReqSeq) {
      loading.value = false;
    }
  }
}

function switchTab(tab: MediaTab) {
  if (tab === activeTab.value) return;
  activeTab.value = tab;
  page.value = 1;
}

function applySearch() {
  keyword.value = keywordInput.value.trim();
  page.value = 1;
}

function resetSearch() {
  keywordInput.value = "";
  keyword.value = "";
  page.value = 1;
}

function resetCreateForm() {
  movieCreateForm.value = emptyMovieForm();
  tvCreateForm.value = emptyTVForm();
  createError.value = "";
  uploadingKey.value = "";
}

async function loadMovieGenreOptions() {
  try {
    const resp = await getMovieGenreList();
    movieGenreOptions.value = normalizeGenreOptions(resp.data?.genres);
  } catch {
    movieGenreOptions.value = [];
  }
}

async function loadTVGenreOptions() {
  try {
    const resp = await getTVGenreList();
    tvGenreOptions.value = normalizeGenreOptions(resp.data?.genres);
  } catch {
    tvGenreOptions.value = [];
  }
}

function openCreatePanel() {
  createPanelVisible.value = true;
  resetCreateForm();
  if (activeTab.value === "movie") {
    void loadMovieGenreOptions();
  } else {
    void loadTVGenreOptions();
  }
}

function closeCreatePanel() {
  createPanelVisible.value = false;
  resetCreateForm();
}

function handleModalKeydown(event: KeyboardEvent) {
  if (event.key !== "Escape") return;
  if (deleteModalVisible.value) {
    closeDeleteModal();
    return;
  }
  if (createPanelVisible.value) {
    closeCreatePanel();
  }
}

function totalPages() {
  return Math.ceil(total.value / pageSize) || 1;
}

function gotoPage(p: number) {
  if (p < 1 || p > totalPages()) return;
  page.value = p;
}

function routeByItem(item: LibraryListItem) {
  if (activeTab.value === "movie") return `/movie/${item.tmdb_id}`;
  return `/tv/${item.tmdb_id}`;
}

function prefetchItemDetail(item: LibraryListItem) {
  prefetchMediaDetail(activeTab.value, Number(item.tmdb_id));
}

function openItemDetail(item: LibraryListItem) {
  void router.push(routeByItem(item));
}

function canDeleteItem(item: LibraryListItem): boolean {
  const id = Number(item?.tmdb_id);
  return Number.isInteger(id) && id !== 0;
}

function parseOptionalInt(raw: string): number | undefined {
  const text = raw.trim();
  if (!text) return undefined;
  const value = Number(text);
  if (!Number.isFinite(value)) return undefined;
  return Math.trunc(value);
}

function parseOptionalFloat(raw: string): number | undefined {
  const text = raw.trim();
  if (!text) return undefined;
  const value = Number(text);
  if (!Number.isFinite(value)) return undefined;
  return value;
}

async function uploadCreateImage(mediaType: MediaTab, field: "poster_path" | "backdrop_path", event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;

  const key = `${mediaType}_${field}` as UploadingKey;
  uploadingKey.value = key;
  createError.value = "";
  try {
    const resp = await uploadAdminImage(file);
    const path = String(resp.data?.path ?? "").trim();
    if (!path) {
      throw new Error("上传成功但未返回图片路径");
    }
    if (mediaType === "movie") {
      movieCreateForm.value[field] = path;
    } else {
      tvCreateForm.value[field] = path;
    }
  } catch {
    /* handled by global toast */
  } finally {
    uploadingKey.value = "";
    input.value = "";
  }
}

async function submitCreate() {
  createError.value = "";

  if (activeTab.value === "movie") {
    const title = movieCreateForm.value.title.trim();
    if (!title) {
      createError.value = "电影标题不能为空";
      return;
    }

    const runtime = parseOptionalInt(movieCreateForm.value.runtime);
    if (movieCreateForm.value.runtime.trim() && runtime === undefined) {
      createError.value = "时长必须是数字";
      return;
    }
    const voteAverage = parseOptionalFloat(movieCreateForm.value.vote_average);
    if (movieCreateForm.value.vote_average.trim() && voteAverage === undefined) {
      createError.value = "评分必须是数字";
      return;
    }
    const popularity = parseOptionalFloat(movieCreateForm.value.popularity);
    if (movieCreateForm.value.popularity.trim() && popularity === undefined) {
      createError.value = "热度必须是数字";
      return;
    }

    const payload: AdminCreateMoviePayload = {
      title,
      original_title: movieCreateForm.value.original_title.trim(),
      release_date: movieCreateForm.value.release_date.trim(),
      status: movieCreateForm.value.status.trim(),
      original_language: movieCreateForm.value.original_language.trim(),
      poster_path: movieCreateForm.value.poster_path.trim(),
      backdrop_path: movieCreateForm.value.backdrop_path.trim(),
      overview: movieCreateForm.value.overview.trim(),
      genre_names: movieCreateForm.value.genre_names,
    };
    if (runtime !== undefined) payload.runtime = runtime;
    if (voteAverage !== undefined) payload.vote_average = voteAverage;
    if (popularity !== undefined) payload.popularity = popularity;

    creating.value = true;
    try {
      const resp = await createMovie(payload);
      const createdID = Number(resp.data?.tmdb_id);
      if (!Number.isInteger(createdID)) {
        throw new Error("创建成功但未返回有效 ID");
      }
      closeCreatePanel();
      await loadData();
      await router.push(`/movie/${createdID}`);
    } catch {
      /* handled by global toast */
    } finally {
      creating.value = false;
    }
    return;
  }

  const name = tvCreateForm.value.name.trim();
  if (!name) {
    createError.value = "剧集名称不能为空";
    return;
  }

  const seasons = parseOptionalInt(tvCreateForm.value.number_of_seasons);
  if (tvCreateForm.value.number_of_seasons.trim() && seasons === undefined) {
    createError.value = "季数必须是数字";
    return;
  }
  const episodes = parseOptionalInt(tvCreateForm.value.number_of_episodes);
  if (tvCreateForm.value.number_of_episodes.trim() && episodes === undefined) {
    createError.value = "集数必须是数字";
    return;
  }
  const voteAverage = parseOptionalFloat(tvCreateForm.value.vote_average);
  if (tvCreateForm.value.vote_average.trim() && voteAverage === undefined) {
    createError.value = "评分必须是数字";
    return;
  }
  const popularity = parseOptionalFloat(tvCreateForm.value.popularity);
  if (tvCreateForm.value.popularity.trim() && popularity === undefined) {
    createError.value = "热度必须是数字";
    return;
  }

  const payload: AdminCreateTVPayload = {
    name,
    original_name: tvCreateForm.value.original_name.trim(),
    first_air_date: tvCreateForm.value.first_air_date.trim(),
    status: tvCreateForm.value.status.trim(),
    type: tvCreateForm.value.type.trim(),
    original_language: tvCreateForm.value.original_language.trim(),
    poster_path: tvCreateForm.value.poster_path.trim(),
    backdrop_path: tvCreateForm.value.backdrop_path.trim(),
    overview: tvCreateForm.value.overview.trim(),
    genre_names: tvCreateForm.value.genre_names,
  };
  if (seasons !== undefined) payload.number_of_seasons = seasons;
  if (episodes !== undefined) payload.number_of_episodes = episodes;
  if (voteAverage !== undefined) payload.vote_average = voteAverage;
  if (popularity !== undefined) payload.popularity = popularity;

  creating.value = true;
  try {
    const resp = await createTV(payload);
    const createdID = Number(resp.data?.tmdb_id);
    if (!Number.isInteger(createdID)) {
      throw new Error("创建成功但未返回有效 ID");
    }
    closeCreatePanel();
    await loadData();
    await router.push(`/tv/${createdID}`);
  } catch {
    /* handled by global toast */
  } finally {
    creating.value = false;
  }
}

function requestDeleteItem(item: LibraryListItem) {
  const id = Number(item?.tmdb_id);
  if (!Number.isInteger(id) || id === 0) {
    return;
  }
  pendingDeleteItem.value = item;
  deleteModalVisible.value = true;
}

function closeDeleteModal() {
  deleteModalVisible.value = false;
  pendingDeleteItem.value = null;
}

async function confirmDeleteItem() {
  const item = pendingDeleteItem.value;
  if (!item) return;
  const id = Number(item.tmdb_id);
  if (!Number.isInteger(id) || id === 0) return;

  deletingId.value = id;
  try {
    if (activeTab.value === "movie") {
      await deleteMovie(id);
    } else {
      await deleteTV(id);
    }

    if (items.value.length <= 1 && page.value > 1) {
      page.value = page.value - 1;
      closeDeleteModal();
      return;
    }
    await loadData();
    closeDeleteModal();
  } catch {
    /* handled by global toast */
  } finally {
    deletingId.value = null;
  }
}

watch(
  () => route.query,
  (query) => {
    const nextTab = normalizeTab(query.tab);
    if (nextTab !== activeTab.value) {
      activeTab.value = nextTab;
      page.value = 1;
      createPanelVisible.value = false;
      createError.value = "";
      if (nextTab === "movie") {
        void loadMovieGenreOptions();
      } else {
        void loadTVGenreOptions();
      }
    }

    const nextPage = normalizePage(query.page);
    if (nextPage !== page.value) {
      page.value = nextPage;
    }

    const nextKeyword = readQueryString(query.q);
    if (nextKeyword !== keyword.value) {
      keyword.value = nextKeyword;
      keywordInput.value = nextKeyword;
    }

    const nextSearchMode = normalizeSearchMode(query.mode);
    if (nextSearchMode !== searchMode.value) {
      searchMode.value = nextSearchMode;
    }

    const nextViewMode = normalizeViewMode(query.view);
    if (nextViewMode !== viewMode.value) {
      viewMode.value = nextViewMode;
    }
  },
);

watch(
  () => createPanelVisible.value || deleteModalVisible.value,
  (visible) => {
    if (visible) {
      previousBodyOverflow = document.body.style.overflow;
      document.body.style.overflow = "hidden";
      window.addEventListener("keydown", handleModalKeydown);
      return;
    }

    document.body.style.overflow = previousBodyOverflow;
    window.removeEventListener("keydown", handleModalKeydown);
  },
);

onBeforeUnmount(() => {
  document.body.style.overflow = previousBodyOverflow;
  window.removeEventListener("keydown", handleModalKeydown);
});

watch([activeTab, page, keyword, searchMode], loadData);
watch([activeTab, page, keyword, searchMode, viewMode], syncLibraryQuery);
onMounted(loadData);
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

  <div
    v-if="deleteModalVisible"
    class="fixed inset-0 z-[1300] flex items-center justify-center p-4"
    role="dialog"
    aria-modal="true"
  >
    <div class="absolute inset-0 bg-black/65 backdrop-blur-[2px]" @click="closeDeleteModal" />
    <section class="panel-glass relative z-10 w-full max-w-md rounded-lg p-5">
      <h3 class="text-base font-semibold text-ink">确认删除</h3>
      <p class="mt-2 text-sm text-black/70">
        将删除本地数据：
        <span class="font-medium text-black">{{
          pendingDeleteItem?.title || pendingDeleteItem?.name || `ID ${pendingDeleteItem?.tmdb_id ?? ""}`
        }}</span>
      </p>
      <p class="mt-1 text-xs text-black/55">删除后不可恢复。</p>
      <div class="mt-5 flex justify-end gap-2">
        <button class="btn-soft" :disabled="deletingId !== null" @click="closeDeleteModal">取消</button>
        <button class="btn-danger-soft disabled:opacity-60" :disabled="deletingId !== null" @click="confirmDeleteItem">
          {{ deletingId !== null ? "删除中..." : "确认删除" }}
        </button>
      </div>
    </section>
  </div>

  <p v-if="loading" class="card mt-4 text-sm text-black/60">加载中...</p>

  <template v-else>
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

    <section v-if="viewMode === 'grid' && items.length" class="mt-4 poster-grid">
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
          @mouseenter="prefetchItemDetail(item)"
          @focus="prefetchItemDetail(item)"
          @touchstart.passive="prefetchItemDetail(item)"
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

    <section v-else class="table-shell">
      <table class="min-w-full text-left text-sm">
        <thead class="table-head text-xs uppercase tracking-wide text-black/60">
          <tr>
            <th class="px-4 py-3">TMDB ID</th>
            <th class="px-4 py-3">名称</th>
            <th class="px-4 py-3">评分</th>
            <th class="px-4 py-3">日期</th>
            <th class="px-4 py-3">类型</th>
            <th class="px-4 py-3">状态</th>
            <th class="px-4 py-3">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.tmdb_id" class="table-row-hover">
            <td class="px-4 py-3">{{ item.tmdb_id }}</td>
            <td class="px-4 py-3">
              <p class="font-medium">{{ item.title || item.name }}</p>
              <p class="text-xs text-black/50">{{ item.original_title || item.original_name }}</p>
            </td>
            <td class="px-4 py-3">
              <span class="rating-badge">{{ (item.vote_average ?? 0).toFixed(1) }} 分</span>
            </td>
            <td class="px-4 py-3">{{ item.release_date || item.first_air_date || "-" }}</td>
            <td class="px-4 py-3">
              <span class="text-xs text-black/70">
                {{ Array.isArray(item.genre_names) && item.genre_names.length ? item.genre_names.join(" / ") : "-" }}
              </span>
            </td>
            <td class="px-4 py-3">
              <span v-if="item.tmdb_id < 0" class="chip-local-new"> 本地新建 </span>
              <span v-else-if="item.is_modified" class="chip-modified"> 已修改 </span>
              <span v-else class="text-xs text-black/45">未修改</span>
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-2">
                <button
                  class="table-action-btn table-action-btn-soft"
                  type="button"
                  data-tooltip="查看详情"
                  aria-label="查看详情"
                  @mouseenter="prefetchItemDetail(item)"
                  @focus="prefetchItemDetail(item)"
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
            <td colspan="7" class="px-4 py-8 text-center text-black/50">无数据</td>
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
  </template>
</template>
