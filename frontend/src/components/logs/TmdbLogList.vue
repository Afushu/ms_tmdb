<script setup lang="ts">
import type { AdminTmdbRequestLogItem } from "@/api/admin";
import {
  bodyMeta,
  formatDateTime,
  formatDuration,
  formatStatusCode,
  statusClass,
  trimMiddle,
  upstreamHost,
} from "@/utils/logFormatters";

defineProps<{
  items: AdminTmdbRequestLogItem[];
  loading: boolean;
}>();

const emit = defineEmits<{
  "open-detail": [item: AdminTmdbRequestLogItem];
}>();
</script>

<template>
  <div class="logs-list-shell">
    <div class="logs-list-head logs-grid-tmdb">
      <span>时间</span>
      <span>上游路径</span>
      <span>状态</span>
      <span>耗时</span>
      <span>响应正文</span>
      <span>操作</span>
    </div>

    <article v-for="item in items" :key="item.id" class="logs-row logs-grid-tmdb">
      <time class="logs-time">{{ formatDateTime(item.created_at) }}</time>

      <div class="logs-main">
        <div class="logs-path-line">
          <span class="logs-method">{{ item.method }}</span>
          <code :title="item.path || '-'">{{ item.path || "-" }}</code>
        </div>
        <p class="logs-host" :title="upstreamHost(item.url)">{{ upstreamHost(item.url) }}</p>
        <p v-if="item.error_message" class="logs-error-line" :title="item.error_message">
          {{ trimMiddle(item.error_message, 120) }}
        </p>
      </div>

      <div>
        <span class="settings-status-pill" :class="statusClass(item.status_code)">
          {{ formatStatusCode(item.status_code) }}
        </span>
      </div>

      <strong class="logs-duration">{{ formatDuration(item.duration_ms) }}</strong>

      <div class="logs-body-cell">
        <span>{{ bodyMeta(item.response_body_bytes, item.response_body_truncated) }}</span>
      </div>

      <button class="btn-soft-xs logs-action" @click="emit('open-detail', item)">详情</button>
    </article>

    <p v-if="!loading && items.length === 0" class="logs-empty">暂无 TMDB 请求日志</p>
    <p v-if="loading" class="logs-empty">日志加载中...</p>
  </div>
</template>
