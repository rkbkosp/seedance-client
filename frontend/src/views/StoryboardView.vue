<template>
  <LegacyStoryboardV2Bridge v-if="isV2" :project-id="projectId" />

  <MainLayout v-else>
    <template #topbar>
      <div class="flex items-center gap-2">
        <RouterLink to="/" class="text-xs text-zinc-400 hover:text-zinc-200">返回项目</RouterLink>
        <span class="text-xs text-zinc-600">|</span>
        <span class="text-xs text-zinc-400">{{ projectTitle }}</span>
        <n-button size="tiny" secondary @click="handleExport">导出 FCPXML</n-button>
      </div>
    </template>

    <template #left>
      <div class="h-full min-h-0 flex flex-col gap-3">
        <section class="panel-surface p-2">
          <n-tabs v-model:value="panelMode" type="line" size="small" animated>
            <n-tab-pane name="breakdown" tab="分镜拆解" />
            <n-tab-pane name="assets" tab="资产管理" />
            <n-tab-pane name="workbench" tab="制作工作台" />
          </n-tabs>
        </section>

        <section class="panel-surface p-3 min-h-0 flex-1 overflow-auto">
          <h3 class="panel-title mb-2">Shots</h3>
          <div class="space-y-2">
            <ShotCard
              v-for="shot in workspace.storyboards"
              :key="shot.id"
              :shot="shot"
              :selected="shot.id === workspace.selectedShotId"
              @select="workspace.selectShot"
            />
          </div>
          <n-button class="mt-3 w-full" size="small" secondary @click="handleCreateShot">+ 新建分镜</n-button>
        </section>
      </div>
    </template>

    <template #center>
      <div v-if="panelMode === 'breakdown'" class="h-full min-h-0 grid grid-cols-1 2xl:grid-cols-2 gap-3 overflow-hidden">
        <section class="panel-surface p-3 overflow-auto">
          <h3 class="panel-title mb-2">分镜编辑</h3>
          <div v-if="!selectedShot" class="text-xs text-zinc-500">暂无分镜</div>
          <div v-else class="space-y-2">
            <div class="grid grid-cols-1 xl:grid-cols-2 gap-2">
              <n-input v-model:value="shotDraft.shot_no" placeholder="镜号" />
              <n-input v-model:value="shotDraft.shot_size" placeholder="景别" />
              <n-input v-model:value="shotDraft.camera_movement" placeholder="运镜" />
              <n-select v-model:value="shotDraft.estimated_duration" :options="durationOptions" />
            </div>
            <n-input v-model:value="shotDraft.frame_content" type="textarea" :autosize="{ minRows: 3, maxRows: 6 }" placeholder="画面内容" />
            <n-input v-model:value="shotDraft.sound_design" type="textarea" :autosize="{ minRows: 2, maxRows: 4 }" placeholder="声音设计" />

            <div v-for="key in refKeys" :key="key" class="panel-surface p-2">
              <div class="text-xs text-zinc-400 mb-2">{{ refLabelMap[key] }}</div>
              <div class="space-y-1">
                <div v-for="(row, index) in shotDraft[key]" :key="`${key}-${index}`" class="grid grid-cols-[1fr_1fr_2fr_auto] gap-1">
                  <n-input v-model:value="row.id" placeholder="id" />
                  <n-input v-model:value="row.name" placeholder="名称" />
                  <n-input v-model:value="row.prompt" placeholder="提示词" />
                  <n-button quaternary type="error" @click="removeRefRow(key, index)">删</n-button>
                </div>
              </div>
              <n-button size="tiny" class="mt-2" @click="addRefRow(key)">+ 新增</n-button>
            </div>

            <div class="flex flex-wrap gap-2 pt-2">
              <n-button type="primary" @click="handleSaveShot">保存</n-button>
              <n-button secondary @click="handleSplitShot">拆分</n-button>
              <n-button secondary @click="handleMergeShot">并入下一镜</n-button>
              <n-button type="error" secondary @click="handleDeleteShot">删除</n-button>
            </div>
          </div>
        </section>

        <section class="panel-surface p-3 overflow-auto">
          <h3 class="panel-title mb-2">文本/Excel 导入</h3>
          <div class="space-y-2">
            <n-select v-model:value="workspace.apiConfig.provider" :options="providerOptions" />
            <n-input v-model:value="workspace.apiConfig.llmModel" placeholder="LLM 模型" />
            <n-input v-model:value="workspace.apiConfig.baseUrl" placeholder="Base URL" :disabled="workspace.apiConfig.provider === 'ark_default'" />
            <n-input v-model:value="workspace.apiConfig.apiKey" type="password" placeholder="独立 API Key" :disabled="workspace.apiConfig.provider === 'ark_default'" />
            <n-checkbox v-model:checked="workspace.apiConfig.replaceExisting">覆盖当前分镜</n-checkbox>
            <n-input
              v-model:value="workspace.decomposeText"
              type="textarea"
              :autosize="{ minRows: 14, maxRows: 20 }"
              placeholder="粘贴 markdown 文本，或先点击导入文件"
            />
            <div class="flex gap-2">
              <n-button secondary @click="handleLoadSourceFile">导入文件</n-button>
              <n-button type="primary" @click="handleDecompose">LLM 拆解</n-button>
            </div>
          </div>
        </section>
      </div>

      <div v-else-if="panelMode === 'assets'" class="h-full min-h-0 grid grid-rows-[auto_minmax(0,1fr)] gap-3">
        <section class="panel-surface p-2">
          <n-tabs v-model:value="assetTab" type="line" size="small">
            <n-tab-pane name="character" tab="角色库" />
            <n-tab-pane name="scene" tab="场景库" />
            <n-tab-pane name="element" tab="物品库" />
            <n-tab-pane name="style" tab="风格库" />
            <n-tab-pane name="frames" tab="首尾帧" />
          </n-tabs>
        </section>

        <section class="panel-surface p-3 overflow-auto">
          <template v-if="assetTab !== 'frames'">
            <div v-if="filteredCatalogs.length === 0" class="text-xs text-zinc-500">该分类暂无资产</div>
            <div v-else class="space-y-3">
              <article v-for="catalog in filteredCatalogs" :key="catalog.id" class="border border-zinc-800 rounded p-2 space-y-2">
                <div class="text-xs text-zinc-500">{{ catalog.asset_code }}</div>
                <n-input v-model:value="getAssetDraft(catalog).name" placeholder="名称" />
                <n-input v-model:value="getAssetDraft(catalog).prompt" type="textarea" :autosize="{ minRows: 2, maxRows: 4 }" placeholder="提示词" />
                <n-input v-model:value="getAssetDraft(catalog).inputImages" type="textarea" :autosize="{ minRows: 2, maxRows: 4 }" placeholder="输入图 URL（多行）" />
                <div class="flex gap-2 flex-wrap">
                  <n-button size="small" @click="handleSaveAsset(catalog.id)">保存</n-button>
                  <n-button size="small" @click="handleUploadAsset(catalog.id)">上传</n-button>
                  <n-button size="small" secondary @click="handleGenerateAsset(catalog.id)">AI 生成</n-button>
                  <n-button size="small" secondary @click="handleGenerateAsset(catalog.id)">重试抽卡</n-button>
                </div>
                <div class="flex gap-1 flex-wrap">
                  <n-tag
                    v-for="version in catalog.versions || []"
                    :key="version.id"
                    :type="version.is_good ? 'success' : 'default'"
                    size="small"
                    class="cursor-pointer"
                    @click="handleToggleAssetGood(version.id)"
                  >
                    V{{ version.version_no }}{{ version.is_good ? '★' : '' }}
                  </n-tag>
                </div>
              </article>
            </div>
          </template>

          <template v-else>
            <div v-for="shot in workspace.storyboards" :key="`frame-${shot.id}`" class="border border-zinc-800 rounded p-2 mb-3">
              <div class="text-xs text-zinc-400 mb-2">{{ shot.shot_no || `Shot ${shot.id}` }}</div>
              <div class="grid grid-cols-1 xl:grid-cols-2 gap-2">
                <div v-for="frameType in ['start','end']" :key="`${shot.id}-${frameType}`" class="panel-surface p-2">
                  <div class="text-xs text-zinc-500 mb-2">{{ frameType === 'start' ? '首帧' : '尾帧' }}</div>
                  <div class="flex gap-1 flex-wrap mb-2">
                    <n-tag
                      v-for="version in getFrameVersions(shot, frameType)"
                      :key="version.id"
                      :type="version.is_good ? 'success' : 'default'"
                      size="small"
                      class="cursor-pointer"
                      @click="handleToggleFrameGood(version.id)"
                    >
                      V{{ version.version_no }}{{ version.is_good ? '★' : '' }}
                    </n-tag>
                  </div>
                  <n-input
                    v-model:value="getFrameDraft(shot.id, frameType, shot.frame_content || '').prompt"
                    type="textarea"
                    :autosize="{ minRows: 2, maxRows: 3 }"
                    placeholder="帧提示词"
                  />
                  <n-input
                    v-model:value="getFrameDraft(shot.id, frameType, shot.frame_content || '').inputImages"
                    class="mt-2"
                    type="textarea"
                    :autosize="{ minRows: 2, maxRows: 3 }"
                    placeholder="输入图 URL（多行）"
                  />
                  <div class="flex gap-2 mt-2">
                    <n-button size="small" @click="handleUploadFrame(shot.id, frameType)">上传</n-button>
                    <n-button size="small" secondary @click="handleGenerateFrame(shot.id, frameType)">AI 生成</n-button>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </section>
      </div>

      <div v-else class="h-full min-h-0 grid grid-rows-[minmax(0,2fr)_minmax(0,1fr)] gap-3">
        <StageMonitor :take="workspace.selectedTake" />
        <TimelineTrack
          :storyboards="workspace.storyboards"
          :selected-shot-id="workspace.selectedShotId"
          @select-shot="workspace.selectShot"
        />
      </div>
    </template>

    <template #right>
      <div class="h-full min-h-0 flex flex-col gap-3 overflow-auto">
        <section class="panel-surface p-3 space-y-2">
          <h2 class="panel-title">Take Inspector</h2>
          <div v-if="!selectedShot || !workspace.selectedTake" class="text-xs text-zinc-500">请选择分镜和 Take</div>
          <template v-else>
            <n-input v-model:value="takeDraft.prompt" type="textarea" :autosize="{ minRows: 5, maxRows: 10 }" placeholder="视频提示词" />
            <n-select v-model:value="takeDraft.modelId" :options="modelOptions" />
            <div class="grid grid-cols-1 xl:grid-cols-2 gap-2">
              <n-select v-model:value="takeDraft.duration" :options="durationOptions" />
              <n-select v-model:value="takeDraft.serviceTier" :options="serviceTierOptions" />
            </div>
            <n-input-number v-model:value="takeDraft.executionExpiresAfter" :min="60" :step="60" :disabled="takeDraft.serviceTier !== 'flex'" />
            <div class="grid grid-cols-1 gap-2">
              <n-checkbox v-model:checked="takeDraft.chainFromPrev">接力上一镜尾帧</n-checkbox>
              <n-checkbox v-model:checked="takeDraft.generateAudio" :disabled="!audioSupported">同步音效</n-checkbox>
            </div>
            <div class="flex gap-2 flex-wrap">
              <n-button type="primary" @click="handleSaveTake">保存为新 Take</n-button>
              <n-button secondary @click="handleGenerateTake(workspace.selectedTake.id)">生成当前 Take</n-button>
            </div>
          </template>
        </section>

        <section class="panel-surface p-3">
          <h2 class="panel-title mb-2">Take Versions</h2>
          <div v-if="!selectedShot" class="text-xs text-zinc-500">暂无选中分镜</div>
          <div v-else class="space-y-2">
            <button
              v-for="take in selectedShot.takes || []"
              :key="take.id"
              class="w-full text-left border rounded px-2 py-1.5 text-xs"
              :class="workspace.selectedTake?.id === take.id ? 'border-zinc-500 bg-zinc-800' : 'border-zinc-800 bg-zinc-900'"
              @click="workspace.selectTake(selectedShot.id, take.id)"
            >
              <div class="flex justify-between items-center">
                <span>{{ take.is_good ? '❤️ ' : '' }}Take #{{ take.id }}</span>
                <span class="text-zinc-400">{{ take.status || 'idle' }}</span>
              </div>
              <div class="flex gap-1 mt-1">
                <n-button size="tiny" quaternary @click.stop="handleToggleGoodTake(take.id)">{{ take.is_good ? '取消 Good' : '标记 Good' }}</n-button>
                <n-button size="tiny" quaternary @click.stop="handleGenerateTake(take.id)">生成</n-button>
              </div>
            </button>
          </div>
        </section>
      </div>
    </template>
  </MainLayout>
</template>

<script setup>
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { RouterLink, useRoute } from 'vue-router';
import {
  NButton,
  NCheckbox,
  NInput,
  NInputNumber,
  NSelect,
  NTabPane,
  NTabs,
  NTag,
  useMessage,
} from 'naive-ui';
import MainLayout from '@/layouts/MainLayout.vue';
import ShotCard from '@/components/Board/ShotCard.vue';
import StageMonitor from '@/components/Player/StageMonitor.vue';
import TimelineTrack from '@/components/Timeline/TimelineTrack.vue';
import LegacyStoryboardV2Bridge from '@/components/LegacyStoryboardV2Bridge.vue';
import { useWorkspaceStore } from '@/stores/workspace';

const route = useRoute();
const workspace = useWorkspaceStore();
const message = useMessage();

const panelMode = ref('workbench');
const assetTab = ref('character');
const isV2 = ref(false);

const shotDraft = reactive({
  shot_no: '',
  shot_size: '',
  camera_movement: '',
  frame_content: '',
  sound_design: '',
  estimated_duration: 5,
  characters: [],
  scenes: [],
  elements: [],
  styles: [],
});

const takeDraft = reactive({
  prompt: '',
  modelId: '',
  duration: 5,
  serviceTier: 'standard',
  executionExpiresAfter: 86400,
  chainFromPrev: false,
  generateAudio: false,
});

const providerOptions = [
  { label: '全局 Ark', value: 'ark_default' },
  { label: '自定义 Ark', value: 'ark_custom' },
  { label: 'OpenAI Compatible', value: 'openai_compatible' },
];

const durationOptions = [
  { label: '5 秒', value: 5 },
  { label: '10 秒', value: 10 },
];

const serviceTierOptions = [
  { label: '在线推理 (standard)', value: 'standard' },
  { label: '离线推理 (flex)', value: 'flex' },
];

const refKeys = ['characters', 'scenes', 'elements', 'styles'];
const refLabelMap = {
  characters: '人物',
  scenes: '场景',
  elements: '特殊元素',
  styles: '风格',
};

const assetDraftById = reactive({});
const frameDraftByKey = reactive({});

const projectId = computed(() => Number(route.params.id || 0));
const selectedShot = computed(() => workspace.selectedShot);
const audioSupported = computed(() => {
  const shotTake = workspace.selectedTake;
  const supported = workspace.audioSupportedModels || [];
  return !!shotTake && supported.includes(shotTake.model_id);
});

const projectTitle = computed(() => {
  if (!workspace.project) return '加载中...';
  return `${workspace.project.name} · #${workspace.project.id} · ${workspace.project.aspect_ratio || '16:9'}`;
});

const filteredCatalogs = computed(() => (workspace.assetCatalogs || []).filter((item) => item.asset_type === assetTab.value));
const modelOptions = computed(() => (workspace.models || []).map((m) => ({ label: m.name, value: m.id })));

function cloneRefRows(rows) {
  if (!rows || rows.length === 0) return [{ id: '', name: '', prompt: '' }];
  return rows.map((row) => ({
    id: row.id || '',
    name: row.name || '',
    prompt: row.prompt || '',
  }));
}

function parseMultilineList(value) {
  return String(value || '')
    .split(/[\n,]/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function getFrameVersions(shot, frameType) {
  if (frameType === 'start') return shot.start_frames || [];
  return shot.end_frames || [];
}

function initShotDraft(shot) {
  if (!shot) return;
  shotDraft.shot_no = shot.shot_no || '';
  shotDraft.shot_size = shot.shot_size || '';
  shotDraft.camera_movement = shot.camera_movement || '';
  shotDraft.frame_content = shot.frame_content || '';
  shotDraft.sound_design = shot.sound_design || '';
  shotDraft.estimated_duration = Number(shot.estimated_duration || 5);
  shotDraft.characters = cloneRefRows(shot.characters);
  shotDraft.scenes = cloneRefRows(shot.scenes);
  shotDraft.elements = cloneRefRows(shot.elements);
  shotDraft.styles = cloneRefRows(shot.styles);
}

function initTakeDraft(take) {
  if (!take) return;
  takeDraft.prompt = take.prompt || '';
  takeDraft.modelId = take.model_id || '';
  takeDraft.duration = Number(take.duration || 5);
  takeDraft.serviceTier = take.service_tier || 'standard';
  takeDraft.executionExpiresAfter = Number(take.expires_after || 86400) || 86400;
  takeDraft.chainFromPrev = !!take.chain_from_prev;
  takeDraft.generateAudio = !!take.generate_audio;
}

function ensureAssetDraft(catalog) {
  if (!assetDraftById[catalog.id]) {
    assetDraftById[catalog.id] = {
      name: catalog.name || '',
      prompt: catalog.prompt || '',
      inputImages: '',
    };
  }
}

function getAssetDraft(catalog) {
  ensureAssetDraft(catalog);
  return assetDraftById[catalog.id];
}

function ensureFrameDraft(shotId, frameType, fallbackPrompt = '') {
  const key = `${shotId}-${frameType}`;
  if (!frameDraftByKey[key]) {
    frameDraftByKey[key] = {
      prompt: fallbackPrompt || '',
      inputImages: '',
    };
  }
}

function getFrameDraft(shotId, frameType, fallbackPrompt = '') {
  ensureFrameDraft(shotId, frameType, fallbackPrompt);
  return frameDraftByKey[`${shotId}-${frameType}`];
}

function addRefRow(key) {
  shotDraft[key].push({ id: '', name: '', prompt: '' });
}

function removeRefRow(key, index) {
  shotDraft[key].splice(index, 1);
  if (shotDraft[key].length === 0) {
    shotDraft[key].push({ id: '', name: '', prompt: '' });
  }
}

async function handleCreateShot() {
  try {
    await workspace.createShot(0);
    message.success('已新建分镜');
  } catch (err) {
    message.error(String(err?.message || err || '新建分镜失败'));
  }
}

async function handleSaveShot() {
  if (!selectedShot.value) return;
  try {
    await workspace.saveShotMetadata({
      storyboard_id: selectedShot.value.id,
      shot_no: shotDraft.shot_no,
      shot_size: shotDraft.shot_size,
      camera_movement: shotDraft.camera_movement,
      frame_content: shotDraft.frame_content,
      sound_design: shotDraft.sound_design,
      estimated_duration: Number(shotDraft.estimated_duration || 5),
      duration_fine: 0,
      characters: shotDraft.characters.filter((r) => r.id || r.name || r.prompt),
      scenes: shotDraft.scenes.filter((r) => r.id || r.name || r.prompt),
      elements: shotDraft.elements.filter((r) => r.id || r.name || r.prompt),
      styles: shotDraft.styles.filter((r) => r.id || r.name || r.prompt),
    });
    message.success('分镜已保存');
  } catch (err) {
    message.error(String(err?.message || err || '保存失败'));
  }
}

async function handleDeleteShot() {
  if (!selectedShot.value) return;
  if (!window.confirm('确认删除这个分镜？')) return;
  try {
    await workspace.deleteShot(selectedShot.value.id);
    message.success('已删除分镜');
  } catch (err) {
    message.error(String(err?.message || err || '删除失败'));
  }
}

async function handleMergeShot() {
  if (!selectedShot.value) return;
  try {
    await workspace.mergeShot(selectedShot.value.id);
    message.success('已并入下一镜');
  } catch (err) {
    message.error(String(err?.message || err || '合并失败'));
  }
}

async function handleSplitShot() {
  if (!selectedShot.value) return;
  const second = window.prompt('请输入拆分后第二镜的画面内容');
  if (second === null) return;
  try {
    await workspace.splitShot(selectedShot.value.id, second);
    message.success('拆分成功');
  } catch (err) {
    message.error(String(err?.message || err || '拆分失败'));
  }
}

async function handleLoadSourceFile() {
  try {
    await workspace.selectStoryboardSourceFile();
  } catch (err) {
    message.error(String(err?.message || err || '导入失败'));
  }
}

async function handleDecompose() {
  try {
    await workspace.decomposeStoryboard();
    message.success('拆解完成');
  } catch (err) {
    message.error(String(err?.message || err || '拆解失败'));
  }
}

async function handleSaveAsset(catalogId) {
  const draft = assetDraftById[catalogId];
  try {
    await workspace.updateAssetCatalog(catalogId, draft?.name || '', draft?.prompt || '');
    message.success('资产已保存');
  } catch (err) {
    message.error(String(err?.message || err || '保存资产失败'));
  }
}

async function handleUploadAsset(catalogId) {
  try {
    await workspace.uploadAssetImage(catalogId);
    message.success('上传完成');
  } catch (err) {
    message.error(String(err?.message || err || '上传失败'));
  }
}

async function handleGenerateAsset(catalogId) {
  const draft = assetDraftById[catalogId];
  try {
    await workspace.generateAssetImage(catalogId, draft?.prompt || '', parseMultilineList(draft?.inputImages || ''));
    message.success('资产生成已提交');
  } catch (err) {
    message.error(String(err?.message || err || '资产生成失败'));
  }
}

async function handleToggleAssetGood(versionId) {
  try {
    await workspace.toggleAssetVersionGood(versionId);
    message.success('已更新 Good 状态');
  } catch (err) {
    message.error(String(err?.message || err || '设置 Good 失败'));
  }
}

async function handleUploadFrame(shotId, frameType) {
  try {
    await workspace.uploadShotFrame(shotId, frameType);
    message.success('帧上传完成');
  } catch (err) {
    message.error(String(err?.message || err || '帧上传失败'));
  }
}

async function handleGenerateFrame(shotId, frameType) {
  const key = `${shotId}-${frameType}`;
  const draft = frameDraftByKey[key] || { prompt: '', inputImages: '' };
  try {
    await workspace.generateShotFrame(shotId, frameType, draft.prompt, parseMultilineList(draft.inputImages));
    message.success('帧生成已提交');
  } catch (err) {
    message.error(String(err?.message || err || '帧生成失败'));
  }
}

async function handleToggleFrameGood(versionId) {
  try {
    await workspace.toggleShotFrameGood(versionId);
    message.success('帧 Good 状态已更新');
  } catch (err) {
    message.error(String(err?.message || err || '设置帧 Good 失败'));
  }
}

async function handleSaveTake() {
  if (!selectedShot.value) return;
  try {
    await workspace.saveTakeAsNew({
      storyboardId: selectedShot.value.id,
      prompt: takeDraft.prompt,
      modelId: takeDraft.modelId,
      duration: takeDraft.duration,
      generateAudio: takeDraft.generateAudio,
      serviceTier: takeDraft.serviceTier,
      executionExpiresAfter: takeDraft.serviceTier === 'flex' ? Math.max(60, Number(takeDraft.executionExpiresAfter || 86400)) : 0,
      chainFromPrev: takeDraft.chainFromPrev,
      firstFramePath: takeDraft.chainFromPrev ? '' : (selectedShot.value.active_start_frame?.image_path || ''),
      lastFramePath: selectedShot.value.active_end_frame?.image_path || '',
    });
    message.success('已保存为新 Take');
  } catch (err) {
    message.error(String(err?.message || err || '保存 Take 失败'));
  }
}

async function handleGenerateTake(takeId) {
  try {
    await workspace.generateTake(takeId);
    message.success('生成任务已提交');
  } catch (err) {
    message.error(String(err?.message || err || '生成失败'));
  }
}

async function handleToggleGoodTake(takeId) {
  try {
    await workspace.toggleGoodTake(takeId);
    message.success('Good Take 状态已更新');
  } catch (err) {
    message.error(String(err?.message || err || '操作失败'));
  }
}

async function handleExport() {
  try {
    await workspace.exportProject();
    message.success('导出任务已触发');
  } catch (err) {
    message.error(String(err?.message || err || '导出失败'));
  }
}

watch(selectedShot, (shot) => {
  if (!shot) return;
  initShotDraft(shot);
}, { immediate: true });

watch(() => workspace.selectedTake, (take) => {
  if (!take) return;
  initTakeDraft(take);
}, { immediate: true });

watch(
  filteredCatalogs,
  (rows) => {
    rows.forEach((catalog) => ensureAssetDraft(catalog));
  },
  { immediate: true },
);

watch(
  () => workspace.storyboards,
  (shots) => {
    (shots || []).forEach((shot) => {
      ensureFrameDraft(shot.id, 'start', shot.frame_content || '');
      ensureFrameDraft(shot.id, 'end', shot.frame_content || '');
    });
  },
  { immediate: true },
);

watch(
  () => route.params.id,
  async (id) => {
    if (!id) return;
    const detail = await workspace.fetchProjectMeta(Number(id));
    isV2.value = (detail?.project?.model_version || 'v1.x') === 'v2.0';
    workspace.stopPolling();
    if (isV2.value) return;
    await workspace.fetchWorkspace(Number(id), { preserveSelection: false });
    workspace.startPolling();
  },
  { immediate: true },
);

onMounted(() => {
  if (!isV2.value) workspace.startPolling();
});

onUnmounted(() => {
  workspace.stopPolling();
});
</script>
