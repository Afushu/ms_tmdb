import { onBeforeUnmount, ref } from "vue";

export type ToastNoticeTone = "info" | "success";

export function useToastNotice(duration = 2600) {
  const toastVisible = ref(false);
  const toastText = ref("");
  const toastTone = ref<ToastNoticeTone>("success");
  let toastTimer: number | undefined;

  function closeToastNotice() {
    if (toastTimer) {
      window.clearTimeout(toastTimer);
      toastTimer = undefined;
    }
    toastVisible.value = false;
  }

  function showToastNotice(message: string, tone: ToastNoticeTone = "success") {
    const text = message.trim();
    if (!text) {
      return;
    }
    if (toastTimer) {
      window.clearTimeout(toastTimer);
    }
    toastText.value = text;
    toastTone.value = tone;
    toastVisible.value = true;
    toastTimer = window.setTimeout(() => {
      toastVisible.value = false;
      toastTimer = undefined;
    }, duration);
  }

  onBeforeUnmount(closeToastNotice);

  return {
    toastVisible,
    toastText,
    toastTone,
    showToastNotice,
    closeToastNotice,
  };
}
