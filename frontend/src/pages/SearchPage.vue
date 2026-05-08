<script setup lang="ts">
import { ref, watch } from "vue";
import { useRoute, useRouter, type LocationQueryRaw } from "vue-router";
import GlassSelect from "@/components/GlassSelect.vue";
import { searchByType, type SearchType } from "@/api/search";
import SearchResultList from "@/components/SearchResultList.vue";
import type { ApiErrorLike, SearchResultItem } from "@/types/media";

const route = useRoute();
const router = useRouter();
const query = ref(readQueryString(route.query.q));
const type = ref<SearchType>(normalizeSearchType(route.query.type));
const loading = ref(false);
const error = ref("");
const results = ref<SearchResultItem[]>([]);
let searchReqSeq = 0;

function readQueryString(value: unknown): string {
  if (Array.isArray(value)) return String(value[0] ?? "").trim();
  return String(value ?? "").trim();
}

function normalizeSearchType(value: unknown): SearchType {
  const text = readQueryString(value);
  if (text === "movie" || text === "tv" || text === "person") return text;
  return "multi";
}

function resolveErrorMessage(err: unknown, fallback: string): string {
  if (err && typeof err === "object" && "message" in err) {
    const message = (err as ApiErrorLike).message;
    if (typeof message === "string" && message.trim()) return message;
  }
  return fallback;
}

const typeOptions = [
  { label: "综合", value: "multi" },
  { label: "电影", value: "movie" },
  { label: "剧集", value: "tv" },
  { label: "人物", value: "person" },
] as const;

function buildSearchQuery(targetType: SearchType, targetQuery: string): LocationQueryRaw {
  const nextQuery: LocationQueryRaw = { q: targetQuery };
  if (targetType !== "multi") nextQuery.type = targetType;
  return nextQuery;
}

function queryValue(value: unknown): string {
  return Array.isArray(value) ? String(value[0] ?? "") : String(value ?? "");
}

function isSameSearchQuery(nextQuery: LocationQueryRaw): boolean {
  const keys = new Set([...Object.keys(route.query), ...Object.keys(nextQuery)]);
  for (const key of keys) {
    if (queryValue(route.query[key]) !== queryValue(nextQuery[key])) return false;
  }
  return true;
}

async function runSearch(targetType: SearchType, targetQuery: string) {
  if (!targetQuery) {
    searchReqSeq++;
    error.value = "请输入关键词";
    results.value = [];
    loading.value = false;
    return;
  }
  const requestSeq = ++searchReqSeq;
  loading.value = true;
  error.value = "";
  try {
    const resp = await searchByType(targetType, targetQuery, 1);
    if (requestSeq !== searchReqSeq) {
      return;
    }
    results.value = resp.data?.results ?? [];
  } catch (err: unknown) {
    if (requestSeq === searchReqSeq) {
      error.value = resolveErrorMessage(err, "搜索失败");
    }
  } finally {
    if (requestSeq === searchReqSeq) {
      loading.value = false;
    }
  }
}

async function handleSearch() {
  const trimmedQuery = query.value.trim();
  const targetType = type.value;
  if (!trimmedQuery) {
    searchReqSeq++;
    error.value = "请输入关键词";
    results.value = [];
    loading.value = false;
    return;
  }

  const nextQuery = buildSearchQuery(targetType, trimmedQuery);
  if (!isSameSearchQuery(nextQuery)) {
    await router.replace({ path: "/search", query: nextQuery });
    return;
  }
  await runSearch(targetType, trimmedQuery);
}

watch(
  () => route.query,
  (routeQuery) => {
    const nextQuery = readQueryString(routeQuery.q);
    const nextType = normalizeSearchType(routeQuery.type);
    query.value = nextQuery;
    type.value = nextType;
    if (!nextQuery) {
      searchReqSeq++;
      results.value = [];
      error.value = "";
      loading.value = false;
      return;
    }
    void runSearch(nextType, nextQuery);
  },
  { immediate: true },
);
</script>

<template>
  <section class="card">
    <div class="mb-4">
      <p class="section-label">Search</p>
      <h2 class="search-page-title">全站搜索</h2>
      <p class="mt-1 text-sm text-black/55">跨电影、剧集和人物搜索 TMDB 数据，并快速进入详情页。</p>
    </div>
    <div class="grid gap-3 md:grid-cols-[140px_1fr_auto]">
      <GlassSelect v-model="type" :options="typeOptions" />
      <input
        v-model="query"
        type="text"
        class="field-control"
        placeholder="输入关键词，例如：Fight Club"
        @keyup.enter="handleSearch"
      />
      <button class="btn-primary" @click="handleSearch">
        {{ loading ? "搜索中..." : "搜索" }}
      </button>
    </div>
    <p v-if="error" class="mt-3 text-sm text-red-600">{{ error }}</p>
  </section>

  <section v-if="results.length" class="card mt-4">
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
      <h3 class="section-title !mb-0">结果</h3>
      <span class="badge">{{ results.length }} 条匹配</span>
    </div>
    <SearchResultList :items="results" :fallback-type="type" :limit="20" />
  </section>
</template>
