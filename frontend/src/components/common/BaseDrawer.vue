<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";
import {
  createOverlayId,
  useFocusTrap,
  useOverlayStack,
} from "@/composables/useFocusTrap";
import { useScrollLock } from "@/composables/useScrollLock";

const props = withDefaults(
  defineProps<{
    visible: boolean;
    title: string;
    description?: string;
    /** 提交中禁止关闭，优先于 closeOnEscape / closeOnOverlay */
    busy?: boolean;
    closeOnOverlay?: boolean;
    closeOnEscape?: boolean;
    initialFocus?: "close" | "primary" | "first";
    showCloseButton?: boolean;
    side?: "right" | "left";
    panelClass?: string;
    headerClass?: string;
    contentClass?: string;
    footerClass?: string;
    overlayClass?: string;
    rootClass?: string;
  }>(),
  {
    description: undefined,
    busy: false,
    closeOnOverlay: true,
    closeOnEscape: true,
    initialFocus: "first",
    showCloseButton: true,
    side: "right",
    panelClass: "admin-preference-drawer",
    headerClass: "admin-preference-header",
    contentClass: "admin-preference-options",
    footerClass: "",
    overlayClass: "admin-preference-mask",
    rootClass: "fixed inset-0 z-[80]",
  },
);

const emit = defineEmits<{
  close: [];
}>();

const overlayId = createOverlayId("drawer");
const titleId = `${overlayId}-title`;
const descriptionId = `${overlayId}-description`;

const containerRef = ref<HTMLElement | null>(null);
const closeButtonRef = ref<HTMLElement | null>(null);
const initialFocusRef = ref<HTMLElement | null | undefined>(null);
const isTopOverlay = ref(false);
/** 焦点会话：与 visible 解耦，确保关闭时先出栈再恢复焦点 */
const sessionOpen = ref(false);
let scrollLocked = false;

const { register, unregister, isTop, subscribe } = useOverlayStack();
const { lock, unlock } = useScrollLock();

const hasDescription = computed(() => Boolean(props.description));
const trapEnabled = computed(() => sessionOpen.value && isTopOverlay.value);
const panelSideClass = computed(() =>
  props.side === "left" ? "base-drawer-panel-left" : "base-drawer-panel-right",
);

useFocusTrap({
  containerRef,
  enabled: trapEnabled,
  open: sessionOpen,
  initialFocusRef,
});

function lockScroll() {
  if (scrollLocked) {
    return;
  }
  lock();
  scrollLocked = true;
}

function unlockScroll() {
  if (!scrollLocked) {
    return;
  }
  unlock();
  scrollLocked = false;
}

function syncTopState() {
  isTopOverlay.value = sessionOpen.value && isTop(overlayId);
}

function resolveInitialFocus() {
  if (props.initialFocus === "close") {
    initialFocusRef.value = closeButtonRef.value;
    return;
  }
  if (props.initialFocus === "primary") {
    initialFocusRef.value =
      containerRef.value?.querySelector<HTMLElement>("[data-dialog-primary]") ?? null;
    return;
  }
  initialFocusRef.value = null;
}

function canClose(): boolean {
  return !props.busy;
}

function requestClose() {
  if (!canClose()) {
    return;
  }
  emit("close");
}

function onOverlayClick() {
  if (!canClose() || !props.closeOnOverlay) {
    return;
  }
  emit("close");
}

function onDocumentKeyDown(event: KeyboardEvent) {
  if (event.key !== "Escape") {
    return;
  }
  if (!sessionOpen.value || !isTopOverlay.value) {
    return;
  }
  // busy 禁止关闭优先于 closeOnEscape
  if (!canClose() || !props.closeOnEscape) {
    return;
  }
  event.preventDefault();
  event.stopPropagation();
  emit("close");
}

function teardownOverlay() {
  unregister(overlayId);
  unlockScroll();
  isTopOverlay.value = false;
  // 出栈后再结束焦点会话，restoreFocus 才能落到新栈顶或触发元素
  sessionOpen.value = false;
}

// 关闭时同步出栈，保证焦点恢复前新栈顶已隔离就绪
watch(
  () => props.visible,
  (visible, wasVisible) => {
    if (!visible && wasVisible) {
      teardownOverlay();
    }
  },
  { flush: "sync" },
);

watch(
  () => props.visible,
  async (visible) => {
    if (!visible) {
      return;
    }
    lockScroll();
    await nextTick();
    if (!props.visible || !containerRef.value) {
      unlockScroll();
      return;
    }
    resolveInitialFocus();
    register(overlayId, containerRef.value);
    sessionOpen.value = true;
    syncTopState();
  },
);

watch(trapEnabled, (enabled) => {
  if (enabled) {
    document.addEventListener("keydown", onDocumentKeyDown, true);
  } else {
    document.removeEventListener("keydown", onDocumentKeyDown, true);
  }
});

const unsubscribeStack = subscribe(() => {
  syncTopState();
});

onBeforeUnmount(() => {
  document.removeEventListener("keydown", onDocumentKeyDown, true);
  unsubscribeStack();
  teardownOverlay();
});
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible"
      :id="overlayId"
      ref="containerRef"
      :class="rootClass"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      :aria-describedby="hasDescription ? descriptionId : undefined"
      tabindex="-1"
    >
      <div :class="overlayClass" aria-hidden="true" @click="onOverlayClick" />

      <aside :class="[panelClass, panelSideClass]">
        <header :class="headerClass">
          <div class="min-w-0">
            <h2 :id="titleId">
              <slot name="title">{{ title }}</slot>
            </h2>
            <p
              v-if="hasDescription"
              :id="descriptionId"
              class="mt-0.5 text-xs text-black/55"
            >
              {{ description }}
            </p>
          </div>
          <button
            v-if="showCloseButton"
            ref="closeButtonRef"
            type="button"
            class="admin-drawer-close"
            :disabled="busy"
            aria-label="关闭"
            @click="requestClose"
          />
        </header>

        <div :class="contentClass">
          <slot />
        </div>

        <div v-if="$slots.footer" :class="footerClass">
          <slot name="footer" />
        </div>
      </aside>
    </div>
  </Teleport>
</template>

<style scoped>
.base-drawer-panel-right {
  inset: 0 0 0 auto;
}

.base-drawer-panel-left {
  inset: 0 auto 0 0;
  border-left: 0;
  border-right: 1px solid var(--border-muted);
  box-shadow: 10px 0 28px rgba(15, 23, 42, 0.14);
}

/* 遮罩改为根容器内 absolute，便于随 Teleport 根节点一并隔离 */
.admin-preference-mask {
  position: absolute;
  inset: 0;
  z-index: 0;
  border: 0;
  background: rgba(0, 0, 0, 0.42);
  cursor: pointer;
}
</style>
