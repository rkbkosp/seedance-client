<template>
  <button
    class="w-full text-left panel-surface p-3 transition border"
    :class="selected ? 'border-zinc-500' : 'border-zinc-800 hover:border-zinc-700'"
    @click="$emit('select', shot.id)"
  >
    <div class="flex items-center justify-between gap-2">
      <span class="text-sm font-medium text-zinc-100">{{ shot.shot_no || `镜头 ${shot.id}` }}</span>
      <span class="text-[11px] px-2 py-0.5 rounded bg-zinc-800 text-zinc-300">{{ statusLabel }}</span>
    </div>

    <p class="mt-2 text-xs text-zinc-400 line-clamp-2">{{ shot.frame_content || '未填写画面描述' }}</p>
  </button>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  shot: {
    type: Object,
    required: true,
  },
  selected: {
    type: Boolean,
    default: false,
  },
});

defineEmits(['select']);

const statusLabel = computed(() => {
  const take = props.shot.active_take || props.shot.takes?.[props.shot.takes.length - 1];
  return take?.status || 'idle';
});
</script>
