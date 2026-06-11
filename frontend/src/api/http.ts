import axios from "axios";
import { showGlobalToast } from "@/composables/useGlobalToast";

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

const http = axios.create({
  baseURL: "/",
  timeout: 15000,
});

http.interceptors.response.use(
  (response) => response,
  (error) => {
    const data = error?.response?.data;
    const raw =
      (typeof data === "string" ? data : "") ||
      data?.status_message ||
      data?.error ||
      data?.message ||
      data?.msg ||
      error?.message ||
      "";

    const status: number | undefined = error?.response?.status;
    const msg = friendlyMessage(status, raw);

    // 全局居中 toast 提示
    showGlobalToast(msg, "error");

    return Promise.reject(new Error(msg));
  },
);

export default http;
