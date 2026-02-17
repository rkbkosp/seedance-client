import { defineStore } from 'pinia';

export const useWorkspaceStore = defineStore('workspace', {
  state: () => ({
    projectId: null,
    project: null,
    models: [],
    audioSupportedModels: [],
    storyboards: [],
    assetCatalogs: [],
    selectedShotId: null,
    selectedTakeByShot: {},
    loading: false,
    error: '',
    pollTimer: null,
    projectVersion: 'v1.x',
    decomposeText: '',
    imageModelDefault: '',
    apiConfig: {
      llmModel: '',
      provider: 'ark_default',
      baseUrl: 'https://ark.cn-beijing.volces.com/api/v3',
      apiKey: '',
      replaceExisting: true,
    },
  }),

  getters: {
    selectedShot(state) {
      return state.storyboards.find((shot) => shot.id === state.selectedShotId) || null;
    },
    selectedTake(state) {
      const shot = state.storyboards.find((item) => item.id === state.selectedShotId);
      if (!shot) return null;
      const selectedTakeId = state.selectedTakeByShot[shot.id];
      if (selectedTakeId) {
        const specific = (shot.takes || []).find((take) => take.id === selectedTakeId);
        if (specific) return specific;
      }
      return shot.active_take || shot.takes?.[shot.takes.length - 1] || null;
    },
    runningTakeIds(state) {
      const ids = [];
      (state.storyboards || []).forEach((shot) => {
        (shot.takes || []).forEach((take) => {
          const status = (take.status || '').toLowerCase();
          if (status === 'running' || status === 'queued') ids.push(take.id);
        });
      });
      return ids;
    },
  },

  actions: {
    selectShot(shotId) {
      this.selectedShotId = shotId;
    },

    selectTake(shotId, takeId) {
      this.selectedTakeByShot = {
        ...this.selectedTakeByShot,
        [shotId]: takeId,
      };
    },

    setApiConfig(partial) {
      this.apiConfig = {
        ...this.apiConfig,
        ...partial,
      };
    },

    async fetchWorkspace(projectId, options = {}) {
      const preserveSelection = options.preserveSelection ?? true;
      this.loading = true;
      this.error = '';
      if (projectId) this.projectId = Number(projectId);

      try {
        const ws = await window.go.main.App.GetV1Workspace(this.projectId);
        this.project = ws.project || null;
        this.projectVersion = ws.project?.model_version || 'v1.x';
        this.models = ws.models || [];
        this.audioSupportedModels = ws.audio_supported_models || [];
        this.storyboards = ws.storyboards || [];
        this.assetCatalogs = ws.asset_catalogs || [];
        this.imageModelDefault = ws.image_model_default || '';
        if (!this.apiConfig.llmModel) {
          this.apiConfig.llmModel = ws.llm_model_default || '';
        }

        const shotIds = this.storyboards.map((shot) => shot.id);
        if (!preserveSelection || !shotIds.includes(this.selectedShotId)) {
          this.selectedShotId = shotIds[0] || null;
        }

        const nextMap = {};
        this.storyboards.forEach((shot) => {
          const prev = this.selectedTakeByShot[shot.id];
          const exists = (shot.takes || []).some((take) => take.id === prev);
          const fallback = shot.active_take?.id || shot.takes?.[shot.takes.length - 1]?.id || null;
          if (exists) nextMap[shot.id] = prev;
          else if (fallback) nextMap[shot.id] = fallback;
        });
        this.selectedTakeByShot = nextMap;
      } catch (err) {
        this.error = String(err?.message || err || '加载工作台失败');
      } finally {
        this.loading = false;
      }
    },

    async fetchProjectMeta(projectId) {
      const data = await window.go.main.App.GetProject(Number(projectId));
      this.projectVersion = data?.project?.model_version || 'v1.x';
      return data;
    },

    async ensureHasGlobalAPIKey() {
      try {
        const ok = await window.go.main.App.HasAPIKey();
        if (ok) return true;
      } catch (err) {
        return true;
      }
      this.error = '[E_APIKEY_MISSING] 未配置 API Key，请先在设置中填写。';
      return false;
    },

    async refreshWorkspace(preserveSelection = true) {
      await this.fetchWorkspace(this.projectId, { preserveSelection });
    },

    async selectStoryboardSourceFile() {
      const result = await window.go.main.App.SelectStoryboardSourceFile();
      if (result?.content) this.decomposeText = result.content;
      return result;
    },

    async decomposeStoryboard() {
      const sourceText = (this.decomposeText || '').trim();
      if (!sourceText) throw new Error('请先输入分镜文案或导入文件');

      if (this.apiConfig.provider === 'ark_default') {
        const ok = await this.ensureHasGlobalAPIKey();
        if (!ok) throw new Error(this.error || '未配置 API Key');
      } else if (!this.apiConfig.apiKey.trim()) {
        throw new Error('请先填写独立 API Key');
      }

      await window.go.main.App.DecomposeStoryboardWithLLM({
        project_id: this.projectId,
        source_text: sourceText,
        llm_model_id: this.apiConfig.llmModel,
        provider: this.apiConfig.provider,
        api_key: this.apiConfig.apiKey || '',
        base_url: this.apiConfig.baseUrl || '',
        replace_existing: !!this.apiConfig.replaceExisting,
      });
      await this.refreshWorkspace(false);
    },

    async createShot(afterStoryboardId = 0) {
      const newShotId = await window.go.main.App.CreateV1Shot({
        project_id: this.projectId,
        after_storyboard_id: afterStoryboardId,
      });
      await this.refreshWorkspace(true);
      this.selectedShotId = Number(newShotId);
      return newShotId;
    },

    async saveShotMetadata(payload) {
      await window.go.main.App.UpdateShotMetadata(payload);
      await this.refreshWorkspace(true);
    },

    async deleteShot(storyboardId) {
      await window.go.main.App.DeleteV1Shot(Number(storyboardId));
      await this.refreshWorkspace(true);
    },

    async mergeShot(storyboardId) {
      await window.go.main.App.MergeShotWithNext(Number(storyboardId));
      await this.refreshWorkspace(true);
    },

    async splitShot(storyboardId, secondContent) {
      await window.go.main.App.SplitShot({
        storyboard_id: Number(storyboardId),
        first_content: '',
        second_content: secondContent,
      });
      await this.refreshWorkspace(true);
    },

    async updateAssetCatalog(catalogId, name, prompt) {
      await window.go.main.App.UpdateAssetCatalog({
        catalog_id: Number(catalogId),
        name: name || '',
        prompt: prompt || '',
      });
      await this.refreshWorkspace(true);
    },

    async uploadAssetImage(catalogId) {
      await window.go.main.App.UploadAssetImage(Number(catalogId));
      await this.refreshWorkspace(true);
    },

    async generateAssetImage(catalogId, prompt, inputImages = []) {
      const ok = await this.ensureHasGlobalAPIKey();
      if (!ok) throw new Error(this.error || '未配置 API Key');
      await window.go.main.App.GenerateAssetImage({
        catalog_id: Number(catalogId),
        model_id: this.imageModelDefault,
        prompt: prompt || '',
        input_images: inputImages,
      });
      await this.refreshWorkspace(true);
    },

    async toggleAssetVersionGood(versionId) {
      await window.go.main.App.ToggleAssetVersionGood(Number(versionId));
      await this.refreshWorkspace(true);
    },

    async uploadShotFrame(storyboardId, frameType) {
      await window.go.main.App.UploadShotFrame({
        storyboard_id: Number(storyboardId),
        frame_type: frameType,
      });
      await this.refreshWorkspace(true);
    },

    async generateShotFrame(storyboardId, frameType, prompt, inputImages = []) {
      const ok = await this.ensureHasGlobalAPIKey();
      if (!ok) throw new Error(this.error || '未配置 API Key');
      await window.go.main.App.GenerateShotFrame({
        storyboard_id: Number(storyboardId),
        frame_type: frameType,
        model_id: this.imageModelDefault,
        prompt: prompt || '',
        input_images: inputImages,
      });
      await this.refreshWorkspace(true);
    },

    async toggleShotFrameGood(versionId) {
      await window.go.main.App.ToggleShotFrameGood(Number(versionId));
      await this.refreshWorkspace(true);
    },

    async saveTakeAsNew({
      storyboardId,
      prompt,
      modelId,
      duration,
      generateAudio,
      serviceTier,
      executionExpiresAfter,
      chainFromPrev,
      firstFramePath,
      lastFramePath,
    }) {
      await window.go.main.App.UpdateStoryboard({
        storyboard_id: Number(storyboardId),
        prompt: prompt || '',
        model_id: modelId || '',
        ratio: this.project?.aspect_ratio || '16:9',
        duration: Number(duration || 5),
        generate_audio: !!generateAudio,
        service_tier: serviceTier || 'standard',
        execution_expires_after: Number(executionExpiresAfter || 0),
        first_frame_path: firstFramePath || '',
        last_frame_path: lastFramePath || '',
        delete_first_frame: false,
        delete_last_frame: false,
        chain_from_prev: !!chainFromPrev,
      });

      await this.refreshWorkspace(true);
      const target = this.storyboards.find((item) => item.id === Number(storyboardId));
      const latestTake = target?.takes?.[target.takes.length - 1];
      if (latestTake) this.selectTake(Number(storyboardId), latestTake.id);
    },

    async generateTake(takeId) {
      const ok = await this.ensureHasGlobalAPIKey();
      if (!ok) throw new Error(this.error || '未配置 API Key');
      await window.go.main.App.GenerateTakeVideo(Number(takeId));
      await this.refreshWorkspace(true);
    },

    async toggleGoodTake(takeId) {
      await window.go.main.App.ToggleGoodTake(Number(takeId));
      await this.refreshWorkspace(true);
    },

    async exportProject() {
      await window.go.main.App.ExportProject(Number(this.projectId));
    },

    updateTakeStatus(takeId, status) {
      this.storyboards.forEach((shot) => {
        const target = (shot.takes || []).find((take) => take.id === takeId);
        if (target) target.status = status;
      });
    },

    startPolling() {
      this.stopPolling();
      this.pollTimer = setInterval(async () => {
        if (!this.projectId) return;
        if (this.runningTakeIds.length === 0) return;

        await Promise.all(this.runningTakeIds.map(async (takeId) => {
          try {
            await window.go.main.App.GetTakeStatus(takeId);
          } catch (err) {
            console.warn('[poll] GetTakeStatus failed', takeId, err);
          }
        }));

        await this.fetchWorkspace(this.projectId, { preserveSelection: true });
      }, 3500);
    },

    stopPolling() {
      if (this.pollTimer) {
        clearInterval(this.pollTimer);
        this.pollTimer = null;
      }
    },
  },
});
