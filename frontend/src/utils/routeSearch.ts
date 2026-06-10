import type { LocationQueryRaw } from "vue-router";
import type { SearchType } from "@/api/search";
import { readQueryString } from "@/utils/routeQuery";
export { queryValue, readQueryString } from "@/utils/routeQuery";

export const searchTypeOptions: ReadonlyArray<{ label: string; value: SearchType }> = [
  { label: "综合", value: "multi" },
  { label: "电影", value: "movie" },
  { label: "剧集", value: "tv" },
  { label: "人物", value: "person" },
];

export function normalizeSearchType(value: unknown): SearchType {
  const text = readQueryString(value);
  if (text === "movie" || text === "tv" || text === "person") return text;
  return "multi";
}

export function buildSearchQuery(targetType: SearchType, targetQuery: string): LocationQueryRaw {
  const nextQuery: LocationQueryRaw = { q: targetQuery };
  if (targetType !== "multi") nextQuery.type = targetType;
  return nextQuery;
}
