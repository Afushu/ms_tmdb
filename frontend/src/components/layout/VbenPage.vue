<script setup lang="ts">
withDefaults(defineProps<{
  description?: string;
  section?: string;
  showHeader?: boolean;
  title: string;
}>(), {
  description: "",
  section: "",
  showHeader: true,
});
</script>

<template>
  <section class="vben-page">
    <header v-if="showHeader" class="vben-page-header">
      <div class="vben-page-heading">
        <p v-if="section" class="vben-page-section">{{ section }}</p>
        <h2 class="vben-page-title">{{ title }}</h2>
        <p v-if="description" class="vben-page-description">{{ description }}</p>
      </div>
      <div v-if="$slots.extra" class="vben-page-extra">
        <slot name="extra"></slot>
      </div>
    </header>

    <div class="vben-page-content">
      <slot></slot>
    </div>

    <footer v-if="$slots.footer" class="vben-page-footer">
      <slot name="footer"></slot>
    </footer>
  </section>
</template>
