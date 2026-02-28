export namespace main {
	
	export class UIPrefs {
	    themeMode: string;
	    palettePreset: string;
	    navCollapsed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UIPrefs(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.themeMode = source["themeMode"];
	        this.palettePreset = source["palettePreset"];
	        this.navCollapsed = source["navCollapsed"];
	    }
	}
	export class Settings {
	    apiBaseUrl: string;
	    authMode: string;
	    apiToken?: string;
	    refreshToken?: string;
	    requestTimeoutSeconds: number;
	    uiPrefs: UIPrefs;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apiBaseUrl = source["apiBaseUrl"];
	        this.authMode = source["authMode"];
	        this.apiToken = source["apiToken"];
	        this.refreshToken = source["refreshToken"];
	        this.requestTimeoutSeconds = source["requestTimeoutSeconds"];
	        this.uiPrefs = this.convertValues(source["uiPrefs"], UIPrefs);
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
	export class UIPrefsPatch {
	    themeMode?: string;
	    palettePreset?: string;
	    navCollapsed?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UIPrefsPatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.themeMode = source["themeMode"];
	        this.palettePreset = source["palettePreset"];
	        this.navCollapsed = source["navCollapsed"];
	    }
	}
	export class SettingsPatch {
	    apiBaseUrl?: string;
	    authMode?: string;
	    apiToken?: string;
	    refreshToken?: string;
	    requestTimeoutSeconds?: number;
	    uiPrefs?: UIPrefsPatch;
	
	    static createFrom(source: any = {}) {
	        return new SettingsPatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apiBaseUrl = source["apiBaseUrl"];
	        this.authMode = source["authMode"];
	        this.apiToken = source["apiToken"];
	        this.refreshToken = source["refreshToken"];
	        this.requestTimeoutSeconds = source["requestTimeoutSeconds"];
	        this.uiPrefs = this.convertValues(source["uiPrefs"], UIPrefsPatch);
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

