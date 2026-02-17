<template>
  <div class="h-screen bg-zinc-950 text-zinc-100 p-6 overflow-auto">
    <div class="max-w-7xl mx-auto space-y-5">
      <div class="flex items-end justify-between gap-4">
        <div>
          <h1 class="text-2xl font-semibold">Projects</h1>
          <p class="text-sm text-zinc-400">管理你的 Seedance 项目</p>
        </div>

        <div class="flex gap-2 items-center">
          <n-input v-model:value="draft.name" placeholder="New Project Name" class="w-56" />
          <n-select v-model:value="draft.version" :options="versionOptions" class="w-40" />
          <n-select v-model:value="draft.ratio" :options="ratioOptions" class="w-28" />
          <n-button type="primary" @click="createProject">创建</n-button>
        </div>
      </div>

      <div v-if="error" class="panel-surface p-3 text-red-300 text-sm">{{ error }}</div>

      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
        <article v-for="project in projects" :key="project.id" class="panel-surface p-4">
          <div class="flex justify-between items-start gap-3">
            <div>
              <h2 class="font-medium text-zinc-100">{{ project.name }}</h2>
              <p class="text-xs text-zinc-400">ID {{ project.id }} · {{ project.aspect_ratio || '16:9' }}</p>
            </div>
            <n-button text type="error" @click="removeProject(project.id)">删除</n-button>
          </div>

          <div class="mt-4 flex justify-between items-center">
            <span class="text-xs text-zinc-500">{{ project.model_version || 'v1.x' }}</span>
            <RouterLink :to="`/projects/${project.id}`" class="text-sm text-zinc-200 hover:text-zinc-50">
              打开 →
            </RouterLink>
          </div>
        </article>
      </div>

      <section class="panel-surface p-4">
        <h2 class="text-sm font-medium text-zinc-200 mb-3">开支统计</h2>
        <div class="grid grid-cols-2 xl:grid-cols-4 gap-3">
          <div class="rounded border border-zinc-800 bg-zinc-950/60 p-3">
            <p class="text-[11px] text-zinc-500">总视频数</p>
            <p class="text-xl font-semibold text-zinc-100">{{ formatCount(stats.total_videos) }}</p>
          </div>
          <div class="rounded border border-zinc-800 bg-zinc-950/60 p-3">
            <p class="text-[11px] text-zinc-500">总 Token</p>
            <p class="text-xl font-semibold text-zinc-100">{{ formatCount(stats.total_token_usage) }}</p>
            <p class="text-[11px] text-zinc-500">输出 Token</p>
          </div>
          <div class="rounded border border-zinc-800 bg-zinc-950/60 p-3">
            <p class="text-[11px] text-zinc-500">总成本</p>
            <p class="text-xl font-semibold text-amber-300">{{ formatCurrency(stats.total_cost) }}</p>
            <p class="text-[11px] text-zinc-500">Estimated</p>
          </div>
          <div class="rounded border border-zinc-800 bg-zinc-950/60 p-3">
            <p class="text-[11px] text-zinc-500">总节省</p>
            <p class="text-xl font-semibold text-emerald-300">{{ formatCurrency(stats.total_savings) }}</p>
            <p class="text-[11px] text-zinc-500">vs Platform Price</p>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { RouterLink } from 'vue-router';
import { onMounted, reactive, ref } from 'vue';
import { NButton, NInput, NSelect } from 'naive-ui';

const projects = ref([]);
const stats = ref({});
const error = ref('');

const draft = reactive({
  name: '',
  version: 'v2.0',
  ratio: '16:9',
});

const versionOptions = [
  { label: 'Seedance 2.0', value: 'v2.0' },
  { label: 'Seedance 1.x', value: 'v1.x' },
];

const ratioOptions = [
  { label: '16:9', value: '16:9' },
  { label: '9:16', value: '9:16' },
  { label: '1:1', value: '1:1' },
  { label: '21:9', value: '21:9' },
];

const countFormatter = new Intl.NumberFormat('zh-CN', {
  maximumFractionDigits: 0,
});

const currencyFormatter = new Intl.NumberFormat('zh-CN', {
  style: 'currency',
  currency: 'CNY',
  minimumFractionDigits: 4,
  maximumFractionDigits: 4,
});

function formatCount(value) {
  return countFormatter.format(Number(value || 0));
}

function formatCurrency(value) {
  return currencyFormatter.format(Number(value || 0));
}

async function loadProjects() {
  error.value = '';
  try {
    const data = await window.go.main.App.GetProjects();
    projects.value = data.projects || [];
    stats.value = data.stats || {};
  } catch (err) {
    error.value = String(err?.message || err || '加载项目失败');
  }
}

async function createProject() {
  const name = draft.name.trim();
  if (!name) return;

  error.value = '';
  try {
    await window.go.main.App.CreateProject({
      name,
      model_version: draft.version,
      aspect_ratio: draft.ratio,
    });
    draft.name = '';
    await loadProjects();
  } catch (err) {
    error.value = String(err?.message || err || '创建项目失败');
  }
}

async function removeProject(id) {
  if (!window.confirm('确认删除该项目？')) return;

  error.value = '';
  try {
    await window.go.main.App.DeleteProject(id);
    await loadProjects();
  } catch (err) {
    error.value = String(err?.message || err || '删除项目失败');
  }
}

onMounted(loadProjects);
</script>
