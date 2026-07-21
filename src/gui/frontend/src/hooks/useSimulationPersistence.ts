import hash from "object-hash";
import { create } from "zustand";
import { LoadSimulationConfig, SaveSimulationConfig } from "../../wailsjs/go/main/App";
import { domain } from "../../wailsjs/go/models";
import { useConfig } from "./useConfig";
import { useDialog } from "./useDialog";
import { useSwarms } from "./useSwarms";
import { useSimulationConfig, DEFAULT_SIMULATION_STATE } from "./useSimulationConfig";

interface State {
  loadConfig(): Promise<void>;
  saveConfig(): Promise<void>;
  newConfig(): Promise<void>;
}

export const useSimulationPersistence = create<State>((_set, get) => ({
  async loadConfig() {
    const simConfig = useSimulationConfig.getState();
    if (simConfig.lastHash || simConfig.lastHash !== simConfig.lastConfig.hash) {
      const action = await useDialog.getState().show({
        title: "Unsaved changes",
        text: "There are unsaved changes. Loading a new configuration will result in the loss of these modifications",
        buttons: [
          { label: "Cancel" },
          { label: "Accept", type: "outlined" },
          { label: "Save & Accept", type: "filled" },
        ],
      });
      switch (action) {
        case 0:
          return;
        case 2:
          await get().saveConfig();
      }
    }

    const state = await LoadSimulationConfig();
    const code = hash(state);

    useSwarms.getState().loadSwarms(state.swarms as any[]);
    useConfig.getState().loadConfig(state.generalConfig as any, state.activeMode);

    useSimulationConfig.getState().setFullState(
      { value: state as any, hash: code },
      code,
      false
    );
  },

  async saveConfig() {
    const simConfig = useSimulationConfig.getState();
    const state = simConfig.lastConfig;
    const savedState = await SaveSimulationConfig(
      state.value.swarms.map((e) => domain.Swarm.createFrom(e)),
      domain.GeneralConfig.createFrom(state.value.generalConfig),
      state.value.activeMode,
    );
    useSwarms.getState().loadSwarms(savedState.swarms as any[]);
    useConfig.getState().loadConfig(savedState.generalConfig as any, savedState.activeMode);
    const code = hash(savedState);
    useSimulationConfig.getState().setFullState(
      { value: savedState as any, hash: code },
      code,
      false
    );
  },

  async newConfig() {
    const simConfig = useSimulationConfig.getState();
    if (!simConfig.lastHash || simConfig.lastHash !== simConfig.lastConfig.hash) {
      const action = await useDialog.getState().show({
        title: "Unsaved changes",
        text: "There are unsaved changes. Loading a new configuration will result in the loss of these modifications",
        buttons: [
          { label: "Cancel" },
          { label: "Accept", type: "outlined" },
          { label: "Save & Accept", type: "filled" },
        ],
      });
      switch (action) {
        case 0:
          return;
        case 2:
          await get().saveConfig();
      }
    }

    if (simConfig.lastConfig.hash !== "") await get().saveConfig();

    const { swarms, generalConfig, activeMode } = DEFAULT_SIMULATION_STATE;
    
    useSimulationConfig.getState().setFullState(
      { value: DEFAULT_SIMULATION_STATE, hash: "" },
      "",
      true
    );

    useSwarms.getState().loadSwarms(swarms);
    useConfig.getState().loadConfig(generalConfig, activeMode);
  },
}));
