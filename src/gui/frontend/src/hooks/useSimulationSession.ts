import { create } from "zustand";
import {
  DiscardCurrentRun,
  SendAlgorithmCommand,
  StartSimulation,
  StopSimulation,
  DownloadLogs,
  BuildImages,
  ExportSimulation,
} from "../../wailsjs/go/main/App";
import { domain } from "../../wailsjs/go/models";
import { useDialog } from "./useDialog";
import { useMap } from "./useMap";
import { useNavigation } from "./useNavigation";
import { useSimulationConfig } from "./useSimulationConfig";
import { useSimulationLog } from "./useSimulationLog";
import { useTelemetry } from "./useTelemetry";
import { useSimulationPersistence } from "./useSimulationPersistence";

interface State {
  isSimulating: boolean;
  setupTime: number;
  simulationTime: number;

  setupFinished(): void;
  simulationFinished(): void;
  startSimulation(): Promise<void>;
  buildImages(): Promise<void>;
  exportSimulation(): Promise<void>;

  start(targets: string[]): void;
  pause(targets: string[]): void;
  stop(targets: string[]): void;
  exit(): Promise<boolean>;
}

export const useSimulationSession = create<State>((set, get) => {
  let setupIntervalId: NodeJS.Timeout | undefined;
  let simulationIntervalId: NodeJS.Timeout | undefined;

  return {
    setupTime: 0,
    simulationTime: 0,
    isSimulating: false,

    setupFinished() {
      clearInterval(setupIntervalId);
      setupIntervalId = undefined;
    },

    simulationFinished() {
      get().setupFinished();
      clearInterval(simulationIntervalId);
      simulationIntervalId = undefined;
    },

    async startSimulation() {
      if (get().isSimulating) {
        const action = await useDialog.getState().show({
          title: "Simulation in progress",
          text: "There is a simulation in progress. Starting a new simulation wil destroy current.",
          buttons: [{ label: "Cancel" }, { label: "Accept", type: "filled" }],
        });
        switch (action) {
          case 0:
            return;
          case 1: {
            const exitSucces = await get().exit();
            if (!exitSucces) return;
          }
        }
      }

      const simConfig = useSimulationConfig.getState();
      if (simConfig.lastConfig.hash !== "") {
        await useSimulationPersistence.getState().saveConfig();
      }

      const config = simConfig.lastConfig.value;

      setupIntervalId = setInterval(() => {
        set((s) => ({ setupTime: s.setupTime + 1 }));
      }, 1000);

      useNavigation.getState().navigateToPath("Simulation");
      await StartSimulation(
        config.swarms.map((e) => domain.Swarm.createFrom(e)),
        domain.GeneralConfig.createFrom(config.generalConfig),
        config.activeMode === "LOCAL",
      );
      useTelemetry.getState().subscribe();
      set({ isSimulating: true });
    },

    async buildImages() {
      const simConfig = useSimulationConfig.getState();
      if (simConfig.lastConfig.hash !== "") {
        await useSimulationPersistence.getState().saveConfig();
      }

      const config = simConfig.lastConfig.value;

      useNavigation.getState().navigateToPath("Simulation");
      await BuildImages(
        config.swarms.map((e) => domain.Swarm.createFrom(e)),
        domain.GeneralConfig.createFrom(config.generalConfig),
        config.activeMode === "LOCAL",
      );
    },

    async exportSimulation() {
      const simConfig = useSimulationConfig.getState();
      if (simConfig.lastConfig.hash !== "") {
        await useSimulationPersistence.getState().saveConfig();
      }

      const config = simConfig.lastConfig.value;

      try {
        await ExportSimulation(
          config.swarms.map((e) => domain.Swarm.createFrom(e)),
          domain.GeneralConfig.createFrom(config.generalConfig),
        );
      } catch (err) {
        console.error("Failed to export simulation:", err);
        useDialog.getState().show({
          title: "Export Error",
          text: `Failed to export simulation: ${err instanceof Error ? err.message : String(err)}`,
          buttons: [{ label: "Ok" }],
        });
      }
    },

    async start(targets: string[]) {
      if (!simulationIntervalId) {
        simulationIntervalId = setInterval(() => {
          set((s) => ({ simulationTime: s.simulationTime + 1 }));
        }, 1000);
      }
      for (const target of targets) {
        handleSendAlgorithmCommand(target, "start");
      }
    },
    async pause(targets: string[]) {
      for (const target of targets) {
        handleSendAlgorithmCommand(target, "pause");
      }
    },
    async stop(targets: string[]) {
      for (const target of targets) {
        handleSendAlgorithmCommand(target, "stop");
      }
    },
    async exit() {
      let keepLogs = true;
      const action = await useDialog.getState().show({
        title: "Are you sure?",
        text: "The current simulation will be closed. Do you want to save logs?",
        buttons: [
          { label: "Cancel" },
          { label: "Don't save", type: "outlined" },
          { label: "Save logs", type: "filled" },
        ],
      });
      switch (action) {
        case 0:
          return false;
        case 1:
          keepLogs = false;
      }

      if (keepLogs) {
        try {
          await DownloadLogs();
        } catch (err) {
          console.error("Failed to download logs:", err);
          useDialog.getState().show({
            title: "Error",
            text: `Failed to save logs: ${err instanceof Error ? err.message : String(err)}`,
            buttons: [{ label: "Ok" }],
          });
        }
      }

      await StopSimulation();
      if (!keepLogs) {
        await DiscardCurrentRun(
          domain.GeneralConfig.createFrom(useSimulationConfig.getState().lastConfig.value.generalConfig),
        );
      }

      get().simulationFinished();
      set({ isSimulating: false, simulationTime: 0, setupTime: 0 });
      useTelemetry.getState().reset();
      useMap.getState().reset();
      useSimulationLog.getState().clear();

      return true;
    },
  };
});

async function handleSendAlgorithmCommand(serviceId: string, command: string) {
  try {
    await SendAlgorithmCommand(serviceId, command);
  } catch (err) {
    console.error("Failed to send algorithm command:", err);
    useDialog.getState().show({
      title: "Error",
      text: `COMMAND_ERROR: ${err instanceof Error ? err.message : String(err)}`,
      buttons: [{ label: "Ok" }],
    });
    throw err;
  }
}
