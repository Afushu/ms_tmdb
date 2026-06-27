<script setup lang="ts">
import type { AdminProxyAccessLogItem } from "@/api/admin";
import {
  accessPath,
  bodyMeta,
  formatDateTime,
  formatDuration,
  formatStatusCode,
  statusClass,
  trimMiddle,
} from "@/utils/logFormatters";

defineProps<{
  items: AdminProxyAccessLogItem[];
  loading: boolean;
}>();

const emit = defineEmits<{
  "open-detail": [item: AdminProxyAccessLogItem];
}>();
</script>

<template>
  <div class="logs-list-shell">
    <div class="logs-list-head logs-grid-access">
      <span>时间</span>
      <span>请求</span>
      <span>状态</span>
      <span>耗时</span>
      <span>正文</span>
      <span>来源</span>
      <span>操作</span>
    </div>

    <article v-for="item in items" :key="item.id" class="logs-row logs-grid-access">
      <time class="logs-time">{{ formatDateTime(item.created_at) }}</time>

      <div class="logs-main">
        <div class="logs-path-line">
          <span class="logs-method">{{ item.method }}</span>
          <code :title="accessPath(item.request_uri)">{{ accessPath(item.request_uri) }}</code>
        </div>
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
        <span>响应 {{ bodyMeta(item.response_body_bytes, item.response_body_truncated) }}</span>
        <small>请求 {{ bodyMeta(item.request_body_bytes, item.request_body_truncated) }}</small>
      </div>

      <div class="logs-source">
        <strong>{{ item.client_ip || "-" }}</strong>
        <span :title="item.user_agent">{{ item.user_agent || "-" }}</span>
      </div>

      <button class="btn-soft-xs logs-action" @click="emit('open-detail', item)">详情</button>
    </article>

    <p v-if="!loading && items.length === 0" class="logs-empty">暂无外部访问日志</p>
    <p v-if="loading" class="logs-empty">日志加载中...</p>
  </div>
</template>
