export namespace domain {
	
	export class Coordinate {
	    lat: number;
	    lon: number;
	    alt: number;
	
	    static createFrom(source: any = {}) {
	        return new Coordinate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lat = source["lat"];
	        this.lon = source["lon"];
	        this.alt = source["alt"];
	    }
	}
	export class DeployedService {
	    instanceId: string;
	    serviceId: string;
	    folderName: string;
	    serviceTitle: string;
	    config: Record<string, any>;
	    nodeLabel?: string;
	
	    static createFrom(source: any = {}) {
	        return new DeployedService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.instanceId = source["instanceId"];
	        this.serviceId = source["serviceId"];
	        this.folderName = source["folderName"];
	        this.serviceTitle = source["serviceTitle"];
	        this.config = source["config"];
	        this.nodeLabel = source["nodeLabel"];
	    }
	}
	export class GeneralConfig {
	    defaultUAVSpeed: number;
	    defaultArduPilotInstance: string;
	    defaultMixer: DeployedService;
	    defaultController: DeployedService;
	    loggingEnabled: boolean;
	    batteryRestricted: boolean;
	    batteryCapacity: number;
	    verboseLogging: boolean;
	    storeLocalData: boolean;
	    windEnabled: boolean;
	    windDirection: number;
	    windSpeed: number;
	    simulationName: string;
	    originalSimulationName: string;
	    dockerHubRepository: string;
	    kubeConfigPath?: string;
	    netsimInstances: number;
	    netsimMode: string;
	    netsimMaxRangeM?: number;
	
	    static createFrom(source: any = {}) {
	        return new GeneralConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaultUAVSpeed = source["defaultUAVSpeed"];
	        this.defaultArduPilotInstance = source["defaultArduPilotInstance"];
	        this.defaultMixer = this.convertValues(source["defaultMixer"], DeployedService);
	        this.defaultController = this.convertValues(source["defaultController"], DeployedService);
	        this.loggingEnabled = source["loggingEnabled"];
	        this.batteryRestricted = source["batteryRestricted"];
	        this.batteryCapacity = source["batteryCapacity"];
	        this.verboseLogging = source["verboseLogging"];
	        this.storeLocalData = source["storeLocalData"];
	        this.windEnabled = source["windEnabled"];
	        this.windDirection = source["windDirection"];
	        this.windSpeed = source["windSpeed"];
	        this.simulationName = source["simulationName"];
	        this.originalSimulationName = source["originalSimulationName"];
	        this.dockerHubRepository = source["dockerHubRepository"];
	        this.kubeConfigPath = source["kubeConfigPath"];
	        this.netsimInstances = source["netsimInstances"];
	        this.netsimMode = source["netsimMode"];
	        this.netsimMaxRangeM = source["netsimMaxRangeM"];
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
	export class LogFilter {
	    InstanceID: string;
	    ServiceID: string;
	    Level: string;
	    EventID: string;
	    SearchText: string;
	
	    static createFrom(source: any = {}) {
	        return new LogFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.InstanceID = source["InstanceID"];
	        this.ServiceID = source["ServiceID"];
	        this.Level = source["Level"];
	        this.EventID = source["EventID"];
	        this.SearchText = source["SearchText"];
	    }
	}
	export class LogMessage {
	    InstanceID: string;
	    ServiceID: string;
	    Level: string;
	    // Go type: time
	    Timestamp: any;
	    Message: string;
	    EventID?: string;
	    // Go type: time
	    ReceivedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new LogMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.InstanceID = source["InstanceID"];
	        this.ServiceID = source["ServiceID"];
	        this.Level = source["Level"];
	        this.Timestamp = this.convertValues(source["Timestamp"], null);
	        this.Message = source["Message"];
	        this.EventID = source["EventID"];
	        this.ReceivedAt = this.convertValues(source["ReceivedAt"], null);
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
	export class ServiceType {
	    id: string;
	    folderName: string;
	    title: string;
	    schemaRaw: string;
	
	    static createFrom(source: any = {}) {
	        return new ServiceType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.folderName = source["folderName"];
	        this.title = source["title"];
	        this.schemaRaw = source["schemaRaw"];
	    }
	}
	export class UAV {
	    id: string;
	    services: DeployedService[];
	    mixer?: DeployedService;
	    controller?: DeployedService;
	    speed?: number;
	    batteryCapacity?: number;
	    homeOverride?: Coordinate;
	    arduPilotInstance?: string;
	    nodeLabel?: string;
	
	    static createFrom(source: any = {}) {
	        return new UAV(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.services = this.convertValues(source["services"], DeployedService);
	        this.mixer = this.convertValues(source["mixer"], DeployedService);
	        this.controller = this.convertValues(source["controller"], DeployedService);
	        this.speed = source["speed"];
	        this.batteryCapacity = source["batteryCapacity"];
	        this.homeOverride = this.convertValues(source["homeOverride"], Coordinate);
	        this.arduPilotInstance = source["arduPilotInstance"];
	        this.nodeLabel = source["nodeLabel"];
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
	export class Swarm {
	    id: string;
	    uavs: UAV[];
	    groundFormation: string;
	    formationCenterLat: number;
	    formationCenterLon: number;
	    formationSpacing: number;
	    formationCenterMode: string;
	
	    static createFrom(source: any = {}) {
	        return new Swarm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.uavs = this.convertValues(source["uavs"], UAV);
	        this.groundFormation = source["groundFormation"];
	        this.formationCenterLat = source["formationCenterLat"];
	        this.formationCenterLon = source["formationCenterLon"];
	        this.formationSpacing = source["formationSpacing"];
	        this.formationCenterMode = source["formationCenterMode"];
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
	export class SimulationState {
	    swarms: Swarm[];
	    generalConfig: GeneralConfig;
	    activeMode: string;
	
	    static createFrom(source: any = {}) {
	        return new SimulationState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.swarms = this.convertValues(source["swarms"], Swarm);
	        this.generalConfig = this.convertValues(source["generalConfig"], GeneralConfig);
	        this.activeMode = source["activeMode"];
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

