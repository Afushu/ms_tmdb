type CacheEntry<T> = {
  data: T;
  expiresAt: number;
  accessedAt: number;
};

const MAX_RESPONSE_CACHE_SIZE = 200;

const responseCache = new Map<string, CacheEntry<unknown>>();
const inflightCache = new Map<string, Promise<unknown>>();
/** 每个缓存键的单调递增代次，用于阻止失效前发出的旧请求回填缓存。 */
const keyGenerations = new Map<string, number>();

function currentGeneration(key: string): number {
  return keyGenerations.get(key) ?? 0;
}

function advanceGeneration(key: string): void {
  keyGenerations.set(key, currentGeneration(key) + 1);
}

function purgeExpired(now: number = Date.now()): void {
  for (const [key, entry] of responseCache) {
    if (entry.expiresAt <= now) {
      responseCache.delete(key);
    }
  }
}

/** 超容量时淘汰 accessedAt 最早的条目；相同时间按 Map 插入顺序（先插入先删）保证确定性。 */
function evictIfOverCapacity(): void {
  while (responseCache.size > MAX_RESPONSE_CACHE_SIZE) {
    let oldestKey: string | undefined;
    let oldestAccessedAt = Infinity;

    for (const [key, entry] of responseCache) {
      if (entry.accessedAt < oldestAccessedAt) {
        oldestAccessedAt = entry.accessedAt;
        oldestKey = key;
      }
    }

    if (oldestKey === undefined) {
      break;
    }
    responseCache.delete(oldestKey);
  }
}

export async function withRequestCache<T>(key: string, loader: () => Promise<T>, ttlMs = 10 * 60 * 1000): Promise<T> {
  const now = Date.now();
  purgeExpired(now);

  const cached = responseCache.get(key);
  if (cached && cached.expiresAt > now) {
    cached.accessedAt = Date.now();
    return cached.data as T;
  }

  const inflight = inflightCache.get(key);
  if (inflight) {
    return inflight as Promise<T>;
  }

  const requestGeneration = currentGeneration(key);

  const request = loader()
    .then((data) => {
      const stillCurrentInflight = inflightCache.get(key) === request;
      const generationMatches = currentGeneration(key) === requestGeneration;

      if (generationMatches && stillCurrentInflight) {
        const writeNow = Date.now();
        const expiresAt = writeNow + ttlMs;
        // 仅在写入时刻条目尚未过期时回填（ttl 非正时拒绝写入）
        if (expiresAt > writeNow) {
          purgeExpired(writeNow);
          responseCache.set(key, {
            data,
            expiresAt,
            accessedAt: writeNow,
          });
          evictIfOverCapacity();
        }
      }

      if (inflightCache.get(key) === request) {
        inflightCache.delete(key);
      }
      return data;
    })
    .catch((error) => {
      if (inflightCache.get(key) === request) {
        inflightCache.delete(key);
      }
      throw error;
    });

  inflightCache.set(key, request);
  return request;
}

export function clearRequestCache(prefix?: string) {
  if (!prefix) {
    const keys = new Set<string>([...responseCache.keys(), ...inflightCache.keys()]);
    responseCache.clear();
    inflightCache.clear();
    for (const key of keys) {
      advanceGeneration(key);
    }
    return;
  }

  const keysToClear = new Set<string>();
  for (const key of responseCache.keys()) {
    if (key.startsWith(prefix)) {
      keysToClear.add(key);
    }
  }
  for (const key of inflightCache.keys()) {
    if (key.startsWith(prefix)) {
      keysToClear.add(key);
    }
  }

  for (const key of keysToClear) {
    responseCache.delete(key);
    inflightCache.delete(key);
    advanceGeneration(key);
  }
}
