<script setup lang="ts">
import { useGlobalToast } from "@/composables/useGlobalToast";

const { state, closeGlobalToast } = useGlobalToast();
</script>

<template>
  <Teleport to="body">
    <Transition name="global-toast-fade">
      <div
        v-if="state.visible"
        class="global-toast-overlay"
        @click.self="closeGlobalToast"
      >
        <div class="global-toast" :class="`global-toast-${state.tone}`" role="alert" aria-live="assertive">
          <span class="global-toast-icon" aria-hidden="true"></span>
          <span class="global-toast-text">{{ state.text }}</span>
          <button class="global-toast-close" type="button" aria-label="关闭提示" @click="closeGlobalToast"></button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
