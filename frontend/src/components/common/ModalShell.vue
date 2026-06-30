<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    visible: boolean;
    title: string;
    maxWidthClass?: string;
    contentClass?: string;
    footerClass?: string;
    variant?: "glass" | "vben";
  }>(),
  {
    maxWidthClass: "max-w-5xl",
    contentClass: "modal-scroll-content max-h-[calc(88vh-120px)] overflow-y-auto px-4 py-4 sm:px-6",
    footerClass: "",
    variant: "glass",
  },
);

const emit = defineEmits<{
  close: [];
}>();
</script>

<template>
  <div
    v-if="visible"
    class="fixed inset-0 z-[1300] flex items-center justify-center p-3 sm:p-6"
    role="dialog"
    aria-modal="true"
  >
    <div class="absolute inset-0 bg-black/60 backdrop-blur-[2px]" @click="emit('close')" />
    <section
      :class="[
        props.variant === 'vben' ? 'vben-modal-shell' : 'panel-glass',
        'relative z-10 w-full overflow-hidden rounded-lg',
        props.maxWidthClass,
      ]"
    >
      <div
        :class="
          props.variant === 'vben'
            ? 'vben-modal-header'
            : 'sticky top-0 z-10 flex items-center justify-between gap-3 border-b border-white/10 bg-black/35 px-4 py-3 backdrop-blur sm:px-6'
        "
      >
        <h3 :class="props.variant === 'vben' ? 'vben-modal-title' : 'text-sm font-semibold'">{{ title }}</h3>
        <button
          :class="props.variant === 'vben' ? 'vben-modal-close' : 'btn-soft px-3 py-1.5 text-xs'"
          :aria-label="props.variant === 'vben' ? '关闭' : undefined"
          @click="emit('close')"
        >
          {{ props.variant === "vben" ? "×" : "关闭" }}
        </button>
      </div>

      <div :class="props.contentClass">
        <slot />
      </div>

      <div v-if="$slots.footer" :class="[props.variant === 'vben' ? 'vben-modal-footer' : '', props.footerClass]">
        <slot name="footer" />
      </div>
    </section>
  </div>
</template>

<style scoped>
.vben-modal-shell {
  color: var(--text-main);
  background: var(--surface);
  border: 1px solid var(--border-muted);
  box-shadow:
    0 12px 28px rgba(0, 0, 0, 0.24),
    0 24px 56px rgba(0, 0, 0, 0.28);
}

.vben-modal-header {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 48px;
  padding: 0 16px 0 20px;
  background: var(--surface);
  border-bottom: 1px solid var(--border-muted);
}

.vben-modal-title {
  color: var(--text-main);
  font-size: 15px;
  font-weight: 600;
  line-height: 1.4;
}

.vben-modal-close {
  display: inline-flex;
  width: 30px;
  height: 30px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 6px;
  color: var(--text-muted);
  background: transparent;
  font-size: 22px;
  line-height: 1;
  transition:
    color 0.16s ease,
    background-color 0.16s ease;
}

.vben-modal-close:hover {
  color: var(--text-main);
  background: var(--surface-muted);
}

.vben-modal-close:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px var(--accent-ring);
}

.vben-modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 20px;
  background: var(--surface);
  border-top: 1px solid var(--border-muted);
}
</style>
