export namespace lecoo {
	
	export class Battery {
	    target: string;
	    effective: string;
	    percent?: number;
	    at: number;
	
	    static createFrom(source: any = {}) {
	        return new Battery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target = source["target"];
	        this.effective = source["effective"];
	        this.percent = source["percent"];
	        this.at = source["at"];
	    }
	}
	export class Capabilities {
	    fanDutyMax: number;
	
	    static createFrom(source: any = {}) {
	        return new Capabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fanDutyMax = source["fanDutyMax"];
	    }
	}
	export class ChargeTarget {
	    kind: string;
	    preset: string;
	    min?: number;
	    max?: number;
	    raw: string;
	
	    static createFrom(source: any = {}) {
	        return new ChargeTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.preset = source["preset"];
	        this.min = source["min"];
	        this.max = source["max"];
	        this.raw = source["raw"];
	    }
	}
	export class CurvePoint {
	    tempC: number;
	    duty: number;
	
	    static createFrom(source: any = {}) {
	        return new CurvePoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tempC = source["tempC"];
	        this.duty = source["duty"];
	    }
	}
	export class FanCurve {
	    enabled: boolean;
	    target: string;
	    points: CurvePoint[];
	    minDuty: number;
	    maxDuty: number;
	    hysteresis: number;
	    panicTemp: number;
	
	    static createFrom(source: any = {}) {
	        return new FanCurve(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.target = source["target"];
	        this.points = this.convertValues(source["points"], CurvePoint);
	        this.minDuty = source["minDuty"];
	        this.maxDuty = source["maxDuty"];
	        this.hysteresis = source["hysteresis"];
	        this.panicTemp = source["panicTemp"];
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
	export class CurveState {
	    curve: FanCurve;
	    tempC?: number;
	    duty: number;
	    appliedPwm: number;
	    lastWriteAt: number;
	    lastError: string;
	
	    static createFrom(source: any = {}) {
	        return new CurveState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.curve = this.convertValues(source["curve"], FanCurve);
	        this.tempC = source["tempC"];
	        this.duty = source["duty"];
	        this.appliedPwm = source["appliedPwm"];
	        this.lastWriteAt = source["lastWriteAt"];
	        this.lastError = source["lastError"];
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
	
	export class FanMode {
	    kind: string;
	    pwm: number;
	
	    static createFrom(source: any = {}) {
	        return new FanMode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.pwm = source["pwm"];
	    }
	}
	export class Health {
	    cliPath: string;
	    cliFound: boolean;
	    daemonUp: boolean;
	    version: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Health(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cliPath = source["cliPath"];
	        this.cliFound = source["cliFound"];
	        this.daemonUp = source["daemonUp"];
	        this.version = source["version"];
	        this.message = source["message"];
	    }
	}
	export class Info {
	    controller: string;
	    controllerVersion: number;
	    hramOffset: string;
	    daemonVersion: string;
	    cliVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.controller = source["controller"];
	        this.controllerVersion = source["controllerVersion"];
	        this.hramOffset = source["hramOffset"];
	        this.daemonVersion = source["daemonVersion"];
	        this.cliVersion = source["cliVersion"];
	    }
	}
	export class MonitorState {
	    running: boolean;
	    rateMs: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new MonitorState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.rateMs = source["rateMs"];
	        this.error = source["error"];
	    }
	}
	export class PowerProfile {
	    kind: string;
	    raw: string;
	
	    static createFrom(source: any = {}) {
	        return new PowerProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.raw = source["raw"];
	    }
	}
	export class Sample {
	    at: number;
	    cpuTempC?: number;
	    sysTempC?: number;
	    cpuRpm?: number;
	    gpuRpm?: number;
	
	    static createFrom(source: any = {}) {
	        return new Sample(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.at = source["at"];
	        this.cpuTempC = source["cpuTempC"];
	        this.sysTempC = source["sysTempC"];
	        this.cpuRpm = source["cpuRpm"];
	        this.gpuRpm = source["gpuRpm"];
	    }
	}
	export class Settings {
	    telemetryEnabled: boolean;
	    telemetryClientId: string;
	    power: PowerProfile;
	    charge: ChargeTarget;
	    fanCpu: FanMode;
	    fanGpu: FanMode;
	    readAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.telemetryEnabled = source["telemetryEnabled"];
	        this.telemetryClientId = source["telemetryClientId"];
	        this.power = this.convertValues(source["power"], PowerProfile);
	        this.charge = this.convertValues(source["charge"], ChargeTarget);
	        this.fanCpu = this.convertValues(source["fanCpu"], FanMode);
	        this.fanGpu = this.convertValues(source["fanGpu"], FanMode);
	        this.readAt = source["readAt"];
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
	
	export class WidgetPosition {
	    x: number;
	    y: number;
	
	    static createFrom(source: any = {}) {
	        return new WidgetPosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	    }
	}

}

