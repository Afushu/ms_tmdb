<script setup lang="ts">
import type { AdminAutoSyncLogItem } from "@/api/admin";
import {
  autoSyncStatusClass,
  formatAutoSyncStatus,
  formatDateTime,
  formatDuration,
  formatMode,
  summarizeMessage,
} from "@/utils/logFormatters";

defineProps<{
  items: AdminAutoSyncLogItem[];
  loading: boolean;
}>();

const emit = defineEmits<{
  "open-detail": [item: AdminAutoSyncLogItem];
}>();
</script>

<template>
  <div class="logs-list-shell">
    <div class="logs-list-head logs-grid-auto-sync">
      <span>时间</span>
      <span>策略</span>
      <span>状态</span>
      <span>耗时</span>
      <span>检查/同步/失败</span>
      <span>摘要</span>
      <span>操作</span>
    </div>

    <article v-for="item in items" :key="item.id" class="logs-row logs-grid-auto-sync">
      <time class="logs-time">{{ formatDateTime(item.triggered_at) }}</time>

      <div class="logs-main">
        <div class="logs-path-line">
          <span class="logs-method">SYNC</span>
          <code :title="formatMode(item.mode)">{{ formatMode(item.mode) }}</code>
        </div>
        <p class="logs-host" :title="item.cron_expr || '-'">{{ item.cron_expr || "-" }} · 批大小 {{ item.batch_size }}</p>
      </div>

      <div>
        <span class="settings-status-pill" :class="autoSyncStatusClass(item.status)">
          {{ formatAutoSyncStatus(item.status) }}
        </span>
      </div>

      <strong class="logs-duration">{{ formatDuration(item.duration_ms) }}</strong>

      <div class="logs-body-cell">
        <span>检查 {{ item.checked }}</span>
        <small>同步 {{ item.synced }} · 失败 {{ item.failed }}</small>
      </div>

      <div class="logs-source">
        <strong :title="item.message">{{ summarizeMessage(item.message) }}</strong>
        <span>{{ formatDateTime(item.finished_at || item.created_at) }}</span>
      </div>

      <button class="btn-soft-xs logs-action" @click="emit('open-detail', item)">详情</button>
    </article>

    <p v-if="!loading && items.length === 0" class="logs-empty">暂无执行日志</p>
    <p v-if="loading" class="logs-empty">日志加载中...</p>
  </div>
</template>
