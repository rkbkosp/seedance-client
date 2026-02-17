<template>
  <div ref="container" class="h-full w-full" />
</template>

<script setup>
import { onMounted, ref } from 'vue';

const props = defineProps({
  projectId: {
    type: Number,
    required: true,
  },
});

const container = ref(null);

onMounted(async () => {
  if (!container.value) return;
  try {
    const { renderStoryboardV2Page } = await import('@/storyboard_v2.js');
    const data = await window.go.main.App.GetProject(props.projectId);
    await renderStoryboardV2Page(container.value, props.projectId, data);
  } catch (err) {
    container.value.innerHTML = `<div style="padding:12px;color:#f87171">加载 v2 页面失败：${String(err?.message || err || '')}</div>`;
  }
});
</script>
