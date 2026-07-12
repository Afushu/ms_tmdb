import { getMovieDetail } from "./movie";
import { getPersonDetail } from "./person";
import { getTVDetail } from "./tv";

export type PrefetchMediaType = "movie" | "tv" | "person";
export type PrefetchKey = `${PrefetchMediaType}:${number}`;

const DEDUPE_WINDOW_MS = 10_000;
const HOVER_FOCUS_DELAY_MS = 120;
const MAX_CONCURRENT = 2;

/** 去重窗口：键 → 首次成功调度时间戳 */
const recentScheduled = new Map<PrefetchKey, number>();
/** 尚未触发的 hover/focus 延迟计时器 */
const pendingTimers = new Map<PrefetchKey, ReturnType<typeof setTimeout>>();
/** 全局进行中的预取数量 */
let concurrentCount = 0;

type NetworkInformationLike = {
  saveData?: boolean;
  effectiveType?: string;
};

function getNetworkConnection(): NetworkInformationLike | undefined {
  if (typeof navigator === "undefined") {
    return undefined;
  }
  const nav = navigator as Navigator & {
    connection?: NetworkInformationLike;
    mozConnection?: NetworkInformationLike;
    webkitConnection?: NetworkInformationLike;
  };
  return nav.connection ?? nav.mozConnection ?? nav.webkitConnection;
}

/** 省流或慢网时跳过；缺少 Network Information API 时允许预取 */
function shouldSkipForNetwork(): boolean {
  const connection = getNetworkConnection();
  if (!connection) {
    return false;
  }
  if (connection.saveData === true) {
    return true;
  }
  const effectiveType = connection.effectiveType;
  return effectiveType === "slow-2g" || effectiveType === "2g";
}

function toPrefetchKey(mediaType: PrefetchMediaType, id: number): PrefetchKey {
  return `${mediaType}:${id}`;
}

function isValidId(id: number): boolean {
  return Number.isFinite(id) && id > 0;
}

function isInDedupeWindow(key: PrefetchKey): boolean {
  const scheduledAt = recentScheduled.get(key);
  if (scheduledAt == null) {
    return false;
  }
  if (Date.now() - scheduledAt >= DEDUPE_WINDOW_MS) {
    recentScheduled.delete(key);
    return false;
  }
  return true;
}

function markScheduled(key: PrefetchKey) {
  if (!isInDedupeWindow(key)) {
    recentScheduled.set(key, Date.now());
  }
}

function unmarkScheduled(key: PrefetchKey) {
  recentScheduled.delete(key);
}

function clearPendingTimer(key: PrefetchKey) {
  const timer = pendingTimers.get(key);
  if (timer == null) {
    return false;
  }
  clearTimeout(timer);
  pendingTimers.delete(key);
  return true;
}

function runPrefetchRequest(mediaType: PrefetchMediaType, id: number) {
  concurrentCount += 1;

  // 预取失败始终静默，不触发全局 Toast，也不影响导航或 UI 状态
  const silent = { showErrorToast: false as const };
  const task =
    mediaType === "movie"
      ? getMovieDetail(id, "zh-CN", "", silent)
      : mediaType === "tv"
        ? getTVDetail(id, "zh-CN", "", silent)
        : getPersonDetail(id, "zh-CN", "", silent);

  void task
    .catch(() => {
      // Prefetch failures should not affect navigation or UI state.
    })
    .finally(() => {
      concurrentCount = Math.max(0, concurrentCount - 1);
    });
}

function tryStartPrefetch(mediaType: PrefetchMediaType, id: number): boolean {
  if (concurrentCount >= MAX_CONCURRENT) {
    return false;
  }
  runPrefetchRequest(mediaType, id);
  return true;
}

/**
 * hover / focus：固定 120ms 延迟后启动。
 * 成功创建计时器即占用 10 秒同键去重窗口；到期前 cancel 则释放窗口。
 * 到期时若并发已满则直接跳过，不排队。
 */
export function schedulePrefetch(mediaType: PrefetchMediaType, id: number) {
  if (!isValidId(id) || shouldSkipForNetwork()) {
    return;
  }

  const key = toPrefetchKey(mediaType, id);
  if (pendingTimers.has(key) || isInDedupeWindow(key)) {
    return;
  }

  const timer = setTimeout(() => {
    pendingTimers.delete(key);
    // 延迟到期时并发已满：跳过且不重试；去重窗口已在创建计时器时占用
    void tryStartPrefetch(mediaType, id);
  }, HOVER_FOCUS_DELAY_MS);

  pendingTimers.set(key, timer);
  markScheduled(key);
}

/**
 * pointerleave / blur：在 120ms 到期前取消，且不占用 10 秒去重窗口。
 */
export function cancelPrefetch(mediaType: PrefetchMediaType, id: number) {
  if (!isValidId(id)) {
    return;
  }
  const key = toPrefetchKey(mediaType, id);
  if (!clearPendingTimer(key)) {
    return;
  }
  // 仅清除尚未触发的延迟任务；已发出的请求不取消
  unmarkScheduled(key);
}

/**
 * touch：立即尝试启动，无额外延迟；仍遵守省流/慢网、同键去重与并发上限。
 * 若存在未触发的延迟计时器，则改为立即启动。
 */
export function prefetchMediaDetail(mediaType: PrefetchMediaType, id: number) {
  if (!isValidId(id) || shouldSkipForNetwork()) {
    return;
  }

  const key = toPrefetchKey(mediaType, id);
  const hadPending = clearPendingTimer(key);

  // 已在去重窗口内且没有待触发计时器：说明此前已成功调度过，跳过
  if (!hadPending && isInDedupeWindow(key)) {
    return;
  }

  if (!tryStartPrefetch(mediaType, id)) {
    // 未能真正启动：不占用去重窗口
    unmarkScheduled(key);
    return;
  }

  markScheduled(key);
}
