import "maplibre-gl/dist/maplibre-gl.css";
import "../../MapLibre.css";
import { useSimulationSession } from "../../hooks/useSimulationSession";
import { useSimulationPageSetup } from "../../hooks/useSimulationPageSetup";
import LogDisplay from "./components/log-display";
import MapControls from "./components/map-controls";
import SimulationControls from "./components/simulation-controls";
import UavTelemetryDisplay from "./components/uav-telemetry-display";

export default function SimulationPage() {
  const simulationFinished = useSimulationSession((s) => s.simulationFinished);
  const { mapContainerRef } = useSimulationPageSetup();
  
  return (
    <main className="flex" style={{ height: "calc(100vh - 60px)" }}>
      <div ref={mapContainerRef} className="h-full relative flex-1 z-30">
        <MapControls className="absolute top-16 left-2 z-50" />
        <SimulationControls className="absolute top-2 left-2 z-50" />
        <LogDisplay
          className="absolute bottom-2 left-2 right-2"
          onFinishReceived={simulationFinished}
        />
      </div>
      <UavTelemetryDisplay />
    </main>
  );
}
