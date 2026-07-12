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
    closeButtonClass?: string;
    closeButtonText?: string;
    maxWidthClass?: string;
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
    closeButtonClass: "btn-soft px-3 py-1.5 text-xs disabled:opacity-60",
    closeButtonText: "关闭",
    maxWidthClass: "max-w-5xl",
    panelClass: "panel-glass",
    headerClass:
      "sticky top-0 z-10 flex items-center justify-between gap-3 border-b border-white/10 bg-black/35 px-4 py-3 backdrop-blur sm:px-6",
    contentClass: "modal-scroll-content max-h-[calc(88vh-120px)] overflow-y-auto px-4 py-4 sm:px-6",
    footerClass: "",
    overlayClass: "absolute inset-0 bg-black/60 backdrop-blur-[2px]",
    rootClass: "fixed inset-0 z-[1300] flex items-center justify-center p-3 sm:p-6",
  },
);

const emit = defineEmits<{
  close: [];
}>();

const overlayId = createOverlayId("dialog");
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

      <section
        :class="[
          panelClass,
          'relative z-10 w-full overflow-hidden rounded-lg',
          maxWidthClass,
        ]"
      >
        <header :class="headerClass">
          <div class="min-w-0">
            <h3 :id="titleId" class="min-w-0">
              <slot name="title">
                <span class="block truncate text-sm font-semibold">{{ title }}</span>
              </slot>
            </h3>
            <p
              v-if="hasDescription"
              :id="descriptionId"
              class="mt-0.5 text-xs text-black/60"
            >
              {{ description }}
            </p>
          </div>
          <button
            v-if="showCloseButton"
            ref="closeButtonRef"
            type="button"
            :class="closeButtonClass"
            :disabled="busy"
            aria-label="关闭"
            @click="requestClose"
          >
            {{ closeButtonText }}
          </button>
        </header>

        <div :class="contentClass">
          <slot />
        </div>

        <div v-if="$slots.footer" :class="footerClass">
          <slot name="footer" />
        </div>
      </section>
    </div>
  </Teleport>
</template>
