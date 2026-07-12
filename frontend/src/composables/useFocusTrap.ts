import { nextTick, onBeforeUnmount, toValue, watch, type MaybeRefOrGetter, type Ref } from "vue";

/**
 * 焦点陷阱 + 顶层弹层栈。
 * 不维护任何业务弹窗可见状态；调用方负责在挂载后 register、关闭时 unregister。
 */

export const ADMIN_MAIN_FOCUS_ID = "admin-main";
const APP_ROOT_ID = "app";

const FOCUSABLE_SELECTOR = [
  "a[href]",
  "area[href]",
  'input:not([disabled]):not([type="hidden"])',
  "select:not([disabled])",
  "textarea:not([disabled])",
  "button:not([disabled])",
  "iframe",
  "object",
  "embed",
  "[contenteditable]:not([contenteditable='false'])",
  '[tabindex]:not([tabindex="-1"])',
].join(",");

export type OverlayStackEntry = {
  id: string;
  container: HTMLElement;
};

const overlayStack: OverlayStackEntry[] = [];
const stackListeners = new Set<() => void>();
let overlayIdCounter = 0;

export function createOverlayId(prefix = "overlay"): string {
  overlayIdCounter += 1;
  return `${prefix}-${overlayIdCounter}`;
}

function supportsInert(): boolean {
  return typeof HTMLElement !== "undefined" && "inert" in HTMLElement.prototype;
}

function isElementVisible(el: HTMLElement): boolean {
  if (!(el.offsetWidth || el.offsetHeight || el.getClientRects().length)) {
    return false;
  }
  const style = window.getComputedStyle(el);
  return style.visibility !== "hidden" && style.display !== "none";
}

function isInertOrHiddenAncestor(el: Element): boolean {
  let current: Element | null = el;
  while (current) {
    if (current instanceof HTMLElement) {
      if (supportsInert() && current.inert) {
        return true;
      }
      if (current.getAttribute("aria-hidden") === "true") {
        return true;
      }
    }
    current = current.parentElement;
  }
  return false;
}

export function getFocusableElements(container: HTMLElement): HTMLElement[] {
  return Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter((el) => {
    if (el.hasAttribute("disabled") || el.getAttribute("aria-disabled") === "true") {
      return false;
    }
    if (el.tabIndex < 0) {
      return false;
    }
    if (!isElementVisible(el)) {
      return false;
    }
    // 排除位于 inert / aria-hidden 子树内的元素（容器自身除外）
    let node: HTMLElement | null = el;
    while (node && node !== container) {
      if (supportsInert() && node.inert) {
        return false;
      }
      if (node.getAttribute("aria-hidden") === "true") {
        return false;
      }
      node = node.parentElement;
    }
    return true;
  });
}

function canReceiveFocus(el: Element | null | undefined): el is HTMLElement {
  if (!(el instanceof HTMLElement)) {
    return false;
  }
  if (!document.contains(el)) {
    return false;
  }
  if (!isElementVisible(el)) {
    return false;
  }
  if (isInertOrHiddenAncestor(el)) {
    return false;
  }
  if ((el as HTMLButtonElement).disabled) {
    return false;
  }
  return true;
}

function tryFocus(el: HTMLElement): boolean {
  if (!canReceiveFocus(el)) {
    return false;
  }
  try {
    el.focus({ preventScroll: true });
  } catch {
    return false;
  }
  return document.activeElement === el || el.contains(document.activeElement);
}

function focusContainer(container: HTMLElement) {
  if (!container.hasAttribute("tabindex")) {
    container.setAttribute("tabindex", "-1");
  }
  tryFocus(container);
}

function focusWithinContainer(container: HTMLElement, initial?: HTMLElement | null) {
  if (initial && container.contains(initial) && tryFocus(initial)) {
    return;
  }
  const focusables = getFocusableElements(container);
  if (focusables.length > 0) {
    tryFocus(focusables[0]);
    return;
  }
  focusContainer(container);
}

function getFallbackFocusTarget(): HTMLElement | null {
  return document.getElementById(ADMIN_MAIN_FOCUS_ID);
}

/**
 * 关闭后恢复焦点：优先触发元素 → 新栈顶容器 → #admin-main。
 * 不聚焦 document.body。
 */
export function restoreFocusTarget(trigger: Element | null | undefined) {
  if (canReceiveFocus(trigger) && tryFocus(trigger)) {
    return;
  }

  const topEntry = overlayStack[overlayStack.length - 1];
  if (topEntry) {
    focusWithinContainer(topEntry.container);
    return;
  }

  const fallback = getFallbackFocusTarget();
  if (fallback) {
    tryFocus(fallback);
  }
}

function setIsolated(el: HTMLElement, isolated: boolean) {
  if (supportsInert()) {
    el.inert = isolated;
  }
  if (isolated) {
    el.setAttribute("aria-hidden", "true");
  } else {
    el.removeAttribute("aria-hidden");
  }
}

function applyOverlayIsolation() {
  if (typeof document === "undefined") {
    return;
  }

  const appRoot = document.getElementById(APP_ROOT_ID);
  const top = overlayStack[overlayStack.length - 1];

  if (!top) {
    if (appRoot) {
      setIsolated(appRoot, false);
    }
    return;
  }

  // Teleport 到 body 的弹层与 #app 为兄弟节点：隔离应用根即可隔离背景。
  // 不得对包含栈顶容器的共同祖先设置 inert。
  if (appRoot) {
    setIsolated(appRoot, true);
  }

  for (const entry of overlayStack) {
    const isTop = entry.id === top.id;
    setIsolated(entry.container, !isTop);
  }
}

function notifyStackListeners() {
  for (const listener of stackListeners) {
    listener();
  }
}

export function useOverlayStack() {
  function register(id: string, container: HTMLElement) {
    const existing = overlayStack.findIndex((entry) => entry.id === id);
    if (existing >= 0) {
      overlayStack.splice(existing, 1);
    }
    overlayStack.push({ id, container });
    applyOverlayIsolation();
    notifyStackListeners();
  }

  function unregister(id: string) {
    const index = overlayStack.findIndex((entry) => entry.id === id);
    if (index < 0) {
      return;
    }
    const [removed] = overlayStack.splice(index, 1);
    setIsolated(removed.container, false);
    applyOverlayIsolation();
    notifyStackListeners();
  }

  function isTop(id: string): boolean {
    const top = overlayStack[overlayStack.length - 1];
    return !!top && top.id === id;
  }

  function top(): OverlayStackEntry | undefined {
    return overlayStack[overlayStack.length - 1];
  }

  function subscribe(listener: () => void): () => void {
    stackListeners.add(listener);
    return () => {
      stackListeners.delete(listener);
    };
  }

  function getDepth(): number {
    return overlayStack.length;
  }

  return {
    register,
    unregister,
    isTop,
    top,
    subscribe,
    getDepth,
  };
}

export type UseFocusTrapOptions = {
  /** 阻断容器根节点 */
  containerRef: Ref<HTMLElement | null | undefined>;
  /**
   * 是否捕获 Tab/Shift+Tab（通常为 visible && isTop）。
   * 仅栈顶容器应启用，以保证键盘焦点不离开栈顶。
   */
  enabled: MaybeRefOrGetter<boolean>;
  /**
   * 是否处于打开会话（用于记录/恢复触发元素）。
   * 默认同 `enabled`。嵌套场景应传 visible：打开时记触发元素，关闭时才恢复焦点；
   * 仅因失去栈顶而 `enabled=false` 时不恢复焦点。
   */
  open?: MaybeRefOrGetter<boolean>;
  /** 可选的初始焦点目标 */
  initialFocusRef?: Ref<HTMLElement | null | undefined>;
};

/**
 * 记录触发元素、初始化焦点、捕获 Tab/Shift+Tab、关闭时恢复焦点。
 * 容器无可聚焦元素时回退聚焦容器自身（tabindex=-1）。
 */
export function useFocusTrap(options: UseFocusTrapOptions) {
  let previouslyFocused: Element | null = null;
  let isOpenSession = false;
  let isTrapListening = false;

  function getContainer(): HTMLElement | null {
    return options.containerRef.value ?? null;
  }

  function isOpenValue(): boolean {
    return toValue(options.open ?? options.enabled);
  }

  function focusInitial() {
    const container = getContainer();
    if (!container) {
      return;
    }
    const initial = options.initialFocusRef?.value ?? null;
    focusWithinContainer(container, initial);
  }

  function onKeyDown(event: KeyboardEvent) {
    if (event.key !== "Tab") {
      return;
    }
    const container = getContainer();
    if (!container) {
      return;
    }

    const focusables = getFocusableElements(container);
    if (focusables.length === 0) {
      event.preventDefault();
      focusContainer(container);
      return;
    }

    const first = focusables[0];
    const last = focusables[focusables.length - 1];
    const current = document.activeElement;

    if (event.shiftKey) {
      if (current === first || !container.contains(current)) {
        event.preventDefault();
        last.focus();
      }
      return;
    }

    if (current === last || !container.contains(current)) {
      event.preventDefault();
      first.focus();
    }
  }

  function ensureFocusInside() {
    const container = getContainer();
    if (!container) {
      return;
    }
    const active = document.activeElement;
    if (active && container.contains(active)) {
      return;
    }
    focusInitial();
  }

  function startListening() {
    if (isTrapListening) {
      void nextTick(() => {
        ensureFocusInside();
      });
      return;
    }
    isTrapListening = true;
    document.addEventListener("keydown", onKeyDown, true);
    void nextTick(() => {
      // 首次成为栈顶时初始化焦点；重新成为栈顶时仅在焦点已离开容器时回收
      ensureFocusInside();
    });
  }

  function stopListening() {
    if (!isTrapListening) {
      return;
    }
    isTrapListening = false;
    document.removeEventListener("keydown", onKeyDown, true);
  }

  function openSession() {
    if (isOpenSession) {
      return;
    }
    isOpenSession = true;
    previouslyFocused = document.activeElement;
  }

  function closeSession() {
    if (!isOpenSession) {
      return;
    }
    isOpenSession = false;
    stopListening();
    restoreFocusTarget(previouslyFocused);
    previouslyFocused = null;
  }

  watch(
    () => ({ open: isOpenValue(), enabled: toValue(options.enabled) }),
    ({ open, enabled }) => {
      if (open) {
        openSession();
      } else {
        closeSession();
        return;
      }

      if (enabled) {
        startListening();
      } else {
        stopListening();
      }
    },
    { immediate: true },
  );

  onBeforeUnmount(() => {
    closeSession();
  });

  return {
    focusInitial,
    openSession,
    closeSession,
    startListening,
    stopListening,
  };
}
