<template>
  <section class="panel-surface h-full p-3 flex flex-col min-h-0">
    <header class="flex items-center justify-between mb-3">
      <h3 class="panel-title">Stage Monitor</h3>
      <span class="text-xs text-zinc-400">{{ statusText }}</span>
    </header>

    <div class="flex-1 min-h-0 rounded-md border border-zinc-800 bg-zinc-950/70 overflow-hidden flex items-center justify-center">
      <video
        v-if="videoSrc"
        :src="videoSrc"
        controls
        class="max-w-full max-h-full"
      />
      <img
        v-else-if="frameSrc"
        :src="frameSrc"
        alt="Last frame"
        class="max-w-full max-h-full object-contain"
      >
      <div v-else class="text-sm text-zinc-500">暂无可预览素材</div>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  take: {
    type: Object,
    default: null,
  },
});

const videoSrc = computed(() => props.take?.local_video_path || props.take?.video_url || '');
const frameSrc = computed(() => props.take?.local_last_frame_path || props.take?.last_frame_url || '');
const statusText = computed(() => props.take?.status || 'idle');
</script>
