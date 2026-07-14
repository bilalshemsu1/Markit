export namespace main {
	
	export class ConversionResult {
	    success: boolean;
	    markdown: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ConversionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.markdown = source["markdown"];
	        this.error = source["error"];
	    }
	}
	export class UpdateCheckResult {
	    hasUpdate: boolean;
	    version: string;
	    url: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasUpdate = source["hasUpdate"];
	        this.version = source["version"];
	        this.url = source["url"];
	        this.description = source["description"];
	    }
	}

}

