import { create } from "zustand";

export interface LogEntry {
  time: string;
  level: string;
  msg: string;
  levelClass: string;
}

interface State {
  logs: LogEntry[];
  simulationStartMs: number | null;

  appendLog(entry: Omit<LogEntry, "time">): void;
  clear(): void;
}

export const useSimulationLog = create<State>((set, get) => ({
  logs: [],
  simulationStartMs: null,

  appendLog(entry) {
    const nowMs = Date.now();
    const startMs = get().simulationStartMs ?? nowMs;

    // Record first log time as t=0
    const nextStartMs = get().simulationStartMs === null ? nowMs : startMs;
    const elapsedSec = Math.floor((nowMs - nextStartMs) / 1000);

    set((s) => ({
      simulationStartMs: nextStartMs,
      // Keep a rolling window of 50 entries to avoid unbounded growth
      logs: [...s.logs.slice(-49), { ...entry, time: formatElapsed(elapsedSec) }],
    }));
  },

  clear() {
    set({ logs: [], simulationStartMs: null });
  },
}));

function formatElapsed(totalSeconds: number): string {
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  const mm = String(minutes).padStart(2, "0");
  const ss = String(seconds).padStart(2, "0");
  return hours > 0 ? `${hours}:${mm}:${ss}` : `${mm}:${ss}`;
}
