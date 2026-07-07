export namespace config {
	
	export class Config {
	    conflictStrategy: string;
	    templateStart: number;
	    templateIncrement: number;
	    templateDigits: number;
	    copyOutputDir: string;
	    cleanKeepTime: boolean;
	    windowWidth: number;
	    windowHeight: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.conflictStrategy = source["conflictStrategy"];
	        this.templateStart = source["templateStart"];
	        this.templateIncrement = source["templateIncrement"];
	        this.templateDigits = source["templateDigits"];
	        this.copyOutputDir = source["copyOutputDir"];
	        this.cleanKeepTime = source["cleanKeepTime"];
	        this.windowWidth = source["windowWidth"];
	        this.windowHeight = source["windowHeight"];
	    }
	}

}

export namespace main {
	
	export class FileView {
	    path: string;
	    oldName: string;
	    preview: string;
	    status: string;
	    err: string;
	
	    static createFrom(source: any = {}) {
	        return new FileView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.oldName = source["oldName"];
	        this.preview = source["preview"];
	        this.status = source["status"];
	        this.err = source["err"];
	    }
	}
	export class PlanOptions {
	    tab: string;
	    template?: renamer.TemplateOptions;
	    replace?: renamer.ReplaceOptions;
	    addRemove?: renamer.AddRemoveOptions;
	    clean: boolean;
	    version: boolean;
	    versionMove: boolean;
	    conflict: string;
	
	    static createFrom(source: any = {}) {
	        return new PlanOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tab = source["tab"];
	        this.template = this.convertValues(source["template"], renamer.TemplateOptions);
	        this.replace = this.convertValues(source["replace"], renamer.ReplaceOptions);
	        this.addRemove = this.convertValues(source["addRemove"], renamer.AddRemoveOptions);
	        this.clean = source["clean"];
	        this.version = source["version"];
	        this.versionMove = source["versionMove"];
	        this.conflict = source["conflict"];
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

export namespace renamer {
	
	export class AddRemoveOptions {
	    Prefix: string;
	    Suffix: string;
	    InsertAt: number;
	    InsertStr: string;
	    RemoveStr: string;
	    RemoveFrom: number;
	    RemoveCount: number;
	
	    static createFrom(source: any = {}) {
	        return new AddRemoveOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Prefix = source["Prefix"];
	        this.Suffix = source["Suffix"];
	        this.InsertAt = source["InsertAt"];
	        this.InsertStr = source["InsertStr"];
	        this.RemoveStr = source["RemoveStr"];
	        this.RemoveFrom = source["RemoveFrom"];
	        this.RemoveCount = source["RemoveCount"];
	    }
	}
	export class ReplaceOptions {
	    Find: string;
	    Replace: string;
	    UseRegex: boolean;
	    IgnoreCase: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ReplaceOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Find = source["Find"];
	        this.Replace = source["Replace"];
	        this.UseRegex = source["UseRegex"];
	        this.IgnoreCase = source["IgnoreCase"];
	    }
	}
	export class TemplateOptions {
	    Pattern: string;
	    Start: number;
	    Increment: number;
	    Digits: number;
	    PadChar: string;
	    Random: boolean;
	    RandomLower: boolean;
	    ExtOverride: string;
	    AutoConflict: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TemplateOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Pattern = source["Pattern"];
	        this.Start = source["Start"];
	        this.Increment = source["Increment"];
	        this.Digits = source["Digits"];
	        this.PadChar = source["PadChar"];
	        this.Random = source["Random"];
	        this.RandomLower = source["RandomLower"];
	        this.ExtOverride = source["ExtOverride"];
	        this.AutoConflict = source["AutoConflict"];
	    }
	}

}

