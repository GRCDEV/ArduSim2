import { create } from "zustand";
import { useSimulationConfig } from "./useSimulationConfig";
import type { DeployedService } from "./useSwarms";

export interface GeneralConfig {
  defaultUAVSpeed?: number;
  defaultArduPilotInstance?: string;
  defaultMixer?: DeployedService;
  defaultController?: DeployedService;
  loggingEnabled?: boolean;
  batteryRestricted?: boolean;
  batteryCapacity?: number;
  verboseLogging?: boolean;
  storeLocalData?: boolean;
  windEnabled?: boolean;
  windDirection?: number;
  windSpeed?: number;
  simulationName?: string;
  originalSimulationName?: string;
  dockerHubRepository?: string;
  kubeConfigPath?: string;
  netsimInstances?: number;
  netsimMode?: string;
  netsimMaxRangeM?: number | null;
}

interface State {
  config: GeneralConfig;
  activeMode: string;
  update(fn: (config: GeneralConfig) => GeneralConfig): void;
  loadConfig(config: GeneralConfig, activeMode: string): void;
  setActieMode(value: string): void;
}

export const useConfig = create<State>((set, get) => ({
  config: { netsimInstances: 1, defaultUAVSpeed: 10.0 },
  activeMode: "LOCAL",

  update(fn: (config: GeneralConfig) => GeneralConfig) {
    const newConfig = fn(get().config);
    set({ config: newConfig });
    useSimulationConfig.getState().update((s) => ({
      ...s,
      generalConfig: newConfig,
    }));
  },

  loadConfig(config: GeneralConfig, activeMode: string) {
    set({ config: config, activeMode: activeMode });
  },

  setActieMode(value: string) {
    set({ activeMode: value });
    useSimulationConfig.getState().update((s) => ({
      ...s,
      activeMode: value,
    }));
  },
}));
