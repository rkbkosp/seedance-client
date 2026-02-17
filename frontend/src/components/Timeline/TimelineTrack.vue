<template>
  <section class="panel-surface h-full p-3 flex flex-col min-h-0">
    <header class="mb-3">
      <h3 class="panel-title">Timeline</h3>
    </header>

    <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-2 overflow-auto pb-2">
      <button
        v-for="shot in storyboards"
        :key="shot.id"
        class="w-full rounded-md border px-3 py-2 text-left transition"
        :class="shot.id === selectedShotId ? 'border-zinc-500 bg-zinc-800/80' : 'border-zinc-800 bg-zinc-900 hover:border-zinc-700'"
        @click="$emit('select-shot', shot.id)"
      >
        <div class="text-xs text-zinc-400">{{ shot.shot_no || `镜头 ${shot.id}` }}</div>
        <div class="mt-1 text-sm truncate text-zinc-200">{{ shot.frame_content || '未填写画面描述' }}</div>
      </button>
    </div>
  </section>
</template>

<script setup>
defineProps({
  storyboards: {
    type: Array,
    default: () => [],
  },
  selectedShotId: {
    type: Number,
    default: null,
  },
});

defineEmits(['select-shot']);
</script>
