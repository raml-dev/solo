export namespace collection {
	
	export class Request {
	    id: string;
	    name: string;
	    url: string;
	    verb: string;
	    body: string;
	    headers: Record<string, string>;
	    cookies: Record<string, string>;
	    settings?: configuration.RequestSettingsOverride;
	    // Go type: time
	    creationTimestamp: any;
	    // Go type: time
	    lastUpdateTimestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new Request(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.verb = source["verb"];
	        this.body = source["body"];
	        this.headers = source["headers"];
	        this.cookies = source["cookies"];
	        this.settings = this.convertValues(source["settings"], configuration.RequestSettingsOverride);
	        this.creationTimestamp = this.convertValues(source["creationTimestamp"], null);
	        this.lastUpdateTimestamp = this.convertValues(source["lastUpdateTimestamp"], null);
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
	export class Collection {
	    // Go type: time
	    creationTimestamp: any;
	    // Go type: time
	    lastUpdateTimestamp: any;
	    requests: Request[];
	    name: string;
	    id: string;
	
	    static createFrom(source: any = {}) {
	        return new Collection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.creationTimestamp = this.convertValues(source["creationTimestamp"], null);
	        this.lastUpdateTimestamp = this.convertValues(source["lastUpdateTimestamp"], null);
	        this.requests = this.convertValues(source["requests"], Request);
	        this.name = source["name"];
	        this.id = source["id"];
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

export namespace configuration {
	
	export class RequestSettings {
	    timeoutSeconds: number;
	    followRedirects: boolean;
	    maxRedirects: number;
	    validateSSL: boolean;
	    defaultUserAgent: string;
	    proxyUrl?: string;
	
	    static createFrom(source: any = {}) {
	        return new RequestSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timeoutSeconds = source["timeoutSeconds"];
	        this.followRedirects = source["followRedirects"];
	        this.maxRedirects = source["maxRedirects"];
	        this.validateSSL = source["validateSSL"];
	        this.defaultUserAgent = source["defaultUserAgent"];
	        this.proxyUrl = source["proxyUrl"];
	    }
	}
	export class GeneralSettings {
	    theme: string;
	    checkForUpdates: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GeneralSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.checkForUpdates = source["checkForUpdates"];
	    }
	}
	export class Configuration {
	    general: GeneralSettings;
	    request: RequestSettings;
	
	    static createFrom(source: any = {}) {
	        return new Configuration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.general = this.convertValues(source["general"], GeneralSettings);
	        this.request = this.convertValues(source["request"], RequestSettings);
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
	
	
	export class RequestSettingsOverride {
	    timeoutSeconds?: number;
	    followRedirects?: boolean;
	    maxRedirects?: number;
	    validateSSL?: boolean;
	    defaultUserAgent?: string;
	    proxyUrl?: string;
	
	    static createFrom(source: any = {}) {
	        return new RequestSettingsOverride(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timeoutSeconds = source["timeoutSeconds"];
	        this.followRedirects = source["followRedirects"];
	        this.maxRedirects = source["maxRedirects"];
	        this.validateSSL = source["validateSSL"];
	        this.defaultUserAgent = source["defaultUserAgent"];
	        this.proxyUrl = source["proxyUrl"];
	    }
	}

}

export namespace environment {
	
	export class ValueType {
	    value: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new ValueType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.type = source["type"];
	    }
	}
	export class Environment {
	    id: string;
	    name: string;
	    values: Record<string, ValueType>;
	    // Go type: time
	    creation_timestamp: any;
	    // Go type: time
	    last_update_timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new Environment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.values = this.convertValues(source["values"], ValueType, true);
	        this.creation_timestamp = this.convertValues(source["creation_timestamp"], null);
	        this.last_update_timestamp = this.convertValues(source["last_update_timestamp"], null);
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
	
	export class RequestOptions {
	    method: string;
	    url: string;
	    headers: Record<string, any>;
	    body: string;
	    settings?: configuration.RequestSettingsOverride;
	
	    static createFrom(source: any = {}) {
	        return new RequestOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.url = source["url"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.settings = this.convertValues(source["settings"], configuration.RequestSettingsOverride);
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

export namespace requester {
	
	export class ResponseData {
	    statusCode: number;
	    headers: Record<string, string>;
	    body: string;
	    duration: number;
	
	    static createFrom(source: any = {}) {
	        return new ResponseData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.duration = source["duration"];
	    }
	}

}

export namespace theme {
	
	export class Theme {
	    name: string;
	    colors: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Theme(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.colors = source["colors"];
	    }
	}

}

