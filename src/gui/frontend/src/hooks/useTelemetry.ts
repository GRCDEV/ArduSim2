import { create } from "zustand";
import { EventsOn } from "../../wailsjs/runtime/runtime";

export interface TelemetryData {
  uav_id?: string;
  position: {
    lat: number;
    lon: number;
    alt: number;
    heading: number;
    relative_alt: number;
  };
  speed: { vx: number; vy: number; vz: number };
  battery: number;
  status: string;
  flight_mode: string;
  nr_gps_online: number;
  time_boot_ms: number;
}

export interface InterpolationNode {
  start: TelemetryData;
  end: TelemetryData;
  startTime: number;
  duration: number;
  trailEmited: boolean;
}

interface State {
  isReady: boolean;
  interpolatedUavs: () => Record<string, TelemetryData>;
  subscribe(): void;
  unsubscribe(): void;
  setOnNewRealPoint(
    fn: (uavId: string, lat: number, lon: number, alt: number) => void,
  ): void;
  notifyAllReady(): void;
  reset(): void;
}

export const useTelemetry = create<State>((set, get) => {
  const rawData = { current: {} as Record<string, TelemetryData> };
  let worker: Worker | undefined;

  let unsubscribeTelemetry: (() => void) | undefined;
  let subscribed = false;
  let onNewRealPoint:
    | ((uavId: string, lat: number, lon: number, alt: number) => void)
    | undefined;

  return {
    isReady: false,

    interpolatedUavs: () => rawData.current,

    notifyAllReady() {
      set({ isReady: true });
      worker?.postMessage({ type: "ALL_READY" });
    },

    setOnNewRealPoint(fn) {
      onNewRealPoint = fn;
    },

    subscribe() {
      if (subscribed) return;
      subscribed = true;

      worker = new Worker(
        new URL("../workers/telemetry-worker.ts", import.meta.url),
        { type: "module" },
      );

      worker.onmessage = (e) => {
        if (e.data.type === "TICK") {
          rawData.current = e.data.interpolated;
        }
        if (e.data.type == "POINT_REACHED") {
          onNewRealPoint?.(e.data.id, e.data.lat, e.data.lon, e.data.alt);
        }
      };

      if (get().isReady) {
        worker.postMessage({ type: "ALL_READY" });
      }

      unsubscribeTelemetry = EventsOn(
        "telemetry_snapshot",
        (payload: { uavs: Record<string, TelemetryData> }) => {
          worker?.postMessage({
            type: "NEW_SNAPSHOT",
            uavs: payload.uavs,
          });
        },
      );
    },

    unsubscribe() {
      unsubscribeTelemetry?.();
      unsubscribeTelemetry = undefined;
      rawData.current = {};
      subscribed = false;
      worker?.terminate();
      worker = undefined;
    },

    reset() {
      get().unsubscribe();
      set({ isReady: false });
    },
  };
});
