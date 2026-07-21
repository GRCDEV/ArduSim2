import { useEffect, useMemo, useState } from "react";
import { useShallow } from "zustand/shallow";
import Button from "../../../components/button";
import Card from "../../../components/card";
import type { SelectableValue } from "../../../components/select";
import SplitButton from "../../../components/split-button";
import { useSwarms, type UAV } from "../../../hooks/useSwarms";
import { useMap } from "../../../hooks/useMap";
import { useSimulationSession } from "../../../hooks/useSimulationSession";
import { cn } from "../../../utils/cn";
import { EventsOff, EventsOn } from "../../../../wailsjs/runtime/runtime";

interface Props {
  className?: string;
}

export default function SimulationControls({ className }: Props) {
  const swarms = useSwarms((s) => s.swarms);
  const fleetUavs = swarms.flatMap(s => s.uavs);
  const availableServices = useMemo(
    () => getAvailabeServices(fleetUavs),
    [fleetUavs],
  );
  const [start, pause, stop, exit] = useSimulationSession(
    useShallow((s) => [s.start, s.pause, s.stop, s.exit]),
  );
  const [allReady, setAllReady] = useState(false);

  useEffect(() => {
    // Reset readiness whenever the simulation resets (isSimulating changes)
    setAllReady(false);

    const READY_EVENT = "simulation:ready";
    EventsOn(READY_EVENT, () => {
      setAllReady(true);
      useSimulationSession.getState().setupFinished();
      useMap.getState().flyToFirstUav();
    });
    return () => EventsOff(READY_EVENT);
  }, []);

  return (
    <Card className={cn("p-1 flex gap-1 w-min overflow-visible ", className)}>
      <span className={cn("flex gap-1", { "opacity-50": !allReady })}>
        <SplitButton
          icon="play_arrow"
          label="Start"
          enabled={allReady}
          options={availableServices}
          onClick={() => start(availableServices.map((e) => e.value))}
          onSelectOption={(e) => start([e])}
        />
        <SplitButton
          icon="pause"
          label="Pause"
          color="#4d4949"
          enabled={allReady}
          options={availableServices}
          onClick={() => pause(availableServices.map((e) => e.value))}
          onSelectOption={(e) => pause([e])}
        />
        <SplitButton
          icon="stop"
          label="Stop"
          enabled={allReady}
          color="#a83e3e"
          options={availableServices}
          onClick={() => stop(availableServices.map((e) => e.value))}
          onSelectOption={(e) => stop([e])}
        />
      </span>
      <Button icon="exit_to_app" label="Exit" onClick={exit} />
    </Card>
  );
}

function getAvailabeServices(uavs: UAV[]): SelectableValue<string>[] {
  const services: Record<string, string> = {};

  for (const uav of uavs) {
    for (const serv of uav.services) {
      services[serv.serviceId] = serv.serviceTitle;
    }
  }

  return Object.entries(services).map(([k, v]) => ({
    value: k,
    label: v,
  }));
}
