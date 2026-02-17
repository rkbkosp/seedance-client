<template>
  <div class="h-screen bg-zinc-950 text-zinc-100 flex flex-col overflow-hidden">
    <header class="h-12 border-b border-zinc-800 bg-zinc-900/90 backdrop-blur px-4 flex items-center justify-between">
      <RouterLink to="/" class="text-sm font-medium tracking-wide text-zinc-100 hover:text-zinc-300">
        Seedance Workbench
      </RouterLink>
      <div class="min-w-0 max-w-[70%] truncate text-right">
        <slot name="topbar" />
      </div>
    </header>

    <main class="flex-1 min-h-0 grid grid-cols-[clamp(170px,20vw,250px)_minmax(0,1fr)_clamp(190px,22vw,290px)] 2xl:grid-cols-[clamp(220px,22vw,280px)_minmax(0,1fr)_clamp(240px,24vw,320px)] overflow-hidden">
      <aside class="border-r border-zinc-800 bg-zinc-900/70 p-3 min-h-0 overflow-hidden">
        <slot name="left" />
      </aside>

      <section class="min-w-0 p-3 overflow-hidden">
        <slot name="center" />
      </section>

      <aside class="border-l border-zinc-800 bg-zinc-900/70 p-3 min-h-0 overflow-hidden">
        <slot name="right" />
      </aside>
    </main>
  </div>
</template>

<script setup>
import { RouterLink } from 'vue-router';
</script>
