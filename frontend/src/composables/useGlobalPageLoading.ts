import { readonly, ref } from "vue";

const MIN_VISIBLE_MS = 260;

const visible = ref(true);
let visibleSince = Date.now();
let hideTimer: ReturnType<typeof setTimeout> | undefined;

export const globalPageLoading = readonly(visible);

export function startGlobalPageLoading() {
  if (hideTimer) {
    clearTimeout(hideTimer);
    hideTimer = undefined;
  }
  visibleSince = Date.now();
  visible.value = true;
}

export function stopGlobalPageLoading() {
  const wait = Math.max(0, MIN_VISIBLE_MS - (Date.now() - visibleSince));

  if (hideTimer) clearTimeout(hideTimer);
  hideTimer = setTimeout(() => {
    visible.value = false;
    hideTimer = undefined;
  }, wait);
}
