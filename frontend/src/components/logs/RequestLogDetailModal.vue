<script setup lang="ts">
import JsonFoldViewer from "@/components/common/JsonFoldViewer.vue";
import type { RequestLogDetail } from "@/utils/logFormatters";
import {
  bodyText,
  detailEndpointDisplay,
  detailEndpointSource,
  detailEndpointTitle,
  formatBytes,
  formatDateTime,
  formatDuration,
  formatStatusCode,
  isAccessDetail,
} from "@/utils/logFormatters";

type RequestLogTab = "access" | "tmdb";

defineProps<{
  visible: boolean;
  loading: boolean;
  detailType: RequestLogTab;
  detail: RequestLogDetail | null;
}>();

const emit = defineEmits<{
  close: [];
}>();
</script>

<template>
  <div
    v-if="visible"
    class="fixed inset-0 z-[1300] flex items-center justify-center bg-black/55 p-3 sm:p-4"
    role="dialog"
    aria-modal="true"
    @click.self="emit('close')"
  >
    <div class="panel-glass logs-detail-modal max-h-[92vh] w-full max-w-6xl overflow-hidden rounded-lg">
      <div class="modal-header-dark">
        <div>
          <p class="section-label">{{ detailType === "access" ? "Access Detail" : "TMDB Detail" }}</p>
          <h4 class="text-base font-semibold">日志详情</h4>
        </div>
        <button class="btn-soft px-3 py-1.5" @click="emit('close')">关闭</button>
      </div>

      <div class="logs-detail-content">
        <p v-if="loading" class="text-sm text-black/60">详情加载中...</p>

        <template v-if="detail">
          <div class="logs-detail-overview">
            <div class="logs-detail-overview-main">
              <span class="logs-method">{{ detail.method }}</span>
              <code :title="detailEndpointTitle(detail)">{{ detailEndpointDisplay(detail) }}</code>
            </div>

            <div class="logs-detail-overview-meta">
              <span
                ><small>状态</small
                ><strong :title="detail.error_message || 'ok'">{{ formatStatusCode(detail.status_code) }}</strong></span
              >
              <span><small>耗时</small><strong>{{ formatDuration(detail.duration_ms) }}</strong></span>
              <span><small>时间</small><strong>{{ formatDateTime(detail.created_at) }}</strong></span>
              <span>
                <small>{{ detailType === "access" ? "来源" : "上游" }}</small>
                <strong>{{ detailEndpointSource(detail) }}</strong>
              </span>
              <span><small>请求正文</small><strong>{{ formatBytes(detail.request_body_bytes) }}</strong></span>
              <span v-if="isAccessDetail(detail)">
                <small>请求 ID</small>
                <strong :title="detail.request_id">{{ detail.request_id || "-" }}</strong>
              </span>
              <span v-if="isAccessDetail(detail)">
                <small>原始请求</small>
                <strong :title="detail.request_uri">{{ detail.request_uri || "-" }}</strong>
              </span>
            </div>
          </div>

          <pre v-if="detail.request_body" class="settings-diff-pre logs-detail-request-body">{{ bodyText(detail.request_body) }}</pre>

          <div class="settings-detail-section logs-detail-response-section">
            <div class="settings-detail-section-header">
              <div>
                <h5 class="text-sm font-semibold">响应正文</h5>
                <p class="settings-note">{{ formatBytes(detail.response_body_bytes) }}</p>
              </div>
            </div>
            <JsonFoldViewer :value="detail.response_body" />
          </div>
        </template>
      </div>
    </div>
  </div>
</template>
