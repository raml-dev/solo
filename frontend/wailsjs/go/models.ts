export namespace collection {
	
	export class Request {
	    name: string;
	    url: string;
	    verb: string;
	    body: string;
	    id: string;
	    headers: Record<string, any>;
	    cookies: Record<string, any>;
	    // Go type: time
	    creationTimestamp: any;
	    // Go type: time
	    lastUpdateTimestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new Request(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.verb = source["verb"];
	        this.body = source["body"];
	        this.id = source["id"];
	        this.headers = source["headers"];
	        this.cookies = source["cookies"];
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

export namespace main {
	
	export class RequestOptions {
	    method: string;
	    url: string;
	    headers: Record<string, any>;
	    body: string;
	
	    static createFrom(source: any = {}) {
	        return new RequestOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.url = source["url"];
	        this.headers = source["headers"];
	        this.body = source["body"];
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

