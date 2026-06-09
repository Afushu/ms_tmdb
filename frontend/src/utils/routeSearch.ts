import type { LocationQueryRaw } from "vue-router";
import type { SearchType } from "@/api/search";

export const searchTypeOptions: ReadonlyArray<{ label: string; value: SearchType }> = [
  { label: "综合", value: "multi" },
  { label: "电影", value: "movie" },
  { label: "剧集", value: "tv" },
  { label: "人物", value: "person" },
];

export function readQueryString(value: unknown): string {
  if (Array.isArray(value)) return String(value[0] ?? "").trim();
  return String(value ?? "").trim();
}

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

export function queryValue(value: unknown): string {
  return Array.isArray(value) ? String(value[0] ?? "") : String(value ?? "");
}
