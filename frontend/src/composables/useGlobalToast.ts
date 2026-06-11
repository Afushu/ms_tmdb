import { reactive, readonly } from "vue";

export type GlobalToastTone = "info" | "success" | "error";

interface GlobalToastState {
  visible: boolean;
  text: string;
  tone: GlobalToastTone;
}

const state = reactive<GlobalToastState>({
  visible: false,
  text: "",
  tone: "success",
});

let hideTimer: number | undefined;

export function showGlobalToast(message: string, tone: GlobalToastTone = "success", duration = 3000) {
  const text = message.trim();
  if (!text) return;
  if (hideTimer) window.clearTimeout(hideTimer);
  state.text = text;
  state.tone = tone;
  state.visible = true;
  hideTimer = window.setTimeout(() => {
    state.visible = false;
    hideTimer = undefined;
  }, duration);
}

function closeGlobalToast() {
  if (hideTimer) {
    window.clearTimeout(hideTimer);
    hideTimer = undefined;
  }
  state.visible = false;
}

export function useGlobalToast() {
  return {
    state: readonly(state),
    showGlobalToast,
    closeGlobalToast,
  };
}
