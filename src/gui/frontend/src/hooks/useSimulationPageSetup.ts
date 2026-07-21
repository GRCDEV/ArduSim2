import { useEffect, useRef } from "react";
import { useShallow } from "zustand/shallow";
import { EventsOff, EventsOn } from "../../wailsjs/runtime/runtime";
import { useMap } from "./useMap";
import { useTelemetry } from "./useTelemetry";

export function useSimulationPageSetup() {
  const mapContainerRef = useRef<HTMLDivElement>(null);

  const [initMap, updateTrails, updateMarkers] = useMap(
    useShallow((s) => [s.init, s.updateTrails, s.updateMarkers]),
  );

  const [interpolatedUavs, setonNewRealPoint, notifyAllReady] =
    useTelemetry(
      useShallow((s) => [
        s.interpolatedUavs,
        s.setOnNewRealPoint,
        s.notifyAllReady,
      ]),
    );

  useEffect(() => {
    const READY_EVENT = "simulation:ready";
    const onReady = () => {
      notifyAllReady();
    };
    EventsOn(READY_EVENT, onReady);
    return () => {
      EventsOff(READY_EVENT);
    };
  }, [notifyAllReady]);

  useEffect(() => {
    if (!mapContainerRef.current) return;
    const cleanup = initMap(mapContainerRef.current);
    return cleanup;
  }, [initMap]);



  useEffect(() => {
    setonNewRealPoint(updateTrails);
  }, [setonNewRealPoint, updateTrails]);

  useEffect(() => {
    let id: number;
    const frame = () => {
      const uavList = Object.values(interpolatedUavs());
      updateMarkers(uavList);
      id = requestAnimationFrame(frame);
    };

    id = requestAnimationFrame(frame);
    return () => cancelAnimationFrame(id);
  }, [interpolatedUavs, updateMarkers]);

  return { mapContainerRef };
}
