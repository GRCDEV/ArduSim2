import hash from "object-hash";
import { create } from "zustand";
import { useConfig, type GeneralConfig } from "./useConfig";
import { useSwarms, type Swarm } from "./useSwarms";

export interface SimulationState {
  swarms: Swarm[];
  generalConfig: GeneralConfig;
  activeMode: string;
}

export const DEFAULT_SIMULATION_STATE: SimulationState = {
  swarms: [],
  generalConfig: {} as GeneralConfig,
  activeMode: "LOCAL",
};

interface ValueHash {
  value: SimulationState;
  hash: string;
}

interface State {
  lastConfig: ValueHash;
  undos: ValueHash[];
  redos: ValueHash[];
  lastHash?: string;
  skipNextUpdate: boolean;

  undo(): void;
  redo(): void;
  update(fn: (state: SimulationState) => SimulationState): void;
  setLastHash(hash: string): void;
  setFullState(config: ValueHash, lastHash: string, skipNextUpdate: boolean): void;
}

export const useSimulationConfig = create<State>((set, get) => ({
  lastConfig: {
    value: DEFAULT_SIMULATION_STATE,
    hash: "",
  },
  lastHash: "",
  undos: [],
  redos: [],
  skipNextUpdate: false,

  undo() {
    const state = get();
    if (state.undos.length === 0) return;

    const prevConfig = state.undos[state.undos.length - 1];
    const newUndos = state.undos.slice(0, -1);
    
    set({
      lastConfig: prevConfig,
      undos: newUndos,
      redos: [...state.redos, state.lastConfig],
      skipNextUpdate: true,
    });

    useSwarms.getState().loadSwarms(prevConfig.value.swarms);
    useConfig
      .getState()
      .loadConfig(
        prevConfig.value.generalConfig,
        prevConfig.value.activeMode,
      );
  },

  redo() {
    const state = get();
    if (state.redos.length === 0) return;

    const nextConfig = state.redos[state.redos.length - 1];
    const newRedos = state.redos.slice(0, -1);

    set({
      lastConfig: nextConfig,
      undos: [...state.undos, state.lastConfig],
      redos: newRedos,
      skipNextUpdate: true,
    });

    useSwarms.getState().loadSwarms(nextConfig.value.swarms);
    useConfig
      .getState()
      .loadConfig(
        nextConfig.value.generalConfig,
        nextConfig.value.activeMode,
      );
  },

  update(fn: (state: SimulationState) => SimulationState) {
    if (get().skipNextUpdate) {
      set({ skipNextUpdate: false });
      return;
    }
    const currentConfig = get().lastConfig;

    const newState = fn(currentConfig.value);
    const code = hash(newState);
    const newConfig = { value: newState, hash: code };

    set((s) => ({
      lastConfig: newConfig,
      undos: [...s.undos, currentConfig],
      redos: [],
    }));
  },

  setLastHash(hash: string) {
    set({ lastHash: hash });
  },

  setFullState(config: ValueHash, lastHash: string, skipNextUpdate: boolean) {
    set({
      undos: [],
      redos: [],
      lastConfig: config,
      lastHash,
      skipNextUpdate,
    });
  }
}));
