<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter, type LocationQueryRaw } from "vue-router";
import { getHomeDashboard, type AdminHomeMediaItem } from "@/api/admin";
import { prefetchMediaDetail } from "@/api/prefetch";
import { searchByType, type SearchType } from "@/api/search";
import { tmdbImg } from "@/api/tmdb";
import GlassSelect from "@/components/GlassSelect.vue";
import SearchResultList from "@/components/SearchResultList.vue";
import type { SearchResultItem } from "@/types/media";
import {
  buildSearchQuery,
  normalizeSearchType,
  readQueryString,
  searchTypeOptions,
} from "@/utils/routeSearch";
import { isSameQuery } from "@/utils/routeQuery";

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const latestMedia = ref<AdminHomeMediaItem[]>([]);
const hotMedia = ref<AdminHomeMediaItem[]>([]);
const searchQuery = ref(readQueryString(route.query.q));
const searchType = ref<SearchType>(normalizeSearchType(route.query.type));
const searching = ref(false);
const searchResults = ref<SearchResultItem[]>([]);
let searchReqSeq = 0;

const hasRouteQuery = computed(() => Boolean(readQueryString(route.query.q)));
const resultSummary = computed(() => {
  if (searching.value) return "检索中";
  if (searchResults.value.length) return `${searchResults.value.length} 条匹配`;
  return hasRouteQuery.value ? "暂无结果" : "等待检索";
});

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

async function loadData() {
  loading.value = true;
  try {
    const resp = await getHomeDashboard();
    latestMedia.value = resp.data?.latest ?? [];
    hotMedia.value = resp.data?.hot ?? [];
  } catch { /* handled by global toast */ } finally {
    loading.value = false;
  }
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
  } catch { /* handled by global toast */ } finally {
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

onMounted(loadData);
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

  <section v-if="hasRouteQuery || searchResults.length" class="card mt-4">
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

  <section class="home-section-head">
    <div>
      <p class="section-label">本地看板</p>
      <h3 class="section-title !mb-0">数据库内容</h3>
    </div>
    <span class="badge">Local Library</span>
  </section>

  <section class="mt-6">
    <h3 class="section-title">最新入库</h3>
    <div v-if="latestMedia.length" class="poster-grid">
      <RouterLink
        v-for="item in latestMedia"
        :key="`${item.media_type}-${item.tmdb_id}`"
        :to="mediaRoute(item)"
        class="poster-card"
        @mouseenter="prefetchListItem(item.media_type, item.tmdb_id)"
        @focus="prefetchListItem(item.media_type, item.tmdb_id)"
        @touchstart.passive="prefetchListItem(item.media_type, item.tmdb_id)"
      >
        <img
          :src="tmdbImg(item.poster_path, 'w342')"
          :srcset="`${tmdbImg(item.poster_path, 'w342')} 1x, ${tmdbImg(item.poster_path, 'w500')} 2x`"
          :alt="displayTitle(item)"
          class="poster-img"
          loading="lazy"
        />
        <div class="poster-info">
          <p class="truncate text-sm font-medium">{{ displayTitle(item) }}</p>
          <p class="poster-meta">
            <span class="poster-rating">评分 {{ item.vote_average.toFixed(1) }}</span>
            <span>{{ mediaTypeLabel(item.media_type) }} {{ yearText(item.air_date) }}</span>
          </p>
        </div>
      </RouterLink>
    </div>
    <p v-else-if="!loading" class="empty-state">本地库暂无数据。</p>
  </section>

  <section class="mt-8">
    <h3 class="section-title">访问热度</h3>
    <div v-if="hotMedia.length" class="poster-grid">
      <RouterLink
        v-for="item in hotMedia"
        :key="`${item.media_type}-${item.tmdb_id}`"
        :to="mediaRoute(item)"
        class="poster-card"
        @mouseenter="prefetchListItem(item.media_type, item.tmdb_id)"
        @focus="prefetchListItem(item.media_type, item.tmdb_id)"
        @touchstart.passive="prefetchListItem(item.media_type, item.tmdb_id)"
      >
        <img
          :src="tmdbImg(item.poster_path, 'w342')"
          :srcset="`${tmdbImg(item.poster_path, 'w342')} 1x, ${tmdbImg(item.poster_path, 'w500')} 2x`"
          :alt="displayTitle(item)"
          class="poster-img"
          loading="lazy"
        />
        <div class="poster-info">
          <p class="truncate text-sm font-medium">{{ displayTitle(item) }}</p>
          <p class="poster-meta">
            <span class="poster-rating">访问 {{ item.visit_count }} 次</span>
            <span>{{ mediaTypeLabel(item.media_type) }} {{ yearText(item.air_date) }}</span>
          </p>
        </div>
      </RouterLink>
    </div>
    <p v-else-if="!loading" class="empty-state">暂无本地访问热度记录。</p>
  </section>
</template>
