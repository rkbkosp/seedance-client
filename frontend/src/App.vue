<template>
  <n-config-provider :theme="darkTheme" :theme-overrides="themeOverrides">
    <n-message-provider>
      <div class="h-full">
        <RouterView />

        <button
          class="fixed right-4 bottom-4 z-50 border border-zinc-700 bg-zinc-900 text-zinc-100 rounded px-2 py-1 text-xs shadow-lg"
          @click="settingsVisible = true"
        >
          Settings
        </button>

        <n-drawer v-model:show="settingsVisible" :width="360" placement="right">
          <n-drawer-content title="设置">
            <div class="space-y-3">
              <div>
                <div class="text-xs text-zinc-500 mb-1">语言</div>
                <n-select v-model:value="language" :options="langOptions" @update:value="handleLanguageChange" />
              </div>

              <div>
                <div class="text-xs text-zinc-500 mb-1">API Key</div>
                <n-input v-model:value="apiKey" type="password" placeholder="输入新的 API Key" />
              </div>

              <n-button type="primary" @click="handleSaveApiKey">保存</n-button>
            </div>
          </n-drawer-content>
        </n-drawer>
      </div>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup>
import { ref } from 'vue';
import { RouterView } from 'vue-router';
import {
  NButton,
  NConfigProvider,
  createDiscreteApi,
  darkTheme,
  NDrawer,
  NDrawerContent,
  NInput,
  NMessageProvider,
  NSelect,
} from 'naive-ui';
import { getLanguage, setLanguage } from '@/i18n.js';

const settingsVisible = ref(false);
const apiKey = ref('');
const language = ref(getLanguage());
const { message } = createDiscreteApi(['message']);

const themeOverrides = {
  common: {
    primaryColor: '#a1a1aa',
    primaryColorHover: '#d4d4d8',
    primaryColorPressed: '#71717a',
    textColorBase: '#f4f4f5',
    bodyColor: '#09090b',
    cardColor: '#18181b',
    borderColor: '#27272a',
    inputColor: '#18181b',
  },
  Input: {
    color: '#18181b',
    colorFocus: '#18181b',
    textColor: '#f4f4f5',
    placeholderColor: '#71717a',
    border: '1px solid #3f3f46',
    borderHover: '1px solid #52525b',
    borderFocus: '1px solid #a1a1aa',
  },
  Select: {
    peers: {
      InternalSelection: {
        color: '#18181b',
        textColor: '#f4f4f5',
        placeholderColor: '#71717a',
        border: '1px solid #3f3f46',
        borderHover: '1px solid #52525b',
        borderFocus: '1px solid #a1a1aa',
      },
    },
  },
  Tabs: {
    tabTextColorActiveLine: '#f4f4f5',
    tabTextColorHoverLine: '#d4d4d8',
    tabTextColorLine: '#a1a1aa',
    barColor: '#a1a1aa',
  },
};

const langOptions = [
  { label: 'English', value: 'en' },
  { label: '简体中文', value: 'zh' },
];

function handleLanguageChange(next) {
  setLanguage(next);
  window.location.reload();
}

async function handleSaveApiKey() {
  const key = (apiKey.value || '').trim();
  if (!key) {
    message.warning('请先填写 API Key');
    return;
  }
  try {
    await window.go.main.App.UpdateAPIKey(key);
    apiKey.value = '';
    settingsVisible.value = false;
    message.success('API Key 已更新');
  } catch (err) {
    message.error(String(err?.message || err || '更新 API Key 失败'));
  }
}
</script>
