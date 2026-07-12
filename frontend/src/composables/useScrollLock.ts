/**
 * 模块级滚动锁定：多层阻断容器共享同一计数，归零后才恢复 body overflow。
 * 不维护任何业务弹窗可见状态。
 */

let lockCount = 0;
let previousOverflow = "";

export function useScrollLock() {
  function lock() {
    if (typeof document === "undefined") {
      return;
    }
    if (lockCount === 0) {
      previousOverflow = document.body.style.overflow;
      document.body.style.overflow = "hidden";
    }
    lockCount += 1;
  }

  function unlock() {
    if (typeof document === "undefined") {
      return;
    }
    if (lockCount <= 0) {
      lockCount = 0;
      return;
    }
    lockCount -= 1;
    if (lockCount === 0) {
      document.body.style.overflow = previousOverflow;
      previousOverflow = "";
    }
  }

  return {
    lock,
    unlock,
  };
}
