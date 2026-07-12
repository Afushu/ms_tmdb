<script setup lang="ts">
import type { AdminAutoSyncLogItem } from "@/api/admin";
import DataListShell from "@/components/common/DataListShell.vue";
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

const columns = ["时间", "策略", "状态", "耗时", "检查/同步/失败", "摘要", "操作"];
</script>

<template>
  <DataListShell
    grid-class="logs-grid-auto-sync"
    :columns="columns"
    :loading="loading"
    :empty="!loading && items.length === 0"
    empty-text="暂无执行日志"
    loading-text="日志加载中..."
  >
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

      <button class="btn-soft-xs logs-action" type="button" @click="emit('open-detail', item)">详情</button>
    </article>
  </DataListShell>
</template>
