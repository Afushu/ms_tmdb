<script setup lang="ts">
defineProps<{
  visible: boolean;
  busy: boolean;
  label: string;
}>();

const emit = defineEmits<{
  close: [];
  confirm: [];
}>();
</script>

<template>
  <div
    v-if="visible"
    class="fixed inset-0 z-[1300] flex items-center justify-center bg-black/45 p-4"
    role="dialog"
    aria-modal="true"
    @click.self="emit('close')"
  >
    <div class="panel-glass w-full max-w-md rounded-lg p-5">
      <h4 class="text-base font-semibold text-red-700">确认清空日志</h4>
      <p class="mt-2 text-sm text-black/70">将清空当前视图的{{ label }}，清空后无法恢复。</p>

      <div class="mt-5 flex items-center justify-end gap-2">
        <button class="btn-soft disabled:opacity-60" :disabled="busy" @click="emit('close')">取消</button>
        <button class="btn-danger-soft disabled:opacity-60" :disabled="busy" @click="emit('confirm')">
          {{ busy ? "清空中..." : "确认清空" }}
        </button>
      </div>
    </div>
  </div>
</template>
