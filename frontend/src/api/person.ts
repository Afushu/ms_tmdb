import http, { type RequestOptions } from "./http";
import { clearRequestCache, withRequestCache } from "./requestCache";

const AUXILIARY_CACHE_TTL = 10 * 60 * 1000;
const DETAIL_CACHE_TTL = 5 * 60 * 1000;

type DetailOptions = RequestOptions & {
  force?: boolean;
};

export function getPopularPeople(page = 1, language = "zh-CN", options?: RequestOptions) {
  return http.get("/api/v3/person/popular", {
    params: { page, language },
    ...options,
  });
}

export function getPersonDetail(id: number, language = "zh-CN", append = "", options: DetailOptions = {}) {
  const { force, ...requestOptions } = options;
  const params = append ? { language, append_to_response: append } : { language };
  const key = `person:detail:${id}:${language}:${append}`;
  if (force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () => http.get(`/api/v3/person/${id}`, { params, ...requestOptions }),
    DETAIL_CACHE_TTL,
  );
}

export function getPersonCombinedCredits(id: number, language = "zh-CN", options: DetailOptions = {}) {
  const { force, ...requestOptions } = options;
  const key = `person:combined_credits:${id}:${language}`;
  if (force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () =>
      http.get(`/api/v3/person/${id}/combined_credits`, {
        params: { language },
        ...requestOptions,
      }),
    AUXILIARY_CACHE_TTL,
  );
}

export function getPersonImages(id: number, options: DetailOptions = {}) {
  const { force, ...requestOptions } = options;
  const key = `person:images:${id}`;
  if (force) {
    clearRequestCache(key);
  }
  return withRequestCache(
    key,
    () => http.get(`/api/v3/person/${id}/images`, { ...requestOptions }),
    AUXILIARY_CACHE_TTL,
  );
}

/**
 * 使 Person 域 Request_Cache 失效。
 * 覆盖：person:detail:*、person:combined_credits:*、person:images:*
 * 传入 id 时仅清理该实体相关键；省略 id 时清理整个 Person 域前缀。
 * Admin 首页看板与媒体列表未接入 Request_Cache，不在此处理。
 *
 * 说明：person:images 现有键为 `person:images:${id}`（id 后无分隔段），
 * clearRequestCache 按 startsWith 匹配，清理 id=1 时可能连带匹配 id=10 等前缀重合键，
 * 属于安全的过度失效，不改变既有读缓存键格式。
 */
export function clearPersonCache(id?: number) {
  if (id !== undefined) {
    clearRequestCache(`person:detail:${id}:`);
    clearRequestCache(`person:combined_credits:${id}:`);
    clearRequestCache(`person:images:${id}`);
    return;
  }
  clearRequestCache("person:detail:");
  clearRequestCache("person:combined_credits:");
  clearRequestCache("person:images:");
}
