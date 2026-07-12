import http, { type RequestOptions } from "./http";
import { clearRequestCache, withRequestCache } from "./requestCache";
import type { MediaGenre, MediaSummary } from "@/types/media";

type PagedResp<T> = {
  page: number;
  total_pages?: number;
  total_results?: number;
  results: T[];
};

type GenreListResp = {
  genres: MediaGenre[];
};

const GENRE_CACHE_TTL = 30 * 60 * 1000;
const AUXILIARY_CACHE_TTL = 10 * 60 * 1000;
const DETAIL_CACHE_TTL = 5 * 60 * 1000;

type DetailOptions = RequestOptions & {
  force?: boolean;
};

export function getPopularTV(page = 1, language = "zh-CN", options?: RequestOptions) {
  return http.get<PagedResp<MediaSummary>>("/api/v3/tv/popular", {
    params: { page, language },
    ...options,
  });
}

export function getTVDetail(id: number, language = "zh-CN", append = "", options: DetailOptions = {}) {
  const { force, ...requestOptions } = options;
  const params = append ? { language, append_to_response: append } : { language };
  const key = `tv:detail:${id}:${language}:${append}`;
  if (force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () =>
      http.get<MediaSummary & Record<string, unknown>>(`/api/v3/tv/${id}`, {
        params,
        ...requestOptions,
      }),
    DETAIL_CACHE_TTL,
  );
}

export function getTVGenreList(language = "zh-CN", options?: RequestOptions) {
  return withRequestCache(
    `genre:tv:${language}`,
    () =>
      http.get<GenreListResp>("/api/v3/genre/tv/list", {
        params: { language },
        ...options,
      }),
    GENRE_CACHE_TTL,
  );
}

export function getTVCredits(id: number, language = "zh-CN", options: DetailOptions = {}) {
  const { force, ...requestOptions } = options;
  const key = `tv:credits:${id}:${language}`;
  if (force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () =>
      http.get<Record<string, unknown>>(`/api/v3/tv/${id}/credits`, {
        params: { language },
        ...requestOptions,
      }),
    AUXILIARY_CACHE_TTL,
  );
}

export function getTVSeasonDetail(
  id: number,
  seasonNumber: number,
  language = "zh-CN",
  append = "",
  options?: RequestOptions,
) {
  const params = append ? { language, append_to_response: append } : { language };
  return http.get<Record<string, unknown>>(`/api/v3/tv/${id}/season/${seasonNumber}`, {
    params,
    ...options,
  });
}

/**
 * 使 TV 域 Request_Cache 失效。
 * 覆盖：tv:detail:*、tv:credits:*、genre:tv:*
 * 传入 id 时仅清理该实体的详情与演员缓存；省略 id 时清理整个 TV 域（含类型列表）。
 * 季详情未走 Request_Cache；Admin 首页看板与媒体列表未接入 Request_Cache，不在此处理。
 */
export function clearTVCache(id?: number) {
  if (id !== undefined) {
    clearRequestCache(`tv:detail:${id}:`);
    clearRequestCache(`tv:credits:${id}:`);
    return;
  }
  clearRequestCache("tv:detail:");
  clearRequestCache("tv:credits:");
  clearRequestCache("genre:tv:");
}
