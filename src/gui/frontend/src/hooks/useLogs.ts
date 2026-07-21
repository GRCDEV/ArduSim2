/* eslint-disable @typescript-eslint/no-explicit-any */
import { create } from "zustand";
import {
  LoadLogEntries,
  SearchLogs
} from "../../wailsjs/go/main/App";
import { domain } from "../../wailsjs/go/models";

interface State {
  logPaths: [string, string][]; // [name, path]
  selectedIdx?: number;
  filter: domain.LogFilter;
  messages: domain.LogMessage[];
  isLoading: boolean;

  loadAll(): Promise<void>;
  selectLog(idx?: number): void;
  setFilter(filter: Partial<domain.LogFilter>): void;
  search(): Promise<void>;
}

export const useLogs = create<State>((set, get) => ({
  logPaths: [],
  filter: new domain.LogFilter({}),
  messages: [],
  isLoading: false,

  async loadAll() {
    const paths = await LoadLogEntries();
    const logPaths: [string, string][] = [];
    for (const path of paths) {
      const name = path.split(/[/\\]/).pop();
      if (!name) continue;
      logPaths.push([name, path]);
    }
    set({ logPaths });
  },

  selectLog(idx?: number) {
    if (idx === undefined) {
      set({ selectedIdx: undefined, messages: [] });
      return;
    }
    set({ selectedIdx: idx });
    get().search();
  },

  setFilter(partial: Partial<domain.LogFilter>) {
    set((state) => ({
      filter: { ...state.filter, ...partial } as domain.LogFilter,
    }));
  },

  async search() {
    const state = get();
    if (state.selectedIdx === undefined) return;
    const path = state.logPaths[state.selectedIdx][1];
    
    set({ isLoading: true });
    try {
      const messages = await SearchLogs(path, state.filter);
      set({ messages: messages || [] });
    } catch (err) {
      console.error("Search failed:", err);
      set({ messages: [] });
    } finally {
      set({ isLoading: false });
    }
  },
}));
