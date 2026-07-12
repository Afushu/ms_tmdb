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

export function getPopularMovies(page = 1, language = "zh-CN", options?: RequestOptions) {
  return http.get<PagedResp<MediaSummary>>("/api/v3/movie/popular", {
    params: { page, language },
    ...options,
  });
}

export function getMovieDetail(id: number, language = "zh-CN", append = "", options: DetailOptions = {}) {
  const { force, ...requestOptions } = options;
  const params = append ? { language, append_to_response: append } : { language };
  const key = `movie:detail:${id}:${language}:${append}`;
  if (force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () =>
      http.get<MediaSummary & Record<string, unknown>>(`/api/v3/movie/${id}`, {
        params,
        ...requestOptions,
      }),
    DETAIL_CACHE_TTL,
  );
}

export function getMovieGenreList(language = "zh-CN", options?: RequestOptions) {
  return withRequestCache(
    `genre:movie:${language}`,
    () =>
      http.get<GenreListResp>("/api/v3/genre/movie/list", {
        params: { language },
        ...options,
      }),
    GENRE_CACHE_TTL,
  );
}

export function getMovieCredits(id: number, language = "zh-CN", options: DetailOptions = {}) {
  const { force, ...requestOptions } = options;
  const key = `movie:credits:${id}:${language}`;
  if (force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () =>
      http.get<Record<string, unknown>>(`/api/v3/movie/${id}/credits`, {
        params: { language },
        ...requestOptions,
      }),
    AUXILIARY_CACHE_TTL,
  );
}

/**
 * 使 Movie 域 Request_Cache 失效。
 * 覆盖：movie:detail:*、movie:credits:*、genre:movie:*
 * 传入 id 时仅清理该实体的详情与演员缓存；省略 id 时清理整个 Movie 域（含类型列表）。
 * Admin 首页看板与媒体列表未接入 Request_Cache，不在此处理。
 */
export function clearMovieCache(id?: number) {
  if (id !== undefined) {
    clearRequestCache(`movie:detail:${id}:`);
    clearRequestCache(`movie:credits:${id}:`);
    return;
  }
  clearRequestCache("movie:detail:");
  clearRequestCache("movie:credits:");
  clearRequestCache("genre:movie:");
}
