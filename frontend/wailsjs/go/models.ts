export namespace applog {
	
	export class LogEntry {
	    // Go type: time
	    timestamp: any;
	    level: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.level = source["level"];
	        this.message = source["message"];
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

export namespace config {
	
	export class Threshold {
	    min: number;
	    max: number;
	    color: string;
	    sound: boolean;
	    blink: boolean;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new Threshold(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.min = source["min"];
	        this.max = source["max"];
	        this.color = source["color"];
	        this.sound = source["sound"];
	        this.blink = source["blink"];
	        this.label = source["label"];
	    }
	}
	export class Config {
	    endpointUrl: string;
	    pollIntervalSeconds: number;
	    jsonPath: string;
	    tlsSkipVerify: boolean;
	    thresholds: Threshold[];
	    soundEnabled: boolean;
	    soundOnChangeOnly: boolean;
	    language: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.endpointUrl = source["endpointUrl"];
	        this.pollIntervalSeconds = source["pollIntervalSeconds"];
	        this.jsonPath = source["jsonPath"];
	        this.tlsSkipVerify = source["tlsSkipVerify"];
	        this.thresholds = this.convertValues(source["thresholds"], Threshold);
	        this.soundEnabled = source["soundEnabled"];
	        this.soundOnChangeOnly = source["soundOnChangeOnly"];
	        this.language = source["language"];
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
	
	export class Status {
	    ordersPending: number;
	    activeThreshold?: config.Threshold;
	    lastPollTime: string;
	    lastPollError?: string;
	    nextPollIn: number;
	    deviceConnected: boolean;
	    deviceName?: string;
	    polling: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ordersPending = source["ordersPending"];
	        this.activeThreshold = this.convertValues(source["activeThreshold"], config.Threshold);
	        this.lastPollTime = source["lastPollTime"];
	        this.lastPollError = source["lastPollError"];
	        this.nextPollIn = source["nextPollIn"];
	        this.deviceConnected = source["deviceConnected"];
	        this.deviceName = source["deviceName"];
	        this.polling = source["polling"];
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

export namespace statuslight {
	
	export class DebugInfo {
	    path: string;
	    openOk: boolean;
	    openError?: string;
	    vendorId?: string;
	    productId?: string;
	    product?: string;
	    outputReportByteLength: number;
	    featureReportByteLength: number;
	    inputReportByteLength: number;
	    usage: number;
	    usagePage: number;
	    writeFileResult: string;
	    setOutputReportResult: string;
	    setFeatureResult: string;
	
	    static createFrom(source: any = {}) {
	        return new DebugInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.openOk = source["openOk"];
	        this.openError = source["openError"];
	        this.vendorId = source["vendorId"];
	        this.productId = source["productId"];
	        this.product = source["product"];
	        this.outputReportByteLength = source["outputReportByteLength"];
	        this.featureReportByteLength = source["featureReportByteLength"];
	        this.inputReportByteLength = source["inputReportByteLength"];
	        this.usage = source["usage"];
	        this.usagePage = source["usagePage"];
	        this.writeFileResult = source["writeFileResult"];
	        this.setOutputReportResult = source["setOutputReportResult"];
	        this.setFeatureResult = source["setFeatureResult"];
	    }
	}
	export class USBDeviceInfo {
	    vendorId: string;
	    productId: string;
	    product: string;
	    manufacturer: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new USBDeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vendorId = source["vendorId"];
	        this.productId = source["productId"];
	        this.product = source["product"];
	        this.manufacturer = source["manufacturer"];
	        this.path = source["path"];
	    }
	}

}

export namespace updater {
	
	export class UpdateStatus {
	    currentVersion: string;
	    latestVersion?: string;
	    updateAvailable: boolean;
	    downloadUrl?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.updateAvailable = source["updateAvailable"];
	        this.downloadUrl = source["downloadUrl"];
	    }
	}

}

