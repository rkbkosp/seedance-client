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
	
	export class CreateStoryboardParams {
	    project_id: number;
	    prompt: string;
	    model_id: string;
	    ratio: string;
	    duration: number;
	    generate_audio: boolean;
	    service_tier: string;
	    first_frame_path: string;
	    last_frame_path: string;
	
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
	        this.first_frame_path = source["first_frame_path"];
	        this.last_frame_path = source["last_frame_path"];
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
	export class UpdateStoryboardParams {
	    storyboard_id: number;
	    prompt: string;
	    model_id: string;
	    ratio: string;
	    duration: number;
	    generate_audio: boolean;
	    service_tier: string;
	    first_frame_path: string;
	    last_frame_path: string;
	    delete_first_frame: boolean;
	    delete_last_frame: boolean;
	
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
	        this.first_frame_path = source["first_frame_path"];
	        this.last_frame_path = source["last_frame_path"];
	        this.delete_first_frame = source["delete_first_frame"];
	        this.delete_last_frame = source["delete_last_frame"];
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

