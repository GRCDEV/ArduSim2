import { create } from "zustand";
import { GetAvailableServices, GetAvailableMixers, GetAvailableControllers } from "../../wailsjs/go/main/App";
import type { domain } from "../../wailsjs/go/models";

interface State {
  services: domain.ServiceType[];
  mixers: domain.ServiceType[];
  controllers: domain.ServiceType[];
  loadServices(): Promise<void>;
}

export const useServices = create<State>((set, get) => ({
  services: [],
  mixers: [],
  controllers: [],
  async loadServices() {
    if (get().services.length !== 0) return;

    const [services, mixers, controllers] = await Promise.all([
      GetAvailableServices(),
      GetAvailableMixers(),
      GetAvailableControllers()
    ]);
    
    set({ services, mixers, controllers });
  },
}));
