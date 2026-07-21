import { type ElementType } from "react";
import { create } from "zustand";
import ConfigPage from "../pages/config/config-page";
import FleetPage from "../pages/fleet/fleet-page";
import SimulationPage from "../pages/simulation/simulation-page";
import LogsPage from "../pages/logs-page";

interface Tab {
  label: string;
  page: ElementType;
}

export const TABS: Tab[] = [
  {
    label: "Config",
    page: ConfigPage,
  },
  {
    label: "Swarms",
    page: FleetPage,
  },
  {
    label: "Simulation",
    page: SimulationPage,
  },
  {
    label: "Logs",
    page: LogsPage
  },
];

export const useNavigation = create<{
  current: Tab;
  navigateTo(label: Tab): void;
  navigateToPath(path: string): void;
}>((set) => ({
  current: TABS[0],
  navigateTo(tab: Tab) {
    set({ current: tab });
  },

  navigateToPath(path: string) {
    const newTab = TABS.find((e) => e.label === path);
    if (!newTab) return;
    set({ current: newTab });
  },
}));
