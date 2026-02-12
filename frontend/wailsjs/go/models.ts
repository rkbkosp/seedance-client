export namespace config {
	
	export class ModelPricing {
	    standard: number;
	    standard_audio?: number;
	    flex: number;
	    flex_audio?: number;
	    platform_price: number;
	
	    static createFrom(source: any = {}) {
	        return new ModelPricing(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.standard = source["standard"];
	        this.standard_audio = source["standard_audio"];
	        this.flex = source["flex"];
	        this.flex_audio = source["flex_audio"];
	        this.platform_price = source["platform_price"];
	    }
	}
	export class Model {
	    id: string;
	    name: string;
	    supports_audio: boolean;
	    default: boolean;
	    pricing: ModelPricing;
	
	    static createFrom(source: any = {}) {
	        return new Model(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.supports_audio = source["supports_audio"];
	        this.default = source["default"];
	        this.pricing = this.convertValues(source["pricing"], ModelPricing);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class AssetVersionResponse {
	    id: number;
	    catalog_id: number;
	    version_no: number;
	    image_path: string;
	    source_type: string;
	    model_id: string;
	    prompt: string;
	    status: string;
	    is_good: boolean;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new AssetVersionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.catalog_id = source["catalog_id"];
	        this.version_no = source["version_no"];
	        this.image_path = source["image_path"];
	        this.source_type = source["source_type"];
	        this.model_id = source["model_id"];
	        this.prompt = source["prompt"];
	        this.status = source["status"];
	        this.is_good = source["is_good"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AssetCatalogResponse {
	    id: number;
	    project_id: number;
	    asset_type: string;
	    asset_code: string;
	    name: string;
	    prompt: string;
	    storyboard_id?: number;
	    versions: AssetVersionResponse[];
	    active?: AssetVersionResponse;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new AssetCatalogResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_id = source["project_id"];
	        this.asset_type = source["asset_type"];
	        this.asset_code = source["asset_code"];
	        this.name = source["name"];
	        this.prompt = source["prompt"];
	        this.storyboard_id = source["storyboard_id"];
	        this.versions = this.convertValues(source["versions"], AssetVersionResponse);
	        this.active = this.convertValues(source["active"], AssetVersionResponse);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class CreateProjectParams {
	    name: string;
	    model_version: string;
	    aspect_ratio: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateProjectParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.model_version = source["model_version"];
	        this.aspect_ratio = source["aspect_ratio"];
	    }
	}
	export class CreateStoryboardParams {
	    project_id: number;
	    prompt: string;
	    model_id: string;
	    ratio: string;
	    duration: number;
	    generate_audio: boolean;
	    service_tier: string;
	    execution_expires_after: number;
	    first_frame_path: string;
	    last_frame_path: string;
	    chain_from_prev: boolean;
	    generation_mode: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateStoryboardParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_id = source["project_id"];
	        this.prompt = source["prompt"];
	        this.model_id = source["model_id"];
	        this.ratio = source["ratio"];
	        this.duration = source["duration"];
	        this.generate_audio = source["generate_audio"];
	        this.service_tier = source["service_tier"];
	        this.execution_expires_after = source["execution_expires_after"];
	        this.first_frame_path = source["first_frame_path"];
	        this.last_frame_path = source["last_frame_path"];
	        this.chain_from_prev = source["chain_from_prev"];
	        this.generation_mode = source["generation_mode"];
	    }
	}
	export class CreateV1ShotParams {
	    project_id: number;
	    after_storyboard_id: number;
	
	    static createFrom(source: any = {}) {
	        return new CreateV1ShotParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_id = source["project_id"];
	        this.after_storyboard_id = source["after_storyboard_id"];
	    }
	}
	export class DecomposeStoryboardParams {
	    project_id: number;
	    source_text: string;
	    llm_model_id: string;
	    provider: string;
	    api_key: string;
	    base_url: string;
	    replace_existing: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DecomposeStoryboardParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_id = source["project_id"];
	        this.source_text = source["source_text"];
	        this.llm_model_id = source["llm_model_id"];
	        this.provider = source["provider"];
	        this.api_key = source["api_key"];
	        this.base_url = source["base_url"];
	        this.replace_existing = source["replace_existing"];
	    }
	}
	export class DeleteTakeResult {
	    success: boolean;
	    storyboard_deleted: boolean;
	    remaining_takes: number;
	
	    static createFrom(source: any = {}) {
	        return new DeleteTakeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.storyboard_deleted = source["storyboard_deleted"];
	        this.remaining_takes = source["remaining_takes"];
	    }
	}
	export class EntityRef {
	    id: string;
	    name: string;
	    prompt: string;
	
	    static createFrom(source: any = {}) {
	        return new EntityRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.prompt = source["prompt"];
	    }
	}
	export class GenerateAssetImageParams {
	    catalog_id: number;
	    model_id: string;
	    prompt: string;
	    input_images: string[];
	
	    static createFrom(source: any = {}) {
	        return new GenerateAssetImageParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.catalog_id = source["catalog_id"];
	        this.model_id = source["model_id"];
	        this.prompt = source["prompt"];
	        this.input_images = source["input_images"];
	    }
	}
	export class GenerateShotFrameParams {
	    storyboard_id: number;
	    frame_type: string;
	    model_id: string;
	    prompt: string;
	    input_images: string[];
	
	    static createFrom(source: any = {}) {
	        return new GenerateShotFrameParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.storyboard_id = source["storyboard_id"];
	        this.frame_type = source["frame_type"];
	        this.model_id = source["model_id"];
	        this.prompt = source["prompt"];
	        this.input_images = source["input_images"];
	    }
	}
	export class TakeResponse {
	    id: number;
	    storyboard_id: number;
	    prompt: string;
	    first_frame_path: string;
	    last_frame_path: string;
	    model_id: string;
	    ratio: string;
	    duration: number;
	    generate_audio: boolean;
	    task_id: string;
	    status: string;
	    video_url: string;
	    last_frame_url: string;
	    local_video_path: string;
	    local_last_frame_path: string;
	    download_status: string;
	    service_tier: string;
	    token_usage: number;
	    expires_after: number;
	    is_good: boolean;
	    chain_from_prev: boolean;
	    generation_mode: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new TakeResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.storyboard_id = source["storyboard_id"];
	        this.prompt = source["prompt"];
	        this.first_frame_path = source["first_frame_path"];
	        this.last_frame_path = source["last_frame_path"];
	        this.model_id = source["model_id"];
	        this.ratio = source["ratio"];
	        this.duration = source["duration"];
	        this.generate_audio = source["generate_audio"];
	        this.task_id = source["task_id"];
	        this.status = source["status"];
	        this.video_url = source["video_url"];
	        this.last_frame_url = source["last_frame_url"];
	        this.local_video_path = source["local_video_path"];
	        this.local_last_frame_path = source["local_last_frame_path"];
	        this.download_status = source["download_status"];
	        this.service_tier = source["service_tier"];
	        this.token_usage = source["token_usage"];
	        this.expires_after = source["expires_after"];
	        this.is_good = source["is_good"];
	        this.chain_from_prev = source["chain_from_prev"];
	        this.generation_mode = source["generation_mode"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StoryboardData {
	    id: number;
	    project_id: number;
	    takes: TakeResponse[];
	    // Go type: time
	    created_at: any;
	    active_take?: TakeResponse;
	
	    static createFrom(source: any = {}) {
	        return new StoryboardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_id = source["project_id"];
	        this.takes = this.convertValues(source["takes"], TakeResponse);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.active_take = this.convertValues(source["active_take"], TakeResponse);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProjectDetail {
	    id: number;
	    name: string;
	    model_version: string;
	    aspect_ratio: string;
	    // Go type: time
	    created_at: any;
	    storyboards: StoryboardData[];
	
	    static createFrom(source: any = {}) {
	        return new ProjectDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.model_version = source["model_version"];
	        this.aspect_ratio = source["aspect_ratio"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.storyboards = this.convertValues(source["storyboards"], StoryboardData);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProjectDetailData {
	    project: ProjectDetail;
	    models: config.Model[];
	    audio_supported_models: string[];
	
	    static createFrom(source: any = {}) {
	        return new ProjectDetailData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project = this.convertValues(source["project"], ProjectDetail);
	        this.models = this.convertValues(source["models"], config.Model);
	        this.audio_supported_models = source["audio_supported_models"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProjectStats {
	    total_videos: number;
	    total_token_usage: number;
	    total_cost: number;
	    total_savings: number;
	    model_video_count: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new ProjectStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_videos = source["total_videos"];
	        this.total_token_usage = source["total_token_usage"];
	        this.total_cost = source["total_cost"];
	        this.total_savings = source["total_savings"];
	        this.model_video_count = source["model_video_count"];
	    }
	}
	export class ProjectsData {
	    projects: models.Project[];
	    stats: ProjectStats;
	
	    static createFrom(source: any = {}) {
	        return new ProjectsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projects = this.convertValues(source["projects"], models.Project);
	        this.stats = this.convertValues(source["stats"], ProjectStats);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ShotFrameVersionResponse {
	    id: number;
	    storyboard_id: number;
	    frame_type: string;
	    version_no: number;
	    image_path: string;
	    source_type: string;
	    model_id: string;
	    prompt: string;
	    status: string;
	    is_good: boolean;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new ShotFrameVersionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.storyboard_id = source["storyboard_id"];
	        this.frame_type = source["frame_type"];
	        this.version_no = source["version_no"];
	        this.image_path = source["image_path"];
	        this.source_type = source["source_type"];
	        this.model_id = source["model_id"];
	        this.prompt = source["prompt"];
	        this.status = source["status"];
	        this.is_good = source["is_good"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SplitShotParams {
	    storyboard_id: number;
	    first_content: string;
	    second_content: string;
	
	    static createFrom(source: any = {}) {
	        return new SplitShotParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.storyboard_id = source["storyboard_id"];
	        this.first_content = source["first_content"];
	        this.second_content = source["second_content"];
	    }
	}
	
	export class StoryboardSourceFile {
	    filename: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new StoryboardSourceFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.content = source["content"];
	    }
	}
	
	export class TakeStatusResult {
	    status: string;
	    video_url: string;
	    last_frame_url: string;
	    poll_interval: number;
	    download_status: string;
	
	    static createFrom(source: any = {}) {
	        return new TakeStatusResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.video_url = source["video_url"];
	        this.last_frame_url = source["last_frame_url"];
	        this.poll_interval = source["poll_interval"];
	        this.download_status = source["download_status"];
	    }
	}
	export class UpdateAssetCatalogParams {
	    catalog_id: number;
	    name: string;
	    prompt: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateAssetCatalogParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.catalog_id = source["catalog_id"];
	        this.name = source["name"];
	        this.prompt = source["prompt"];
	    }
	}
	export class UpdateShotParams {
	    storyboard_id: number;
	    shot_no: string;
	    shot_size: string;
	    camera_movement: string;
	    frame_content: string;
	    characters: EntityRef[];
	    scenes: EntityRef[];
	    elements: EntityRef[];
	    styles: EntityRef[];
	    sound_design: string;
	    estimated_duration: number;
	    duration_fine: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdateShotParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.storyboard_id = source["storyboard_id"];
	        this.shot_no = source["shot_no"];
	        this.shot_size = source["shot_size"];
	        this.camera_movement = source["camera_movement"];
	        this.frame_content = source["frame_content"];
	        this.characters = this.convertValues(source["characters"], EntityRef);
	        this.scenes = this.convertValues(source["scenes"], EntityRef);
	        this.elements = this.convertValues(source["elements"], EntityRef);
	        this.styles = this.convertValues(source["styles"], EntityRef);
	        this.sound_design = source["sound_design"];
	        this.estimated_duration = source["estimated_duration"];
	        this.duration_fine = source["duration_fine"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UpdateStoryboardParams {
	    storyboard_id: number;
	    prompt: string;
	    model_id: string;
	    ratio: string;
	    duration: number;
	    generate_audio: boolean;
	    service_tier: string;
	    execution_expires_after: number;
	    first_frame_path: string;
	    last_frame_path: string;
	    delete_first_frame: boolean;
	    delete_last_frame: boolean;
	    chain_from_prev: boolean;
	    generation_mode: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateStoryboardParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.storyboard_id = source["storyboard_id"];
	        this.prompt = source["prompt"];
	        this.model_id = source["model_id"];
	        this.ratio = source["ratio"];
	        this.duration = source["duration"];
	        this.generate_audio = source["generate_audio"];
	        this.service_tier = source["service_tier"];
	        this.execution_expires_after = source["execution_expires_after"];
	        this.first_frame_path = source["first_frame_path"];
	        this.last_frame_path = source["last_frame_path"];
	        this.delete_first_frame = source["delete_first_frame"];
	        this.delete_last_frame = source["delete_last_frame"];
	        this.chain_from_prev = source["chain_from_prev"];
	        this.generation_mode = source["generation_mode"];
	    }
	}
	export class UploadShotFrameParams {
	    storyboard_id: number;
	    frame_type: string;
	
	    static createFrom(source: any = {}) {
	        return new UploadShotFrameParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.storyboard_id = source["storyboard_id"];
	        this.frame_type = source["frame_type"];
	    }
	}
	export class V1ProjectData {
	    id: number;
	    name: string;
	    model_version: string;
	    aspect_ratio: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new V1ProjectData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.model_version = source["model_version"];
	        this.aspect_ratio = source["aspect_ratio"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class V1ShotData {
	    id: number;
	    project_id: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    shot_order: number;
	    shot_no: string;
	    shot_size: string;
	    camera_movement: string;
	    frame_content: string;
	    characters: EntityRef[];
	    scenes: EntityRef[];
	    elements: EntityRef[];
	    styles: EntityRef[];
	    sound_design: string;
	    estimated_duration: number;
	    duration_fine: number;
	    takes: TakeResponse[];
	    active_take?: TakeResponse;
	    start_frames: ShotFrameVersionResponse[];
	    end_frames: ShotFrameVersionResponse[];
	    active_start_frame?: ShotFrameVersionResponse;
	    active_end_frame?: ShotFrameVersionResponse;
	
	    static createFrom(source: any = {}) {
	        return new V1ShotData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_id = source["project_id"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.shot_order = source["shot_order"];
	        this.shot_no = source["shot_no"];
	        this.shot_size = source["shot_size"];
	        this.camera_movement = source["camera_movement"];
	        this.frame_content = source["frame_content"];
	        this.characters = this.convertValues(source["characters"], EntityRef);
	        this.scenes = this.convertValues(source["scenes"], EntityRef);
	        this.elements = this.convertValues(source["elements"], EntityRef);
	        this.styles = this.convertValues(source["styles"], EntityRef);
	        this.sound_design = source["sound_design"];
	        this.estimated_duration = source["estimated_duration"];
	        this.duration_fine = source["duration_fine"];
	        this.takes = this.convertValues(source["takes"], TakeResponse);
	        this.active_take = this.convertValues(source["active_take"], TakeResponse);
	        this.start_frames = this.convertValues(source["start_frames"], ShotFrameVersionResponse);
	        this.end_frames = this.convertValues(source["end_frames"], ShotFrameVersionResponse);
	        this.active_start_frame = this.convertValues(source["active_start_frame"], ShotFrameVersionResponse);
	        this.active_end_frame = this.convertValues(source["active_end_frame"], ShotFrameVersionResponse);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class V1WorkspaceData {
	    project: V1ProjectData;
	    storyboards: V1ShotData[];
	    asset_catalogs: AssetCatalogResponse[];
	    models: config.Model[];
	    audio_supported_models: string[];
	    llm_model_default: string;
	    image_model_default: string;
	
	    static createFrom(source: any = {}) {
	        return new V1WorkspaceData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project = this.convertValues(source["project"], V1ProjectData);
	        this.storyboards = this.convertValues(source["storyboards"], V1ShotData);
	        this.asset_catalogs = this.convertValues(source["asset_catalogs"], AssetCatalogResponse);
	        this.models = this.convertValues(source["models"], config.Model);
	        this.audio_supported_models = source["audio_supported_models"];
	        this.llm_model_default = source["llm_model_default"];
	        this.image_model_default = source["image_model_default"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace models {
	
	export class Take {
	    id: number;
	    storyboard_id: number;
	    prompt: string;
	    first_frame_path: string;
	    last_frame_path: string;
	    model_id: string;
	    ratio: string;
	    duration: number;
	    generate_audio: boolean;
	    task_id: string;
	    status: string;
	    video_url: string;
	    last_frame_url: string;
	    local_video_path: string;
	    local_last_frame_path: string;
	    download_status: string;
	    service_tier: string;
	    token_usage: number;
	    expires_after: number;
	    is_good: boolean;
	    chain_from_prev: boolean;
	    generation_mode: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Take(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.storyboard_id = source["storyboard_id"];
	        this.prompt = source["prompt"];
	        this.first_frame_path = source["first_frame_path"];
	        this.last_frame_path = source["last_frame_path"];
	        this.model_id = source["model_id"];
	        this.ratio = source["ratio"];
	        this.duration = source["duration"];
	        this.generate_audio = source["generate_audio"];
	        this.task_id = source["task_id"];
	        this.status = source["status"];
	        this.video_url = source["video_url"];
	        this.last_frame_url = source["last_frame_url"];
	        this.local_video_path = source["local_video_path"];
	        this.local_last_frame_path = source["local_last_frame_path"];
	        this.download_status = source["download_status"];
	        this.service_tier = source["service_tier"];
	        this.token_usage = source["token_usage"];
	        this.expires_after = source["expires_after"];
	        this.is_good = source["is_good"];
	        this.chain_from_prev = source["chain_from_prev"];
	        this.generation_mode = source["generation_mode"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Storyboard {
	    id: number;
	    project_id: number;
	    takes: Take[];
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    shot_order: number;
	    shot_no: string;
	    shot_size: string;
	    camera_movement: string;
	    frame_content: string;
	    characters_json: string;
	    scenes_json: string;
	    elements_json: string;
	    styles_json: string;
	    sound_design: string;
	    estimated_duration: number;
	    duration_fine: number;
	    active_take?: Take;
	
	    static createFrom(source: any = {}) {
	        return new Storyboard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_id = source["project_id"];
	        this.takes = this.convertValues(source["takes"], Take);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.shot_order = source["shot_order"];
	        this.shot_no = source["shot_no"];
	        this.shot_size = source["shot_size"];
	        this.camera_movement = source["camera_movement"];
	        this.frame_content = source["frame_content"];
	        this.characters_json = source["characters_json"];
	        this.scenes_json = source["scenes_json"];
	        this.elements_json = source["elements_json"];
	        this.styles_json = source["styles_json"];
	        this.sound_design = source["sound_design"];
	        this.estimated_duration = source["estimated_duration"];
	        this.duration_fine = source["duration_fine"];
	        this.active_take = this.convertValues(source["active_take"], Take);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Project {
	    id: number;
	    name: string;
	    model_version: string;
	    aspect_ratio: string;
	    // Go type: time
	    created_at: any;
	    storyboards: Storyboard[];
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.model_version = source["model_version"];
	        this.aspect_ratio = source["aspect_ratio"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.storyboards = this.convertValues(source["storyboards"], Storyboard);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

