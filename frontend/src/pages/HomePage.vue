<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter, type LocationQueryRaw } from "vue-router";
import { getHomeDashboard, type AdminHomeMediaItem } from "@/api/admin";
import { prefetchMediaDetail } from "@/api/prefetch";
import { searchByType, type SearchType } from "@/api/search";
import { tmdbImg } from "@/api/tmdb";
import GlassSelect from "@/components/GlassSelect.vue";
import SearchResultList from "@/components/SearchResultList.vue";
import type { SearchResultItem } from "@/types/media";
import { buildSearchQuery, normalizeSearchType, readQueryString, searchTypeOptions } from "@/utils/routeSearch";
import { isSameQuery } from "@/utils/routeQuery";

const route = useRoute();
const router = useRouter();

type HomeMediaViewMode = "immersive" | "compact" | "list";
type HomeMediaSectionKey = "latest" | "hot";

type HomeMediaSection = {
  key: HomeMediaSectionKey;
  title: string;
  emptyText: string;
  items: AdminHomeMediaItem[];
};

const homeMediaViewStorageKey = "ms_tmdb_home_media_view_mode";
const homeDashboardBaseLimit = 15;
const homeMediaViewOptions: Array<{ label: string; value: HomeMediaViewMode }> = [
  { label: "沉浸", value: "immersive" },
  { label: "紧凑", value: "compact" },
  { label: "列表", value: "list" },
];

const loading = ref(false);
const latestMedia = ref<AdminHomeMediaItem[]>([]);
const hotMedia = ref<AdminHomeMediaItem[]>([]);
const searchQuery = ref(readQueryString(route.query.q));
const searchType = ref<SearchType>(normalizeSearchType(route.query.type));
const searching = ref(false);
const searchResults = ref<SearchResultItem[]>([]);
const homeMediaViewMode = ref<HomeMediaViewMode>(readHomeMediaViewMode());
const homeMediaColumnCount = ref(5);
let searchReqSeq = 0;
let homeDashboardReady = false;

const hasRouteQuery = computed(() => Boolean(readQueryString(route.query.q)));
const showSearchResults = computed(() => hasRouteQuery.value || searchResults.value.length > 0);
const homeDashboardLimit = computed(() => {
  const columns = Math.max(homeMediaColumnCount.value, 1);
  return Math.ceil(homeDashboardBaseLimit / columns) * columns;
});
const homeMediaSections = computed<HomeMediaSection[]>(() => [
  {
    key: "latest",
    title: "最新入库",
    emptyText: "本地库暂无数据。",
    items: latestMedia.value,
  },
  {
    key: "hot",
    title: "访问热度",
    emptyText: "暂无本地访问热度记录。",
    items: hotMedia.value,
  },
]);
const resultSummary = computed(() => {
  if (searching.value) return "检索中";
  if (searchResults.value.length) return `${searchResults.value.length} 条匹配`;
  return hasRouteQuery.value ? "暂无结果" : "等待检索";
});

function normalizeHomeMediaViewMode(value: unknown): HomeMediaViewMode {
  return value === "immersive" || value === "list" ? value : "compact";
}

function readHomeMediaViewMode(): HomeMediaViewMode {
  if (typeof window === "undefined") return "compact";
  try {
    return normalizeHomeMediaViewMode(window.localStorage.getItem(homeMediaViewStorageKey));
  } catch {
    return "compact";
  }
}

function setHomeMediaViewMode(mode: HomeMediaViewMode) {
  homeMediaViewMode.value = mode;
  try {
    window.localStorage.setItem(homeMediaViewStorageKey, mode);
  } catch {
    /* local preference is optional */
  }
  syncHomeMediaColumnCount();
}

function displayTitle(item: AdminHomeMediaItem): string {
  return item.title || item.original_title || `ID ${item.tmdb_id}`;
}

function mediaTypeLabel(mediaType: string): string {
  return mediaType === "tv" ? "剧集" : "电影";
}

function mediaRoute(item: AdminHomeMediaItem): string {
  return item.media_type === "tv" ? `/tv/${item.tmdb_id}` : `/movie/${item.tmdb_id}`;
}

function yearText(value: string): string {
  return value ? value.slice(0, 4) : "";
}

function ratingText(item: AdminHomeMediaItem): string {
  return `${item.vote_average.toFixed(1)} 分`;
}

function visitText(item: AdminHomeMediaItem): string {
  return `访问 ${item.visit_count} 次`;
}

function overviewText(item: AdminHomeMediaItem): string {
  return item.overview.trim() || "暂无简介";
}

async function loadData() {
  loading.value = true;
  try {
    const resp = await getHomeDashboard(homeDashboardLimit.value);
    latestMedia.value = resp.data?.latest ?? [];
    hotMedia.value = resp.data?.hot ?? [];
  } catch {
    /* handled by global toast */
  } finally {
    loading.value = false;
  }
}

function currentHomeMediaColumnCount(): number {
  if (typeof window === "undefined") {
    return homeMediaViewMode.value === "compact" ? 5 : homeMediaViewMode.value === "immersive" ? 3 : 1;
  }

  if (homeMediaViewMode.value === "compact") {
    if (window.innerWidth >= 768) return 5;
    if (window.innerWidth >= 640) return 3;
    return 2;
  }

  if (homeMediaViewMode.value === "immersive") {
    if (window.innerWidth >= 1024) return 3;
    if (window.innerWidth >= 640) return 2;
  }

  return 1;
}

function syncHomeMediaColumnCount() {
  homeMediaColumnCount.value = currentHomeMediaColumnCount();
}

function isSameSearchQuery(nextQuery: LocationQueryRaw): boolean {
  return isSameQuery(route.query, nextQuery);
}

async function runSearch(targetType: SearchType, targetQuery: string) {
  if (!targetQuery) {
    searchReqSeq++;
    searchResults.value = [];
    searching.value = false;
    return;
  }

  const requestSeq = ++searchReqSeq;
  searching.value = true;
  searchResults.value = [];
  try {
    const resp = await searchByType(targetType, targetQuery, 1);
    if (requestSeq !== searchReqSeq) return;
    searchResults.value = resp.data?.results ?? [];
  } catch {
    /* handled by global toast */
  } finally {
    if (requestSeq === searchReqSeq) {
      searching.value = false;
    }
  }
}

async function handleHomeSearch() {
  const trimmedQuery = searchQuery.value.trim();
  const targetType = searchType.value;
  if (!trimmedQuery) {
    searchReqSeq++;
    searchResults.value = [];
    searching.value = false;
    if (route.fullPath !== "/") {
      await router.replace("/");
    }
    return;
  }

  const nextQuery = buildSearchQuery(targetType, trimmedQuery);
  if (!isSameSearchQuery(nextQuery)) {
    await router.replace({ path: "/", query: nextQuery });
    return;
  }
  await runSearch(targetType, trimmedQuery);
}

function prefetchListItem(mediaType: "movie" | "tv", id: number | undefined) {
  prefetchMediaDetail(mediaType, Number(id));
}

watch(
  () => route.query,
  (routeQuery) => {
    const nextQuery = readQueryString(routeQuery.q);
    const nextType = normalizeSearchType(routeQuery.type);
    searchQuery.value = nextQuery;
    searchType.value = nextType;
    if (!nextQuery) {
      searchReqSeq++;
      searchResults.value = [];
      searching.value = false;
      return;
    }
    void runSearch(nextType, nextQuery);
  },
  { immediate: true },
);

watch(homeDashboardLimit, () => {
  if (homeDashboardReady) {
    void loadData();
  }
});

onMounted(() => {
  syncHomeMediaColumnCount();
  homeDashboardReady = true;
  window.addEventListener("resize", syncHomeMediaColumnCount, { passive: true });
  void loadData();
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", syncHomeMediaColumnCount);
});
</script>

<template>
  <section class="home-workbench">
    <div class="home-search-panel card">
      <div class="home-panel-head">
        <div class="min-w-0">
          <p class="section-label">工作台</p>
          <h2 class="home-workbench-title">媒体数据检索</h2>
        </div>
        <button class="btn-soft btn-primary-compact" :disabled="loading" @click="loadData">
          {{ loading ? "刷新中..." : "刷新数据" }}
        </button>
      </div>

      <div class="search-toolbar-form mt-4">
        <GlassSelect v-model="searchType" :options="searchTypeOptions" />
        <input
          v-model="searchQuery"
          type="text"
          class="field-control"
          placeholder="搜索电影、剧集、人物..."
          @keyup.enter="handleHomeSearch"
        />
        <button class="btn-primary" :disabled="searching" @click="handleHomeSearch">
          {{ searching ? "检索中..." : "检索" }}
        </button>
      </div>
    </div>
  </section>

  <section v-if="showSearchResults" class="card mt-4">
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
      <h3 class="section-title !mb-0">搜索结果</h3>
      <span class="badge">{{ resultSummary }}</span>
    </div>
    <p v-if="searching" class="empty-state">检索中...</p>
    <SearchResultList
      v-else
      :items="searchResults"
      :fallback-type="searchType"
      :limit="20"
      empty-text="未找到结果，请尝试更换关键词。"
    />
  </section>

  <template v-else>
    <section class="home-section-head">
      <div>
        <p class="section-label">本地看板</p>
        <h3 class="section-title !mb-0">数据库内容</h3>
      </div>
      <div class="home-section-actions">
        <div class="home-view-switch" role="group" aria-label="首页媒体显示模式">
          <button
            v-for="option in homeMediaViewOptions"
            :key="option.value"
            type="button"
            class="home-view-switch-btn"
            :class="homeMediaViewMode === option.value ? 'home-view-switch-btn-active' : ''"
            :aria-pressed="homeMediaViewMode === option.value"
            @click="setHomeMediaViewMode(option.value)"
          >
            {{ option.label }}
          </button>
        </div>
      </div>
    </section>

    <section
      v-for="section in homeMediaSections"
      :key="section.key"
      class="home-media-section"
      :class="section.key === 'latest' ? 'mt-6' : 'mt-8'"
    >
      <h3 class="section-title">{{ section.title }}</h3>
      <div v-if="section.items.length" class="home-media-board" :class="`home-media-board-${homeMediaViewMode}`">
        <RouterLink
          v-for="item in section.items"
          :key="`${item.media_type}-${item.tmdb_id}`"
          :to="mediaRoute(item)"
          class="home-media-card"
          :class="`home-media-card-${homeMediaViewMode}`"
          @mouseenter="prefetchListItem(item.media_type, item.tmdb_id)"
          @focus="prefetchListItem(item.media_type, item.tmdb_id)"
          @touchstart.passive="prefetchListItem(item.media_type, item.tmdb_id)"
        >
          <img
            :src="tmdbImg(item.poster_path, 'w342')"
            :srcset="`${tmdbImg(item.poster_path, 'w342')} 1x, ${tmdbImg(item.poster_path, 'w500')} 2x`"
            :alt="displayTitle(item)"
            class="home-media-poster"
            loading="lazy"
          />
          <div class="home-media-info">
            <p class="home-media-title">{{ displayTitle(item) }}</p>
            <p class="home-media-meta">
              <span class="home-media-stat-group">
                <span class="rating-badge">{{ ratingText(item) }}</span>
                <span v-if="section.key === 'hot'" class="home-media-metric">{{ visitText(item) }}</span>
              </span>
              <span>{{ mediaTypeLabel(item.media_type) }} {{ yearText(item.air_date) }}</span>
            </p>
            <p v-if="homeMediaViewMode === 'list'" class="home-media-overview">{{ overviewText(item) }}</p>
          </div>
        </RouterLink>
      </div>
      <p v-else-if="!loading" class="empty-state">{{ section.emptyText }}</p>
    </section>
  </template>
</template>
