import axios, { type AxiosRequestConfig } from "axios";
import { showGlobalToast } from "@/composables/useGlobalToast";

declare module "axios" {
  interface AxiosRequestConfig {
    /** 是否展示全局错误 Toast；默认 true（`!== false` 时展示） */
    showErrorToast?: boolean;
  }
}

/** 读取 API 可选请求配置，经 API 层透传给 Axios */
export type RequestOptions = Pick<AxiosRequestConfig, "showErrorToast">;

/** 将 HTTP 错误映射为用户友好的中文提示 */
function friendlyMessage(status?: number, raw?: string): string {
  if (status) {
    if (status >= 500) return "服务器异常，请稍后再试";
    if (status === 404) return "请求的资源不存在";
    if (status === 403) return "没有操作权限";
    if (status === 401) return "未授权，请重新登录";
    if (status === 400) return "请求参数有误";
    if (status === 429) return "请求过于频繁，请稍后再试";
  }
  // 网络错误 / 超时
  if (raw && /timeout/i.test(raw)) return "请求超时，请检查网络";
  if (raw && /network/i.test(raw)) return "网络异常，请检查连接";
  return "请求失败，请稍后再试";
}

function messageText(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

/** 优先读取后端业务错误消息，避免把可操作提示泛化成“服务器异常”。 */
function backendMessage(data: unknown): string {
  if (!data) return "";
  if (typeof data === "string") {
    try {
      return backendMessage(JSON.parse(data) as unknown);
    } catch {
      const text = data.trim();
      return /^<!doctype|^<html/i.test(text) ? "" : text;
    }
  }
  if (typeof data !== "object") return "";

  const payload = data as Record<string, unknown>;
  const candidates = [payload.msg, payload.message, payload.status_message, payload.error];
  for (const item of candidates) {
    const text = messageText(item);
    if (text) return text;
  }
  return "";
}

const http = axios.create({
  baseURL: "/",
  timeout: 15000,
});

http.interceptors.response.use(
  (response) => response,
  (error) => {
    const data = error?.response?.data;
    const status: number | undefined = error?.response?.status;
    const msg = backendMessage(data) || friendlyMessage(status, error?.message || "");

    // 默认展示全局 Toast；显式 showErrorToast: false 时静默，由页面构建区域失败态
    if (error?.config?.showErrorToast !== false) {
      showGlobalToast(msg, "error");
    }

    return Promise.reject(new Error(msg));
  },
);

export default http;
