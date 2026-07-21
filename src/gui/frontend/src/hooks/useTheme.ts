import { create } from "zustand";
import { persist } from "zustand/middleware";

interface State {
  isNight: boolean;
  toggleTheme(): void;
}

export const useTheme = create<State, [["zustand/persist", State]]>(
  persist(
    (set) => ({
      isNight: false,
      toggleTheme() {
        set((s) => ({ isNight: !s.isNight }));
      },
    }),
    { name: "ardusim-theme-storage" },
  ),
);
